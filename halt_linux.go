//go:build !windows
// +build !windows

package main

import "time"

func do_halt() error {
	_, err := shell.
		Command("/sbin/poweroff", "now").
		SetTimeout(DEFAULT_EXEC_TIMEOUT * time.Second).
		CombinedOutput()
	return err
}
