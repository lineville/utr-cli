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

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
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

// -----------------------------------------------------------------------------
// Type Definitions
// -----------------------------------------------------------------------------

type model struct {
	showSearch     bool
	searchQuery    textinput.Model
	selectedPlayer list.Item
	list           list.Model
	quitting       bool
}

type PlayerSearchResult struct {
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

type PlayerProfileResult struct {
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	Gender      string  `json:"gender"`
	City        string  `json:"city"`
	State       string  `json:"state"`
	Nationality string  `json:"nationality"`
	SinglesUTR  float64 `json:"singlesUtr"`
	DoublesUTR  float64 `json:"doublesUtr"`
}

// Needed to implement the list.Item interface
func (p Player) FilterValue() string {
	return p.Source.DisplayName
}

// -----------------------------------------------------------------------------
// Initial App State
// -----------------------------------------------------------------------------

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Player name"
	ti.Focus()
	ti.CharLimit = 156

	return model{
		showSearch:     true,
		searchQuery:    ti,
		list:           list.New(nil, itemDelegate{}, 40, 16),
		quitting:       false,
		selectedPlayer: nil,
	}
}

// Init initializes the model with some reasonable defaults
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles events from the UI and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.showSearch {
				playerSearchResult, err := searchPlayers(m.searchQuery.Value())
				if err != nil || playerSearchResult.Total == 0 {
					fmt.Println("No player found.\n", err)
					return m, tea.Quit
				}

				// Only one result found, open the player profile
				if playerSearchResult.Total == 1 {
					m.selectedPlayer = playerSearchResult.Players[0]
					m.showSearch = false
				} else { // Multiple results found, show the list of options
					players := make([]list.Item, playerSearchResult.Total)
					for i, player := range playerSearchResult.Players {
						players[i] = list.Item(player)
					}
					if err != nil {
						fmt.Println("Error searching for player:", err)
					}
					m.list.SetItems(players)
					m.showSearch = false
				}
			} else {
				i, ok := m.list.SelectedItem().(Player)
				if ok {
					m.selectedPlayer = i
				}
			}
		}
	}

	var cmd tea.Cmd
	m.searchQuery, _ = m.searchQuery.Update(msg)
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the UI from the current model
func (m model) View() string {

	// Show text input if we are searching
	if m.showSearch {
		return fmt.Sprintf(
			"Search for a player by name\n\n%s\n\n%s",
			m.searchQuery.View(),
			"(esc to quit)",
		) + "\n"
	}
	// A player has been selected, show their results
	if m.selectedPlayer != nil {
		// Open the player profile
		//  Theres really not a lot of data in the profile but we can put some of it in the title of the list
		// instead lets search for the players results and show them in a navigatable list
		profile, err := playerProfile(m.selectedPlayer.(Player).Source.Id)
		if err != nil {
			fmt.Println("Error getting player profile:", err)
		}

		m.list.Title = m.selectedPlayer.(Player).Source.DisplayName + " (" + m.selectedPlayer.(Player).Source.Location.Display + ")\nSingles UTR: " + strconv.FormatFloat(profile.SinglesUTR, 'f', 2, 64) + " Doubles UTR: " + strconv.FormatFloat(profile.DoublesUTR, 'f', 2, 64)
		m.list.Styles.Title = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			AlignHorizontal(lipgloss.Center)
		m.list.SetShowStatusBar(false)

		return "\n" + m.list.View()
	}

	// Show the list of players from the search results
	return "\n" + m.list.View()
}

// Searches for players by name (only shows 5 results)
func searchPlayers(player string) (result PlayerSearchResult, err error) {

	r := PlayerSearchResult{}
	req, err := http.NewRequest("GET", "https://app.universaltennis.com/api/v2/search/players?query="+url.PathEscape(player), nil)
	if err != nil {
		fmt.Println("Error forming the request url:", err)
		return r, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("Error making request:", err)
		return r, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return r, err
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		fmt.Println("Error parsing response body:", err)
		return r, err
	}
	return r, nil
}

// Gets a player's profile by id
func playerProfile(playerId int) (result PlayerProfileResult, err error) {
	r := PlayerProfileResult{}
	req, err := http.NewRequest("GET", "https://app.universaltennis.com/api/v1/player/"+strconv.Itoa(playerId), nil)
	if err != nil {
		fmt.Println("Error forming the request url:", err)
		return r, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("Error making request:", err)
		return r, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return r, err
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		fmt.Println("Error parsing response body:", err)
		return r, err
	}
	return r, nil
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
