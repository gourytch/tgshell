// +build !windows

package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func scrot(lossless bool) (data []byte, fname string, err error) {
	fname := RealPath(config.Data_Dir + "/scrot")
	CheckDatadir()
	out, err := shell.SetEnv("DISPLAY", config.Display).
		SetDir(config.Data_Dir).
		Command("scrot", full_name).
		CombinedOutput()
	if err != nil {
		return
	}
	data, err = ioutil.ReadFile(full_name)
	if err != nil {
		return
	}
}
