package main

import (
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	EXEC_TIMEOUT       = 5 * time.Second
	EXEC_SEND_DELAY    = 1 * time.Second
	EXEC_SEND_LIMIT    = 4000
	RECONNECT_INTERVAL = 15 * time.Second
	CONNKEY_SIZE       = 9
	USE_SPEW           = true
)

type Config struct {
	Token     string  `json:"token"`
	Master    int64   `json:"master"`
	Allow_New bool    `json:"allow_new"`
	Users     []int64 `json:"users"`
	Shell     string  `json:"shell"`
	Data_Dir  string  `json:"datadir"`
	Display   string  `jdon:"display"`
	Host      string  `json:"-"`
}

type HandlerProc func(m *tgbotapi.Message)

type Handler struct {
	proc HandlerProc
	info string
}

var handlers map[string]Handler = make(map[string]Handler)

var config Config
var bot *tgbotapi.BotAPI
var job_counter int
var shell *sh.Session
var connect_key string
var exitcode int
var sigchan chan os.Signal

func ppj(v interface{}) string {
	if USE_SPEW {
		return spew.Sdump(v)
	} else {
		b, _ := json.MarshalIndent(v, "", "  ")
		return string(b)
	}
}

func random_string(n int) string {
	var key []byte
	count := (n + 3) * 3 / 4
	for i := 0; i < count; i++ {
		key = append(key, byte(rand.Int()&0xFF))
	}
	return base64.StdEncoding.EncodeToString(key)[:n]
}

func generate_key() {
	connect_key = random_string(CONNKEY_SIZE)
}

var splitRx *regexp.Regexp = regexp.MustCompile("(?sm)\\A\\s*([\\S]+)\\s*(.*)\\z")

func Split2(text string) (token, rest string) {
	//r := regexp.MustCompile("(?sm)\\A\\s*([\\S]+)\\s*(.*)\\z")
	//v := r.FindStringSubmatch(text)
	v := splitRx.FindStringSubmatch(text)
	if v == nil {
		return "", ""
	}
	//v := strings.SplitN(text+" ", " ", 2)
	//return strings.TrimSpace(v[0]), strings.TrimSpace(v[1])
	return v[1], v[2]
}

func GetFirstToken(text string) string {
	token, _ := Split2(text)
	return token
}

func isMaster(id int64) bool {
	if config.Master == id {
		return true
	}
	return false
}

func isUser(id int64) bool {
	if config.Master == id {
		return true
	}
	for _, i := range config.Users {
		if id == i {
			return true
		}
	}
	return false
}

func addHandler(name string, proc HandlerProc, info string) {
	handlers[strings.ToUpper(name)] = Handler{proc: proc, info: info}
}
