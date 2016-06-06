package main

import (
	"fmt"
	"log"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_shexec(m *tgbotapi.Message) {
	if config.Shell == "" {
		send_reply(m, "shell executable is not set", true)
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
	send_reply(m, sout, true)
}

func handle_setsh(m *tgbotapi.Message) {
	_, shell := Split2(m.Text)
	if shell == "" {
		send_reply(m, "shell required", true)
		return
	}
	config.Shell = shell
	SaveConfig()
	send_reply(m, fmt.Sprintf("shell set to <%s>", config.Shell), true)
}

func handle_unsetsh(m *tgbotapi.Message) {
	config.Shell = ""
	SaveConfig()
	send_reply(m, fmt.Sprintf("shell set to empty string"), true)
}

func register_sh() {
	addHandler("SH", handle_shexec,
		"SH params[...]\n"+
			"execute shell sequence on remote system",
		ACL_EXEC)
	addHandler("SETSH", handle_setsh,
		"SETSH path/to/the/shell/executable\n"+
			"set shell executable for SH command",
		ACL_ADMIN)
	addHandler("UNSETSH", handle_unsetsh,
		"UNSETSH\n"+
			"set shell executable for SH command to empty string",
		ACL_ADMIN)
}
