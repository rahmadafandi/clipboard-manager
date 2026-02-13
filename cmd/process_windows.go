//go:build windows

package cmd

import (
	"os"
	"os/exec"
)

func configureSysProcAttr(cmd *exec.Cmd) {
	// No Setsid on Windows
}

func isProcessRunning(process *os.Process) bool {
	// On Windows, finding a process that doesn't exist might return an error
	// or returning a process struct that fails on signal.
	// For now, we assume if we found it (via FindProcess logic in daemon), it might be running.
	// Proper check would require Windows API calls.
	return true
}

func terminateProcess(process *os.Process) error {
	return process.Kill()
}
