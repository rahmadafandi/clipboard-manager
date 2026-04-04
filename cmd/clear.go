package cmd

import (
	"fmt"

	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
)

var clearAll bool

var clearCmd = &cobra.Command{
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

func init() {
	clearCmd.Flags().BoolVar(&clearAll, "all", false, "Clear all items including pinned")
	rootCmd.AddCommand(clearCmd)
}
