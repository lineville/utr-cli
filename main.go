package main

import (
	"fmt"
	"os"
	"time"

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

type model struct {
	searching      bool
	loading        bool
	searchQuery    textinput.Model
	playerList     list.Model
	resultsList    list.Model
	selectedPlayer list.Item
	spinner        spinner.Model
}

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
		searching:      true,
		loading:        false,
		searchQuery:    ti,
		playerList:     playerList,
		resultsList:    resultsList,
		selectedPlayer: nil,
		spinner:        s,
	}
	// If a name was passed in via cli args search for it
	if len(args) > 0 {
		m.searchQuery.SetValue(args[0])
		ti.SetValue(args[0])
		m.searching = false
		m.loading = true
		return m
	}

	return m
}

// Init initializes the model with some reasonable defaults
func (m model) Init() tea.Cmd {
	// If a name was passed in via cli args search for it
	if m.searchQuery.Value() != "" {
		return tea.Batch(
			m.spinner.Tick,
			internal.SearchPlayers(m.searchQuery.Value()),
		)
	}

	return tea.Batch(
		m.spinner.Tick,
		textinput.Blink,
	)
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
			if m.selectedPlayer == nil {
				m.loading = true
			}
			if m.searching {
				m.searching = false
				return m, internal.SearchPlayers(m.searchQuery.Value())
			} else {
				m.selectedPlayer, _ = m.playerList.SelectedItem().(internal.Player)
				return m, internal.PlayerProfile(m.selectedPlayer.(internal.Player).Source.Id)
			}

		case "esc":
			if m.searching {
				return m, tea.Quit
			} else {
				if m.selectedPlayer == nil {
					m.searching = true
					m.playerList.SetItems(nil)
					return m, nil
				} else {
					m.selectedPlayer = nil
					m.resultsList.SetItems(nil)
					return m, internal.SearchPlayers(m.searchQuery.Value())
				}
			}
		}

	case internal.PlayerSearchResults:
		// sleep for a couple seconds to simulate loading
		time.Sleep(2 * time.Second)
		players := make([]list.Item, len(msg.Players))
		for i, player := range msg.Players {
			players[i] = list.Item(player)
		}
		m.playerList.SetItems(players)
		m.loading = false
		return m, nil

	case internal.MatchResults:
		matches := make([]list.Item, len(msg.Events))
		for i, event := range msg.Events {
			matches[i] = list.Item(event)
		}
		m.resultsList.SetItems(matches)
		m.resultsList.NewStatusMessage(fmt.Sprintf("\n\n Win/Loss ( %s)", msg.WinLossString))
		m.loading = false
		return m, nil

	case internal.Profile:
		m.resultsList.Title = m.selectedPlayer.(internal.Player).Source.DisplayName + "'s Match Results"
		m.resultsList.Title += fmt.Sprintf("\n\nUTR (Singles: %.2f / Doubles: %.2f)", msg.SinglesUTR, msg.DoublesUTR)
		return m, internal.PlayerResults(m.selectedPlayer.(internal.Player).Source.Id)

	}

	if m.searching {
		// Perform the default update to the text input
		m.searchQuery, cmd = m.searchQuery.Update(msg)
		return m, cmd
	} else {
		// No player is selected this is just moving cursor around the player list
		if m.selectedPlayer == nil {
			m.playerList, cmd = m.playerList.Update(msg)
			return m, cmd
		} else {
			// A player is selected, this is moving around the results list
			m.resultsList, cmd = m.resultsList.Update(msg)
			return m, cmd
		}
	}

}

// View renders the UI from the current model
func (m model) View() string {
	if m.searching {
		return "\n  Search for a Tennis or PickleBall player\n\n" + m.searchQuery.View()
	}
	if m.loading {
		if m.selectedPlayer != nil {
			return "\n\n" + m.spinner.View() + " Searching for " + m.selectedPlayer.(internal.Player).Source.DisplayName + "'s match results"
		} else {
			return "\n\n" + m.spinner.View() + " Searching for " + m.searchQuery.Value()
		}
	}
	if m.selectedPlayer != nil {
		return "\n\n" + m.resultsList.View()
	}

	return "\n\n" + m.playerList.View()
}

func main() {
	if _, err := tea.NewProgram(initialModel(os.Args[1:]...)).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
