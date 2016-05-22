package main

import (
	"fmt"
	"log"
	"os/user"
	//"path/filepath"
	//"syscall"
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
			log.Printf("inform failed:\nmessage=%s\nerror=%s", ppj(msg), err)
		} else {
			log.Printf("message sent:\n%s", ppj(msg))
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
	/*
		var pcstr string
		if cn, err := syscall.ComputerName(); err == nil {
			pcstr = fmt.Sprintf("computer: <%s>", cn)
		} else {
			pcstr = "<unknown computer>"
		}
	*/
	inform(fmt.Sprintf("bot %s\nstarted\n"+
		"and ready to serve.\n"+
		"%s\n"+
		// "%s\n"+
		"connect key is:\n"+
		"%s", mestr,
		ustr,
		// pcstr,
		connect_key))
}

func inform_at_stop() {
	inform("bot stopped")
}

func send_reply(m *tgbotapi.Message, text string) {
	msg := tgbotapi.NewMessage(m.Chat.ID, text)
	msg.ReplyToMessageID = m.MessageID

	if _, err := bot.Send(msg); err != nil {
		log.Printf("send_reply failed:\nmessage=%s\nerror=%s", ppj(msg), err)
	} else {
		log.Printf("reply sent:\n%s", ppj(msg))
	}
}

func send_reply_document(m *tgbotapi.Message, fname string, data []byte) {
	var msg tgbotapi.DocumentConfig
	if data == nil {
		msg = tgbotapi.NewDocumentUpload(m.Chat.ID, fname)
	} else {
		fb := tgbotapi.FileBytes{Name: fname, Bytes: data}
		msg = tgbotapi.NewDocumentUpload(m.Chat.ID, fb)
	}
	msg.ReplyToMessageID = m.MessageID
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send_reply_document failed:\nmessage=%#v\nerror=%#v", msg, err)
	} else {
		log.Printf("document reply sent:\n%#v", msg)
	}
}

func send_reply_image(m *tgbotapi.Message, fname string, data []byte) {
	var msg tgbotapi.PhotoConfig
	if data == nil {
		msg = tgbotapi.NewPhotoUpload(m.Chat.ID, fname)
	} else {
		fb := tgbotapi.FileBytes{Name: fname, Bytes: data}
		msg = tgbotapi.NewPhotoUpload(m.Chat.ID, fb)
	}
	msg.ReplyToMessageID = m.MessageID
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send_reply_image failed:\nmessage=%#v\nerror=%#v", msg, err)
	} else {
		log.Printf("image reply sent:\n%#v", msg)
	}
}
