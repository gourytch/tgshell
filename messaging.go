package main

import (
	"fmt"
	"log"
	"os/user"
	"syscall"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

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
