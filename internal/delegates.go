package internal

// -----------------------------------------------------------------------------
// Handles the List Item Rendering and other details
// -----------------------------------------------------------------------------

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style definitions
var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2).PaddingRight(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).PaddingRight(2).Foreground(lipgloss.Color("#25CCF7")).Border(lipgloss.RoundedBorder())
	greenStyle        = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#0be881"))
	redStyle          = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#f53b57"))
)

// -----------------------------------------------------------------------------
// Player Item Implementation
// -----------------------------------------------------------------------------

func (d Player) Height() int                             { return 1 }
func (d Player) Spacing() int                            { return 0 }
func (d Player) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (p Player) FilterValue() string {
	return p.Source.DisplayName + " (" + p.Source.Location.Display + ")"
}

// Renders a player item
func (d Player) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Player)
	if !ok {
		return
	}

	player := fmt.Sprintf("%d. %s (%s) [Age: %s]", index+1, i.Source.DisplayName, i.Source.Location.Display, i.Source.AgeRange)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("→ " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(player))
}

// -----------------------------------------------------------------------------
// Event Item Implementation
// -----------------------------------------------------------------------------

func (e Event) Title() string {
	return e.Name + " (" + formatDate(e.StartDate) + " - " + formatDate(e.EndDate) + ")"
}

func (e Event) Height() int  { return 1 }
func (e Event) Spacing() int { return 1 }

func (Event) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (e Event) FilterValue() string {
	return e.Title()
}

// Renders an event item
func (d Event) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	e, ok := listItem.(Event)
	if !ok {
		return
	}

	playerName := strings.Split(m.Title, "'s Match Results")[0]

	draws := make([]string, len(e.Draws)+1)
	draws[0] = fmt.Sprintf("%d. %s", index+1, e.Title())
	for i, d := range e.Draws {
		if d.Name == "" {
			draws[i+1] = d.TeamType + " " + formatDrawWinLoss(d, playerName)
		} else {
			draws[i+1] = d.Name + " " + formatDrawWinLoss(d, playerName)
		}
	}

	fn := func(s ...string) string { return itemStyle.Render(strings.Join(s, "\n   • ")) }

	if index == m.Index() {
		fn = func(s ...string) string {
			eventPrintout := selectedItemStyle.Render(fmt.Sprintf("→ %d. %s", index+1, e.Title()))
			for _, d := range e.Draws {
				if d.Name == "" {
					eventPrintout += itemStyle.Render("\n   • " + d.TeamType + " " + formatDrawWinLoss(d, playerName))
				} else {
					eventPrintout += itemStyle.Render("\n   • " + d.Name + " " + formatDrawWinLoss(d, playerName))
				}
				for _, result := range d.Results {
					eventPrintout += formatMatchScore(result, playerName)
				}
			}

			return eventPrintout
		}
	}

	fmt.Fprint(w, fn(draws...))
}

// -----------------------------------------------------------------------------
// Rendering Helpers
// -----------------------------------------------------------------------------

// Formats the match score for a draw
func formatMatchScore(match Match, playerName string) string {
	winner := match.Players.Winner1.FirstName + " " + match.Players.Winner1.LastName
	if match.Players.Winner2.FirstName != "" {
		winner += " / " + match.Players.Winner2.FirstName + " " + match.Players.Winner2.LastName
	}
	loser := match.Players.Loser1.FirstName + " " + match.Players.Loser1.LastName
	if match.Players.Loser2.FirstName != "" {
		loser += " / " + match.Players.Loser2.FirstName + " " + match.Players.Loser2.LastName
	}
	scoreString := fmt.Sprintf("%s def. %s", winner, loser)
	if match.Score.FirstSet.WinnerScore != 0 {
		scoreString += fmt.Sprintf(" (%d-%d", match.Score.FirstSet.WinnerScore, match.Score.FirstSet.LoserScore)
	} else {
		scoreString += " (ff"
	}

	if match.Score.SecondSet.WinnerScore != 0 {
		scoreString += fmt.Sprintf(", %d-%d", match.Score.SecondSet.WinnerScore, match.Score.SecondSet.LoserScore)
	}
	if match.Score.ThirdSet.WinnerScore != 0 {
		if match.Score.ThirdSet.WinnerScore == 1 {
			scoreString += fmt.Sprintf(", %d-%d", match.Score.ThirdSet.WinnerTieBreakScore, match.Score.ThirdSet.LoserTieBreakScore)
		} else {
			scoreString += fmt.Sprintf(", %d-%d", match.Score.ThirdSet.WinnerScore, match.Score.ThirdSet.LoserScore)
		}
	}
	scoreString += ")"

	if strings.Contains(strings.ToLower(winner), strings.ToLower(playerName)) {
		return greenStyle.Render("\n   ✅ " + scoreString)
	} else {
		return redStyle.Render("\n   ❌ " + scoreString)
	}
}

// Formats the win-loss record for a draw
func formatDrawWinLoss(draw Draw, playerName string) string {
	results := draw.Results
	wins := 0
	losses := 0
	for _, result := range results {
		name1 := result.Players.Winner1.FirstName + " " + result.Players.Winner1.LastName
		name2 := result.Players.Winner2.FirstName + " " + result.Players.Winner2.LastName
		if name1 == playerName || name2 == playerName {
			wins++
		} else {
			losses++
		}
	}
	return "(" + strconv.Itoa(wins) + " - " + strconv.Itoa(losses) + ")"
}

// Formats a date from the API to a more readable format
func formatDate(date string) string {
	layout := "2006-01-02T15:04:05"
	t, err := time.Parse(layout, date)
	if err != nil {
		fmt.Println(err)
	}
	return t.Format("01/02/2006")
}
