package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var setupRemove bool

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup autostart and Super+V global shortcut (--remove to undo)",
	Run: func(cmd *cobra.Command, args []string) {
		if setupRemove {
			if err := removeAutostart(); err != nil {
				fmt.Println("Failed to remove autostart:", err)
			} else {
				fmt.Println("✔ Autostart removed")
			}
			if err := removeKeybinding(); err != nil {
				fmt.Println("Failed to remove keybinding:", err)
			} else {
				fmt.Println("✔ Global shortcut removed")
			}
			return
		}

		if err := installDependencies(); err != nil {
			fmt.Println("Failed to install dependencies:", err)
			return
		}

		if err := setupAutostart(); err != nil {
			fmt.Println("Failed to setup autostart:", err)
			return
		}
		fmt.Println("✔ Autostart configured")

		if err := setupKeybinding(); err != nil {
			fmt.Println("Failed to setup keybinding:", err)
			return
		}
		fmt.Println("✔ Global shortcut Super+V configured")

		fmt.Println("\nSetup complete! Log out and back in, or run 'clipboard-manager start' now.")
	},
}

func init() {
	setupCmd.Flags().BoolVar(&setupRemove, "remove", false, "Remove autostart and keybinding")
	rootCmd.AddCommand(setupCmd)
}
