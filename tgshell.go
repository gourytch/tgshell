package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	COMMAND_EXEC       = "exec "
	COMMAND_KEYGEN     = "keygen"
	COMMAND_EXIT       = "exit"
	EXEC_TIMEOUT       = 5 * time.Second
	EXEC_SEND_DELAY    = 1 * time.Second
	EXEC_SEND_LIMIT    = 4000
	RECONNECT_INTERVAL = 15 * time.Second
	CONNKEY_SIZE       = 9
)

type Config struct {
	Token     string  `json:"token"`
	Master    int64   `json:"master"`
	Allow_New bool    `json:"allow_new"`
	Users     []int64 `json:"users"`
	Host      string  `json:"-"`
}

var config Config
var bot *tgbotapi.BotAPI
var job_counter int
var shell *sh.Session
var connect_key string

func LoadConfig() {
	exe, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	fname := exe + ".config"
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}
	config.Host, err = os.Hostname()
	if err != nil {
		log.Printf("os.Hostname() error: %s", err)
		config.Host = "unknown"
	}
}

func SaveConfig() {
	data, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Marshal error: %s", err)
	}
	exe, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatalf("Abs(%s) error: %s", os.Args[0], err)
	}
	fname := exe + ".config"
	if _, err = os.Stat(fname); !os.IsNotExist(err) {
		// make backup
		fname_backup := fname + ".bak"
		if _, err := os.Stat(fname_backup); !os.IsNotExist(err) {
			err = os.Remove(fname_backup)
			if err != nil {
				log.Fatalf("Remove(%s) failed: %s", fname_backup, err)
			}
		}
		err = os.Rename(fname, fname_backup)
		if err != nil {
			log.Fatalf("Rename(%s, %s) failed: %s", fname, fname_backup, err)
		}
	}
	err = ioutil.WriteFile(fname, data, 0600)
}

func isMaster(id int64) bool {
	if config.Master == id {
		return true
	}
	return false
}

func isUser(id int64) bool {
	if config.Master == id {
		return true
	}
	for _, i := range config.Users {
		if id == i {
			return true
		}
	}
	return false
}

func isCommand(text, command string) bool {
	if text == "" {
		return false
	}
	chunks := strings.Split(text, " ")
	return strings.EqualFold(chunks[0], command)
}

func inform(text string) {
	log.Print(text)
	now := time.Now()
	s := fmt.Sprintf("%s @ #%s\n%s", now.Format("2006/01/02 15:04:05 MST"), config.Host, text)
	ids := append(config.Users, config.Master)
	for _, id := range ids {
		msg := tgbotapi.NewMessage(id, s)
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("Send to %v failed: %s", id, err)
		}
	}
}

func generate_key() {
	var key []byte
	for i := 0; i < CONNKEY_SIZE; i++ {
		key = append(key, byte(rand.Int()&0xFF))
	}
	connect_key = base64.StdEncoding.EncodeToString(key)
}

func inform_at_start() {
	if connect_key == "" {
		generate_key()
	}
	me, err := bot.GetMe()
	var mestr string
	if err == nil {
		mestr = fmt.Sprintf("%v/%s", me.ID, me.UserName)
	} else {
		mestr = "<unknown>"
	}
	u, err := user.Current()
	var ustr string
	if err == nil {
		ustr = fmt.Sprintf("%s [%s:%s]", u.Username, u.Uid, u.Gid)
	} else {
		ustr = "<unknown>"
	}
	inform(fmt.Sprintf("bot %s\nstarted as user %s\n"+
		"and ready to serve.\n"+
		"connect key is:\n"+
		"%s", mestr, ustr, connect_key))
}

func handle_guest(m *tgbotapi.Message) {
	var msg tgbotapi.MessageConfig
	if config.Allow_New && strings.HasPrefix(m.Text, connect_key) {
		config.Users = append(config.Users, m.Chat.ID)
		SaveConfig()
		msg = tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf(
			"Ave, %s! Your id=%v|%v",
			m.Chat.UserName, m.Chat.ID, m.From.ID))
	} else {
		msg = tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf(
			"Hi, %s! your id=%v|%v", m.Chat.UserName, m.Chat.ID, m.From.ID))
	}
	msg.ReplyToMessageID = m.MessageID
	bot.Send(msg)
}

func handle_keygen(m *tgbotapi.Message) {
	if !isMaster(m.Chat.ID) {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Insufficient permissions")
		msg.ReplyToMessageID = m.MessageID
		bot.Send(msg)
		return
	}
	generate_key()
	msg := tgbotapi.NewMessage(m.Chat.ID, connect_key)
	msg.ReplyToMessageID = m.MessageID
	bot.Send(msg)
}

func handle_exit(m *tgbotapi.Message) {
	if !isMaster(m.Chat.ID) {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Insufficient permissions")
		msg.ReplyToMessageID = m.MessageID
		bot.Send(msg)
		return
	}
	tail := strings.TrimSpace(strings.SplitN(m.Text+" ", " ", 2)[1])
	exitcode, err := strconv.Atoi(tail)
	if err != nil {
		exitcode = 0
	}
	msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("EXIT %d", exitcode))
	msg.ReplyToMessageID = m.MessageID
	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("Send reply failed: %s", err)
	}
	time.Sleep(1 * time.Second)
	os.Exit(exitcode)
}

func handle_exec(m *tgbotapi.Message) {
	if m.Text == "" {
		return
	}
	cmd := strings.TrimSpace(strings.SplitN(m.Text, " ", 2)[1])
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]
	log.Printf("execute '%s' ...", cmd)
	out, err := shell.Command(head, parts).SetTimeout(EXEC_TIMEOUT).Output()
	limit := len(out)
	if EXEC_SEND_LIMIT < limit {
		limit = EXEC_SEND_LIMIT
	}
	sout := fmt.Sprintf("err:%v\nresult\n%s", err, out[:limit])
	log.Print(sout)
	msg := tgbotapi.NewMessage(m.Chat.ID, sout)
	msg.ReplyToMessageID = m.MessageID
	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("Send reply failed: %s", err)
	}
}

func workSession() {
	log.Println("start work session")
	var err error
	for {
		bot, err = tgbotapi.NewBotAPI(config.Token)
		if err == nil {
			break
		}
		log.Println("tgbotapi.NewBotAPI() got error", err, ". sleep and re-init")
		time.Sleep(RECONNECT_INTERVAL)
	}
	//bot.Debug = true
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		sig := <-c
		inform(fmt.Sprintf("Got signal <%s>.\nExecution terminated", sig))
		os.Exit(0)
	}()
	shell = sh.NewSession()
	log.Printf("authiorized as %s", bot.Self.UserName)
	inform_at_start()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	var updates <-chan tgbotapi.Update
	for {
		updates, err = bot.GetUpdatesChan(u)
		if err == nil {
			break
		}
		log.Printf("GetUpdatesChan() got error %s\nsleep and try again", err)
		time.Sleep(RECONNECT_INTERVAL)
	}
	for update := range updates {
		m := update.Message
		switch {
		case !isUser(m.Chat.ID):
			go handle_guest(m)
		case isCommand(m.Text, COMMAND_KEYGEN):
			go handle_keygen(m)
		case isCommand(m.Text, COMMAND_EXEC):
			go handle_exec(m)
		case isCommand(m.Text, COMMAND_EXIT):
			go handle_exit(m)
		default:
			msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf(
				"U said: <<%s>>", m.Text))
			msg.ReplyToMessageID = m.MessageID
			bot.Send(msg)
		}
	}
	log.Println("finish work session")
}

func main() {
	LoadConfig()
	for {
		workSession()
		log.Println("delay and return to work")
		time.Sleep(RECONNECT_INTERVAL)
	}
}
