//go:build darwin

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const launchAgentPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.clipboard-manager.watcher</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>start</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
`

func installDependencies() error {
	if _, err := exec.LookPath("choose"); err == nil {
		return nil
	}

	if _, err := exec.LookPath("brew"); err != nil {
		fmt.Println("  Homebrew not found. Please install choose-gui manually:")
		fmt.Println("    brew install choose-gui")
		return nil
	}

	fmt.Println("Installing choose-gui via Homebrew...")
	cmd := exec.Command("brew", "install", "choose-gui")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install choose-gui: %w", err)
	}
	fmt.Println("✔ Dependencies installed")
	return nil
}

func isSetupDone() bool {
	path, err := getLaunchAgentPath()
	if err != nil {
		return false
	}
	if _, err := os.Stat(path); err != nil {
		return false
	}
	if _, err := exec.LookPath("choose"); err != nil {
		return false
	}
	return true
}

func getLaunchAgentPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "com.clipboard-manager.watcher.plist"), nil
}

func setupAutostart() error {
	exe, err := getExePath()
	if err != nil {
		return err
	}

	path, err := getLaunchAgentPath()
	if err != nil {
		return err
	}

	content := fmt.Sprintf(launchAgentPlist, exe)
	return os.WriteFile(path, []byte(content), 0644)
}

func removeAutostart() error {
	path, err := getLaunchAgentPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func setupKeybinding() error {
	fmt.Println("  Note: macOS does not support automatic global shortcut registration.")
	fmt.Println("  To bind Super+V, create an Automator Quick Action and assign it in")
	fmt.Println("  System Settings → Keyboard → Shortcuts → Services.")
	return nil
}

func removeKeybinding() error {
	return nil
}
