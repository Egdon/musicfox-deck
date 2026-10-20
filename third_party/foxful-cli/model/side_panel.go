package model

import (
	"fmt"
	"image/color"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/layout"
	"github.com/anhoder/foxful-cli/style"
	"github.com/charmbracelet/x/ansi"
)

// SidePanelLayout is how a side panel wants to be laid out in a window.
type SidePanelLayout struct {
	// Width is the number of columns the panel takes at the right edge,
	// next to the menu.
	Width int
	// Height, when Width is zero, is the number of rows the panel takes
	// under the menu, across the full width: a dock for narrow windows.
	// With neither, the panel is hidden and Main lays out as usual.
	Height int
	// FooterRows is the number of rows at the bottom where the components
	// render across the full width, below the menu and the panel.
	FooterRows int
}

func (l SidePanelLayout) shown() bool  { return l.Width > 0 || l.Height > 0 }
func (l SidePanelLayout) docked() bool { return l.Width <= 0 && l.Height > 0 }

// SidePanel draws a column at the right of Main, next to the menu, or a dock
// under it. While it shows, the menu starts near the top and the components
// form a footer.
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
	docked     bool // under the menu rather than beside it
}

func (m *Main) sidePanelLayout(w, h int) SidePanelLayout {
	// Tabs and the centered layout keep the full width.
	if m.options.SidePanel == nil || m.app == nil || m.options.EnableTabs || m.options.CenterEverything {
		return SidePanelLayout{}
	}
	l := m.options.SidePanel.Layout(m.app, w, h)
	l.FooterRows = max(l.FooterRows, 0)
	switch {
	case l.Width > 0 && l.Width < w:
		l.Height = 0
		return l
	case l.docked() && l.Height+l.FooterRows < h:
		l.Width = 0
		return l
	}
	return SidePanelLayout{}
}

// PanelVisible reports whether the side panel shows, beside the menu or
// under it.
func (m *Main) PanelVisible() bool {
	return m.sidePanelLayout(m.app.WindowWidth(), m.app.WindowHeight()).shown()
}

// PanelRect returns the cells the side panel occupies.
func (m *Main) PanelRect() (x, y, w, h int, ok bool) {
	g, ok := m.panelGeometry(m.app.WindowWidth(), m.app.WindowHeight())
	return g.x, g.y, g.w, g.h, ok
}

func (m *Main) panelGeometry(w, h int) (panelGeometry, bool) {
	l := m.sidePanelLayout(w, h)
	if !l.shown() {
		return panelGeometry{}, false
	}
	top := m.topBarRows() + m.leftLeadRows()
	bottom := l.FooterRows
	if m.statusBar != nil && m.options.StatusBarPosition == StatusBarBottom {
		bottom++
	}
	if l.docked() {
		g := panelGeometry{x: 0, y: h - bottom - l.Height, w: w, h: l.Height, footerRows: l.FooterRows, docked: true}
		return g, g.y > top
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
func (m *Main) joinPanel(left []string, panel string, g panelGeometry, bgSGR string) []string {
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
		out[i] = m.fitCached(line, g.x, bgSGR) + cursorColumn(g.x) + m.fitCached(p, g.w, bgSGR)
	}
	m.fitPrev, m.fitCur = m.fitCur, map[fitKey]string{}
	return out
}

type fitKey struct {
	line, pad string
	w         int
}

// fitCached fits a line, reusing the result from this or the last frame:
// most lines do not change between frames.
func (m *Main) fitCached(line string, w int, pad string) string {
	k := fitKey{line, pad, w}
	if s, ok := m.fitCur[k]; ok {
		return s
	}
	s, ok := m.fitPrev[k]
	if !ok {
		s = ansix.Fit(line, w, pad)
	}
	if m.fitCur == nil {
		m.fitCur = map[fitKey]string{}
	}
	m.fitCur[k] = s
	return s
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
	left := m.stackMenu(m.animateMenu(a, m.menuSections(a)))
	lines = append(lines, strings.Split(left, "\n")...)
	if g.docked {
		lines = fitRows(lines, g.y)
		dock := fitRows(strings.Split(m.options.SidePanel.View(a, m, g.w, g.h), "\n"), g.h)
		for _, line := range dock {
			lines = append(lines, m.fitCached(line, g.w, bgSGR))
		}
		m.fitPrev, m.fitCur = m.fitCur, map[fitKey]string{}
	} else {
		lines = fitRows(lines, g.y+g.h)
		lines = m.joinPanel(lines, m.options.SidePanel.View(a, m, g.w, g.h), g, bgSGR)
	}

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
	view := renderAppBackgroundWith(strings.Join(lines, "\n"), w, ss.AppBackground, a.appBackgroundExclusion, m.widths.width)
	m.widths.next()
	return view
}

// stackMenu stacks the menu's sections into one block, reusing last frame's
// block when they did not change, as they mostly do not.
func (m *Main) stackMenu(sections []string) string {
	if m.menuOut == "" || !slices.Equal(sections, m.menuIn) {
		m.menuIn, m.menuOut = slices.Clone(sections), layout.JoinVertical(lipgloss.Left, sections...)
	}
	return m.menuOut
}

// widthMemo remembers the widths of rows measured in this and the last
// frame: most rows do not change between frames, and measuring styled text
// by graphemes is costly.
type widthMemo struct{ cur, prev map[string]int }

// width is ansi.StringWidth(s).
func (w *widthMemo) width(s string) int {
	if n, ok := w.cur[s]; ok {
		return n
	}
	n, ok := w.prev[s]
	if !ok {
		n = ansi.StringWidth(s)
	}
	if w.cur == nil {
		w.cur = map[string]int{}
	}
	w.cur[s] = n
	return n
}

// next starts a frame: rows not measured in the last two are forgotten.
func (w *widthMemo) next() { w.prev, w.cur = w.cur, nil }

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
