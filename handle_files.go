package main

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_get(m *tgbotapi.Message) {
	_, txt := Split2(m.Text)
	if txt == "" {
		send_reply(m, false, "file name required")
		return
	}
	fnames := strings.Fields(txt)
	for _, fname := range fnames {
		send_reply_document(m, fname, nil)
	}
}

func handle_put(m *tgbotapi.Message) {
	send_reply(m, false, "NIY")
}

func register_files() {
	addHandler("GET", handle_get, "GET /path/to/file [...]\n"+
		"retrieve file(s) from remote system",
		ACL_FILES)
	addHandler("PUT", handle_put, "PUT /path/to/file\n"+
		"store file to remote system",
		ACL_FILES)
}
