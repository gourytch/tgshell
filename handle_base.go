package main

import (
	"fmt"
	"html"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_list(m *tgbotapi.Message) {
	id := m.From.ID
	var commands []string
	for name, h := range handlers {
		if acl_can(id, h.perm) {
			commands = append(commands, name)
		}
	}
	send_reply(m, false, fmt.Sprintf("available commands:\n%q", commands))
}

func handle_help(m *tgbotapi.Message) {
	_, cmd := Split2(m.Text)
	cmd = strings.ToUpper(cmd)
	var reply string
	if handler, ok := handlers[cmd]; ok {
		reply = fmt.Sprintf("help for <b>%s</b>:\n%s", html.EscapeString(cmd), html.EscapeString(handler.info))
	} else {
		reply = fmt.Sprintf(
			"no such command: &lt;<b>%s</b>&rt;\n"+
				"enter <b>LIST</b> for list of available commands", html.EscapeString(cmd))
	}
	send_reply(m, false, reply)
}

func handle_exit(m *tgbotapi.Message) {
	e := acl_entry(m.From.ID)
	_, tail := Split2(m.Text)
	exitcode, err := strconv.Atoi(tail) // var exitcode is global
	if err != nil {
		exitcode = 0
	}
	inform(fmt.Sprintf("EXIT %d FROM %v", exitcode, e.String()))
	time.Sleep(1 * time.Second)
	//	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	os.Exit(exitcode)
}

func handle_keygen(m *tgbotapi.Message) {
	generate_key()
	send_reply(m, true, connect_key)
}

func handle_me(m *tgbotapi.Message) {
	e := acl_entry(m.From.ID)
	send_reply(m, false, fmt.Sprintf("You %v in %v#%v",
		e.String(), m.Chat.Type, m.Chat.ID))
}

func register_base() {
	addHandler("ME", handle_me,
		"ME\n"+
			"show your ID",
		ACL_ANY)
	addHandler("KEYGEN", handle_keygen, "KEYGEN\n"+
		"generate and show new access key. invalidate old access key",
		ACL_ADMIN)
	addHandler("UPTIME", handle_uptime,
		"UPTIME\n"+
			"show utime for tgshell",
		ACL_ANY)
	addHandler("EXIT", handle_exit,
		"EXIT [<num:exitcode>]\n"+
			"invoke tgshell exit/restart routine",
		ACL_ADMIN)
	addHandler("LIST", handle_list,
		"LIST\n"+
			"show list of available commands",
		ACL_ANY)
	addHandler("HELP", handle_help,
		"HELP command\n"+
			"show command usage",
		ACL_ANY)
}
