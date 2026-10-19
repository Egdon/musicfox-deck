// Widgets example — demonstrates the five new composable widgets in foxful-cli.
//
// This example showcases Form, Table, Tabs, Tree, and FilePicker in a single
// layout. Press Tab to cycle through widgets. Each widget is self-contained
// and follows the standard Focus/Blur/Update/View interface.
//
// Navigation:
//
//	Tab/Shift+Tab — cycle focus between widgets
//	q/Ctrl+C      — quit
//
// Within each widget, standard navigation keys apply (up/down/left/right, etc.)
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
)

// focusable is the common interface satisfied by all five widgets.
type focusable interface {
	Focus()
	Blur()
	Focused() bool
	SetSize(width, height int)
	Update(tea.Msg) tea.Cmd
	View() string
}

type widgetModel struct {
	widgets []focusable
	titles  []string
	focus   int
	width   int
	height  int
}

func newWidgetModel() widgetModel {
	// Form
	form := model.NewForm([]model.FormField{
		{Key: "name", Label: "Name", Required: true},
		{Key: "email", Label: "Email", Required: true},
		{Key: "notes", Label: "Notes"},
	})

	// Table
	table := model.NewTable(
		[]model.Column{{Title: "ID", Width: 5}, {Title: "Name", Width: 16}, {Title: "Status", Width: 0}},
		[][]string{
			{"1", "Alice", "Active"},
			{"2", "Bob", "Inactive"},
			{"3", "Carol", "Active"},
			{"4", "Dave", "Pending"},
		},
	)

	// Tabs
	tabs := model.NewTabs([]string{"Overview", "Details", "Settings", "Logs"})

	// Tree
	tree := model.NewTree(&model.TreeNode{
		Label:    "project",
		Expanded: true,
		Children: []*model.TreeNode{
			{Label: "src", Expanded: true, Children: []*model.TreeNode{
				{Label: "main.go"},
				{Label: "util.go"},
			}},
			{Label: "docs", Children: []*model.TreeNode{
				{Label: "README.md"},
			}},
			{Label: "tests"},
		},
	})

	// FilePicker — start in current directory
	startDir, _ := os.Getwd()
	fp := model.NewFilePicker(startDir)

	widgets := []focusable{form, table, tabs, tree, fp}
	titles := []string{"Form", "Table", "Tabs", "Tree", "FilePicker"}

	// Focus first widget
	widgets[0].Focus()

	return widgetModel{widgets: widgets, titles: titles}
}

func (m widgetModel) Init() tea.Cmd {
	return nil
}

func (m widgetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m = m.shiftFocus(1)
			return m, nil
		case "shift+tab":
			m = m.shiftFocus(-1)
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		panelH := max(3, (m.height-4)/len(m.widgets))
		for _, w := range m.widgets {
			w.SetSize(m.width-4, panelH-2)
		}
		return m, nil
	}

	// Delegate unhandled keys to the focused widget
	cmd := m.widgets[m.focus].Update(msg)
	return m, cmd
}

func (m widgetModel) shiftFocus(delta int) widgetModel {
	m.widgets[m.focus].Blur()
	m.focus = (m.focus + delta + len(m.widgets)) % len(m.widgets)
	m.widgets[m.focus].Focus()
	return m
}

func (m widgetModel) View() tea.View {
	if m.width == 0 {
		return tea.NewView("")
	}

	ss := style.CurrentStyleSet()
	header := ss.Title.Render(" foxful-cli Widget Showcase ")
	hint := ss.Muted.Render(" Tab/Shift+Tab: cycle focus • q: quit")

	var panels []string
	for i, w := range m.widgets {
		var titleSt lipgloss.Style
		if i == m.focus {
			titleSt = ss.SelectedItem
		} else {
			titleSt = ss.Subtitle
		}
		title := titleSt.Render(fmt.Sprintf(" %s ", m.titles[i]))

		content := w.View()
		if content == "" {
			content = ss.Muted.Render("(empty)")
		}

		inner := lipgloss.JoinVertical(lipgloss.Left, title, content)

		borderFg := ss.Border.GetForeground()
		if i == m.focus {
			borderFg = ss.SelectedItem.GetForeground()
		}
		panel := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderFg).
			Padding(0, 1).
			Width(m.width - 4).
			Render(inner)

		panels = append(panels, panel)
	}

	full := lipgloss.JoinVertical(lipgloss.Left,
		header,
		hint,
		"",
		lipgloss.JoinVertical(lipgloss.Left, panels...),
	)
	return tea.NewView(full)
}

func main() {
	p := tea.NewProgram(newWidgetModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
