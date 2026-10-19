// Layout example — demonstrates using the layout package inside a real
// bubbletea v2 TUI application. Shows how layout primitives (Join, Place,
// Overlay, Compositor) are called inside View() every frame with live data.
//
// Keys:
//
//	1-6  — switch between layout demos
//	q    — quit
//	↑↓   — move selection (demo 1)
//	←→   — change layer count (demo 6)
package main

import (
	"fmt"
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/anhoder/foxful-cli/layout"
	"github.com/anhoder/foxful-cli/style"
)

// ── color palette ──

var (
	cAccent     color.Color = lipgloss.BrightCyan
	cPrimary    color.Color = lipgloss.BrightGreen
	cDim        color.Color = lipgloss.BrightBlack
	cBlue       color.Color = lipgloss.BrightBlue
	cYellow     color.Color = lipgloss.BrightYellow
	cRed        color.Color = lipgloss.BrightRed
	cMagenta    color.Color = lipgloss.BrightMagenta
	cBg         color.Color = lipgloss.Color("#222222")
	cBgSelected color.Color = lipgloss.Color("#444444")
	cBgHeader   color.Color = lipgloss.Color("#333333")
	cBgTab      color.Color = lipgloss.Color("#555555")
)

// ── styles ──

var (
	headerStyle = lipgloss.NewStyle().
			Background(cBgHeader).
			Padding(0, 1)

	tabActiveStyle = lipgloss.NewStyle().
			Padding(0, 2).
			MarginRight(1).
			Foreground(cPrimary).
			Background(cBgTab).
			Bold(true)

	tabInactiveStyle = lipgloss.NewStyle().
			Padding(0, 2).
			MarginRight(1).
			Foreground(cDim)

	footerStyle = lipgloss.NewStyle().
			Foreground(cDim).
			Padding(0, 1)

	infoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBlue).
			Padding(1, 2)

	selectedBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cYellow).
			Padding(1, 2)

	dimBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cDim).
			Padding(1, 2)
)

type model struct {
	width  int
	height int

	activeDemo int // 1-6
	selection  int // for demo 1 (list)
	layerCount int // for demo 4 (compositor)
}

// ── Init ──

func (m model) Init() tea.Cmd {
	return nil
}

// ── Update ──

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "Q", "ctrl+c":
			return m, tea.Quit

		case "1", "2", "3", "4", "5", "6":
			m.activeDemo = int(msg.String()[0] - '0')

		case "up", "k", "K":
			if m.selection > 0 {
				m.selection--
			}
		case "down", "j", "J":
			if m.selection < 3 {
				m.selection++
			}

		case "left", "h", "H":
			if m.layerCount > 0 {
				m.layerCount--
			}
		case "right", "l", "L":
			if m.layerCount < 3 {
				m.layerCount++
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// ── View ──

func (m model) View() tea.View {
	var v tea.View
	v.AltScreen = true

	if m.width <= 0 {
		return v
	}

	// Header: tabs built with JoinHorizontal
	header := m.renderHeader()
	headerLine := lipgloss.NewStyle().Width(m.width).Render(header)

	// Body: switch between layout demos
	var body string
	switch m.activeDemo {
	case 1:
		body = m.demoJoin()
	case 2:
		body = m.demoPlace()
	case 3:
		body = m.demoOverlay()
	case 4:
		body = m.demoCompositor()
	case 5:
		body = m.demoPlaceCustom()
	case 6:
		body = m.demoMeasure()
	}

	// Footer: controls
	footer := m.renderFooter()

	v.SetContent(layout.JoinVertical(layout.Left, headerLine, body, footer))
	return v
}

// ── header: six tabs via JoinHorizontal ──

func (m model) renderHeader() string {
	labels := []string{"1.Join", "2.Place", "3.Overlay", "4.Compositor", "5.Place+", "6.Width"}
	var tabs []string
	for i, label := range labels {
		if i+1 == m.activeDemo {
			tabs = append(tabs, tabActiveStyle.Render(label))
		} else {
			tabs = append(tabs, tabInactiveStyle.Render(label))
		}
	}
	return layout.JoinHorizontal(layout.Top,
		headerStyle.Render(" Layout Demo "),
		layout.JoinHorizontal(layout.Top, tabs...),
		lipgloss.NewStyle().Foreground(cDim).Padding(0, 1).Render(
			fmt.Sprintf("%dx%d", m.width, m.height),
		),
	)
}

// ── footer ──

func (m model) renderFooter() string {
	hint := footerStyle.Render("q:quit  1-6:demo  ↑↓:select  ←→:layers")
	return lipgloss.NewStyle().Width(m.width).Render(hint)
}

// ── Demo 1: JoinHorizontal + JoinVertical (grid) ──

func (m model) demoJoin() string {
	desc := infoBoxStyle.Render(
		"JoinHorizontal + JoinVertical build a 2×2 grid.\n" +
			"The selected box highlights dynamically.\n" +
			"Use ↑↓ to move selection — View() re-evaluates each frame.",
	)

	type item struct {
		label string
		clr   color.Color
	}
	items := []item{
		{"Songs", cRed},
		{"Albums", lipgloss.BrightGreen},
		{"Artists", lipgloss.BrightBlue},
		{"Playlists", cMagenta},
	}

	var boxes []string
	for i, it := range items {
		s := dimBoxStyle.Copy().BorderForeground(it.clr)
		if i == m.selection {
			s = selectedBoxStyle.Copy().BorderForeground(it.clr).
				Background(cBgSelected)
		}
		boxes = append(boxes, s.Width(28).Height(5).Align(lipgloss.Center, lipgloss.Center).Render(
			fmt.Sprintf("%s\n(%d)", it.label, i+1),
		))
	}

	topRow := layout.JoinHorizontal(layout.Top, boxes[0], boxes[1])
	bottomRow := layout.JoinHorizontal(layout.Top, boxes[2], boxes[3])
	grid := layout.JoinVertical(layout.Left, topRow, bottomRow)

	return layout.JoinVertical(layout.Left,
		desc, "",
		lipgloss.NewStyle().Width(62).Align(lipgloss.Center).Render(grid),
	)
}

// ── Demo 2: Place + PlaceCenter ──

func (m model) demoPlace() string {
	desc := infoBoxStyle.Render(
		"Place centers a string in a bounding box (width×height).\n" +
			"PlaceCenter is shorthand for Place(w, h, Center, Center, content).\n" +
			"Use for: splash screens, centered dialogs, logo areas.",
	)

	logo := lipgloss.NewStyle().
		Foreground(cPrimary).
		Bold(true).
		Render("┌─────────────┐\n│ foxful-cli   │\n│ Layout Demo  │\n└─────────────┘")

	centered := layout.PlaceCenter(logo, 30, 8)
	bordered := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cBlue).
		Render(centered)

	return layout.JoinVertical(layout.Left, desc, "",
		lipgloss.NewStyle().Width(40).Align(lipgloss.Center).Render(bordered),
	)
}

// ── Demo 3: Overlay ──

func (m model) demoOverlay() string {
	desc := infoBoxStyle.Render(
		"Overlay layers one string on top of another at (x, y).\n" +
			"Background outside the overlay region is preserved.\n" +
			"Use for: popups, tooltips, picture-in-picture.",
	)

	bg := dimBoxStyle.Copy().Width(60).Height(10).Render(
		"Background content — this is the main app view",
	)

	overlay := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cPrimary).
		Background(cBg).
		Padding(1, 2).
		Align(lipgloss.Center).
		Render("▲ Popup ▲\nOverlay at (15, 3)")

	return layout.JoinVertical(layout.Left, desc, "",
		layout.Overlay(bg, overlay, 15, 3),
	)
}

// ── Demo 4: Compositor (multi-layer) ──

func (m model) demoCompositor() string {
	desc := infoBoxStyle.Copy().Width(70).Render(
		fmt.Sprintf(
			"Compositor stacks multiple Layer objects back-to-front.\n"+
				"Layers: %d (←→ to change). Each has its own (x, y) position.\n"+
				"Use for: sidebars, multi-panel layouts, complex overlays.",
			m.layerCount+1,
		),
	)

	layerColors := []color.Color{cBlue, cMagenta, cPrimary, cYellow}

	bg := dimBoxStyle.Copy().Width(66).Height(10).Render("Layer 0: Background")

	// Build all layers at once — v2 Compositor takes them in the constructor
	bgLayer := layout.NewLayer(bg)
	topLayers := make([]*layout.Layer, 0, m.layerCount)
	for i := 0; i < m.layerCount; i++ {
		layer := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(layerColors[i]).
			Background(cBg).
			Width(24).Height(5).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("Layer %d\n(%d,%d)", i+1, 4+i*10, 2+i*2))
		x := 4 + i*10
		y := 2 + i*2
		topLayers = append(topLayers, layout.NewLayer(layer).X(x).Y(y))
	}

	allLayers := append([]*layout.Layer{bgLayer}, topLayers...)
	return layout.JoinVertical(layout.Left, desc, "",
		layout.NewCompositor(allLayers...).Render(),
	)
}

// ── Demo 5: Place with WithWhitespaceChars ──

func (m model) demoPlaceCustom() string {
	desc := infoBoxStyle.Render(
		"Place with WithWhitespaceChars fills empty space with custom\n" +
			"characters instead of spaces. WithWhitespaceStyle colors them.\n" +
			"Use for: decorative borders, loading screens, patterned filler.",
	)

	filled := layout.Place(
		40, 6,
		layout.Center, layout.Center,
		lipgloss.NewStyle().Foreground(cPrimary).Bold(true).Render(" ★ foxful-cli ★ "),
		layout.WithWhitespaceChars("·"),
		layout.WithWhitespaceStyle(lipgloss.NewStyle().Foreground(cDim)),
	)

	bordered := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cBlue).
		Render(filled)

	return layout.JoinVertical(layout.Left, desc, "", bordered)
}

// ── Demo 6: Width measurement ──

func (m model) demoMeasure() string {
	desc := infoBoxStyle.Render(
		"Width measures visual string width, ignoring ANSI escapes.\n" +
			"CJK chars are 2 cells; emoji are 2 cells. Essential for\n" +
			"dynamic layout calculations in View().",
	)

	type row struct {
		label   string
		content string
		note    string
	}
	examples := []row{
		{"Plain ASCII", "Hello, World!", "13 chars = 13 cells"},
		{"Styled", style.FG("Hello, World!", cPrimary), "13 cells (escapes excluded)"},
		{"CJK", "你好，世界！", "6 chars = 12 cells"},
		{"Mixed", "Track 01 — 晴天", "14 cells (latin + CJK)"},
		{"Emoji", "🎵 ▶️ 🎶", "5 chars (emoji are 2-wide)"},
		{"Empty", "", "0 cells"},
	}

	var rows []string
	for _, ex := range examples {
		w := layout.Width(ex.content)
		rows = append(rows, layout.JoinHorizontal(layout.Top,
			lipgloss.NewStyle().Width(30).Render(fmt.Sprintf("  %s", ex.label)),
			lipgloss.NewStyle().Width(12).Render(fmt.Sprintf("  %d", w)),
			lipgloss.NewStyle().Width(30).Render(style.FG(ex.note, cDim)),
		))
	}

	table := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cBlue).
		Padding(1, 2).
		Render(layout.JoinVertical(layout.Left, rows...))

	return layout.JoinVertical(layout.Left, desc, "", table)
}

// ── main ──

func newModel() model {
	return model{activeDemo: 1, layerCount: 2}
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
