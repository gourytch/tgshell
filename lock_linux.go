// +build !windows

package main

// rundll32.exe user32.dll,LockWorkStation

func do_lock() error {
	_, err := shell.
		Command("xdg-screensaver", "lock").
		SetTimeout(EXEC_TIMEOUT).
		CombinedOutput()
	return err
}
