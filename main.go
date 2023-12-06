package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// -----------------------------------------------------------------------------
// Main handles the main program flow Model, View and Update based on Elm architecture
// -----------------------------------------------------------------------------

type model struct {
	searching      bool
	searchQuery    textinput.Model
	playerList     list.Model
	resultsList    list.Model
	selectedPlayer list.Item
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Player name"
	ti.Focus()
	ti.CharLimit = 156

	playerList := list.New(nil, Player{}, 40, 20)
	playerList.Title = "Select a player"
	playerList.SetShowStatusBar(false)

	ps := spinner.New()
	ps.Spinner = spinner.Dot
	ps.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	playerList.SetSpinner(ps.Spinner)

	resultsList := list.New(nil, Event{}, 40, 20)
	resultsList.Title = ""
	resultsList.SetShowStatusBar(false)
	rs := spinner.New()
	rs.Spinner = spinner.Dot
	rs.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	resultsList.SetSpinner(rs.Spinner)

	return model{
		searching:      true,
		searchQuery:    ti,
		playerList:     playerList,
		resultsList:    resultsList,
		selectedPlayer: nil,
	}
}

// Init initializes the model with some reasonable defaults
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles events from the UI and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.playerList.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			if m.searching {
				m.searching = false
				m.playerList.ToggleSpinner()
				return m, searchPlayers(m.searchQuery.Value())
			} else {
				m.selectedPlayer, _ = m.playerList.SelectedItem().(Player)
				m.resultsList.ToggleSpinner()
				return m, playerProfile(m.selectedPlayer.(Player).Source.Id)
			}
		}

	case PlayerSearchResults:
		players := make([]list.Item, len(msg.Players))
		for i, player := range msg.Players {
			players[i] = list.Item(player)
		}
		m.playerList.SetItems(players)
		m.playerList.ToggleSpinner()
		return m, nil

	case MatchResults:
		matches := make([]list.Item, len(msg.Events))
		for i, event := range msg.Events {
			matches[i] = list.Item(event)
		}
		m.resultsList.SetItems(matches)
		m.resultsList.ToggleSpinner()
		return m, nil

	case Profile:
		m.resultsList.Title = m.selectedPlayer.(Player).Source.DisplayName + "'s Match Results"
		return m, playerResults(m.selectedPlayer.(Player).Source.Id)

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

	// Show text input if we are searching
	if m.searching {
		return fmt.Sprintf(
			"Search for a player by name\n\n%s\n\n%s",
			m.searchQuery.View(),
			"(esc to quit)",
		) + "\n"
	}
	// A player has been selected, show their results
	if m.selectedPlayer != nil {
		return "\n" + m.resultsList.View()
	} else {
		return "\n" + m.playerList.View()
	}
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
