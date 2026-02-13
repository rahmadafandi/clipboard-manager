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

## Usage

### 1. Start the Watcher
Start the background recording process:

#### Linux / macOS
```bash
./clipboard-manager start
```

#### Windows (PowerShell)
```powershell
.\clipboard-manager.exe start
```

### 2. Stop the Watcher
To stop the background process:

#### Linux / macOS
```bash
./clipboard-manager stop
```

#### Windows (PowerShell)
```powershell
.\clipboard-manager.exe stop
```

### 3. Select from History
To pick an item, simply run:

#### Linux / macOS
```bash
./clipboard-manager pick
```

#### Windows (PowerShell)
```powershell
.\clipboard-manager.exe pick
```
(or just `.\clipboard-manager.exe` without arguments)

Use arrow keys to navigate and **Enter** to select. The item will be copied to your clipboard.

## Global Keyboard Shortcuts

Since this is a CLI tool, if you bind it to a global shortcut (e.g., `Super+V`), you must run it **inside a terminal**.

### Linux (GNOME / Unity / etc.)
- **Command**: `gnome-terminal -- /path/to/clipboard-manager pick`
- Or if you use Alacritty: `alacritty -e /path/to/clipboard-manager pick`

### macOS
- You can use **Automator** to create a "Service" or "Quick Action" that runs a shell script:
  ```bash
  open -a Terminal /path/to/clipboard-manager --args pick
  ```
- Bind this service to a keyboard shortcut in System Settings -> Keyboard -> Shortcuts.

### Windows
- Create a Shortcut to `clipboard-manager.exe`.
- Right-click the shortcut -> Properties.
- Set "Shortcut key" to your desired combination (e.g., `Ctrl + Alt + V`).
- Note: This will open a console window briefly. To hide it, you might need a wrapper script or utility like AutoHotkey.
