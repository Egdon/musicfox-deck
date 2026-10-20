package model

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/layout"
	"github.com/anhoder/foxful-cli/style"
)

// SidePanelLayout is how a side panel wants to be laid out in a window.
type SidePanelLayout struct {
	// Width is the number of columns the panel takes at the right edge;
	// zero or less hides it and Main lays out as usual.
	Width int
	// FooterRows is the number of rows at the bottom where the components
	// render across the full width, below the menu and the panel.
	FooterRows int
}

// SidePanel draws a column at the right of Main, next to the menu. While it
// shows, the menu starts near the top and the components form a footer.
type SidePanel interface {
	// Layout is asked every frame; it must be cheap.
	Layout(a *App, w, h int) SidePanelLayout
	// Update receives every message Main receives, like Component.Update.
	Update(msg tea.Msg, a *App)
	// View returns exactly h lines of exactly w cells.
	View(a *App, m *Main, w, h int) string
}

// SidePanelMouse is implemented by panels that handle the mouse. Events are
// given in panel coordinates.
type SidePanelMouse interface {
	HandleMouse(a *App, m *Main, msg tea.MouseMsg, x, y int) SidePanelMouseResult
	// MouseLeave tells the panel the pointer left it.
	MouseLeave(a *App, m *Main)
}

// SidePanelMouseResult tells Main what the panel did with a mouse event.
type SidePanelMouseResult struct {
	Handled bool // the event is consumed
	Capture bool // keep sending events while a button is held, even outside
	Page    Page // switch to this page
	Cmd     tea.Cmd
	Pointer string // pointer shape over the panel, "" for the default
}

// panelGeometry is where the side panel sits in the frame.
type panelGeometry struct {
	x, y, w, h int
	footerRows int
}

func (m *Main) sidePanelLayout(w, h int) SidePanelLayout {
	// Tabs and the centered layout keep the full width.
	if m.options.SidePanel == nil || m.app == nil || m.options.EnableTabs || m.options.CenterEverything {
		return SidePanelLayout{}
	}
	l := m.options.SidePanel.Layout(m.app, w, h)
	if l.Width <= 0 || l.Width >= w {
		return SidePanelLayout{}
	}
	l.FooterRows = max(l.FooterRows, 0)
	return l
}

// PanelVisible reports whether the side panel shows.
func (m *Main) PanelVisible() bool {
	return m.sidePanelLayout(m.app.WindowWidth(), m.app.WindowHeight()).Width > 0
}

// PanelRect returns the cells the side panel occupies.
func (m *Main) PanelRect() (x, y, w, h int, ok bool) {
	g, ok := m.panelGeometry(m.app.WindowWidth(), m.app.WindowHeight())
	return g.x, g.y, g.w, g.h, ok
}

func (m *Main) panelGeometry(w, h int) (panelGeometry, bool) {
	l := m.sidePanelLayout(w, h)
	if l.Width <= 0 {
		return panelGeometry{}, false
	}
	top := m.topBarRows() + m.leftLeadRows()
	bottom := l.FooterRows
	if m.statusBar != nil && m.options.StatusBarPosition == StatusBarBottom {
		bottom++
	}
	g := panelGeometry{x: w - l.Width, y: top, w: l.Width, h: h - top - bottom, footerRows: l.FooterRows}
	return g, g.h > 0
}

// topBarRows counts the title or top status bar row.
func (m *Main) topBarRows() int {
	if m.statusBar != nil && m.options.StatusBarPosition == StatusBarTop {
		return 1
	}
	if m.options.WhetherDisplayTitle {
		return 1
	}
	return 0
}

// leftLeadRows counts the blank rows between the top bar and the menu title.
func (m *Main) leftLeadRows() int {
	if t := m.computeTitleStartRow(); t > 1 {
		return t
	}
	return 0
}

// cursorColumn moves the cursor to a column, so the panel starts where it
// must even if the terminal measured the left part differently.
func cursorColumn(x int) string { return fmt.Sprintf("\x1b[%dG", x+1) }

// joinPanel places the panel's lines to the right of the left column, from
// the top of the panel.
func joinPanel(left []string, panel string, g panelGeometry, bgSGR string) []string {
	pl := strings.Split(panel, "\n")
	out := make([]string, len(left))
	for i, line := range left {
		row := i - g.y
		if row < 0 || row >= g.h {
			out[i] = line
			continue
		}
		p := ""
		if row < len(pl) {
			p = pl[row]
		}
		out[i] = ansix.Fit(line, g.x, bgSGR) + cursorColumn(g.x) + ansix.Fit(p, g.w, bgSGR)
	}
	return out
}

// panelView lays Main out next to the side panel: the top bar, the menu
// block on the left with the panel on its right, the components as a footer
// across the full width, and the bottom status bar.
func (m *Main) panelView(a *App, w, h int, g panelGeometry) string {
	ss := style.CurrentStyleSet()
	bgSGR := backgroundSGR(ss.AppBackground.GetBackground())

	var lines []string
	if m.statusBar != nil && m.options.StatusBarPosition == StatusBarTop {
		lines = append(lines, strings.Split(m.statusBar.View(a, m), "\n")...)
	} else if m.options.WhetherDisplayTitle {
		lines = append(lines, m.TitleView(a))
	}
	left := layout.JoinVertical(lipgloss.Left, m.menuSections(a)...)
	lines = append(lines, strings.Split(left, "\n")...)
	lines = fitRows(lines, g.y+g.h)
	lines = joinPanel(lines, m.options.SidePanel.View(a, m, g.w, g.h), g, bgSGR)

	var footer []string
	for _, c := range m.components {
		if c == nil {
			continue
		}
		if v, _ := c.View(a, m); v != "" {
			footer = append(footer, strings.Split(v, "\n")...)
		}
	}
	// The footer keeps to the bottom: pad above, drop from the top.
	if len(footer) > g.footerRows {
		footer = footer[len(footer)-g.footerRows:]
	}
	for len(footer) < g.footerRows {
		footer = append([]string{""}, footer...)
	}
	lines = append(lines, footer...)

	if m.statusBar != nil && m.options.StatusBarPosition == StatusBarBottom {
		lines = append(lines, strings.Split(m.statusBar.View(a, m), "\n")...)
	}
	lines = fitRows(lines, h)
	return renderAppBackground(strings.Join(lines, "\n"), w, ss.AppBackground, a.appBackgroundExclusion)
}

// fitRows pads lines with blank rows or cuts them to n.
func fitRows(lines []string, n int) []string {
	if len(lines) > n {
		return lines[:n]
	}
	for len(lines) < n {
		lines = append(lines, "")
	}
	return lines
}

// backgroundSGR returns the SGR sequence that paints c as background, or ""
// for no color.
func backgroundSGR(c color.Color) string {
	if isNoColor(c) {
		return ""
	}
	r, g, b, a := c.RGBA()
	if a == 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r>>8, g>>8, b>>8)
}

// routePanelMouse gives mouse events over the panel, or captured by it, to
// the panel. handled stops Main and its controllers from seeing the event.
func (m *Main) routePanelMouse(msg tea.MouseMsg, a *App) (handled bool, page Page, cmd tea.Cmd) {
	pm, ok := m.options.SidePanel.(SidePanelMouse)
	if !ok {
		return false, nil, nil
	}
	g, visible := m.panelGeometry(a.WindowWidth(), a.WindowHeight())
	mouse := msg.Mouse()
	inside := visible && mouse.X >= g.x && mouse.X < g.x+g.w && mouse.Y >= g.y && mouse.Y < g.y+g.h
	if !inside && !m.panelCapture {
		if m.panelHover {
			m.panelHover = false
			pm.MouseLeave(a, m)
			return false, nil, m.setPanelPointer("")
		}
		return false, nil, nil
	}
	m.panelHover = inside
	res := pm.HandleMouse(a, m, msg, mouse.X-g.x, mouse.Y-g.y)
	m.panelCapture = res.Capture
	return true, res.Page, tea.Batch(res.Cmd, m.setPanelPointer(res.Pointer))
}

// setPanelPointer changes the pointer shape when the panel asks for another.
func (m *Main) setPanelPointer(shape string) tea.Cmd {
	if shape == m.panelPointer {
		return nil
	}
	m.panelPointer = shape
	if shape == "" {
		shape = "default"
	}
	return setMousePointer(shape)
}
