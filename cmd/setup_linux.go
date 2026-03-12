//go:build linux

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const autostartDesktop = `[Desktop Entry]
Type=Application
Name=Clipboard Manager
Comment=Clipboard history watcher
Exec=%s start
Terminal=false
X-GNOME-Autostart-enabled=true
`

const gsettingsSchema = "org.gnome.settings-daemon.plugins.media-keys"
const gsettingsBindingSchema = "org.gnome.settings-daemon.plugins.media-keys.custom-keybinding"
const gsettingsPath = "/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings"

func getExePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}
	return exe, nil
}

func getAutostartPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, "autostart")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "clipboard-manager.desktop"), nil
}

func setupAutostart() error {
	exe, err := getExePath()
	if err != nil {
		return err
	}

	path, err := getAutostartPath()
	if err != nil {
		return err
	}

	content := fmt.Sprintf(autostartDesktop, exe)
	return os.WriteFile(path, []byte(content), 0644)
}

func removeAutostart() error {
	path, err := getAutostartPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func detectTerminal() string {
	// Try common terminals in order of preference
	terminals := []struct {
		bin  string
		args []string
	}{
		{"x-terminal-emulator", []string{"-e"}},
		{"gnome-terminal", []string{"--"}},
		{"konsole", []string{"-e"}},
		{"alacritty", []string{"-e"}},
		{"kitty", nil},
		{"xfce4-terminal", []string{"-e"}},
		{"xterm", []string{"-e"}},
	}

	for _, t := range terminals {
		if _, err := exec.LookPath(t.bin); err == nil {
			parts := []string{t.bin}
			parts = append(parts, t.args...)
			return strings.Join(parts, " ")
		}
	}
	return "x-terminal-emulator -e"
}

func findExistingBinding() string {
	// Check if we already have a clipboard-manager keybinding
	out, err := exec.Command("gsettings", "get", gsettingsSchema, "custom-keybindings").Output()
	if err != nil {
		return ""
	}

	raw := strings.TrimSpace(string(out))
	if raw == "@as []" || raw == "[]" {
		return ""
	}

	// Parse paths from the array
	re := regexp.MustCompile(`'([^']+)'`)
	matches := re.FindAllStringSubmatch(raw, -1)
	for _, m := range matches {
		path := m[1]
		nameOut, err := exec.Command("gsettings", "get",
			gsettingsBindingSchema+":"+path+"/", "name").Output()
		if err == nil && strings.Contains(string(nameOut), "Clipboard Manager") {
			return path
		}
	}
	return ""
}

func nextCustomSlot() (string, int) {
	out, err := exec.Command("gsettings", "get", gsettingsSchema, "custom-keybindings").Output()
	if err != nil {
		return gsettingsPath + "/custom0", 0
	}

	raw := strings.TrimSpace(string(out))
	if raw == "@as []" || raw == "[]" {
		return gsettingsPath + "/custom0", 0
	}

	// Find the highest existing customN index
	re := regexp.MustCompile(`custom(\d+)`)
	matches := re.FindAllStringSubmatch(raw, -1)
	maxIdx := -1
	for _, m := range matches {
		idx, _ := strconv.Atoi(m[1])
		if idx > maxIdx {
			maxIdx = idx
		}
	}

	next := maxIdx + 1
	return fmt.Sprintf("%s/custom%d", gsettingsPath, next), next
}

func setupKeybinding() error {
	// Check if gsettings is available (GNOME)
	if _, err := exec.LookPath("gsettings"); err != nil {
		return fmt.Errorf("gsettings not found — only GNOME/Unity desktops are supported for automatic keybinding setup")
	}

	exe, err := getExePath()
	if err != nil {
		return err
	}

	terminal := detectTerminal()
	command := fmt.Sprintf("%s %s pick", terminal, exe)

	// Check if binding already exists
	existingPath := findExistingBinding()
	var bindingPath string

	if existingPath != "" {
		bindingPath = existingPath
	} else {
		newPath, _ := nextCustomSlot()
		bindingPath = newPath
	}

	dconfPath := bindingPath + "/"

	// Set the keybinding properties
	cmds := [][]string{
		{"gsettings", "set", gsettingsBindingSchema + ":" + dconfPath, "name", "Clipboard Manager"},
		{"gsettings", "set", gsettingsBindingSchema + ":" + dconfPath, "command", command},
		{"gsettings", "set", gsettingsBindingSchema + ":" + dconfPath, "binding", "<Super>v"},
	}

	for _, c := range cmds {
		if out, err := exec.Command(c[0], c[1:]...).CombinedOutput(); err != nil {
			return fmt.Errorf("%s: %s", err, string(out))
		}
	}

	// If this is a new binding, add it to the custom-keybindings list
	if existingPath == "" {
		out, err := exec.Command("gsettings", "get", gsettingsSchema, "custom-keybindings").Output()
		if err != nil {
			return err
		}

		raw := strings.TrimSpace(string(out))
		var newList string
		if raw == "@as []" || raw == "[]" {
			newList = fmt.Sprintf("['%s/']", bindingPath)
		} else {
			// Remove trailing ] and append
			newList = strings.TrimSuffix(raw, "]")
			newList += fmt.Sprintf(", '%s/']", bindingPath)
		}

		if out, err := exec.Command("gsettings", "set", gsettingsSchema, "custom-keybindings", newList).CombinedOutput(); err != nil {
			return fmt.Errorf("%s: %s", err, string(out))
		}
	}

	return nil
}

func removeKeybinding() error {
	if _, err := exec.LookPath("gsettings"); err != nil {
		return nil // Not GNOME, nothing to remove
	}

	existingPath := findExistingBinding()
	if existingPath == "" {
		return nil // No binding found
	}

	// Remove the binding from the list
	out, err := exec.Command("gsettings", "get", gsettingsSchema, "custom-keybindings").Output()
	if err != nil {
		return err
	}

	raw := strings.TrimSpace(string(out))

	// Remove our entry from the list
	re := regexp.MustCompile(`'` + regexp.QuoteMeta(existingPath) + `/?'`)
	cleaned := re.ReplaceAllString(raw, "")
	// Clean up leftover commas
	cleaned = strings.ReplaceAll(cleaned, ", ,", ",")
	cleaned = strings.ReplaceAll(cleaned, "[,", "[")
	cleaned = strings.ReplaceAll(cleaned, ",]", "]")
	cleaned = strings.ReplaceAll(cleaned, ", ]", "]")
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "[]" || cleaned == "" {
		cleaned = "@as []"
	}

	if _, err := exec.Command("gsettings", "set", gsettingsSchema, "custom-keybindings", cleaned).CombinedOutput(); err != nil {
		return err
	}

	// Reset the keybinding slot
	dconfPath := existingPath + "/"
	exec.Command("gsettings", "reset", gsettingsBindingSchema+":"+dconfPath, "name").Run()
	exec.Command("gsettings", "reset", gsettingsBindingSchema+":"+dconfPath, "command").Run()
	exec.Command("gsettings", "reset", gsettingsBindingSchema+":"+dconfPath, "binding").Run()

	return nil
}
