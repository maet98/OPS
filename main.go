package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
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

type state int

const (
	FETCHING state = iota
	CHOOSE
	CONFIRM
)

type model struct {
	list     list.Model
	spinner  spinner.Model
	quitting bool
	state    state
}

func (m model) Init() tea.Cmd {
	go m.fetchEpisodes()
	return m.spinner.Tick
}

type found tea.Msg

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.state == FETCHING {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			default:
				return m, nil
			}
		case found:
			var cmd tea.Cmd
			m.state = CHOOSE
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd

		default:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			value := m.list.Items()[m.list.Index()].FilterValue()
			m.list.NewStatusMessage(fmt.Sprintf("Episode: %s selected", value))
		}
		if msg.String() == "c" {
			if m.state == CHOOSE {
				m.state = CONFIRM
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	if m.state == CHOOSE {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.state == CHOOSE {
		return docStyle.Render(m.list.View())
	} else {
		str := fmt.Sprintf("\n\n   %s Loading forever...press q to quit\n\n", m.spinner.View(), m.state)
		return str
	}
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

func (model *model) fetchEpisodes() {
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

	model.list = list.New(items, list.NewDefaultDelegate(), 0, 0)
	found := true
	model.Update(found)
}

func getInitialModel() model {
	// url := "https://w44.1piecemanga.com"
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := model{spinner: s, state: FETCHING}
	m.list.Title = "One piece episode"

	return m
}

func main() {
	m := getInitialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
