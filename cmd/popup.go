package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rahmadafandi/clipboard-manager/internal/storage"
)

type launcher struct {
	bin      string
	args     func() []string
	imgArgs  func() []string
	fmtLine  func(idx int, item storage.ClipItem, imgPath string) string
	parseIdx func(output string) (int, bool)
}

func defaultFmtLine(idx int, item storage.ClipItem, imgPath string) string {
	pin := ""
	if item.Pinned {
		pin = "[*] "
	}
	if item.Type == storage.Text {
		preview := strings.ReplaceAll(item.TextContent, "\n", " ")
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		return fmt.Sprintf("%d. %s%s", idx+1, pin, preview)
	}
	return fmt.Sprintf("%d. %s[Image] %d bytes", idx+1, pin, len(item.ImageData))
}

func sortPinnedFirst(items []storage.ClipItem) []storage.ClipItem {
	var pinned, unpinned []storage.ClipItem
	for i := len(items) - 1; i >= 0; i-- {
		if items[i].Pinned {
			pinned = append(pinned, items[i])
		} else {
			unpinned = append(unpinned, items[i])
		}
	}
	return append(pinned, unpinned...)
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

