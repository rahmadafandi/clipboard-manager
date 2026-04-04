package cmd

import (
	"fmt"
	"strings"

	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
)

var searchCopy bool

var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search clipboard history and optionally copy the first match",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyword := strings.Join(args, " ")
		lower := strings.ToLower(keyword)

		s, err := storage.NewFileStorage()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		items, err := s.Load()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		var matches []storage.ClipItem
		for i := len(items) - 1; i >= 0; i-- {
			item := items[i]
			if item.Type == storage.Text && strings.Contains(strings.ToLower(item.TextContent), lower) {
				matches = append(matches, item)
			}
		}

		if len(matches) == 0 {
			fmt.Printf("No matches for '%s'\n", keyword)
			return
		}

		if searchCopy {
			writeTextToClipboard(matches[0].TextContent)
			preview := matches[0].TextContent
			if len(preview) > 60 {
				preview = preview[:60] + "..."
			}
			fmt.Printf("Copied: %s\n", preview)
			return
		}

		for i, m := range matches {
			preview := strings.ReplaceAll(m.TextContent, "\n", " ")
			if len(preview) > 80 {
				preview = preview[:80] + "..."
			}
			fmt.Printf("  %d. %s  (%s)\n", i+1, preview, m.Timestamp.Format("15:04:05"))
		}
		fmt.Printf("\n%d matches. Use --copy to copy the first match.\n", len(matches))
	},
}

func init() {
	searchCmd.Flags().BoolVar(&searchCopy, "copy", false, "Copy the first match to clipboard")
	rootCmd.AddCommand(searchCmd)
}
