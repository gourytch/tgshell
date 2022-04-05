package main

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_halt(m *tgbotapi.Message) {
	err := do_halt()
	if err != nil {
		send_reply(m, true, fmt.Sprintf("session not halted: %s", err))
	} else {
		send_reply(m, true, "halt session invoked")
	}
}

func register_halt() {
	addHandler("HALT", handle_halt, "HALT\n"+
		"halt and shutdown the computer",
		ACL_HALT)
}
