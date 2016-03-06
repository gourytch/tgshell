package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	PREFIX_EXEC        = "exec "
	EXEC_TIMEOUT       = 5 * time.Second
	EXEC_SEND_DELAY    = 1 * time.Second
	EXEC_SEND_LIMIT    = 4000
	RECONNECT_INTERVAL = 15 * time.Second
)

type Config struct {
	Token string `json:"token"`
	Owner int    `json:"owner"`
	Host  string `json:"-"`
}

var config Config
var bot *tgbotapi.BotAPI
var job_counter int
var shell *sh.Session

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

func inform(text string) {
	log.Print(text)
	now := time.Now()
	s := fmt.Sprintf("%s @ #%s\n%s", now.Format("2006/01/02 15:04:05 MST"), config.Host, text)
	msg := tgbotapi.NewMessage(config.Owner, s)
	bot.Send(msg)
}

func inform_at_start() {
	u, err := user.Current()
	var ustr string
	if err == nil {
		ustr = fmt.Sprintf("[uid=%s:gid=%s] %s", u.Uid, u.Gid, u.Username)
	} else {
		ustr = "<unknown>"
	}
	inform(fmt.Sprintf("bot started as user\n%s\nand ready to serve", ustr))
}

func handle_exec(m tgbotapi.Message) {
	cmd := strings.TrimPrefix(m.Text, PREFIX_EXEC)
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
		log.Printf("bot.Send() error: %s", err)
	} else {
		log.Print("bot.Send() without errors")
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
		log.Println("tgbotapi.GetUpdatesChan() got error", err, ", sleep and try again")
		time.Sleep(RECONNECT_INTERVAL)
	}
	for update := range updates {
		m := update.Message
		if m.Chat.ID != config.Owner {
			msg := tgbotapi.NewMessage(m.Chat.ID, "Hi, "+m.Chat.UserName)
			msg.ReplyToMessageID = m.MessageID
			bot.Send(msg)
		} else {
			if strings.HasPrefix(m.Text, PREFIX_EXEC) {
				go handle_exec(m)
			} else {
				msg := tgbotapi.NewMessage(m.Chat.ID, m.Text)
				msg.ReplyToMessageID = m.MessageID
				bot.Send(msg)
			}
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
