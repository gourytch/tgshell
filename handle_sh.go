package main

import (
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_shexec(m *tgbotapi.Message) {
	if config.Shell == "" {
		send_reply(m, true, "shell executable is not set")
		return
	}
	_, script := Split2(m.Text)
	log.Printf("execute shell script: %s", script)
	session := shell.Command(config.Shell).SetInput(script)
	if config.Exec.Timeout > 0 {
		session = session.SetTimeout(time.Duration(config.Exec.Timeout) * time.Second)
	}
	tStart := time.Now().UTC()
	out, err := session.CombinedOutput()
	tFinish := time.Now().UTC()
	limit := len(out)
	if config.Exec.SendLimit < limit {
		limit = config.Exec.SendLimit
		out = out[:limit]
	}
	sout := ExecFmt(tFinish.Sub(tStart)/time.Millisecond, err, out)
	log.Print(sout)
	send_reply(m, true, sout...)
}

func handle_setsh(m *tgbotapi.Message) {
	_, shell := Split2(m.Text)
	if shell == "" {
		send_reply(m, true, "shell required")
		return
	}
	config.Shell = shell
	SaveConfig()
	send_reply(m, true, fmt.Sprintf("shell set to <%s>", config.Shell))
}

func handle_unsetsh(m *tgbotapi.Message) {
	config.Shell = ""
	SaveConfig()
	send_reply(m, true, fmt.Sprintf("shell set to empty string"))
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
