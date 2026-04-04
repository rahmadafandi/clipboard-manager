//go:build darwin

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.design/x/clipboard"
)

var darwinLaunchers = []launcher{
	{
		bin: "choose",
		args: func() []string {
			return []string{"-p", "Clipboard"}
		},
		imgArgs:  func() []string { return nil },
		fmtLine:  defaultFmtLine,
		parseIdx: func(output string) (int, bool) { return parseIdxFromNumber(output) },
	},
}

func detectLauncher() (*launcher, error) {
	for i := range darwinLaunchers {
		if _, err := exec.LookPath(darwinLaunchers[i].bin); err == nil {
			return &darwinLaunchers[i], nil
		}
	}
	return nil, fmt.Errorf("no supported launcher found (install choose-gui: brew install choose-gui)")
}

func writeTextToClipboard(text string) {
	if bin, err := exec.LookPath("pbcopy"); err == nil {
		cmd := exec.Command(bin)
		cmd.Stdin = strings.NewReader(text)
		if cmd.Run() == nil {
			return
		}
	}
	if err := clipboard.Init(); err == nil {
		clipboard.Write(clipboard.FmtText, []byte(text))
	}
}

func writeImageToClipboard(data []byte) {
	// Write to temp file, then use osascript to set clipboard
	tmpFile := filepath.Join(os.TempDir(), "clipboard-manager-img.png")
	if err := os.WriteFile(tmpFile, data, 0644); err == nil {
		defer os.Remove(tmpFile)
		script := fmt.Sprintf(`set the clipboard to (read (POSIX file "%s") as «class PNGf»)`, tmpFile)
		if exec.Command("osascript", "-e", script).Run() == nil {
			return
		}
	}
	if err := clipboard.Init(); err == nil {
		clipboard.Write(clipboard.FmtImage, data)
	}
}

func showNotification(msg string) {
	script := fmt.Sprintf(`display notification "%s" with title "Clipboard Manager"`, msg)
	exec.Command("osascript", "-e", script).Run()
}
