package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func getPidFilePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cacheDir, "clipboard-manager")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "watcher.pid"), nil
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the clipboard watcher daemon",
	Run: func(cmd *cobra.Command, args []string) {
		// Auto-setup on first run
		if !isSetupDone() {
			fmt.Println("First run detected, running setup...")
			if err := installDependencies(); err != nil {
				fmt.Println("Failed to install dependencies:", err)
				return
			}
			if err := setupAutostart(); err != nil {
				fmt.Println("Warning: autostart setup failed:", err)
			} else {
				fmt.Println("✔ Autostart configured")
			}
			if err := setupKeybinding(); err != nil {
				fmt.Println("Warning: keybinding setup failed:", err)
			} else {
				fmt.Println("✔ Global shortcut Super+V configured")
			}
		}

		pidFile, err := getPidFilePath()
		if err != nil {
			fmt.Println("Error getting pid file path:", err)
			return
		}

		// Check if already running
		if _, err := os.Stat(pidFile); err == nil {
			// File exists, check if process actually runs
			pidBytes, _ := os.ReadFile(pidFile) // ignore error
			pid, _ := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
			process, err := os.FindProcess(pid)
			if err == nil {
				// Check if running using platform-specific method
				if isProcessRunning(process) {
					fmt.Printf("Watcher is already running (PID: %d)\n", pid)
					return
				}
			}
			// Stale PID file
			os.Remove(pidFile)
		}

		// Use the same executable
		exe, err := os.Executable()
		if err != nil {
			fmt.Println("Error getting executable path:", err)
			return
		}

		// Start watch command
		// We use "watch" subcommand
		cmdArgs := []string{"watch"}
		watchCmd := exec.Command(exe, cmdArgs...)
		configureSysProcAttr(watchCmd)

		if err := watchCmd.Start(); err != nil {
			fmt.Println("Failed to start watcher:", err)
			return
		}

		// Save PID
		pid := watchCmd.Process.Pid
		if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
			fmt.Println("Failed to write PID file:", err)
			// Try to kill if we can't save PID
			watchCmd.Process.Kill()
			return
		}

		fmt.Printf("Clipboard watcher started [PID: %d]\n", pid)
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the clipboard watcher daemon",
	Run: func(cmd *cobra.Command, args []string) {
		pidFile, err := getPidFilePath()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		data, err := os.ReadFile(pidFile)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Watcher is not running (no PID file)")
				return
			}
			fmt.Println("Error reading PID file:", err)
			return
		}

		pidStr := strings.TrimSpace(string(data))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			fmt.Println("Invalid PID in file:", err)
			return
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Println("Process not found:", err)
			os.Remove(pidFile)
			return
		}

		if err := terminateProcess(process); err != nil {
			fmt.Println("Failed to stop process:", err)
			return
		}

		os.Remove(pidFile)
		fmt.Println("Clipboard watcher stopped")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
}
