package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tebeka/selenium"
)

// -----------------------------------------------------------------------------
// API Calls
// -----------------------------------------------------------------------------

const utrBaseUrl = "https://app.universaltennis.com/api"
const ustaBaseUrl = "https://www.usta.com/en/home/play/rankings.html?"
const htmlTarget = "v-grid-cell__content"

var sectionCodes = map[string]string{
	"Eastern":             "S10",
	"Florida":             "S15",
	"Hawaii Pacific":      "S20",
	"Intermountain":       "S25",
	"Mid-Atlantic":        "S30",
	"Middle States":       "S35",
	"Midwest":             "S85",
	"Missouri Valley":     "S40",
	"New England":         "S45",
	"Northern California": "S50",
	"Northern":            "S55",
	"Pacific NW":          "S60",
	"Southern California": "S65",
	"Southern":            "S70",
	"Southwest":           "S75",
	"Texas":               "S80",
	"Unassigned":          "SS00",
}

// Searches for players by name (only shows 5 results)
func SearchUTRPlayers(player string) tea.Cmd {
	return func() tea.Msg {
		r := PlayerSearchResults{}
		req, err := http.NewRequest("GET", utrBaseUrl+"/v2/search/players?query="+url.PathEscape(player), nil)
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

// Searches for players in the USTA
func SearchUSTAPlayers(player string, format string, gender string, level string, section string) tea.Cmd {
	return func() tea.Msg {
		driver, err := CreateDriver()
		if err != nil {
			return errorMessage{err}
		}

		driver.Get(ustaBaseUrl + "ntrp-searchText" + url.PathEscape(player) + "ntrp-matchFormat" + url.PathEscape(format) + "ntrp-rankListGender" + url.PathEscape(gender) + "ntrp-ntrpPlayerLevel" + url.PathEscape(level) + "ntrp-sectionCode" + url.PathEscape(sectionCodes[section]))

		elements, err := driver.FindElements(selenium.ByClassName, htmlTarget)
		if err != nil {
			return errorMessage{err}
		}

		fmt.Println(elements)
		return nil
	}
}

// Gets a player's profile by id
func PlayerProfile(playerId int) tea.Cmd {
	return func() tea.Msg {
		r := Profile{}
		req, err := http.NewRequest("GET", utrBaseUrl+"/v1/player/"+strconv.Itoa(playerId), nil)
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
func PlayerResults(playerId int) tea.Cmd {
	return func() tea.Msg {
		r := MatchResults{}
		req, err := http.NewRequest("GET", utrBaseUrl+"/v1/player/"+strconv.Itoa(playerId)+"/results", nil)
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
