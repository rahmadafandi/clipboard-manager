package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
)

type launcher struct {
	bin      string
	args     func() []string
	imgArgs  func() []string
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
	Short: "Pick from clipboard history using a lightweight popup",
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

func init() {
	rootCmd.AddCommand(popupCmd)
}
