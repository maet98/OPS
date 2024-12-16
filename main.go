package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type LineItem struct {
	title, desc string
}

func (i LineItem) Title() string       { return i.title }
func (i LineItem) Description() string { return i.desc }
func (i LineItem) FilterValue() string { return i.title }

type model struct {
	list list.Model
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
			value := m.list.Items()[m.list.Index()].FilterValue()
			m.list.NewStatusMessage(fmt.Sprintf("Episode: %s selected", value))
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

func getEpisodeNumber(url string) string {
	splits := strings.Split(url, "-")
	for i, value := range splits {
		if value == "chapter" {
			return strings.TrimSuffix(splits[i+1], "/")
		}
	}
	return ""
}

func getHomePage() []string {
	var answer []string
	for i := 1; i < 200; i++ {
		episodeNumber := fmt.Sprintf("chapter-%d", i)
		answer = append(answer, episodeNumber)
	}
	return answer
}

func main() {
	// url := "https://w44.1piecemanga.com"
	episodes := getHomePage()
	var items []list.Item
	for _, episode := range episodes {
		episodeNumber := getEpisodeNumber(episode)
		item := LineItem{
			title: episodeNumber,
			desc:  fmt.Sprintf("chapter episode %s", episodeNumber),
		}
		items = append(items, item)
	}

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "One piece episode"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
