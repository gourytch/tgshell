package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
	PREFIX_EXEC = "exec "
)

type Config struct {
	Token string `json:"token"`
	Owner int    `json:"owner"`
	Host  string `json:"-"`
}

var config Config
var bot *tgbotapi.BotAPI
var job_counter int
shell := sh.NewSession()

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

func at_start() {
	u, err := user.Current()
	var ustr string
	if err == nil {
		ustr = fmt.Sprintf("[%s:%s]%s", u.Uid, u.Gid, u.Username)
	} else {
		ustr = "<unknown>"
	}
	inform(fmt.Sprintf("bot started as user %s and ready to serve", ustr))
}

func handle_exec(m tgbotapi.Message) {
	cmd := strings.Trim(chat_text[len(PREFIX_EXEC):len(chat_text)])
	parts := strings.Fields(command)
	head := parts[0]
	parts = parts[1:len(parts)]
	out, err := session.Command(head, parts...).SetTimeout(EXEC_TIMEOUT).Output()
	var result string
	if err == nil {
		result = fmt.Sprintf("RESULT:\n%s", out)
	} else {
		result = fmt.Sprintf("ERROR: %s", err)
	}
	msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Srintf("COMMAND:\n%s\n%s", cmd, result))
	msg.ReplyToMessageID = m.MessageID
	bot.Send(msg)
}

func main() {
	LoadConfig()
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Fatal(err)
	}
	//bot.Debug = true
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		sig := <-c
		inform(fmt.Sprintf("execution interrupted by %s", sig))
		os.Exit(0)
	}()

	log.Printf("authiorized as %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		chat_id := update.Message.Chat.ID
		chat_text := update.Message.Text
		if strings.HasPrefix(chat_text, PREFIX_EXEC) {
			go handle_exec(update.Message)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}
	}
}
