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
	id := m.From.ID
	name := m.From.String()
	var reply string
	if !acl_exists(id) {
		reply = fmt.Sprintf("Hello, %s! I will memorize you", name)
		acl_touch(id, name)
	} else {
		reply = fmt.Sprintf("I know you, %s.", name)
	}
	send_reply(m, reply, false)
}

func handle_uptime(m *tgbotapi.Message) {
	send_reply(m, "UPTIME NIY", false)
}

func dispatch(m *tgbotapi.Message) bool {
	id := m.From.ID
	name := m.From.String()
	cmd := strings.ToUpper(GetFirstToken(m.Text))
	handler, ok := handlers[cmd]
	if ok {
		if acl_can(id, handler.perm) {
			go handler.proc(m)
		} else {
			send_reply(m, fmt.Sprintf("[%v]%v cannot into %s",
				id, name, handler.perm), true)
		}
		return true
	}
	if m.Chat.Type == "private" {
		send_reply(m, fmt.Sprintf("I don't understood, what %v is", cmd), true)
	}
	return false
}

func register_all() {
	register_base()
	register_exec()
	register_sh()
	register_files()
	register_screenshot()
	register_acl()
	register_lock()
	register_halt()
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
		log.Printf("got message:\n%s", ppj(m))
		if m == nil {
			continue // we have no deals with malformed messages
		}
		from := m.From
		if from == nil {
			continue // we have no deals with anonymous messages
		}
		acl_touch(from.ID, from.String()) // remember/refresh acl
		if !dispatch(m) {
			if m.Chat.Type == "private" {
				msg := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf(
					"U said: <<%s>>", m.Text))
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
