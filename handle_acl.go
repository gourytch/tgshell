package main

import (
	"fmt"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_acl(m *tgbotapi.Message) {
	_, cmd := Split2(m.Text)
	cmd = strings.ToUpper(cmd)
	if cmd == "" {
		send_reply(m, fmt.Sprintf("ACCESS CONTROL LIST:\nMaster:"), true)
	}
}

func register_acl() {
	addHandler("acl", handle_acl,
		"ACL\n"+
			"access control list management",
		ACL_ADMIN)
}
