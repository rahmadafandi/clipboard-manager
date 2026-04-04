package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
)

var mergeSep string

var mergeCmd = &cobra.Command{
	Use:   "merge <index1> <index2> [index3...]",
	Short: "Merge multiple text items and copy to clipboard",
	Long: `Merge multiple clipboard history items into one.
Indices are 1-based (newest first, as shown in 'pick' or 'popup').
Only text items can be merged.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
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

		if len(items) == 0 {
			fmt.Println("Clipboard history is empty")
			return
		}

		// Parse indices (1-based, newest first)
		var parts []string
		for _, arg := range args {
			idx, err := strconv.Atoi(arg)
			if err != nil || idx < 1 || idx > len(items) {
				fmt.Printf("Invalid index: %s (valid: 1-%d)\n", arg, len(items))
				return
			}
			// Convert from 1-based newest-first to actual index
			actualIdx := len(items) - idx
			item := items[actualIdx]
			if item.Type != storage.Text {
				fmt.Printf("Item %d is an image, cannot merge\n", idx)
				return
			}
			parts = append(parts, item.TextContent)
		}

		merged := strings.Join(parts, mergeSep)
		writeTextToClipboard(merged)

		preview := merged
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		fmt.Printf("Merged %d items and copied to clipboard:\n%s\n", len(parts), preview)
	},
}

func init() {
	mergeCmd.Flags().StringVar(&mergeSep, "sep", "\n", "Separator between merged items")
	rootCmd.AddCommand(mergeCmd)
}
