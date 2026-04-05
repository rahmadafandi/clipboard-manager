# Go Clipboard Manager (CLI)

A lightweight cross-platform clipboard manager built with Go.

## Features
- **Clipboard Watcher**: Background daemon that records clipboard history (text & images)
- **Popup Picker**: Lightweight popup (Super+V) via wofi/rofi (Linux), choose (macOS), PowerShell (Windows)
- **TUI Picker**: Interactive terminal UI with search, delete, and pin
- **Pin Items**: Pin frequently used items so they stay at the top
- **Auto-expire**: Automatically remove old clipboard entries
- **Snippets**: Save and recall permanent text snippets
- **Configurable**: Max history size, auto-expire, preview settings
- **Auto-setup**: One command installs dependencies, autostart, and keybinding
- **Systemd Service**: Optional systemd user service on Linux
- **Cross-platform**: Linux, macOS, Windows

## Installation

```bash
go install github.com/rahmadafandi/clipboard-manager@latest
```

Or build from source:
```bash
git clone https://github.com/rahmadafandi/clipboard-manager.git
cd clipboard-manager
go build -o clipboard-manager
```

## Quick Setup

```bash
clipboard-manager start
```

On first run, this automatically:
- Installs dependencies (wofi/rofi, wl-clipboard, etc.)
- Configures autostart on login
- Binds **Super+V** to the popup picker (Linux/GNOME)
- Enables systemd user service (Linux, if available)

To remove: `clipboard-manager unsetup`

## Usage

### Popup Picker (Super+V)
```bash
clipboard-manager popup
```
Lightweight popup — pinned items shown first, image previews supported.

### TUI Picker
```bash
clipboard-manager pick
```
| Key | Action |
|-----|--------|
| Arrow keys | Navigate |
| Enter | Copy to clipboard |
| d / Delete | Delete item |
| p | Pin / unpin item |
| q / Ctrl+C | Quit |

### Daemon Control
```bash
clipboard-manager start    # Start watcher (auto-setup on first run)
clipboard-manager stop     # Stop watcher
```

### Clear History
```bash
clipboard-manager clear        # Clear history (keeps pinned)
clipboard-manager clear --all  # Clear everything
```

### Snippets
```bash
clipboard-manager snippet add myemail user@example.com
clipboard-manager snippet list
clipboard-manager snippet copy myemail
clipboard-manager snippet remove myemail
```

### Version
```bash
clipboard-manager --version
```

## Configuration

Config file: `~/.config/clipboard-manager/config.json`

```json
{
  "max_history": 50,
  "auto_expire_hours": 0,
  "keybinding": "<Super>v"
}
```

| Setting | Default | Description |
|---------|---------|-------------|
| `max_history` | 50 | Maximum number of clipboard items to keep |
| `auto_expire_hours` | 0 | Auto-delete items after N hours (0 = disabled) |
| `keybinding` | `<Super>v` | GNOME keybinding for popup picker (e.g. `<Ctrl><Alt>v`) |

After changing `keybinding`, re-run `clipboard-manager setup` to apply.

## Shell Completion

Enable tab completion for your shell:

**Bash:**
```bash
clipboard-manager completion bash | sudo tee /etc/bash_completion.d/clipboard-manager
```

**Zsh:**
```bash
clipboard-manager completion zsh > "${fpath[1]}/_clipboard-manager"
```

**Fish:**
```bash
clipboard-manager completion fish > ~/.config/fish/completions/clipboard-manager.fish
```

**PowerShell:**
```powershell
clipboard-manager completion powershell | Out-String | Invoke-Expression
```

Run `clipboard-manager completion <shell> --help` for more details.

## Platform Support

| Feature | Linux | macOS | Windows |
|---------|-------|-------|---------|
| Popup picker | wofi/rofi/dmenu/fuzzel | choose-gui (brew) | PowerShell WinForms |
| Image preview | wofi/rofi | - | - |
| Image clipboard persist | wl-copy / xclip | osascript | Native |
| Autostart | .desktop + systemd | LaunchAgent | Startup folder |
| Global shortcut | gsettings (GNOME) | Manual (Automator) | Manual (shortcut key) |
| Auto-install deps | pkexec / sudo | brew | Not needed |
