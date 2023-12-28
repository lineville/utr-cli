package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lineville/utr-cli/internal"
)

// -----------------------------------------------------------------------------
// Main handles the main program flow Model, View and Update based on Elm arch
// -----------------------------------------------------------------------------

const (
	initialSelection = iota
	searchingUTRPlayer
	selectingUTRPlayer
	viewingUTRPlayerResults
	searchingUSTAPlayer
	searchingUSTAFormat
	searchingUSTAGender
	searchingUSTALevel
	searchingUSTASection
	searchingUSTARankingsBySection
	loading
)

type model struct {
	mode             int
	searchQuery      textinput.Model
	playerList       list.Model
	resultsList      list.Model
	selectedPlayer   list.Item
	selectedFormat   string
	selectedGender   string
	selectedLevel    string
	selectedSection  string
	spinner          spinner.Model
	commandSelection string
	cursor           int
}

var commandChoices = []string{"Match Results (UTR)", "Ranking (USTA)", "Rankings by Section (USTA)"}
var formatChoices = []string{"Singles", "Doubles"}
var genderChoices = []string{"M", "F"}
var levelChoices = []string{"3.0", "3.5", "4.0", "4.5", "5.0"}
var sectionChoices = []string{"Eastern", "Florida", "Hawaii Pacific", "Intermountain", "Mid-Atlantic", "Middle States", "Midwest", "Missouri Valley", "New England", "Northern California", "Northern", "Pacific NW", "Southern", "Southern California", "Southwest", "Texas", "Unassigned"}

// Defines the initial model of the program
func initialModel(args ...string) model {
	ti := textinput.New()
	ti.Prompt = "🔍 » "
	ti.PromptStyle.PaddingLeft(2)
	ti.Placeholder = "Player name"
	ti.Focus()
	ti.CharLimit = 156

	playerList := list.New(nil, internal.Player{}, 40, 20)
	playerList.Title = "Select a player"
	playerList.Styles.TitleBar.PaddingLeft(2)
	playerList.Styles.Title.AlignHorizontal(lipgloss.Center)
	playerList.Styles.Title.Background(lipgloss.Color("#25CCF7"))
	playerList.SetStatusBarItemName("player", "players")

	resultsList := list.New(nil, internal.Event{}, 40, 20)
	resultsList.Title = ""
	resultsList.Styles.TitleBar.PaddingLeft(2)
	resultsList.Styles.Title.AlignHorizontal(lipgloss.Center)
	resultsList.Styles.TitleBar.PaddingLeft(2)
	resultsList.Styles.Title.Background(lipgloss.Color("#25CCF7"))
	resultsList.SetStatusBarItemName("event", "events")

	s := spinner.New()
	s.Spinner = spinner.Meter
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#25CCF7"))

	playerList.SetSpinner(s.Spinner)
	resultsList.SetSpinner(s.Spinner)

	additionalKeyBindings := func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Back")),
		}
	}

	resultsList.AdditionalFullHelpKeys = additionalKeyBindings
	resultsList.AdditionalShortHelpKeys = additionalKeyBindings
	playerList.AdditionalFullHelpKeys = additionalKeyBindings
	playerList.AdditionalShortHelpKeys = additionalKeyBindings

	m := model{
		mode:           initialSelection,
		searchQuery:    ti,
		playerList:     playerList,
		resultsList:    resultsList,
		selectedPlayer: nil,
		spinner:        s,
	}

	return m
}

// Init initializes the model with some reasonable defaults
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles events from the UI and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.playerList.SetWidth(msg.Width)
		m.resultsList.SetWidth(msg.Width)
		m.playerList.SetHeight(msg.Height - 24)
		m.resultsList.SetHeight(msg.Height - 24)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			switch m.mode {

			case initialSelection:
				m.commandSelection = commandChoices[m.cursor]

				switch m.commandSelection {

				case "Match Results (UTR)":
					m.mode = searchingUTRPlayer
					return m, tea.Batch(
						m.spinner.Tick,
						textinput.Blink,
					)

				case "Ranking (USTA)":
					m.mode = searchingUSTAPlayer
					return m, tea.Batch(
						m.spinner.Tick,
						textinput.Blink,
					)

				case "Rankings by Section (USTA)":
					m.mode = searchingUSTARankingsBySection
					return m, tea.Batch(
						m.spinner.Tick,
						textinput.Blink,
					)
				}

			case searchingUTRPlayer:
				m.mode = loading
				return m, internal.SearchUTRPlayers(m.searchQuery.Value())

			case searchingUSTAPlayer:
				m.mode = searchingUSTAFormat
				m.cursor = 0
				return m, nil

			case searchingUSTAFormat:
				m.mode = searchingUSTAGender
				m.cursor = 0
				return m, nil

			case searchingUSTAGender:
				m.mode = searchingUSTALevel
				m.cursor = 0
				return m, nil

			case searchingUSTALevel:
				m.mode = searchingUSTASection
				m.cursor = 0
				return m, nil

			case searchingUSTASection:
				m.mode = loading
				m.cursor = 0
				return m, internal.SearchUSTAPlayers(m.searchQuery.Value(), m.selectedFormat, m.selectedGender, m.selectedLevel, m.selectedSection)

			case selectingUTRPlayer:
				m.mode = viewingUTRPlayerResults
				m.selectedPlayer, _ = m.playerList.SelectedItem().(internal.Player)
				return m, internal.PlayerProfile(m.selectedPlayer.(internal.Player).Source.Id)
			}

		case "up", "k":
			m.cursor--
			switch m.mode {
			case initialSelection:
				if m.cursor < 0 {
					m.cursor = len(commandChoices) - 1
				}
			case searchingUSTAFormat:
				if m.cursor < 0 {
					m.cursor = len(formatChoices) - 1
				}
			case searchingUSTAGender:
				if m.cursor < 0 {
					m.cursor = len(genderChoices) - 1
				}
			case searchingUSTALevel:
				if m.cursor < 0 {
					m.cursor = len(levelChoices) - 1
				}
			case searchingUSTASection:
				if m.cursor < 0 {
					m.cursor = len(sectionChoices) - 1
				}
			}
			return m, nil

		case "down", "j":
			m.cursor++
			switch m.mode {
			case initialSelection:
				if m.cursor >= len(commandChoices) {
					m.cursor = 0
				}
			case searchingUSTAFormat:
				if m.cursor >= len(formatChoices) {
					m.cursor = 0
				}
			case searchingUSTAGender:
				if m.cursor >= len(genderChoices) {
					m.cursor = 0
				}
			case searchingUSTALevel:
				if m.cursor >= len(levelChoices) {
					m.cursor = 0
				}
			case searchingUSTASection:
				if m.cursor >= len(sectionChoices) {
					m.cursor = 0
				}
			}
			return m, nil

		case "esc":
			switch m.mode {
			case initialSelection:
				return m, tea.Quit

			case searchingUTRPlayer:
				m.mode = initialSelection
				m.searchQuery.SetValue("")
				return m, nil

			case selectingUTRPlayer:
				m.mode = searchingUTRPlayer
				m.selectedPlayer = nil
				m.resultsList.SetItems(nil)
				return m, tea.Batch(
					m.spinner.Tick,
					textinput.Blink,
				)

			case viewingUTRPlayerResults:
				m.mode = selectingUTRPlayer
				m.resultsList.SetItems(nil)
				return m, internal.SearchUTRPlayers(m.searchQuery.Value())

			case searchingUSTAPlayer:
				m.mode = initialSelection
				m.searchQuery.SetValue("")
				return m, nil

			case searchingUSTAFormat:
				m.mode = searchingUSTAPlayer
				m.selectedFormat = ""
				return m, nil

			case searchingUSTAGender:
				m.mode = searchingUSTAFormat
				m.selectedGender = ""
				return m, nil

			case searchingUSTALevel:
				m.mode = searchingUSTAGender
				m.selectedLevel = ""
				return m, nil

			case searchingUSTASection:
				m.mode = searchingUSTALevel
				m.selectedSection = ""
				return m, nil

			case searchingUSTARankingsBySection:
				m.mode = initialSelection
				m.searchQuery.SetValue("")
				return m, nil
			}

		}

	case internal.PlayerSearchResults:
		players := make([]list.Item, len(msg.Players))
		for i, player := range msg.Players {
			players[i] = list.Item(player)
		}
		m.playerList.SetItems(players)
		m.mode = selectingUTRPlayer
		return m, nil

	case internal.MatchResults:
		matches := make([]list.Item, len(msg.Events))
		for i, event := range msg.Events {
			matches[i] = list.Item(event)
		}
		m.resultsList.SetItems(matches)
		m.resultsList.NewStatusMessage(fmt.Sprintf("\n\n Win/Loss ( %s)", msg.WinLossString))
		m.mode = viewingUTRPlayerResults
		return m, nil

	case internal.Profile:
		m.resultsList.Title = m.selectedPlayer.(internal.Player).Source.DisplayName + "'s Match Results"
		m.resultsList.Title += fmt.Sprintf("\n\nUTR (Singles: %.2f / Doubles: %.2f)", msg.SinglesUTR, msg.DoublesUTR)
		return m, internal.PlayerResults(m.selectedPlayer.(internal.Player).Source.Id)

	}

	switch m.mode {
	case searchingUTRPlayer:
		m.searchQuery, cmd = m.searchQuery.Update(msg)
		return m, cmd

	case searchingUSTAPlayer:
		m.searchQuery, cmd = m.searchQuery.Update(msg)
		return m, cmd

	case searchingUSTARankingsBySection:
		m.searchQuery, cmd = m.searchQuery.Update(msg)
		return m, cmd

	case selectingUTRPlayer:
		m.playerList, cmd = m.playerList.Update(msg)
		return m, cmd

	case viewingUTRPlayerResults:
		m.resultsList, cmd = m.resultsList.Update(msg)
		return m, cmd

	default:
		return m, nil
	}

}

// View renders the UI from the current model
func (m model) View() string {
	switch m.mode {
	case initialSelection:
		return internal.CreatePrompt("\n  What kind of search would you like to perform?\n\n\n", commandChoices, m.cursor)

	case searchingUTRPlayer:
		return "\n  Search for a Tennis or PickleBall player\n\n" + m.searchQuery.View()

	case selectingUTRPlayer:
		return "\n\n" + m.playerList.View()

	case viewingUTRPlayerResults:
		return "\n\n" + m.resultsList.View()

	case searchingUSTAPlayer:
		return "\n Search for a USTA Player\n\n" + m.searchQuery.View()

	case searchingUSTAFormat:
		return internal.CreatePrompt("\n  Select a match format\n\n\n", formatChoices, m.cursor)

	case searchingUSTAGender:
		return internal.CreatePrompt("\n  Select a gender\n\n\n", genderChoices, m.cursor)

	case searchingUSTALevel:
		return internal.CreatePrompt("\n  Select a level\n\n\n", levelChoices, m.cursor)

	case searchingUSTASection:
		return internal.CreatePrompt("\n  Select a section\n\n\n", sectionChoices, m.cursor)

	case searchingUSTARankingsBySection:
		return "\n Search for USTA Rankings by Section\n\n" + m.searchQuery.View()

	case loading:
		return "\n\n" + m.spinner.View() + " Searching..."

	default:
		return ""
	}
}

func main() {
	if _, err := tea.NewProgram(initialModel(os.Args[1:]...)).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
