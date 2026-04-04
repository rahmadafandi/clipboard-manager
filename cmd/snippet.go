package cmd

import (
	"fmt"
	"strings"

	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
)

var snippetCmd = &cobra.Command{
	Use:   "snippet",
	Short: "Manage saved text snippets",
}

var snippetAddCmd = &cobra.Command{
	Use:   "add <name> <content>",
	Short: "Save a text snippet",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		content := strings.Join(args[1:], " ")
		s, err := storage.NewFileStorage()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if err := s.AddSnippet(name, content); err != nil {
			fmt.Println("Error saving snippet:", err)
			return
		}
		fmt.Printf("Snippet '%s' saved\n", name)
	},
}

var snippetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved snippets",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := storage.NewFileStorage()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		snippets, err := s.LoadSnippets()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if len(snippets) == 0 {
			fmt.Println("No snippets saved")
			return
		}
		for _, sn := range snippets {
			preview := sn.Content
			if len(preview) > 60 {
				preview = preview[:60] + "..."
			}
			fmt.Printf("  %s: %s\n", sn.Name, preview)
		}
	},
}

var snippetRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a snippet",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s, err := storage.NewFileStorage()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if err := s.RemoveSnippet(args[0]); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("Snippet '%s' removed\n", args[0])
	},
}

var snippetCopyCmd = &cobra.Command{
	Use:   "copy <name>",
	Short: "Copy a snippet to clipboard",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s, err := storage.NewFileStorage()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		snippets, err := s.LoadSnippets()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		for _, sn := range snippets {
			if sn.Name == args[0] {
				writeTextToClipboard(sn.Content)
				fmt.Printf("Snippet '%s' copied to clipboard\n", sn.Name)
				return
			}
		}
		fmt.Printf("Snippet '%s' not found\n", args[0])
	},
}

func init() {
	snippetCmd.AddCommand(snippetAddCmd)
	snippetCmd.AddCommand(snippetListCmd)
	snippetCmd.AddCommand(snippetRemoveCmd)
	snippetCmd.AddCommand(snippetCopyCmd)
	rootCmd.AddCommand(snippetCmd)
}
