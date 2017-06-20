package main

import (
	"fmt"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_lock(m *tgbotapi.Message) {
	err := do_lock()
	if err != nil {
		send_reply(m, fmt.Sprintf("session not locked: %s", err), true)
	} else {
		send_reply(m, "lock session invoked", true)
	}
}

func register_lock() {
	addHandler("LOCK", handle_lock, "LOCK\n"+
		"lock current session",
		ACL_LOCK)
}
