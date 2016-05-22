package main

import (
	"fmt"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func handle_scrot(m *tgbotapi.Message) {
	if !isUser(m.Chat.ID) {
		send_reply(m, "Insufficient permissions")
		return
	}
	_, dsp := Split2(m.Text)
	if dsp == "" {
		dsp = config.Display
	} else {
		if config.Display != dsp {
			config.Display = dsp
			SaveConfig()
		}
	}
	CheckDatadir()
	fname := "scrot_" + time.Now().Format("20060102_150405") + ".jpg"
	out, err := shell.SetEnv("DISPLAY", dsp).
		SetDir(config.Data_Dir).
		Command("scrot", fname).
		CombinedOutput()
	if err != nil {
		send_reply(m, fmt.Sprintf("scrot error: %s\nmessage:\n%s", err, out))
	} else {
		//send_reply_image(m, config.Data_Dir+"/"+fname, nil)
		send_reply_document(m, config.Data_Dir+"/"+fname, nil)
	}

}

func register_screenshot() {
	addHandler("SCROT", handle_scrot, "SCROT [<display>]\n"+
		"take screenshot")
}
