package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func SplitOutput(limit int, data []byte) []string {
	r := bytes.NewBuffer(data)
	scanner := bufio.NewScanner(r)
	ret := []string{}
	cur := bytes.Buffer{}
	for scanner.Scan() {
		t := scanner.Text()
		if limit-cur.Len() < len(t) {
			// the new line doesn't fit to the limit
			ret = append(ret, cur.String()) // flush current text
			cur = bytes.Buffer{}
			for limit < len(t) { // if line is toooooo big - tear it to pieces
				ret = append(ret, t[:limit])
				t = t[limit+1:]
			}
		}
		// add line
		cur.WriteString(t)
		cur.WriteString("\n")
	}
	if cur.Len() > 0 {
		ret = append(ret, cur.String()) // flush current text
	}
	return ret
}

func ExecFmt(ms time.Duration, err error, out []byte) []string {
	var serr string
	if err != nil {
		serr = err.Error()
	} else {
		serr = "no error"
	}
	chunks := SplitOutput(config.Exec.SendLimit, out)
	ret := []string{}
	txt := fmt.Sprintf("<b>dt:</b> %.4f\n<b>err: </b>%v\n", float64(ms)/1000.0, html.EscapeString(serr))
	switch len(chunks) {
	case 0:
		ret = []string{txt + "no output"}
	case 1:
		ret = []string{txt + fmt.Sprintf("<b>output:</b><pre>%s</pre>", html.EscapeString(chunks[0]))}
	default:
		ret = []string{txt + fmt.Sprintf("<b>output (%d pages):</b>", len(chunks))}
		for _, chunk := range chunks {
			ret = append(ret, fmt.Sprintf("<pre>%s</pre>", html.EscapeString(chunk)))
		}
	}
	return ret
}

func handle_exec(m *tgbotapi.Message) {
	_, cmdtext := Split2(m.Text)
	parts := strings.Fields(cmdtext)
	if len(parts) == 0 {
		send_reply(m, false, "command required")
		return
	}
	cmd := parts[0]
	args := parts[1:]
	log.Printf("execute '%s' %v ...", cmd, args)
	sh := shell.Command(cmd, args)
	if config.Exec.Timeout > 0 {
		sh.SetTimeout(time.Duration(config.Exec.Timeout) * time.Second)
	}
	tStart := time.Now().UTC()
	out, err := sh.CombinedOutput()
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

func register_exec() {
	addHandler("EXEC", handle_exec,
		"EXEC command [params...]\n"+
			"execute noninteractive command on remote system.",
		ACL_EXEC)
}
