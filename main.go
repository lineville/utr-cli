package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// -----------------------------------------------------------------------------
// List Item Implementation
// -----------------------------------------------------------------------------

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

type playerItemDelegate struct{}

func (d playerItemDelegate) Height() int                             { return 1 }
func (d playerItemDelegate) Spacing() int                            { return 0 }
func (d playerItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d playerItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Player)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Source.DisplayName+" ("+i.Source.Location.Display+")")

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type eventItemDelegate struct{}

func (d eventItemDelegate) Height() int                             { return 1 }
func (d eventItemDelegate) Spacing() int                            { return 0 }
func (d eventItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d eventItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Event)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Name)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// -----------------------------------------------------------------------------
// Type Definitions
// -----------------------------------------------------------------------------

type model struct {
	searching      bool
	searchQuery    textinput.Model
	playerList     list.Model
	resultsList    list.Model
	selectedPlayer list.Item
}

type PlayerSearchResults struct {
	Players []Player `json:"hits"`
	Total   int      `json:"total"`
}

type Player struct {
	Source struct {
		Id          int    `json:"id"`
		DisplayName string `json:"displayName"`
		Gender      string `json:"gender"`
		AgeRange    string `json:"ageRange"`
		Location    struct {
			Display string `json:"display"`
		} `json:"location"`
	} `json:"source"`
}

type Profile struct {
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	Gender      string  `json:"gender"`
	City        string  `json:"city"`
	State       string  `json:"state"`
	Nationality string  `json:"nationality"`
	SinglesUTR  float64 `json:"singlesUtr"`
	DoublesUTR  float64 `json:"doublesUtr"`
}

type MatchResults struct {
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	Events        []Event `json:"events"`
	WinLossString string  `json:"winLossString"`
}

type Event struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Draws     []Draw `json:"draws"`
}

type Draw struct {
	Id       int     `json:"id"`
	Name     string  `json:"name"`
	TeamType string  `json:"teamType"`
	Gender   string  `json:"gender"`
	Results  []Match `json:"results"`
}

type Match struct {
	Id       int               `json:"id"`
	Date     string            `json:"date"`
	Players  MatchParticipants `json:"players"`
	IsWinner bool              `json:"isWinner"`
	Score    Score             `json:"score"`
}

type MatchParticipants struct {
	Winner1 Profile `json:"winner1"`
	Winner2 Profile `json:"winner2"`
	Loser1  Profile `json:"loser1"`
	Loser2  Profile `json:"loser2"`
}

type Score struct {
	FirstSet  Set `json:"1"`
	SecondSet Set `json:"2"`
	ThirdSet  Set `json:"3"`
}

type Set struct {
	WinnerScore         int `json:"winner"`
	LoserScore          int `json:"loser"`
	LoserTieBreakScore  int `json:"tiebreak"`
	WinnerTieBreakScore int `json:"winnerTiebreak"`
}

type errorMessage struct {
	err error
}

// Needed to implement the list.Item interface
func (p Player) FilterValue() string {
	return p.Source.DisplayName
}

func (m Event) FilterValue() string {
	return m.Name
}

// -----------------------------------------------------------------------------
// Initial App State
// -----------------------------------------------------------------------------

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Player name"
	ti.Focus()
	ti.CharLimit = 156

	playerList := list.New(nil, playerItemDelegate{}, 40, 20)
	playerList.Title = "Select a player"
	playerList.SetShowStatusBar(false)

	resultsList := list.New(nil, eventItemDelegate{}, 40, 20)
	resultsList.SetShowStatusBar(false)

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
				m.playerList.StartSpinner()
				return m, searchPlayers(m.searchQuery.Value())
			} else {
				i, ok := m.playerList.SelectedItem().(Player)
				if ok {
					m.selectedPlayer = i
				}
				m.resultsList.StartSpinner()
				return m, playerProfile(m.selectedPlayer.(Player).Source.Id)
			}
		}

	case PlayerSearchResults:
		players := make([]list.Item, msg.Total)
		for i, player := range msg.Players {
			players[i] = list.Item(player)
		}
		m.playerList.SetItems(players)
		m.playerList.StopSpinner()
		return m, nil

	case MatchResults:
		matches := make([]list.Item, len(msg.Events))
		for i, event := range msg.Events {
			matches[i] = list.Item(event)
		}
		m.resultsList.SetItems(matches)
		m.resultsList.StopSpinner()
		return m, nil

	case Profile:
		m.resultsList.Title = m.selectedPlayer.(Player).Source.DisplayName + "'s Match Results"

		return m, playerResults(m.selectedPlayer.(Player).Source.Id)

	}

	if m.searching {
		m.searchQuery, cmd = m.searchQuery.Update(msg)
		return m, cmd
	} else {

		if m.selectedPlayer == nil {
			m.playerList, cmd = m.playerList.Update(msg)
			return m, cmd
		} else {
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

// Searches for players by name (only shows 5 results)
func searchPlayers(player string) tea.Cmd {
	return func() tea.Msg {
		r := PlayerSearchResults{}
		req, err := http.NewRequest("GET", "https://app.universaltennis.com/api/v2/search/players?query="+url.PathEscape(player), nil)
		if err != nil {
			return errorMessage{err}
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != 200 {
			return errorMessage{err}
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errorMessage{err}
		}

		err = json.Unmarshal(body, &r)
		if err != nil {
			return errorMessage{err}
		}
		return r
	}
}

// Gets a player's profile by id
func playerProfile(playerId int) tea.Cmd {
	return func() tea.Msg {
		r := Profile{}
		req, err := http.NewRequest("GET", "https://app.universaltennis.com/api/v1/player/"+strconv.Itoa(playerId), nil)
		if err != nil {
			return errorMessage{err}
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != 200 {
			return errorMessage{err}
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errorMessage{err}
		}

		err = json.Unmarshal(body, &r)
		if err != nil {
			return errorMessage{err}
		}
		return r
	}
}

// Gets a player's match results by id
func playerResults(playerId int) tea.Cmd {
	return func() tea.Msg {
		r := MatchResults{}
		req, err := http.NewRequest("GET", "https://app.universaltennis.com/api/v1/player/"+strconv.Itoa(playerId)+"/results", nil)
		if err != nil {
			return errorMessage{err}
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != 200 {
			return errorMessage{err}
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errorMessage{err}
		}

		err = json.Unmarshal(body, &r)
		if err != nil {
			return errorMessage{err}
		}
		return r
	}
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
