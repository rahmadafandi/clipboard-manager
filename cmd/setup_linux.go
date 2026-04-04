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

type pkgInfo struct {
	bin    string // binary to check in PATH
	aptPkg string
	dnfPkg string
	pacPkg string
	zypper string
}

func requiredPkgs() []pkgInfo {
	// Pick the right launcher based on display server
	launcherPkg := pkgInfo{"rofi", "rofi", "rofi", "rofi", "rofi"}
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		launcherPkg = pkgInfo{"wofi", "wofi", "wofi", "wofi", "wofi"}
	}
	return []pkgInfo{
		launcherPkg,
		{"notify-send", "libnotify-bin", "libnotify", "libnotify", "libnotify-tools"},
	}
}

func detectElevator() string {
	// pkexec shows a GUI password dialog (polkit) — works without a terminal
	if _, err := exec.LookPath("pkexec"); err == nil {
		return "pkexec"
	}
	return "sudo"
}

func detectPkgManager() (name string, installCmd []string) {
	elevator := detectElevator()
	managers := []struct {
		bin  string
		name string
		args []string
	}{
		{"apt-get", "apt", []string{elevator, "apt-get", "install", "-y"}},
		{"dnf", "dnf", []string{elevator, "dnf", "install", "-y"}},
		{"pacman", "pacman", []string{elevator, "pacman", "-S", "--noconfirm"}},
		{"zypper", "zypper", []string{elevator, "zypper", "install", "-y"}},
	}
	for _, m := range managers {
		if _, err := exec.LookPath(m.bin); err == nil {
			return m.name, m.args
		}
	}
	return "", nil
}

func pkgNameForManager(pkg pkgInfo, manager string) string {
	switch manager {
	case "apt":
		return pkg.aptPkg
	case "dnf":
		return pkg.dnfPkg
	case "pacman":
		return pkg.pacPkg
	case "zypper":
		return pkg.zypper
	}
	return pkg.aptPkg
}

func installDependencies() error {
	var missing []pkgInfo

	for _, pkg := range requiredPkgs() {
		if _, err := exec.LookPath(pkg.bin); err != nil {
			missing = append(missing, pkg)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	manager, installCmd := detectPkgManager()
	if installCmd == nil {
		names := make([]string, len(missing))
		for i, m := range missing {
			names[i] = m.bin
		}
		return fmt.Errorf("missing dependencies: %s\nCould not detect package manager. Please install manually",
			strings.Join(names, ", "))
	}

	// Collect package names to install
	var pkgNames []string
	for _, m := range missing {
		pkgNames = append(pkgNames, pkgNameForManager(m, manager))
	}

	fmt.Printf("Installing dependencies: %s ...\n", strings.Join(pkgNames, ", "))

	args := append(installCmd, pkgNames...)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}
	fmt.Println("✔ Dependencies installed")
	return nil
}

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

func isSetupDone() bool {
	path, err := getAutostartPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
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
		path := strings.TrimSuffix(m[1], "/")
		dconf := path + "/"
		nameOut, err := exec.Command("gsettings", "get",
			gsettingsBindingSchema+":"+dconf, "name").Output()
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

	command := fmt.Sprintf("%s popup", exe)

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
