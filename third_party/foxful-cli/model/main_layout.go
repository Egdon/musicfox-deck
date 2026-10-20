package model

// ContentWidth returns the columns the menu and the components lay out in:
// the window width less the side panel.
func (m *Main) ContentWidth() int {
	return m.contentWidthFor(m.app.WindowWidth(), m.app.WindowHeight())
}

// contentWidthFor returns the content width of a w×h window.
func (m *Main) contentWidthFor(w, h int) int {
	return max(w-m.panelWidthFor(w, h), 0)
}

// panelWidthFor returns the side panel's width in a w×h window; 0 means no
// panel.
func (m *Main) panelWidthFor(w, h int) int {
	return m.sidePanelLayout(w, h).Width
}
