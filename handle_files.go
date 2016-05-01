package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_get(m *tgbotapi.Message) {
	send_reply(m, "NIY")
}

func handle_put(m *tgbotapi.Message) {
	send_reply(m, "NIY")
}

func register_files() {
	addHandler("GET", handle_get, "GET /path/to/file [...]\n"+
		"retrieve file(s) from remote system")
	addHandler("PUT", handle_put, "PUT /path/to/file\n"+
		"store file to remote system")
}
