// +build !windows

package main

import (
	"io/ioutil"
)

func scrot(lossless bool) (data []byte, err error) {
	full_name := RealPath(config.Data_Dir + scrotFName(lossless))
	CheckDatadir()
	_, err := shell.SetEnv("DISPLAY", config.Display).
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
	return
}
