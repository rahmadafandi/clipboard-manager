//go:build windows

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

func installDependencies() error {
	return nil
}

func isSetupDone() bool {
	path, err := getStartupPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func getStartupPath() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("APPDATA environment variable not set")
	}
	dir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	return filepath.Join(dir, "clipboard-manager.bat"), nil
}

func setupAutostart() error {
	exe, err := getExePath()
	if err != nil {
		return err
	}

	path, err := getStartupPath()
	if err != nil {
		return err
	}

	content := fmt.Sprintf("@echo off\nstart \"\" \"%s\" start\n", exe)
	return os.WriteFile(path, []byte(content), 0644)
}

func removeAutostart() error {
	path, err := getStartupPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func setupKeybinding() error {
	fmt.Println("  Note: Windows does not support automatic global shortcut registration.")
	fmt.Println("  To bind a shortcut, create a shortcut to clipboard-manager.exe,")
	fmt.Println("  right-click → Properties → set Shortcut Key (e.g., Ctrl+Alt+V).")
	return nil
}

func removeKeybinding() error {
	return nil
}
