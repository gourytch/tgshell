package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_list(m *tgbotapi.Message) {
	var commands []string
	for name, _ := range handlers {
		commands = append(commands, name)
	}
	send_reply(m, fmt.Sprintf("available commands:\n%q", commands))
}

func handle_help(m *tgbotapi.Message) {
	_, cmd := Split2(m.Text)
	cmd = strings.ToUpper(cmd)
	var reply string
	if handler, ok := handlers[cmd]; ok {
		reply = fmt.Sprintf("help for %s:\n%s", cmd, handler.info)
	} else {
		reply = fmt.Sprintf(
			"no such command: <%s>\n"+
				"enter LIST for list of available commands", cmd)
	}
	send_reply(m, reply)
}

func handle_exit(m *tgbotapi.Message) {
	if !isMaster(m.Chat.ID) {
		send_reply(m, "Insufficient permissions")
		return
	}
	_, tail := Split2(m.Text)
	exitcode, err := strconv.Atoi(tail) // var exitcode is global
	if err != nil {
		exitcode = 0
	}
	inform(fmt.Sprintf("EXIT %d", exitcode))
	time.Sleep(1 * time.Second)
	//	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	os.Exit(exitcode)
}

func handle_keygen(m *tgbotapi.Message) {
	if !isMaster(m.Chat.ID) {
		send_reply(m, "Insufficient permissions")
		return
	}
	generate_key()
	send_reply(m, connect_key)
}

func handle_me(m *tgbotapi.Message) {
	send_reply(m, fmt.Sprintf("user id=%v\nchat id=%v",
		m.Contact.UserID, m.Chat.ID))
}

func register_base() {
	addHandler("ME", handle_me,
		"ME\n"+
			"show your ID")
	addHandler("KEYGEN", handle_keygen, "KEYGEN\n"+
		"generate and show new access key. invalidate old access key")
	addHandler("UPTIME", handle_uptime,
		"UPTIME\n"+
			"show utime for tgshell")
	addHandler("EXIT", handle_exit,
		"EXIT [<num:exitcode>]\n"+
			"invoke tgshell exit/restart routine")
	addHandler("LIST", handle_list,
		"LIST\n"+
			"show list of available commands")
	addHandler("HELP", handle_help,
		"HELP command\n"+
			"show command usage")

}
