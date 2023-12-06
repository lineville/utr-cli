package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// -----------------------------------------------------------------------------
// Main handles the main program flow Model, View and Update based on Elm arch
// -----------------------------------------------------------------------------

type model struct {
	searching      bool
	searchQuery    textinput.Model
	playerList     list.Model
	resultsList    list.Model
	selectedPlayer list.Item
}

func initialModel(args ...string) model {
	ti := textinput.New()
	ti.Prompt = "🔍 » "
	ti.PromptStyle.PaddingLeft(2)
	ti.Placeholder = "Player name"
	ti.Focus()
	ti.CharLimit = 156

	playerList := list.New(nil, Player{}, 40, 20)
	playerList.Title = "Select a player"
	playerList.Styles.TitleBar.PaddingLeft(2)
	playerList.SetStatusBarItemName("player", "players")

	resultsList := list.New(nil, Event{}, 40, 20)
	resultsList.Title = ""
	resultsList.Styles.TitleBar.PaddingLeft(2)
	resultsList.SetStatusBarItemName("event", "events")

	m := model{
		searching:      true,
		searchQuery:    ti,
		playerList:     playerList,
		resultsList:    resultsList,
		selectedPlayer: nil,
	}
	// If a name was passed in via cli args search for it
	if len(args) > 0 {
		m.searchQuery.SetValue(args[0])
		ti.SetValue(args[0])
		m.searching = false
		return m
	}

	return m
}

// Init initializes the model with some reasonable defaults
func (m model) Init() tea.Cmd {
	// If a name was passed in via cli args search for it
	if m.searchQuery.Value() != "" {
		return searchPlayers(m.searchQuery.Value())
	}

	// Default to text input blinking
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
				return m, searchPlayers(m.searchQuery.Value())
			} else {
				m.selectedPlayer, _ = m.playerList.SelectedItem().(Player)
				return m, playerProfile(m.selectedPlayer.(Player).Source.Id)
			}
		}

	case PlayerSearchResults:
		players := make([]list.Item, len(msg.Players))
		for i, player := range msg.Players {
			players[i] = list.Item(player)
		}
		m.playerList.SetItems(players)
		return m, nil

	case MatchResults:
		matches := make([]list.Item, len(msg.Events))
		for i, event := range msg.Events {
			matches[i] = list.Item(event)
		}
		m.resultsList.SetItems(matches)
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
	if m.searching {
		return "\n  Search for a Tennis or PickleBall player\n\n" + m.searchQuery.View()
	}
	if m.selectedPlayer != nil {
		return "\n" + m.resultsList.View()
	} else {
		return "\n" + m.playerList.View()
	}
}

func main() {
	if _, err := tea.NewProgram(initialModel(os.Args[1:]...)).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
