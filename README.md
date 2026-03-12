# Go Clipboard Manager (CLI)

A lightweight CLI clipboard manager built with Cobra and Bubbletea.

## Features
- **Watcher**: Daemon that records clipboard history.
- **Picker**: Interactive TUI to select and paste items.
- **Daemon Control**: Built-in `start` and `stop` commands.
- **Storage**: JSON file in user cache.

## Installation

### Option 1: Go Install (Recommended)
You can install the tool directly to your `$GOPATH/bin` (make sure it's in your `$PATH`):

```bash
go install github.com/rahmadafandi/clipboard-manager@latest
```

### Option 2: Build from Source

#### Linux / macOS
```bash
git clone https://github.com/rahmadafandi/clipboard-manager.git
cd clipboard-manager
go mod tidy
go build -o clipboard-manager
```

#### Windows (PowerShell)
```powershell
git clone https://github.com/rahmadafandi/clipboard-manager.git
cd clipboard-manager
go mod tidy
go build -o clipboard-manager.exe
```

## Quick Setup

Run the setup command to automatically configure **autostart on login** and **Super+V global shortcut**:

```bash
clipboard-manager setup
```

This will:
- Register the watcher daemon to start automatically on login
- Bind **Super+V** to open a lightweight popup picker (Linux/GNOME only, requires `rofi`, `dmenu`, `wofi`, or `fuzzel`)

To remove the setup:
```bash
clipboard-manager unsetup
```

## Usage

### 1. Start the Watcher
Start the background recording process:

```bash
clipboard-manager start
```

### 2. Stop the Watcher
```bash
clipboard-manager stop
```

### 3. Select from History
```bash
clipboard-manager pick
```
(or just `clipboard-manager` without arguments)

Use arrow keys to navigate and **Enter** to select. The item will be copied to your clipboard.

### 4. Popup Picker (lightweight, no terminal)
```bash
clipboard-manager popup
```
Opens a lightweight popup window (like Windows' Win+V) using rofi/dmenu/wofi/fuzzel. This is what **Super+V** uses after running `setup`.

### Platform Notes

#### Linux (GNOME)
`setup` fully supports autostart and Super+V keybinding via gsettings.

#### macOS
`setup` configures autostart via LaunchAgent. For the global shortcut, use Automator to create a Quick Action:
```bash
open -a Terminal /path/to/clipboard-manager --args pick
```
Then bind it in System Settings → Keyboard → Shortcuts.

#### Windows
`setup` configures autostart via the Startup folder. For the global shortcut, create a shortcut to `clipboard-manager.exe`, right-click → Properties → set Shortcut Key.
