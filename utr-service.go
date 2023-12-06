package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

const baseUrl = "https://app.universaltennis.com/api"

// Searches for players by name (only shows 5 results)
func searchPlayers(player string) tea.Cmd {
	return func() tea.Msg {
		r := PlayerSearchResults{}
		req, err := http.NewRequest("GET", baseUrl+"/v2/search/players?query="+url.PathEscape(player), nil)
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
		req, err := http.NewRequest("GET", baseUrl+"/v1/player/"+strconv.Itoa(playerId), nil)
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
		req, err := http.NewRequest("GET", baseUrl+"/v1/player/"+strconv.Itoa(playerId)+"/results", nil)
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
