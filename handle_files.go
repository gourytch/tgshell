package main

import (
	//"fmt"
	//"io/ioutil"
	//"log"
	//"path"
	//"path/filepath"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_get(m *tgbotapi.Message) {
	_, txt := Split2(m.Text)
	if txt == "" {
		send_reply(m, "file name required")
		return
	}
	fnames := strings.Fields(txt)
	for _, fname := range fnames {
		send_reply_document(m, fname, nil)
		/*
			if data, err := ioutil.ReadFile(fname); err != nil {
				send_reply(m, fmt.Sprintf("file <%s> load error: %s", fname, err))
			} else {
				log.Printf("upload file <%s>", fname)

			}
		*/
	}
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
