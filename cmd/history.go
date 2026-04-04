package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Manage clipboard history (clear, search, export, import)",
}

// --- clear ---

var clearAll bool

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear clipboard history (keeps pinned items unless --all)",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := storage.NewFileStorage()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if clearAll {
			if err := s.ClearAll(); err != nil {
				fmt.Println("Error clearing history:", err)
				return
			}
			fmt.Println("All clipboard history cleared")
		} else {
			if err := s.Clear(); err != nil {
				fmt.Println("Error clearing history:", err)
				return
			}
			fmt.Println("Clipboard history cleared (pinned items kept)")
		}
	},
}

// --- search ---

var searchCopy bool

var historySearchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search clipboard history",
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

// --- export ---

var historyExportCmd = &cobra.Command{
	Use:   "export [file]",
	Short: "Export clipboard history to JSON (stdout or file)",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := storage.NewFileStorage()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		items, err := s.Load()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		if len(args) > 0 {
			if err := os.WriteFile(args[0], data, 0644); err != nil {
				fmt.Fprintln(os.Stderr, "Error writing file:", err)
				os.Exit(1)
			}
			fmt.Printf("Exported %d items to %s\n", len(items), args[0])
		} else {
			fmt.Println(string(data))
		}
	},
}

// --- import ---

var historyImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import clipboard history from JSON file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading file:", err)
			os.Exit(1)
		}

		var imported []storage.ClipItem
		if err := json.Unmarshal(data, &imported); err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing JSON:", err)
			os.Exit(1)
		}

		s, err := storage.NewFileStorage()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		existing, _ := s.Load()
		merged := append(existing, imported...)

		if err := s.Save(merged); err != nil {
			fmt.Fprintln(os.Stderr, "Error saving:", err)
			os.Exit(1)
		}

		fmt.Printf("Imported %d items (%d total now)\n", len(imported), len(merged))
	},
}

func init() {
	historyClearCmd.Flags().BoolVar(&clearAll, "all", false, "Clear all items including pinned")
	historySearchCmd.Flags().BoolVar(&searchCopy, "copy", false, "Copy the first match to clipboard")

	historyCmd.AddCommand(historyClearCmd)
	historyCmd.AddCommand(historySearchCmd)
	historyCmd.AddCommand(historyExportCmd)
	historyCmd.AddCommand(historyImportCmd)
	rootCmd.AddCommand(historyCmd)
}
