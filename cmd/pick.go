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
	item storage.ClipItem
}

func (i itemWrapper) FilterValue() string {
	if i.item.Type == storage.Text {
		return i.item.TextContent
	}
	return "Image"
}

func (i itemWrapper) Title() string {
	if i.item.Type == storage.Text {
		// Truncate for display
		content := i.item.TextContent
		if len(content) > 50 {
			content = content[:50] + "..."
		}
		return content
	}
	return fmt.Sprintf("[Image] %d bytes", len(i.item.ImageData))
}

func (i itemWrapper) Description() string {
	return i.item.Timestamp.Format("15:04:05")
}

// Delegate to render items
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
			teaItems = append(teaItems, itemWrapper{items[i]})
		}
		
		// Use default delegate for simplicity first
		l := list.New(teaItems, list.NewDefaultDelegate(), 0, 0)
		l.Title = "Clipboard History"

		m := model{list: l}

		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}

type model struct {
	list list.Model
	choice *storage.ClipItem
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			i, ok := m.list.SelectedItem().(itemWrapper)
			if ok {
				m.choice = &i.item
				
				// Paste logic
				// Note: clipboard write requires Init()
				// We do it here or in main
				// But Init() might stick the thread. 
				// The clipboard lib is a bit tricky with threads. 
				// Ensure we Init in the main function before running any command if possible, 
				// or just do it here.
				err := clipboard.Init()
				if err == nil { // Ignore if already init
					if i.item.Type == storage.Text {
						clipboard.Write(clipboard.FmtText, []byte(i.item.TextContent))
					} else {
						clipboard.Write(clipboard.FmtImage, i.item.ImageData)
					}
				}
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}
