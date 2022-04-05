package main

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_lock(m *tgbotapi.Message) {
	err := do_lock()
	if err != nil {
		send_reply(m, true, fmt.Sprintf("session not locked: %s", err))
	} else {
		send_reply(m, true, "lock session invoked")
	}
}

func register_lock() {
	addHandler("LOCK", handle_lock, "LOCK\n"+
		"lock current session",
		ACL_LOCK)
}
