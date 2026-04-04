package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
	"golang.design/x/clipboard"
)

var (
	pickTUI   bool
	pickPaste bool
)

var pickCmd = &cobra.Command{
	Use:   "pick",
	Short: "Pick an item from clipboard history",
	Long:  `Opens a popup picker by default. Use --tui for terminal UI mode.`,
	Run: func(cmd *cobra.Command, args []string) {
		if pickTUI {
			runTUI()
			return
		}

		// Try popup first, fall back to TUI
		if _, err := detectLauncher(); err == nil {
			runPopup()
		} else {
			runTUI()
		}
	},
}

func init() {
	pickCmd.Flags().BoolVar(&pickTUI, "tui", false, "Force terminal UI mode")
	pickCmd.Flags().BoolVar(&pickPaste, "paste", false, "Auto-paste after selecting (simulate Ctrl+V)")
}

// --- Popup mode ---

func runPopup() {
	s, err := storage.NewFileStorage()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error accessing storage:", err)
		os.Exit(1)
	}

	items, err := s.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading history:", err)
		os.Exit(1)
	}

	if len(items) == 0 {
		showNotification("Clipboard history is empty")
		return
	}

	s.PurgeExpired()

	items, err = s.Load()
	if err != nil || len(items) == 0 {
		showNotification("Clipboard history is empty")
		return
	}

	l, err := detectLauncher()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sorted := sortPinnedFirst(items)

	tmpDir, imgPaths := saveImageThumbs(sorted)
	if tmpDir != "" {
		defer os.RemoveAll(tmpDir)
	}

	var lines []string
	for i, item := range sorted {
		lines = append(lines, l.fmtLine(i, item, imgPaths[i]))
	}

	input := strings.Join(lines, "\n")

	launcherArgs := l.args()
	if len(imgPaths) > 0 && l.imgArgs != nil {
		if extra := l.imgArgs(); len(extra) > 0 {
			launcherArgs = append(launcherArgs, extra...)
		}
	}

	c := exec.Command(l.bin, launcherArgs...)
	c.Stdin = strings.NewReader(input)
	c.Stderr = os.Stderr

	out, err := c.Output()
	if err != nil {
		return
	}

	result := strings.TrimSpace(string(out))
	if result == "" {
		return
	}

	idx, ok := l.parseIdx(result)
	if !ok || idx < 0 || idx >= len(sorted) {
		return
	}

	selected := sorted[idx]

	if selected.Type == storage.Text {
		writeTextToClipboard(selected.TextContent)
	} else {
		writeImageToClipboard(selected.ImageData)
	}

	if pickPaste {
		autoPaste()
	} else {
		if selected.Type == storage.Text {
			showNotification("Copied to clipboard")
		} else {
			showNotification("Image copied to clipboard")
		}
	}
}

// --- TUI mode ---

var docStyle = lipgloss.NewStyle().Margin(1, 2)
var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginLeft(2)

type itemWrapper struct {
	item    storage.ClipItem
	origIdx int
}

func (i itemWrapper) FilterValue() string {
	if i.item.Type == storage.Text {
		return i.item.TextContent
	}
	return "Image"
}

func (i itemWrapper) Title() string {
	prefix := ""
	if i.item.Pinned {
		prefix = "[*] "
	}
	if i.item.Type == storage.Text {
		content := i.item.TextContent
		if len(content) > 50 {
			content = content[:50] + "..."
		}
		return prefix + content
	}
	return fmt.Sprintf("%s[Image] %d bytes", prefix, len(i.item.ImageData))
}

func (i itemWrapper) Description() string {
	ts := i.item.Timestamp.Format("15:04:05")
	if i.item.Pinned {
		ts += " (pinned)"
	}
	return ts
}

func runTUI() {
	s, err := storage.NewFileStorage()
	if err != nil {
		fmt.Println("Error accessing storage:", err)
		return
	}

	items, err := s.Load()
	if err != nil {
		fmt.Println("Error loading history:", err)
		return
	}

	if len(items) == 0 {
		fmt.Println("Clipboard history is empty.")
		return
	}

	var teaItems []list.Item
	for i := len(items) - 1; i >= 0; i-- {
		teaItems = append(teaItems, itemWrapper{item: items[i], origIdx: i})
	}

	l := list.New(teaItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Clipboard History"

	m := pickModel{list: l, storage: s}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

type pickModel struct {
	list    list.Model
	storage *storage.FileStorage
}

func (m pickModel) Init() tea.Cmd { return nil }

func (m pickModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(itemWrapper)
			if ok {
				if err := clipboard.Init(); err == nil {
					if i.item.Type == storage.Text {
						clipboard.Write(clipboard.FmtText, []byte(i.item.TextContent))
					} else {
						clipboard.Write(clipboard.FmtImage, i.item.ImageData)
					}
				}
			}
			return m, tea.Quit
		case "d", "delete":
			i, ok := m.list.SelectedItem().(itemWrapper)
			if ok {
				m.storage.Delete(i.origIdx)
				return m, m.reloadItems()
			}
		case "p":
			i, ok := m.list.SelectedItem().(itemWrapper)
			if ok {
				m.storage.TogglePin(i.origIdx)
				return m, m.reloadItems()
			}
		}
	case reloadMsg:
		items, err := m.storage.Load()
		if err != nil || len(items) == 0 {
			return m, tea.Quit
		}
		var teaItems []list.Item
		for i := len(items) - 1; i >= 0; i-- {
			teaItems = append(teaItems, itemWrapper{item: items[i], origIdx: i})
		}
		m.list.SetItems(teaItems)
		return m, nil
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

type reloadMsg struct{}

func (m pickModel) reloadItems() tea.Cmd {
	return func() tea.Msg { return reloadMsg{} }
}

func (m pickModel) View() string {
	help := helpStyle.Render("enter: copy • d: delete • p: pin/unpin • q: quit")
	return docStyle.Render(m.list.View()) + "\n" + help
}
