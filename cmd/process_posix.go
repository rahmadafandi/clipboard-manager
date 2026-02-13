//go:build !windows

package cmd

import (
	"os"
	"os/exec"
	"syscall"
)

func configureSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Detach from terminal
	}
}

func isProcessRunning(process *os.Process) bool {
	// On Unix, FindProcess always succeeds, so we need to send signal 0
	if err := process.Signal(syscall.Signal(0)); err == nil {
		return true
	}
	return false
}

func terminateProcess(process *os.Process) error {
	return process.Signal(syscall.SIGTERM)
}
