package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var Version = ""

func getVersion() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

var rootCmd = &cobra.Command{
	Use:   "clipboard-manager",
	Short: "A CLI clipboard manager",
	Long:  `A clipboard manager that watches your clipboard history and allows you to select and paste items.`,
	Run: func(cmd *cobra.Command, args []string) {
		pickCmd.Run(cmd, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = getVersion()
	rootCmd.AddCommand(pickCmd)

	// watch is internal — hide from help
	watchCmd.Hidden = true
	rootCmd.AddCommand(watchCmd)
}
