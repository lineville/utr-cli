package main

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

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	greenStyle        = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("10"))
	blueStyle         = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("14"))
)

// -----------------------------------------------------------------------------
// Player Item Implementation
// -----------------------------------------------------------------------------

func (d Player) Height() int                             { return 1 }
func (d Player) Spacing() int                            { return 0 }
func (d Player) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (p Player) FilterValue() string {
	return p.Source.DisplayName
}

// Renders a player item
func (d Player) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Player)
	if !ok {
		return
	}

	player := fmt.Sprintf("%d. %s", index+1, i.Source.DisplayName+" ("+i.Source.Location.Display+")")

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
	return e.Name + " [" + formatDate(e.StartDate) + " - " + formatDate(e.EndDate) + "]"
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
	i, ok := listItem.(Event)
	if !ok {
		return
	}

	eventRows := make([]string, len(i.Draws)+1)
	event := fmt.Sprintf("%d. %s", index+1, i.Title())
	eventRows[0] = event

	playerName := strings.Split(m.Title, "'s Match Results")[0]
	for i, draw := range i.Draws {
		eventRows[i+1] = draw.Name + " " + formatDrawWinLoss(draw, playerName)
	}

	fn := func(s ...string) string { return itemStyle.Render(strings.Join(s, "\n   • ")) }
	if index == m.Index() {
		fn = func(s ...string) string {
			eventPrintout := selectedItemStyle.Render("→ " + event)
			for i := range eventRows {
				if i == 0 {
					continue
				}
				if i%2 == 0 {
					eventPrintout += greenStyle.Render("\n   • " + eventRows[i])
				} else {
					eventPrintout += blueStyle.Render("\n   • " + eventRows[i])
				}
			}

			return eventPrintout
		}
	}

	fmt.Fprint(w, fn(eventRows...))
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
