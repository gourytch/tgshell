package main

import (
	"fmt"
	"log"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_shexec(m *tgbotapi.Message) {
	if !isUser(m.Chat.ID) {
		send_reply(m, "Insufficient permissions")
		return
	}
	if config.Shell == "" {
		send_reply(m, "shell executable is not set")
		return
	}
	_, script := Split2(m.Text)
	log.Printf("execute shell script: %s", script)
	out, err := shell.Command(config.Shell).SetInput(script).SetTimeout(EXEC_TIMEOUT).CombinedOutput()
	limit := len(out)
	if EXEC_SEND_LIMIT < limit {
		limit = EXEC_SEND_LIMIT
		out = out[:limit]
	}
	sout := fmt.Sprintf("err:%v\nresult\n%s", err, out)
	log.Print(sout)
	send_reply(m, sout)
}

func handle_setsh(m *tgbotapi.Message) {
	_, shell := Split2(m.Text)
	if !isMaster(m.Chat.ID) {
		send_reply(m, "Insufficient permissions")
		return
	}
	if shell == "" {
		send_reply(m, "shell required")
		return
	}
	config.Shell = shell
	SaveConfig()
	send_reply(m, fmt.Sprintf("shell set to <%s>", config.Shell))
}

func handle_unsetsh(m *tgbotapi.Message) {
	if !isMaster(m.Chat.ID) {
		send_reply(m, "Insufficient permissions")
		return
	}
	config.Shell = ""
	SaveConfig()
	send_reply(m, fmt.Sprintf("shell set to empty string"))
}

func register_sh() {
	addHandler("SH", handle_shexec,
		"SH params[...]\n"+
			"execute shell sequence on remote system")
	addHandler("SETSH", handle_setsh,
		"SETSH path/to/the/shell/executable\n"+
			"set shell executable for SH command")
	addHandler("UNSETSH", handle_unsetsh,
		"UNSETSH\n"+
			"set shell executable for SH command to empty string")
}
