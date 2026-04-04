//go:build linux

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"golang.design/x/clipboard"
)

var waylandLaunchers = []launcher{
	{
		bin: "wofi",
		args: func() []string {
			return []string{"--dmenu", "-i", "-p", "Clipboard", "--lines", "10"}
		},
		imgArgs: func() []string {
			return []string{"--allow-images"}
		},
		fmtLine: func(idx int, item storage.ClipItem, imgPath string) string {
			if item.Type == storage.Image && imgPath != "" {
				return fmt.Sprintf("img:%s:text:%d. [Image]", imgPath, idx+1)
			}
			return defaultFmtLine(idx, item, imgPath)
		},
		parseIdx: func(output string) (int, bool) {
			s := output
			if strings.HasPrefix(s, "img:") {
				if i := strings.Index(s, ":text:"); i != -1 {
					s = s[i+len(":text:"):]
				}
			}
			parts := strings.SplitN(s, ".", 2)
			if len(parts) < 2 {
				return 0, false
			}
			num, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return 0, false
			}
			return num - 1, true
		},
	},
	{
		bin: "fuzzel",
		args: func() []string {
			return []string{"--dmenu", "--prompt", "Clipboard > ", "--lines", "10"}
		},
		imgArgs:  func() []string { return nil },
		fmtLine:  defaultFmtLine,
		parseIdx: func(output string) (int, bool) { return parseIdxFromNumber(output) },
	},
	{
		bin:  "rofi",
		args: rofiArgs,
		imgArgs: func() []string {
			return []string{"-show-icons"}
		},
		fmtLine: rofiImgFmtLine,
		parseIdx: func(output string) (int, bool) {
			num, err := strconv.Atoi(strings.TrimSpace(output))
			if err != nil {
				return 0, false
			}
			return num, true
		},
	},
	{
		bin: "dmenu",
		args: func() []string {
			return []string{"-i", "-l", "10", "-p", "Clipboard"}
		},
		imgArgs:  func() []string { return nil },
		fmtLine:  defaultFmtLine,
		parseIdx: func(output string) (int, bool) { return parseIdxFromNumber(output) },
	},
}

var x11Launchers = []launcher{
	{
		bin:  "rofi",
		args: rofiArgs,
		imgArgs: func() []string {
			return []string{"-show-icons"}
		},
		fmtLine: rofiImgFmtLine,
		parseIdx: func(output string) (int, bool) {
			num, err := strconv.Atoi(strings.TrimSpace(output))
			if err != nil {
				return 0, false
			}
			return num, true
		},
	},
	{
		bin: "dmenu",
		args: func() []string {
			return []string{"-i", "-l", "10", "-p", "Clipboard"}
		},
		imgArgs:  func() []string { return nil },
		fmtLine:  defaultFmtLine,
		parseIdx: func(output string) (int, bool) { return parseIdxFromNumber(output) },
	},
	{
		bin: "wofi",
		args: func() []string {
			return []string{"--dmenu", "-i", "-p", "Clipboard", "--lines", "10"}
		},
		imgArgs: func() []string {
			return []string{"--allow-images"}
		},
		fmtLine: func(idx int, item storage.ClipItem, imgPath string) string {
			if item.Type == storage.Image && imgPath != "" {
				return fmt.Sprintf("img:%s:text:%d. [Image]", imgPath, idx+1)
			}
			return defaultFmtLine(idx, item, imgPath)
		},
		parseIdx: func(output string) (int, bool) {
			s := output
			if strings.HasPrefix(s, "img:") {
				if i := strings.Index(s, ":text:"); i != -1 {
					s = s[i+len(":text:"):]
				}
			}
			parts := strings.SplitN(s, ".", 2)
			if len(parts) < 2 {
				return 0, false
			}
			num, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return 0, false
			}
			return num - 1, true
		},
	},
	{
		bin: "fuzzel",
		args: func() []string {
			return []string{"--dmenu", "--prompt", "Clipboard > ", "--lines", "10"}
		},
		imgArgs:  func() []string { return nil },
		fmtLine:  defaultFmtLine,
		parseIdx: func(output string) (int, bool) { return parseIdxFromNumber(output) },
	},
}

func rofiArgs() []string {
	return []string{
		"-dmenu", "-i", "-p", "Clipboard", "-format", "i",
		"-hover-select",
		"-me-select-entry", "",
		"-me-accept-entry", "MousePrimary",
		"-kb-accept-entry", "Return",
		"-theme-str", "window {width: 400px;} listview {lines: 10;}",
	}
}

func rofiImgFmtLine(idx int, item storage.ClipItem, imgPath string) string {
	label := defaultFmtLine(idx, item, imgPath)
	if item.Type == storage.Image && imgPath != "" {
		return label + "\x00icon\x1f" + imgPath
	}
	return label
}

func isWayland() bool {
	return os.Getenv("WAYLAND_DISPLAY") != ""
}

func detectLauncher() (*launcher, error) {
	list := x11Launchers
	if isWayland() {
		list = waylandLaunchers
	}
	for i := range list {
		if _, err := exec.LookPath(list[i].bin); err == nil {
			return &list[i], nil
		}
	}
	hint := "rofi, dmenu, wofi, or fuzzel"
	if isWayland() {
		hint = "wofi, fuzzel, or rofi-wayland"
	}
	return nil, fmt.Errorf("no supported launcher found (install %s)", hint)
}

func writeTextToClipboard(text string) {
	if isWayland() {
		if bin, err := exec.LookPath("wl-copy"); err == nil {
			cmd := exec.Command(bin)
			cmd.Stdin = strings.NewReader(text)
			if cmd.Run() == nil {
				return
			}
		}
	} else {
		if bin, err := exec.LookPath("xclip"); err == nil {
			cmd := exec.Command(bin, "-selection", "clipboard")
			cmd.Stdin = strings.NewReader(text)
			if cmd.Run() == nil {
				return
			}
		}
	}
	if err := clipboard.Init(); err == nil {
		clipboard.Write(clipboard.FmtText, []byte(text))
	}
}

func writeImageToClipboard(data []byte) {
	if isWayland() {
		if bin, err := exec.LookPath("wl-copy"); err == nil {
			cmd := exec.Command(bin, "--type", "image/png")
			cmd.Stdin = bytes.NewReader(data)
			if cmd.Run() == nil {
				return
			}
		}
	} else {
		if bin, err := exec.LookPath("xclip"); err == nil {
			cmd := exec.Command(bin, "-selection", "clipboard", "-t", "image/png", "-i")
			cmd.Stdin = bytes.NewReader(data)
			if cmd.Run() == nil {
				return
			}
		}
	}
	if err := clipboard.Init(); err == nil {
		clipboard.Write(clipboard.FmtImage, data)
	}
}

func showNotification(msg string) {
	if path, err := exec.LookPath("notify-send"); err == nil {
		exec.Command(path, "-t", "2000", "Clipboard Manager", msg).Run()
	}
}
