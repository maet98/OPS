package main

import (
	"flag"
	"fmt"
	"log"
	"maet98/scrapper/internal/config"
	"maet98/scrapper/internal/merge"
	"maet98/scrapper/internal/scrap"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error

var docStyle = lipgloss.NewStyle().Margin(1, 2)
var selected []string

type LineItem struct {
	title, desc string
}

func (i LineItem) Title() string       { return i.title }
func (i LineItem) Description() string { return i.desc }
func (i LineItem) FilterValue() string { return i.desc }

type state int

const (
	FETCHING state = iota
	CHOOSE
	CONFIRM
	DOWNLOADED
)

type model struct {
	list         list.Model
	spinner      spinner.Model
	quitting     bool
	err          error
	state        state
	windowWidth  int
	windowHeight int
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

type EpisodeFound struct {
	list list.Model
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if s, ok := msg.(string); ok {
		if s == "FETCHED" {
			m.state = DOWNLOADED
			return m, nil
		}
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth, m.windowHeight = msg.Width, msg.Height
	}

	switch m.state {
	case DOWNLOADED:
		switch msg.(type) {
		case tea.KeyMsg:
			return m, tea.Quit
		}
	case FETCHING:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			default:
				return m, nil
			}
		case errMsg:
			m.err = msg
			return m, nil
		case EpisodeFound:
			m.state = CHOOSE
			m.list = msg.list
			return m, nil
		default:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case CHOOSE:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "c" {
				if m.state == CHOOSE {
					m.state = CONFIRM
					return m, cmd
				}
			}
			if msg.String() == "enter" {
				value := m.list.Items()[m.list.Index()].FilterValue()
				m.list.NewStatusMessage(fmt.Sprintf("Episode: %s selected", value))
				selected = append(selected, value)
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		case tea.WindowSizeMsg:
			h, v := docStyle.GetFrameSize()
			m.list.SetSize(msg.Width-h, msg.Height-v)
		}
	case CONFIRM:
		spinner, cmd := m.spinner.Update(msg)
		m.spinner = spinner
		return m, cmd
	}

	return m, nil
}

func (m model) renderDetailView() string {
	// Use Lip Gloss to style your detail page
	header := lipgloss.NewStyle().Bold(true).Render("hola:")
	footer := "\n\n(press 'esc' to go back)"

	// Center the view on the screen
	content := lipgloss.JoinVertical(lipgloss.Left, header, "", footer)
	return lipgloss.Place(m.windowWidth, m.windowHeight, lipgloss.Center, lipgloss.Center, content)
}

func (m model) View() string {
	if m.state == DOWNLOADED {
		return "Episode have been downloaded\n\n. Press any key to leave."
	}
	if m.state == CHOOSE {
		return docStyle.Render(m.list.View())
	}

	if m.err != nil {
		return m.err.Error()
	}

	str := fmt.Sprintf("\n\n   %s...press q to quit\n\n", m.spinner.View())
	if m.quitting {
		return str + "\n"
	}
	return str
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

func fetchEpisodes() EpisodeFound {
	url := config.GetConfig().BaseUrl
	episodes := scrap.GetHomePage(url)

	var items []list.Item
	for _, episode := range episodes {
		episodeNumber := getEpisodeNumber(episode)
		item := LineItem{
			title: episodeNumber,
			desc:  episode,
		}
		items = append(items, item)
	}

	list := list.New(items, list.NewDefaultDelegate(), 0, 0)
	list.Title = "One piece episode"

	msg := EpisodeFound{list: list}

	return msg
}

func download() tea.Msg {
	log.Println("Initialize Download")
	for _, episode := range selected {
		log.Println(episode)
		episodeNumber := scrap.GetEpisode(episode)
		log.Println("episode number finished ", episodeNumber)
		merge.MergeToPdf(episodeNumber)
		log.Println("episode number merged to pdf ", episodeNumber)
	}
	return "FETCHED"
}

func getInitialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	episodes := fetchEpisodes()

	m := model{spinner: s, list: episodes.list, state: CHOOSE}

	return m
}

func start_tui() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()
	model := getInitialModel()
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

}

func getChapterUrl(chapterNumber int, episodes []string) string {
	chapter := strconv.Itoa(chapterNumber)
	for _, value := range episodes {
		if strings.Contains(value, chapter) {
			return value
		}
	}
	return ""
}

func rangeDownload(from, to int) {
	for from <= to {
		singleDownload(from)
		from++
	}
}

func singleDownload(chapter int) {
	url := config.GetConfig().BaseUrl
	episodes := scrap.GetHomePage(url)

	episodeUrl := getChapterUrl(chapter, episodes)
	episodeNumber := scrap.GetEpisode(episodeUrl)
	merge.MergeToPdf(episodeNumber)
	fmt.Println("The episode was downloaded")
}

func main() {
	chapter := flag.Int("chapter", -1, "Number of the chapter to download")
	from := flag.Int("from", -1, "Range of chapter to download")
	to := flag.Int("to", -1, "Range of chapter to download")
	flag.Parse()

	if *chapter == -1 && *from == -1 && *to == -1 {
		fmt.Print("You have to provide a chapter")
		return
	}

	if *chapter != -1 {
		singleDownload(*chapter)
	}

	if *from != -1 && *to != -1 {
		rangeDownload(*from, *to)
	}

	fmt.Println("You must provide a valid param.")

}
