package main

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_acl_list(m *tgbotapi.Message) {
	var b bytes.Buffer
	for _, e := range config.Users {
		if isMaster(e.Id) {
			fmt.Fprintf(&b, "%s: MASTER\n", e.String())
		} else {
			fmt.Fprintf(&b, "%s: %v\n", e.String(), e.Allow)
		}
	}
	send_reply(m, b.String(), true)
}

func handle_acl_abilities(m *tgbotapi.Message) {
	abilities := acl_abilities()
	var b bytes.Buffer
	for _, ability := range abilities {
		fmt.Fprintf(&b, "%s\n", ability)
	}
	send_reply(m, b.String(), true)
}

func handle_acl_update(m *tgbotapi.Message) {
	_, cmdtext := Split2(m.Text)
	if cmdtext == "" {
		send_reply(m, "malformed command", true)
		return
	}
	idstr, L := Split2(cmdtext)
	id, err := strconv.Atoi(idstr)
	if err != nil {
		send_reply(m, fmt.Sprintf("malformed id `%v` in string `%v`",
			idstr, cmdtext), true)
		return
	}
	e := acl_entry(id)
	if e == nil {
		send_reply(m, fmt.Sprintf("unknown id `%v`", id), true)
		return
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "alter acl for %v:\n", e.String())
	dirty := false
	for {
		var s string
		s, L = Split2(L)
		if s == "" {
			break
		}
		op := s[0]
		ability := s[1:]
		switch {
		case op == '+':
			if acl_grant(id, ability) {
				fmt.Fprintf(&b, "... grant %s\n", ability)
				dirty = true
			}
		case op == '-':
			if acl_revoke(id, ability) {
				fmt.Fprintf(&b, "... revoke %s\n", ability)
				dirty = true
			}
		default:
			fmt.Fprintf(&b, "... skip op=`%v`, ab=`%s`\n", op, ability)

		}
	}
	if dirty {
		SaveConfig()
	}
	send_reply(m, b.String(), true)
}

func register_acl() {
	addHandler("acl.list", handle_acl_list,
		"ACL.LIST\n"+
			"display all users with abilities",
		ACL_ADMIN)
	addHandler("acl.abilities", handle_acl_abilities,
		"ACL.LIST\n"+
			"display all active abilities",
		ACL_ADMIN)
	addHandler("acl.update", handle_acl_update,
		"ACL.UPDATE id (+|-)(*|rulename) ...\n"+
			"change abilities",
		ACL_ADMIN)
}
