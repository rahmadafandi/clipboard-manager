//go:build linux

package cmd

import (
	"os"
	"os/exec"
	"time"
)

func autoPaste() {
	// Small delay to let clipboard settle
	time.Sleep(50 * time.Millisecond)

	if os.Getenv("WAYLAND_DISPLAY") != "" {
		if bin, err := exec.LookPath("wtype"); err == nil {
			exec.Command(bin, "-M", "ctrl", "-k", "v", "-m", "ctrl").Run()
			return
		}
		if bin, err := exec.LookPath("ydotool"); err == nil {
			exec.Command(bin, "key", "29:1", "47:1", "47:0", "29:0").Run()
			return
		}
	}
	if bin, err := exec.LookPath("xdotool"); err == nil {
		exec.Command(bin, "key", "--clearmodifiers", "ctrl+v").Run()
	}
}
