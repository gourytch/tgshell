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
	log.Printf("func SplitOutput(limit=%d, datasize=%d)", limit, len(data))
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
	log.Printf("func SplitOutput() returned %d pages", len(ret))
	return ret
}

func ExecFmt(ms time.Duration, err error, out []byte) []string {
	var serr string
	if err != nil {
		serr = fmt.Sprintf("<b>%s</b>", html.EscapeString(err.Error()))
	} else {
		serr = "<em>no error</em>"
	}
	chunks := SplitOutput(config.Exec.PageSize, out)
	ret := []string{}
	txt := fmt.Sprintf("<b>dt:</b> %.4f\n%s\n", float64(ms)/1000.0, serr)
	switch len(chunks) {
	case 0:
		ret = []string{txt + "<em>no output</em>"}
	case 1:
		ret = []string{txt + fmt.Sprintf("<b>output:</b><pre>%s</pre>", html.EscapeString(chunks[0]))}
	default:
		numPages := len(chunks)
		if numPages > config.Exec.MaxPages {
			ret = []string{txt + fmt.Sprintf("<b>output (%d pages, truncated):</b>", numPages)}
			numPages = config.Exec.MaxPages
		} else {
			ret = []string{txt + fmt.Sprintf("<b>output (%d pages):</b>", numPages)}
		}
		for i := 0; i < numPages; i++ {
			ret = append(ret, fmt.Sprintf("<b>PAGE %d</b><pre>%s</pre>", i+1, html.EscapeString(chunks[i])))
		}
		if numPages < len(chunks) {
			ret = append(ret, fmt.Sprintf("<em>... %d pages were skipped</em>", len(chunks)-numPages))
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
	sout := ExecFmt(tFinish.Sub(tStart)/time.Millisecond, err, out)
	send_reply(m, true, sout...)
	send_reply_document(m, fmt.Sprintf("exec-%s.log", tStart.Format("20060102_150405")), out)
}

func register_exec() {
	addHandler("EXEC", handle_exec,
		"EXEC command [params...]\n"+
			"execute noninteractive command on remote system.",
		ACL_EXEC)
}
