package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_exec(m *tgbotapi.Message) {
	if !isUser(m.Chat.ID) {
		send_reply(m, "Insufficient permissions")
		return
	}
	_, cmdtext := Split2(m.Text)
	parts := strings.Fields(cmdtext)
	if len(parts) == 0 {
		send_reply(m, "command required")
		return
	}
	cmd := parts[0]
	args := parts[1:len(parts)]
	log.Printf("execute '%s' %v ...", cmd, args)
	out, err := shell.Command(cmd, args).SetTimeout(EXEC_TIMEOUT).CombinedOutput()
	limit := len(out)
	if EXEC_SEND_LIMIT < limit {
		limit = EXEC_SEND_LIMIT
		out = out[:limit]
	}
	sout := fmt.Sprintf("err:%v\nresult\n%s", err, out)
	log.Print(sout)
	send_reply(m, sout)
}

func register_exec() {
	addHandler("EXEC", handle_exec,
		"EXEC command [params...]\n"+
			"execute noninteractive command on remote system.")
}
