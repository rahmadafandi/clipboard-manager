package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
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

var importCmd = &cobra.Command{
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
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
}
