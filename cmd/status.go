package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rahmadafandi/clipboard-manager/internal/config"
	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show watcher status, history stats, and config",
	Run: func(cmd *cobra.Command, args []string) {
		// Watcher status
		pidFile, err := getPidFilePath()
		if err == nil {
			if data, err := os.ReadFile(pidFile); err == nil {
				pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
				process, err := os.FindProcess(pid)
				if err == nil && isProcessRunning(process) {
					fmt.Printf("Watcher:  running (PID %d)\n", pid)
				} else {
					fmt.Println("Watcher:  not running (stale PID file)")
				}
			} else {
				fmt.Println("Watcher:  not running")
			}
		}

		// History stats
		s, err := storage.NewFileStorage()
		if err != nil {
			fmt.Println("Storage:  error:", err)
			return
		}

		items, _ := s.Load()
		textCount, imageCount, pinnedCount := 0, 0, 0
		for _, item := range items {
			if item.Type == storage.Text {
				textCount++
			} else {
				imageCount++
			}
			if item.Pinned {
				pinnedCount++
			}
		}
		fmt.Printf("Items:    %d total (%d text, %d image, %d pinned)\n",
			len(items), textCount, imageCount, pinnedCount)

		// Disk usage
		if info, err := os.Stat(s.Path()); err == nil {
			fmt.Printf("Disk:     %s\n", formatBytes(info.Size()))
		}
		fmt.Printf("Path:     %s\n", s.Path())

		// Snippets
		snippets, _ := s.LoadSnippets()
		if len(snippets) > 0 {
			fmt.Printf("Snippets: %d\n", len(snippets))
		}

		// Config
		cfg, _ := config.Load()
		fmt.Printf("\nConfig:\n")
		fmt.Printf("  max_history:       %d\n", cfg.MaxHistory)
		if cfg.AutoExpireHours > 0 {
			fmt.Printf("  auto_expire_hours: %d\n", cfg.AutoExpireHours)
		} else {
			fmt.Printf("  auto_expire_hours: disabled\n")
		}
	},
}

func formatBytes(b int64) string {
	switch {
	case b >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
	case b >= 1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
