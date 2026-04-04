package cmd

import (
	"os"
	"path/filepath"
)

func getExePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}
