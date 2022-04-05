//go:build !windows
// +build !windows

package main

import "time"

// rundll32.exe user32.dll,LockWorkStation

func do_lock() error {
	_, err := shell.
		Command("xdg-screensaver", "lock").
		SetTimeout(DEFAULT_EXEC_TIMEOUT * time.Second).
		CombinedOutput()
	return err
}
