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
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
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
	Shell     string  `json:"shell"`
}

var config Config
var bot *tgbotapi.BotAPI
var job_counter int
var shell *sh.Session
var connect_key string
var exitcode int
var sigchan chan os.Signal

func Split2(text string) (token, rest string) {
	r := regexp.MustCompile("(?sm)\\A\\s*([\\S]+)\\s*(.*)\\z")
	v := r.FindStringSubmatch(text)
	if v == nil {
		return "", ""
	}
	//v := strings.SplitN(text+" ", " ", 2)
	//return strings.TrimSpace(v[0]), strings.TrimSpace(v[1])
	return v[1], v[2]
}

func GetFirstToken(text string) string {
	token, _ := Split2(text)
	return token
}

func ExeName() string {
	exe, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	return exe
}

func AppDir() string {
	dir, err := filepath.Abs(filepath.Dir(ExeName()))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func AppBaseFileName() string {
	r, _ := regexp.Compile("^(.*?)(?:\\.exe|\\.EXE|)$")
	return r.FindStringSubmatch(ExeName())[1]
}

func GetConfigName() string {
	return AppBaseFileName() + ".config"
}

func LoadConfig() {
	data, err := ioutil.ReadFile(GetConfigName())
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
	fname := GetConfigName()
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

func isCommand(text string, command string) bool {
	return strings.EqualFold(strings.TrimSpace(
		strings.SplitN(text+" ", " ", 2)[0]),
		command)
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

	var mestr string
	if me, err := bot.GetMe(); err == nil {
		mestr = fmt.Sprintf("%v/%s", me.ID, me.UserName)
	} else {
		mestr = "<unknown>"
	}

	var ustr string
	if u, err := user.Current(); err == nil {
		ustr = fmt.Sprintf("user:%s\n  uid=%s\n  gid=%s",
			u.Username, u.Uid, u.Gid)
	} else {
		ustr = "<unknown user>"
	}

	var pcstr string
	if cn, err := syscall.ComputerName(); err == nil {
		pcstr = fmt.Sprintf("computer: <%s>", cn)
	} else {
		pcstr = "<unknown computer>"
	}
	inform(fmt.Sprintf("bot %s\nstarted\n"+
		"and ready to serve.\n"+
		"%s\n"+
		"%s\n"+
		"connect key is:\n"+
		"%s", mestr, ustr, pcstr, connect_key))
}

func inform_at_stop() {
	inform("bot stopped")
}

func send_reply(m *tgbotapi.Message, text string) {
	msg := tgbotapi.NewMessage(m.Chat.ID, text)
	msg.ReplyToMessageID = m.MessageID

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Send reply failed: %s", err)
	}
}

func handle_guest(m *tgbotapi.Message) {
	var reply string
	if config.Allow_New && strings.HasPrefix(m.Text, connect_key) {
		config.Users = append(config.Users, m.Chat.ID)
		SaveConfig()
		reply = fmt.Sprintf("Ave, %s! Your id=%v|%v", m.Chat.UserName, m.Chat.ID, m.From.ID)
	} else {
		reply = fmt.Sprintf("Hi, %s! your id=%v|%v", m.Chat.UserName, m.Chat.ID, m.From.ID)
	}
	send_reply(m, reply)
}

func handle_keygen(m *tgbotapi.Message) {
	if !isMaster(m.Chat.ID) {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Insufficient permissions")
		msg.ReplyToMessageID = m.MessageID
		bot.Send(msg)
		return
	}
	generate_key()
	send_reply(m, connect_key)
}

func handle_uptime(m *tgbotapi.Message) {
	send_reply(m, "UPTIME NIY")
}

func handle_list(m *tgbotapi.Message) {
	var commands []string
	for name, _ := range handlers {
		commands = append(commands, name)
	}
	send_reply(m, fmt.Sprintf("available commands:\n%q", commands))
}

func handle_help(m *tgbotapi.Message) {
	_, cmd := Split2(m.Text)
	cmd = strings.ToUpper(cmd)
	var reply string
	if handler, ok := handlers[cmd]; ok {
		reply = fmt.Sprintf("help for %s:\n%s", cmd, handler.info)
	} else {
		reply = fmt.Sprintf(
			"no such command: <%s>\n"+
				"enter LIST for list of available commands", cmd)
	}
	send_reply(m, reply)
}

func handle_exit(m *tgbotapi.Message) {
	if !isMaster(m.Chat.ID) {
		send_reply(m, "Insufficient permissions")
		return
	}
	_, tail := Split2(m.Text)
	exitcode, err := strconv.Atoi(tail) // var exitcode is global
	if err != nil {
		exitcode = 0
	}
	inform(fmt.Sprintf("EXIT %d", exitcode))
	time.Sleep(1 * time.Second)
	//	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	os.Exit(exitcode)
}

func execute(cmd string, args []string) (out []byte, err error) {
	log.Printf("execute '%s' %v ...", cmd, args)
	out, err = shell.Command(cmd, args).SetTimeout(EXEC_TIMEOUT).CombinedOutput()
	limit := len(out)
	if EXEC_SEND_LIMIT < limit {
		limit = EXEC_SEND_LIMIT
		out = out[:limit]
	}
	sout := fmt.Sprintf("err:%v\nresult\n%s", err, out)
	log.Print(sout)
	return
}

func handle_exec(m *tgbotapi.Message) {
	_, cmd := Split2(m.Text)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		msg := tgbotapi.NewMessage(m.Chat.ID, "command missing")
		msg.ReplyToMessageID = m.MessageID
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("Send reply failed: %s", err)
		}
		return
	}
	head := parts[0]
	parts = parts[1:len(parts)]
	out, err := execute(head, parts)
	sout := fmt.Sprintf("err:%v\nresult\n%s", err, out)
	log.Print(sout)
	send_reply(m, sout)
}

func handle_shexec(m *tgbotapi.Message) {
	if config.Shell == "" {
		send_reply(m, "shell executable is not set")
		return
	}
	_, script := Split2(m.Text)
	log.Printf("execute shell script: %s", script)
	out, err := shell.Command(config.Shell).SetInput(script).SetTimeout(EXEC_TIMEOUT).CombinedOutput()
	limit := len(out)
	if EXEC_SEND_LIMIT < limit {
		limit = EXEC_SEND_LIMIT
		out = out[:limit]
	}
	sout := fmt.Sprintf("err:%v\nresult\n%s", err, out)
	log.Print(sout)
	send_reply(m, sout)
}

func handle_setsh(m *tgbotapi.Message) {
	_, shell := Split2(m.Text)
	if !isMaster(m.Chat.ID) {
		send_reply(m, "Insufficient permissions")
		return
	}
	if shell == "" {
		send_reply(m, "shell required")
		return
	}
	config.Shell = shell
	SaveConfig()
	send_reply(m, fmt.Sprintf("shell set to <%s>", config.Shell))
}

func handle_unsetsh(m *tgbotapi.Message) {
	if !isMaster(m.Chat.ID) {
		send_reply(m, "Insufficient permissions")
		return
	}
	config.Shell = ""
	SaveConfig()
	send_reply(m, fmt.Sprintf("shell set to empty string"))
}

func handle_get(m *tgbotapi.Message) {
}

func handle_put(m *tgbotapi.Message) {
}

type HandlerProc func(m *tgbotapi.Message)
type Handler struct {
	proc HandlerProc
	info string
}

var handlers map[string]Handler = make(map[string]Handler)

func addHandler(name string, proc HandlerProc, info string) {
	handlers[strings.ToUpper(name)] = Handler{proc: proc, info: info}
}

func dispatch(m *tgbotapi.Message) bool {
	cmd := strings.ToUpper(GetFirstToken(m.Text))
	handler, ok := handlers[cmd]
	if ok {
		go handler.proc(m)
		return true
	} else {
		return false
	}
	return false
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
	sigchan = make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, syscall.SIGTERM)
	//defer Stop(sigchan)
	go func() {
		sig := <-sigchan
		inform(fmt.Sprintf("Got signal <%s>.\nExecution terminated", sig))
		os.Exit(exitcode)
	}()
	shell = sh.NewSession()
	log.Printf("authiorized as %s", bot.Self.UserName)
	inform_at_start()
	defer inform_at_stop()

	addHandler("KEYGEN", handle_keygen,
		"KEYGEN\n"+
			"generate and show new access key. invalidate old access key")
	addHandler("UPTIME", handle_uptime,
		"UPTIME\n"+
			"show utime for tgshell")
	addHandler("EXIT", handle_exit,
		"EXIT [<num:exitcode>]\n"+
			"invoke tgshell exit/restart routine")
	addHandler("LIST", handle_list,
		"LIST\n"+
			"show list of available commands")
	addHandler("HELP", handle_help,
		"HELP command\n"+
			"show command usage")
	addHandler("EXEC", handle_exec,
		"EXEC command [params...]\n"+
			"execute noninteractive command on remote system.")
	addHandler("SH", handle_shexec,
		"SH params[...]\n"+
			"execute shell sequence on remote system")
	addHandler("SETSH", handle_setsh,
		"SETSH path/to/the/shell/executable\n"+
			"set shell executable for SH command")
	addHandler("UNSETSH", handle_unsetsh,
		"UNSETSH\n"+
			"set shell executable for SH command to empty string")

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
		if !dispatch(m) {
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
