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
```bash
git clone https://github.com/rahmadafandi/clipboard-manager.git
cd clipboard-manager
go mod tidy
go build -o clipboard-manager
```

## Usage

### 1. Start the Watcher
Start the background recording process:

```bash
./clipboard-manager start
```

To stop it later:
```bash
./clipboard-manager stop
```

### 2. Select from History
To pick an item, simply run:

```bash
./clipboard-manager
```
(or `./clipboard-manager pick`)

Use arrow keys to navigate and **Enter** to select. The item will be copied to your clipboard.

### 3. Global Keyboard Shortcut
Since this is a CLI tool, if you bind it to a global shortcut (e.g., `Super+V`), you must run it **inside a terminal**.

For example, on Ubuntu/GNOME:
- **Command**: `gnome-terminal -- /path/to/clipboard-manager pick`
- Or if you use Alacritty: `alacritty -e /path/to/clipboard-manager pick`

If you just run `/path/to/clipboard-manager`, it will run in the background and you won't see the UI.
