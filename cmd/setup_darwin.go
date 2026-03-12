//go:build darwin

package cmd

import (
	"fmt"
	"os"
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
	return nil
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
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
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
