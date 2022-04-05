//go:build windows

package main

import (
	"github.com/mitchellh/go-homedir"
)

// default data directory for windows service or windows user
func GetDefaultDataDir() string {
	dir, err := homedir.Dir()
	if err != nil {
		dir = "C:/TEMP/TGSHELL"
	} else {
		dir += "/tgshell"
	}
	return RealPath(dir)
}
