package main

import (
	"fmt"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func scrotFName(lossless bool) string {
	fname := "scrot_" + time.Now().Format("20060102_150405")
	if lossless {
		fname += ".png"
	} else {
		fname += ".jpg"
	}
	return fname
}

func do_scrot(m *tgbotapi.Message, lossless bool) {
	_, dsp := Split2(m.Text)
	if dsp == "" {
		dsp = config.Display
	} else {
		if config.Display != dsp {
			config.Display = dsp
			SaveConfig()
		}
	}
	data, err := scrot(lossless)
	fname := scrotFName(lossless)
	if err != nil {
		send_reply(m, fmt.Sprintf("scrot error: %s", err), true)
	} else {
		if lossless {
			send_reply_document(m, fname, data)
		} else {
			send_reply_image(m, fname, data)
		}
	}
}

func handle_scrot(m *tgbotapi.Message) {
	do_scrot(m, false)
}

func handle_hqscrot(m *tgbotapi.Message) {
	do_scrot(m, true)
}

func register_screenshot() {
	addHandler("SCROT", handle_scrot, "SCROT [<display>]\n"+
		"take screenshot",
		ACL_SCROT)
	addHandler("HQSCROT", handle_hqscrot, "HQSCROT [<display>]\n"+
		"take screenshot in high quality",
		ACL_SCROT)

}
