package model

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/ansix"
	"github.com/anhoder/foxful-cli/style"
)

// ListStyle is how Main lays out the menu list.
type ListStyle uint8

const (
	// ListClassic is the original list: one or two columns of "N. title".
	ListClassic ListStyle = iota
	// ListTracklist is a single left-aligned table: index, title, a
	// subtitle column and right-aligned Meta, under a section header.
	ListTracklist
)

// MenuHeader is the section header above a tracklist:
// "07 ── Title  Note ……… Meta".
type MenuHeader struct {
	Kicker string // section number; the parent's selected index by default
	Title  string
	Note   string // muted text after the title
	Meta   string // right-aligned; the number of items by default
}

const (
	trackMarginLeft  = 2
	trackMarginRight = 2
	trackGap         = 2
)

func (m *Main) useTracklist() bool {
	return m.options.ListStyle == ListTracklist && !m.options.CenterEverything
}

// trackColumns are the column widths of one page of the tracklist.
type trackColumns struct {
	start   int // first column of the row, where the left cap goes
	index   int // column of the index
	idxW    int
	titleW  int
	artistW int
	metaW   int
	end     int // column after the right cap
}

func (m *Main) trackColumns(items []MenuItem) trackColumns {
	c := trackColumns{start: m.menuStartColumn - 4, index: m.menuStartColumn, idxW: m.getMaxIndexWidth()}
	c.end = m.ContentWidth() - trackMarginRight
	maxSub := 0
	for _, it := range items {
		c.metaW = max(c.metaW, ansix.Width(it.Meta))
		maxSub = max(maxSub, ansix.Width(ansix.Normalize(it.Subtitle)))
	}
	rest := (c.end - 1) - (c.index + c.idxW + trackGap)
	if c.metaW > 0 {
		rest -= c.metaW + trackGap
	}
	if maxSub > 0 {
		c.artistW = min(maxSub, min(max(rest*2/5, 8), 32))
		c.titleW = rest - trackGap - c.artistW
		if c.titleW < 10 {
			c.artistW -= 10 - c.titleW
			c.titleW = 10
		}
		if c.artistW < 6 {
			c.artistW = 0
		}
	}
	if c.artistW == 0 {
		c.titleW = rest
	}
	c.titleW = max(c.titleW, 1)
	return c
}

type trackRowKey struct {
	item     MenuItem
	index    int
	selected bool
	hovered  bool
	cols     trackColumns
	width    int
	styleGen uint64
}

// trackPad fits plain text into w cells, cut with an ellipsis or padded.
func trackPad(s string, w int, right bool) string {
	if w <= 0 {
		return ""
	}
	s = ansix.Truncate(ansix.Normalize(s), w, "…")
	pad := strings.Repeat(" ", max(w-ansix.Width(s), 0))
	if right {
		return pad + s
	}
	return s + pad
}

// trackRow renders one row of the tracklist, exactly ContentWidth cells.
func (m *Main) trackRow(index int, cols trackColumns) string {
	item := m.menuList[index]
	selected, hovered := m.isSelected(index), index == m.hoveredMenuItemIdx
	key := trackRowKey{item, index, selected, hovered, cols, m.ContentWidth(), style.StyleGeneration()}
	if e, ok := m.trackCache[index]; ok && e.key == key {
		return e.row
	}

	ss := style.CurrentStyleSet()
	appBg := ss.AppBackground.GetBackground()
	bg := appBg
	if selected {
		bg = ss.SelectedItem.GetBackground()
	}
	paint := func(s string, fg color.Color, under bool) string {
		st := lipgloss.NewStyle().Underline(under)
		if !isNoColor(fg) {
			st = st.Foreground(fg)
		}
		if !isNoColor(bg) {
			st = st.Background(bg)
		}
		return st.Render(s)
	}
	textFg := ss.MenuItem.GetForeground()
	if selected {
		textFg = ss.SelectedItem.GetForeground()
	} else if hovered {
		textFg = ss.MenuItemHover.GetForeground()
	}
	indexFg := ss.MenuIndex.GetForeground()
	if isNoColor(indexFg) {
		indexFg = ss.Subtitle.GetForeground()
	}
	subFg := ss.Subtitle.GetForeground()

	var b strings.Builder
	b.WriteString(paintBg(strings.Repeat(" ", cols.start), appBg))
	if selected {
		b.WriteString(capCell(ss.MenuSelectedSepLeft, ss.SelectedItem.GetBackground(), appBg))
		arrow := ss.MenuArrow.GetForeground()
		if isNoColor(arrow) {
			arrow = textFg
		}
		marker := []rune(ss.MenuSelectedArrow)
		b.WriteString(paint(string(marker[0]), arrow, false))
		b.WriteString(paint(strings.Repeat(" ", cols.index-cols.start-2), nil, false))
	} else {
		b.WriteString(paint(strings.Repeat(" ", cols.index-cols.start), nil, false))
	}
	b.WriteString(paint(trackPad(strconv.Itoa(index), cols.idxW, true), indexFg, false))
	b.WriteString(paint(strings.Repeat(" ", trackGap), nil, false))
	title := trackPad(item.Title, cols.titleW, false)
	if hovered && !selected {
		t := strings.TrimRight(title, " ")
		b.WriteString(paint(t, textFg, true))
		b.WriteString(paint(strings.Repeat(" ", ansix.Width(title)-ansix.Width(t)), nil, false))
	} else {
		b.WriteString(paint(title, textFg, false))
	}
	if cols.artistW > 0 {
		b.WriteString(paint(strings.Repeat(" ", trackGap), nil, false))
		b.WriteString(paint(trackPad(item.Subtitle, cols.artistW, false), subFg, false))
	}
	if cols.metaW > 0 {
		b.WriteString(paint(strings.Repeat(" ", trackGap), nil, false))
		b.WriteString(paint(trackPad(item.Meta, cols.metaW, true), indexFg, false))
	}
	if selected {
		b.WriteString(capCell(ss.MenuSelectedSepRight, ss.SelectedItem.GetBackground(), appBg))
	} else {
		b.WriteString(paintBg(" ", appBg))
	}
	row := ansix.Fit(b.String(), m.ContentWidth(), backgroundSGR(appBg))

	if m.trackCache == nil {
		m.trackCache = map[int]trackRowEntry{}
	}
	m.trackCache[index] = trackRowEntry{key: key, row: row}
	return row
}

type trackRowEntry struct {
	key trackRowKey
	row string
}

func paintBg(s string, bg color.Color) string {
	if isNoColor(bg) {
		return s
	}
	return lipgloss.NewStyle().Background(bg).Render(s)
}

// capCell draws a rounded or square end of the selection band; without a
// configured glyph it is a plain band cell.
func capCell(glyph string, band, appBg color.Color) string {
	if glyph == "" || ansix.Width(glyph) != 1 {
		return paintBg(" ", band)
	}
	st := lipgloss.NewStyle()
	if !isNoColor(band) {
		st = st.Foreground(band)
	}
	if !isNoColor(appBg) {
		st = st.Background(appBg)
	}
	return st.Render(glyph)
}

// tracklistView renders the rows of the current page.
func (m *Main) tracklistView(lines int) []string {
	menus := m.getCurPageMenus()
	cols := m.trackColumns(menus)
	start := m.getPageStartIndex()
	rows := make([]string, 0, lines)
	for i := range min(lines, len(menus)) {
		rows = append(rows, m.trackRow(start+i, cols))
	}
	return rows
}

// tracklistBounds is the clickable span of a tracklist row: all of it.
func (m *Main) tracklistBounds(index int) (start, end int, ok bool) {
	cols := m.trackColumns(m.getCurPageMenus())
	return cols.start, cols.end, true
}

// menuHeader returns the header of the current menu.
func (m *Main) menuHeader(a *App, mt *MenuItem) MenuHeader {
	h := MenuHeader{Title: mt.Title, Note: mt.Subtitle, Meta: strconv.Itoa(len(m.menuList))}
	if m.menuStack.Len() > 0 {
		if top, ok := m.menuStack.Peek().(*menuStackItem); ok {
			h.Kicker = fmt.Sprintf("%02d", top.selectedIndex)
		}
	}
	if f := m.options.MenuHeader; f != nil {
		h = f(a, m.menu, h)
	}
	if m.loadingTips != "" {
		h.Meta = m.loadingTips
	}
	return h
}

// tracklistHeader renders the section header row:
// "← 07 ── 飙升榜  [更新27首] ……… 100 首".
func (m *Main) tracklistHeader(a *App, mt *MenuItem) string {
	ss := style.CurrentStyleSet()
	appBg := ss.AppBackground.GetBackground()
	h := m.menuHeader(a, mt)
	cols := m.trackColumns(m.getCurPageMenus())
	muted := ss.Muted.GetForeground()
	paint := func(s string, st lipgloss.Style) string {
		if !isNoColor(appBg) {
			st = st.Background(appBg)
		}
		return st.Render(s)
	}
	fg := func(c color.Color) lipgloss.Style {
		st := lipgloss.NewStyle()
		if !isNoColor(c) {
			st = st.Foreground(c)
		}
		return st
	}
	indexFg := ss.MenuIndex.GetForeground()
	if isNoColor(indexFg) {
		indexFg = ss.Subtitle.GetForeground()
	}

	var b strings.Builder
	b.WriteString(paintBg(strings.Repeat(" ", max(m.menuStartColumn-backButtonWidth, 0)), appBg))
	if m.menuStack.Len() > 0 {
		b.WriteString(m.backButtonIcon())
		b.WriteString(paintBg(" ", appBg))
	} else {
		b.WriteString(paintBg(strings.Repeat(" ", backButtonWidth), appBg))
	}
	avail := cols.end - m.menuStartColumn
	meta := ansix.Normalize(h.Meta)
	note := ansix.Normalize(h.Note)
	kicker := ""
	if h.Kicker != "" {
		kicker = h.Kicker + " ── "
	}
	title := ansix.Normalize(h.Title)
	need := func() int {
		n := ansix.Width(kicker) + ansix.Width(title)
		if note != "" {
			n += 2 + ansix.Width(note)
		}
		if meta != "" {
			n += 3 + ansix.Width(meta)
		}
		return n
	}
	if need() > avail {
		note = ""
	}
	if need() > avail {
		meta = ""
	}
	if need() > avail {
		title = ansix.Truncate(title, max(avail-ansix.Width(kicker), 1), "…")
	}
	if kicker != "" {
		b.WriteString(paint(h.Kicker, fg(indexFg)))
		b.WriteString(paint(" ── ", fg(muted)))
	}
	b.WriteString(paint(title, ss.MenuTitle.Bold(true)))
	used := ansix.Width(kicker) + ansix.Width(title)
	if note != "" {
		b.WriteString(paint("  "+note, fg(ss.Subtitle.GetForeground())))
		used += 2 + ansix.Width(note)
	}
	if meta != "" {
		b.WriteString(paint(strings.Repeat(" ", avail-used-ansix.Width(meta)), lipgloss.NewStyle()))
		b.WriteString(paint(meta, fg(indexFg)))
	}
	return ansix.Fit(b.String(), m.ContentWidth(), backgroundSGR(appBg))
}

// tracklistRule is the hairline under the section header.
func (m *Main) tracklistRule() string {
	ss := style.CurrentStyleSet()
	appBg := ss.AppBackground.GetBackground()
	cols := m.trackColumns(m.getCurPageMenus())
	st := lipgloss.NewStyle()
	if c := ss.Muted.GetForeground(); !isNoColor(c) {
		st = st.Foreground(c)
	}
	if !isNoColor(appBg) {
		st = st.Background(appBg)
	}
	line := paintBg(strings.Repeat(" ", cols.start+1), appBg) + st.Render(strings.Repeat("─", max(cols.end-cols.start-1, 0)))
	return ansix.Fit(line, m.ContentWidth(), backgroundSGR(appBg))
}
