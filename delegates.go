package main

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
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
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
// Event Item Implementation
// -----------------------------------------------------------------------------

func (e Event) Title() string {
	return e.Name + " (" + formatDate(e.StartDate) + " - " + formatDate(e.EndDate) + ")"
}

func (e Event) Description() string {
	return strconv.Itoa(len(e.Draws)) + " draws"
}

func (e Event) Height() int  { return 1 }
func (e Event) Spacing() int { return 0 }

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

	str := fmt.Sprintf("%d. %s", index+1, i.Title()+" ("+i.Description()+")")

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("🎾 " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func formatDate(date string) string {
	layout := "2006-01-02T15:04:05"
	t, err := time.Parse(layout, date)
	if err != nil {
		fmt.Println(err)
	}
	return t.Format("January 2, 2006")
}
