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
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#98FB98"))
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

func (d Event) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Event)
	if !ok {
		return
	}

	strs := make([]string, len(i.Draws)+1)
	event := fmt.Sprintf("%d. %s", index+1, i.Title())
	strs[0] = event
	for i, draw := range i.Draws {
		results := draw.Results
		wins := 0
		losses := 0
		for _, result := range results {
			if result.IsWinner {
				wins++
			} else {
				losses++
			}
		}
		strs[i+1] = draw.Name + " ( " + strconv.Itoa(wins) + " - " + strconv.Itoa(losses) + " )"
	}

	fn := func(s ...string) string { return itemStyle.Render(strings.Join(s, "\n   • ")) }
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("→ " + strings.Join(s, "\n   • "))
		}
	}

	fmt.Fprint(w, fn(strs...))
}

func formatDate(date string) string {
	layout := "2006-01-02T15:04:05"
	t, err := time.Parse(layout, date)
	if err != nil {
		fmt.Println(err)
	}
	return t.Format("January 2, 2006")
}
