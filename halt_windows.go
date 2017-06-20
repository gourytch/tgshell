// +build windows

package main

import (
	"log"
)

// shutdown /p /f

func do_halt() error {
	out, err := shell.
		Command("SHUTDOWN.EXE", "/p", "/f").
		SetTimeout(EXEC_TIMEOUT).
		CombinedOutput()
	log.Printf("*** SHUTDOWN INVOKED ***")
	log.Printf("err = %v", err)
	log.Printf("out = %v", string(out))
	/* for Win9x =)
	if err != nil {
		out, err = shell.
			Command("RUNDLL32.EXE", "shell32.dll,SHExitWindowsEx 8").
			SetTimeout(EXEC_TIMEOUT).
			CombinedOutput()
		log.Printf("*** SHExitWindowsEx INVOKED ***")
		log.Printf("err = %v", err)
		log.Printf("out = %v", string(out))
	}
	*/
	return err
}
