package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
	"golang.design/x/clipboard"
)

type launcher struct {
	bin  string
	args func() []string
}

var launchers = []launcher{
	{"rofi", func() []string {
		return []string{"-dmenu", "-i", "-p", "Clipboard", "-format", "i",
			"-theme-str", "window {width: 400px;} listview {lines: 10;}"}
	}},
	{"dmenu", func() []string {
		return []string{"-i", "-l", "10", "-p", "Clipboard"}
	}},
	{"wofi", func() []string {
		return []string{"--dmenu", "-i", "-p", "Clipboard", "--lines", "10"}
	}},
	{"fuzzel", func() []string {
		return []string{"--dmenu", "--prompt", "Clipboard > ", "--lines", "10"}
	}},
}

func detectLauncher() (*launcher, error) {
	for i := range launchers {
		if _, err := exec.LookPath(launchers[i].bin); err == nil {
			return &launchers[i], nil
		}
	}
	return nil, fmt.Errorf("no supported launcher found (install rofi, dmenu, wofi, or fuzzel)")
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

		// Build display lines (newest first)
		var lines []string
		reversed := make([]storage.ClipItem, len(items))
		for i, item := range items {
			reversed[len(items)-1-i] = item
		}

		for i, item := range reversed {
			var label string
			if item.Type == storage.Text {
				// Single line preview, truncate
				preview := strings.ReplaceAll(item.TextContent, "\n", " ")
				if len(preview) > 80 {
					preview = preview[:80] + "..."
				}
				label = fmt.Sprintf("%d. %s", i+1, preview)
			} else {
				label = fmt.Sprintf("%d. [Image] %d bytes", i+1, len(item.ImageData))
			}
			lines = append(lines, label)
		}

		input := strings.Join(lines, "\n")

		// Run launcher
		launcherArgs := l.args()
		c := exec.Command(l.bin, launcherArgs...)
		c.Stdin = strings.NewReader(input)
		c.Stderr = os.Stderr

		out, err := c.Output()
		if err != nil {
			// User pressed Escape or closed the popup
			return
		}

		result := strings.TrimSpace(string(out))
		if result == "" {
			return
		}

		// Determine selected index
		var idx int
		if l.bin == "rofi" {
			// rofi with -format i returns the index directly
			idx, err = strconv.Atoi(result)
			if err != nil {
				return
			}
		} else {
			// dmenu/wofi/fuzzel return the text — parse the number prefix
			parts := strings.SplitN(result, ".", 2)
			if len(parts) < 2 {
				return
			}
			num, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return
			}
			idx = num - 1
		}

		if idx < 0 || idx >= len(reversed) {
			return
		}

		selected := reversed[idx]

		if err := clipboard.Init(); err != nil {
			fmt.Fprintln(os.Stderr, "Clipboard error:", err)
			os.Exit(1)
		}

		if selected.Type == storage.Text {
			clipboard.Write(clipboard.FmtText, []byte(selected.TextContent))
			showNotification("Copied to clipboard")
		} else {
			clipboard.Write(clipboard.FmtImage, selected.ImageData)
			showNotification("Image copied to clipboard")
		}
	},
}

func showNotification(msg string) {
	// Best-effort notification
	if path, err := exec.LookPath("notify-send"); err == nil {
		exec.Command(path, "-t", "2000", "Clipboard Manager", msg).Run()
	}
}

func init() {
	rootCmd.AddCommand(popupCmd)
}
