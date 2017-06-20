// +build windows

package main

// rundll32.exe user32.dll,LockWorkStation

func do_lock() error {
	_, err := shell.
		Command("RUNDLL32.EXE", "user32.dll,LockWorkStation").
		SetTimeout(EXEC_TIMEOUT).
		CombinedOutput()
	return err
}
