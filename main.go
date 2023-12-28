package main

import (
	"fmt"
	"os"
	"strings"

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
	searchingUSTARanking
	searchingUSTARankingsBySection
	loading
)

type model struct {
	mode             int
	searchQuery      textinput.Model
	playerList       list.Model
	resultsList      list.Model
	selectedPlayer   list.Item
	spinner          spinner.Model
	commandSelection string
	cursor           int
}

var commandChoices = []string{"Match Results (UTR)", "Ranking (USTA)", "Rankings by Section (USTA)"}

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

				case "Ranking (USTA)": // TODO
					return m, tea.Quit

				case "Rankings by Section (USTA)": // TODO
					return m, tea.Quit
				}

			case searchingUTRPlayer:
				m.mode = loading
				return m, internal.SearchPlayers(m.searchQuery.Value())

			case selectingUTRPlayer:
				m.mode = viewingUTRPlayerResults
				m.selectedPlayer, _ = m.playerList.SelectedItem().(internal.Player)
				return m, internal.PlayerProfile(m.selectedPlayer.(internal.Player).Source.Id)
			}

		case "up", "k":
			if m.mode == initialSelection {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(commandChoices) - 1
				}
				return m, nil
			}

		case "down", "j":
			if m.mode == initialSelection {
				m.cursor++
				if m.cursor >= len(commandChoices) {
					m.cursor = 0
				}
				return m, nil
			}

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
				return m, internal.SearchPlayers(m.searchQuery.Value())
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
		s := strings.Builder{}
		s.WriteString("\n  What kind of search would you like to perform?\n\n\n")

		for i := 0; i < len(commandChoices); i++ {
			if m.cursor == i {
				s.WriteString("→  ")
			} else {
				s.WriteString(" ")
			}
			s.WriteString(commandChoices[i])
			s.WriteString("\n\n")
		}

		return s.String()

	case searchingUTRPlayer:
		return "\n  Search for a Tennis or PickleBall player\n\n" + m.searchQuery.View()

	case selectingUTRPlayer:
		return "\n\n" + m.playerList.View()

	case viewingUTRPlayerResults:
		return "\n\n" + m.resultsList.View()

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
