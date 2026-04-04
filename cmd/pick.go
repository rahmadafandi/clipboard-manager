package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
	"golang.design/x/clipboard"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type itemWrapper struct {
	item     storage.ClipItem
	origIdx  int // original index in storage (for delete/pin)
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

type delegate struct{}

func (d delegate) Height() int                               { return 1 }
func (d delegate) Spacing() int                              { return 0 }
func (d delegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d delegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(itemWrapper)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Title())

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + s[0])
		}
	}

	fmt.Fprint(w, fn(str))
}

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginLeft(2)
)

var pickCmd = &cobra.Command{
	Use:   "pick",
	Short: "Pick an item from history",
	Run: func(cmd *cobra.Command, args []string) {
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

		// Reverse items to show newest first
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
	},
}

type pickModel struct {
	list    list.Model
	storage *storage.FileStorage
}

func (m pickModel) Init() tea.Cmd {
	return nil
}

func (m pickModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(itemWrapper)
			if ok {
				err := clipboard.Init()
				if err == nil {
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
				// Reload list
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
	return func() tea.Msg {
		return reloadMsg{}
	}
}

func (m pickModel) View() string {
	help := helpStyle.Render("enter: copy • d: delete • p: pin/unpin • q: quit")
	return docStyle.Render(m.list.View()) + "\n" + help
}
