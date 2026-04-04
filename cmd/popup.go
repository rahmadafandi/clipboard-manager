package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
	"golang.design/x/clipboard"
)

type launcher struct {
	bin      string
	args     func() []string
	imgArgs  func() []string                          // extra args when images are present
	fmtLine  func(idx int, item storage.ClipItem, imgPath string) string
	parseIdx func(output string) (int, bool)
}

func defaultFmtLine(idx int, item storage.ClipItem, imgPath string) string {
	if item.Type == storage.Text {
		preview := strings.ReplaceAll(item.TextContent, "\n", " ")
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		return fmt.Sprintf("%d. %s", idx+1, preview)
	}
	return fmt.Sprintf("%d. [Image] %d bytes", idx+1, len(item.ImageData))
}

func parseIdxFromNumber(output string) (int, bool) {
	parts := strings.SplitN(output, ".", 2)
	if len(parts) < 2 {
		return 0, false
	}
	num, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, false
	}
	return num - 1, true
}

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
			// wofi returns the full line; for img lines, parse after "text:"
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
		fmtLine: func(idx int, item storage.ClipItem, imgPath string) string {
			label := defaultFmtLine(idx, item, imgPath)
			if item.Type == storage.Image && imgPath != "" {
				// rofi dmenu icon format: text\0icon\x1fpath
				return label + "\x00icon\x1f" + imgPath
			}
			return label
		},
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
		fmtLine: func(idx int, item storage.ClipItem, imgPath string) string {
			label := defaultFmtLine(idx, item, imgPath)
			if item.Type == storage.Image && imgPath != "" {
				return label + "\x00icon\x1f" + imgPath
			}
			return label
		},
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

// saveImageThumbs writes image items to temp PNG files and returns a map of index -> path.
func saveImageThumbs(items []storage.ClipItem) (tmpDir string, paths map[int]string) {
	paths = make(map[int]string)
	hasImages := false
	for _, item := range items {
		if item.Type == storage.Image {
			hasImages = true
			break
		}
	}
	if !hasImages {
		return "", paths
	}

	tmpDir, err := os.MkdirTemp("", "clipboard-manager-thumbs-")
	if err != nil {
		return "", paths
	}

	for i, item := range items {
		if item.Type != storage.Image || len(item.ImageData) == 0 {
			continue
		}
		path := filepath.Join(tmpDir, fmt.Sprintf("thumb_%d.png", i))
		if err := os.WriteFile(path, item.ImageData, 0644); err == nil {
			paths[i] = path
		}
	}
	return tmpDir, paths
}

var popupCmd = &cobra.Command{
	Use:   "popup",
	Short: "Pick from clipboard history using a popup (rofi/dmenu/wofi)",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := storage.NewFileStorage()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error accessing storage:", err)
			os.Exit(1)
		}

		items, err := s.Load()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error loading history:", err)
			os.Exit(1)
		}

		if len(items) == 0 {
			showNotification("Clipboard history is empty")
			return
		}

		l, err := detectLauncher()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Reverse items (newest first)
		reversed := make([]storage.ClipItem, len(items))
		for i, item := range items {
			reversed[len(items)-1-i] = item
		}

		// Save image thumbnails to temp files
		tmpDir, imgPaths := saveImageThumbs(reversed)
		if tmpDir != "" {
			defer os.RemoveAll(tmpDir)
		}

		// Build display lines with image paths
		var lines []string
		for i, item := range reversed {
			lines = append(lines, l.fmtLine(i, item, imgPaths[i]))
		}

		input := strings.Join(lines, "\n")

		// Build launcher args, add image flags if there are images
		launcherArgs := l.args()
		if len(imgPaths) > 0 && l.imgArgs != nil {
			if extra := l.imgArgs(); len(extra) > 0 {
				launcherArgs = append(launcherArgs, extra...)
			}
		}

		c := exec.Command(l.bin, launcherArgs...)
		c.Stdin = strings.NewReader(input)
		c.Stderr = os.Stderr

		out, err := c.Output()
		if err != nil {
			return
		}

		result := strings.TrimSpace(string(out))
		if result == "" {
			return
		}

		idx, ok := l.parseIdx(result)
		if !ok || idx < 0 || idx >= len(reversed) {
			return
		}

		selected := reversed[idx]

		if selected.Type == storage.Text {
			writeTextToClipboard(selected.TextContent)
			showNotification("Copied to clipboard")
		} else {
			writeImageToClipboard(selected.ImageData)
			showNotification("Image copied to clipboard")
		}
	},
}

// writeTextToClipboard writes text using native tools, fallback to clipboard lib.
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
	// Fallback
	if err := clipboard.Init(); err == nil {
		clipboard.Write(clipboard.FmtText, []byte(text))
	}
}

// writeImageToClipboard writes image data using wl-copy/xclip so clipboard
// persists after the process exits.
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
	// Fallback: clipboard lib (image may not persist after exit)
	if err := clipboard.Init(); err == nil {
		clipboard.Write(clipboard.FmtImage, data)
	}
}

func showNotification(msg string) {
	if path, err := exec.LookPath("notify-send"); err == nil {
		exec.Command(path, "-t", "2000", "Clipboard Manager", msg).Run()
	}
}

func init() {
	rootCmd.AddCommand(popupCmd)
}
