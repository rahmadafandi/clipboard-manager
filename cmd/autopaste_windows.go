//go:build windows

package cmd

import (
	"os/exec"
	"time"
)

func autoPaste() {
	time.Sleep(50 * time.Millisecond)
	exec.Command("powershell.exe", "-NoProfile", "-Command",
		`Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait("^v")`).Run()
}
