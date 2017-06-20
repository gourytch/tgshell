package main

import (
	"fmt"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_halt(m *tgbotapi.Message) {
	err := do_halt()
	if err != nil {
		send_reply(m, fmt.Sprintf("session not halted: %s", err), true)
	} else {
		send_reply(m, "halt session invoked", true)
	}
}

func register_halt() {
	addHandler("HALT", handle_halt, "HALT\n"+
		"halt and shutdown the computer",
		ACL_HALT)
}
