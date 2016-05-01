package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

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

func handle_uptime(m *tgbotapi.Message) {
	send_reply(m, "UPTIME NIY")
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

func register_all() {
	register_base()
	register_exec()
	register_sh()
	register_files()
	register_screenshot()
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

	register_all()

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
