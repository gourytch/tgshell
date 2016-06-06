package main

import (
	"fmt"
	"log"
	"os/user"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func inform(text string) {
	log.Print(text)
	now := time.Now()
	s := fmt.Sprintf("%s @ #%s\n%s", now.Format("2006/01/02 15:04:05 MST"), config.Host, text)
	ids := acl_all([]string{ACL_INFORM})
	for _, id := range ids {
		msg := tgbotapi.NewMessage(int64(id), s)
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("inform failed:\nmessage=%s\nerror=%s", ppj(msg), err)
		} else {
			log.Printf("message sent:\n%s", ppj(msg))
		}
	}
}

func skip_noticing(id int, m *tgbotapi.Message) bool {
	// we do not need to send notice for:
	// 1. ... author
	if m.From != nil && m.From.ID == id {
		return true
	}
	// 2. ... addressat in private chat
	if m.Chat.ID == int64(id) {
		return true
	}
	// 3. ... private replying
	if m.ReplyToMessage != nil && m.ReplyToMessage.Chat.ID == int64(id) {
		return true
	}
	// 3. ... public replying
	if m.ReplyToMessage != nil && m.ReplyToMessage.From != nil && m.ReplyToMessage.From.ID == id {
		return true
	}
	return false
}

func notice_forward(m *tgbotapi.Message) {
	for _, id := range acl_all([]string{ACL_SUPERVISE}) {
		if skip_noticing(id, m) {
			continue
		}
		msg := tgbotapi.NewForward(int64(id), m.Chat.ID, m.MessageID)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("forward-notice to %v not sent: %s", id, err)
		}
	}
}

func notice_interaction(request, responce *tgbotapi.Message) {
	for _, id := range acl_all([]string{ACL_SUPERVISE}) {
		if skip_noticing(id, request) || skip_noticing(id, responce) {
			continue
		}
		msg := tgbotapi.NewForward(int64(id), request.Chat.ID, request.MessageID)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("notice-forward request to %v not sent: %s", id, err)
		}
		msg = tgbotapi.NewForward(int64(id), responce.Chat.ID, responce.MessageID)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("notice-forward responce to %v not sent: %s", id, err)
		}
	}
}

func notice(request *tgbotapi.Message, text string) {
	e := acl_entry(request.From.ID)
	report := fmt.Sprintf("User %s: %s", e.String(), text)
	for _, id := range acl_all([]string{ACL_SUPERVISE}) {
		if skip_noticing(id, request) {
			continue
		}
		msg := tgbotapi.NewMessage(int64(id), report)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("notice to %v not sent: %s", id, err)
		}
	}
}

func inform_at_start() {
	if connect_key == "" {
		generate_key()
	}

	var mestr string
	if me, err := bot.GetMe(); err == nil {
		mestr = fmt.Sprintf("%v/%s", me.ID, me.String())
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

func send_reply(m *tgbotapi.Message, text string, notable bool) {
	msg := tgbotapi.NewMessage(m.Chat.ID, text)
	msg.ReplyToMessageID = m.MessageID

	if reply, err := bot.Send(msg); err != nil {
		log.Printf("send_reply failed:\nmessage=%s\nerror=%s", ppj(msg), err)
	} else {
		log.Printf("reply sent:\n%s", ppj(msg))
		if notable {
			notice_interaction(m, &reply)
		}
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
	if reply, err := bot.Send(msg); err != nil {
		log.Printf("send_reply_document failed:\nmessage=%#v\nerror=%#v", msg, err)
	} else {
		log.Printf("document reply sent:\n%#v", msg)
		notice_interaction(m, &reply)
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
	if reply, err := bot.Send(msg); err != nil {
		log.Printf("send_reply_image failed:\nmessage=%#v\nerror=%#v", msg, err)
	} else {
		log.Printf("image reply sent:\n%#v", msg)
		notice_interaction(m, &reply)
	}
}
