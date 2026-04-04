//go:build darwin

package cmd

import (
	"os/exec"
	"time"
)

func autoPaste() {
	time.Sleep(50 * time.Millisecond)
	exec.Command("osascript", "-e",
		`tell application "System Events" to keystroke "v" using command down`).Run()
}
