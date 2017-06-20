// +build !windows

package main

func do_halt() error {
	_, err := shell.
		Command("/sbin/poweroff", "now").
		SetTimeout(EXEC_TIMEOUT).
		CombinedOutput()
	return err
}
