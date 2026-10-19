package model

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func newKeyMsg(s string) tea.KeyMsg {
	switch s {
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "pgup":
		return tea.KeyPressMsg{Code: tea.KeyPgUp}
	case "pgdown":
		return tea.KeyPressMsg{Code: tea.KeyPgDown}
	case "home":
		return tea.KeyPressMsg{Code: tea.KeyHome}
	case "end":
		return tea.KeyPressMsg{Code: tea.KeyEnd}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	default:
		r := []rune(s)
		return tea.KeyPressMsg{Code: r[0], Text: s}
	}
}

func sampleColumns() []Column {
	return []Column{
		{Title: "Name", Width: 10},
		{Title: "Age", Width: 5},
	}
}

func sampleRows(n int) [][]string {
	rows := make([][]string, n)
	for i := 0; i < n; i++ {
		rows[i] = []string{"row", "val"}
	}
	return rows
}

func TestTableNewTable(t *testing.T) {
	tests := []struct {
		name     string
		columns  []Column
		rows     [][]string
		wantCols int
		wantRows int
	}{
		{"empty", nil, nil, 0, 0},
		{"columns only", sampleColumns(), nil, 2, 0},
		{"full", sampleColumns(), sampleRows(3), 2, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := NewTable(tt.columns, tt.rows)
			if tbl == nil {
				t.Fatal("NewTable returned nil")
			}
			if len(tbl.columns) != tt.wantCols {
				t.Errorf("columns = %d, want %d", len(tbl.columns), tt.wantCols)
			}
			if len(tbl.rows) != tt.wantRows {
				t.Errorf("rows = %d, want %d", len(tbl.rows), tt.wantRows)
			}
			if tbl.Focused() {
				t.Error("new table should not be focused")
			}
		})
	}
}

func TestTableFocusBlur(t *testing.T) {
	tbl := NewTable(sampleColumns(), sampleRows(3))

	if tbl.Focused() {
		t.Error("new table should not be focused")
	}

	tbl.Focus()
	if !tbl.Focused() {
		t.Error("table should be focused after Focus()")
	}

	tbl.Blur()
	if tbl.Focused() {
		t.Error("table should not be focused after Blur()")
	}
}

func TestTableSelectedRow(t *testing.T) {
	tests := []struct {
		name string
		rows [][]string
		want int
	}{
		{"empty rows", nil, -1},
		{"with rows", sampleRows(3), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := NewTable(sampleColumns(), tt.rows)
			if got := tbl.SelectedRow(); got != tt.want {
				t.Errorf("SelectedRow() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTableNavigationClamping(t *testing.T) {
	tbl := NewTable(sampleColumns(), sampleRows(3))
	tbl.SetSize(40, 10)
	tbl.Focus()

	// Cannot move above first row.
	tbl.Update(newKeyMsg("up"))
	if got := tbl.SelectedRow(); got != 0 {
		t.Errorf("after up at top, SelectedRow() = %d, want 0", got)
	}

	// Move down through the rows.
	tbl.Update(newKeyMsg("down"))
	if got := tbl.SelectedRow(); got != 1 {
		t.Errorf("after down, SelectedRow() = %d, want 1", got)
	}
	tbl.Update(newKeyMsg("down"))
	if got := tbl.SelectedRow(); got != 2 {
		t.Errorf("after down, SelectedRow() = %d, want 2", got)
	}

	// Cannot move below last row.
	tbl.Update(newKeyMsg("down"))
	if got := tbl.SelectedRow(); got != 2 {
		t.Errorf("after down at bottom, SelectedRow() = %d, want 2", got)
	}
}

func TestTableHomeEnd(t *testing.T) {
	tbl := NewTable(sampleColumns(), sampleRows(5))
	tbl.SetSize(40, 10)
	tbl.Focus()

	// Jump to end.
	tbl.Update(newKeyMsg("end"))
	if got := tbl.SelectedRow(); got != 4 {
		t.Errorf("after end, SelectedRow() = %d, want 4", got)
	}

	// Jump to home.
	tbl.Update(newKeyMsg("home"))
	if got := tbl.SelectedRow(); got != 0 {
		t.Errorf("after home, SelectedRow() = %d, want 0", got)
	}

	// Vim-style keys.
	tbl.Update(newKeyMsg("G"))
	if got := tbl.SelectedRow(); got != 4 {
		t.Errorf("after G, SelectedRow() = %d, want 4", got)
	}
	tbl.Update(newKeyMsg("g"))
	if got := tbl.SelectedRow(); got != 0 {
		t.Errorf("after g, SelectedRow() = %d, want 0", got)
	}
}

func TestTableUpdateWhenBlurred(t *testing.T) {
	tbl := NewTable(sampleColumns(), sampleRows(3))
	tbl.SetSize(40, 10)
	// Not focused: navigation should be ignored.
	tbl.Update(newKeyMsg("down"))
	if got := tbl.SelectedRow(); got != 0 {
		t.Errorf("blurred table moved selection to %d, want 0", got)
	}
}

func TestTableSetRows(t *testing.T) {
	tbl := NewTable(sampleColumns(), sampleRows(5))
	tbl.SetSize(40, 10)
	tbl.Focus()

	tbl.Update(newKeyMsg("end"))
	if got := tbl.SelectedRow(); got != 4 {
		t.Fatalf("setup failed, SelectedRow() = %d, want 4", got)
	}

	// SetRows resets selection.
	tbl.SetRows(sampleRows(2))
	if got := tbl.SelectedRow(); got != 0 {
		t.Errorf("after SetRows, SelectedRow() = %d, want 0", got)
	}
	if got := len(tbl.Rows()); got != 2 {
		t.Errorf("after SetRows, Rows() len = %d, want 2", got)
	}
}

func TestTableRowsAccessor(t *testing.T) {
	rows := sampleRows(3)
	tbl := NewTable(sampleColumns(), rows)
	got := tbl.Rows()
	if len(got) != len(rows) {
		t.Errorf("Rows() len = %d, want %d", len(got), len(rows))
	}
}

func TestTableViewNonEmpty(t *testing.T) {
	tbl := NewTable(sampleColumns(), sampleRows(3))
	tbl.SetSize(40, 10)
	view := tbl.View()
	if view == "" {
		t.Error("View() returned empty string for populated table")
	}
	if !strings.Contains(view, "Name") {
		t.Error("View() should contain column title 'Name'")
	}
}

func TestTableEmptyState(t *testing.T) {
	tests := []struct {
		name    string
		columns []Column
		rows    [][]string
		want    string
	}{
		{"no columns", nil, nil, "No columns"},
		{"no rows", sampleColumns(), nil, "No data"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tbl := NewTable(tt.columns, tt.rows)
			tbl.SetSize(40, 10)
			view := tbl.View()
			if !strings.Contains(view, tt.want) {
				t.Errorf("View() = %q, want to contain %q", view, tt.want)
			}
		})
	}
}

func TestTableColumnTruncation(t *testing.T) {
	columns := []Column{
		{Title: "Name", Width: 5},
	}
	rows := [][]string{
		{"a-very-long-value-that-overflows"},
	}
	tbl := NewTable(columns, rows)
	tbl.SetSize(40, 10)
	tbl.Focus()
	view := tbl.View()

	if view == "" {
		t.Fatal("View() returned empty string")
	}
	// The overflowing value should be truncated with an ellipsis.
	if !strings.Contains(view, "…") {
		t.Error("View() should truncate long cell with ellipsis")
	}
	// The full untruncated value should not appear.
	if strings.Contains(view, "a-very-long-value-that-overflows") {
		t.Error("View() should not contain the full untruncated value")
	}
}

func TestTableColumnPadding(t *testing.T) {
	columns := []Column{
		{Title: "Name", Width: 10},
		{Title: "Age", Width: 5},
	}
	rows := [][]string{
		{"Bob", "30"},
	}
	tbl := NewTable(columns, rows)
	tbl.SetSize(40, 10)

	widths := tbl.computeColumnWidths()
	if len(widths) != 2 {
		t.Fatalf("computeColumnWidths() len = %d, want 2", len(widths))
	}
	if widths[0] != 10 {
		t.Errorf("column 0 width = %d, want 10", widths[0])
	}
	if widths[1] != 5 {
		t.Errorf("column 1 width = %d, want 5", widths[1])
	}
}

func TestTableScrolling(t *testing.T) {
	tbl := NewTable(sampleColumns(), sampleRows(50))
	tbl.SetSize(40, 8) // 6 visible rows
	tbl.Focus()

	// Navigate to end; scroll offset must advance to keep selection visible.
	tbl.Update(newKeyMsg("end"))
	if tbl.scrollOffset == 0 {
		t.Error("scrollOffset should advance when navigating to end of a long list")
	}
	if got := tbl.SelectedRow(); got != 49 {
		t.Errorf("SelectedRow() = %d, want 49", got)
	}

	// Selected row must remain within the visible window.
	visible := tbl.visibleRows()
	if tbl.selectedRow < tbl.scrollOffset || tbl.selectedRow >= tbl.scrollOffset+visible {
		t.Errorf("selected row %d not visible in window [%d, %d)", tbl.selectedRow, tbl.scrollOffset, tbl.scrollOffset+visible)
	}
}

func TestTableHorizontalOverflow(t *testing.T) {
	columns := []Column{
		{Title: "A", Width: 20},
		{Title: "B", Width: 20},
		{Title: "C", Width: 20},
	}
	rows := [][]string{
		{"aaa", "bbb", "ccc"},
	}
	tbl := NewTable(columns, rows)
	tbl.SetSize(15, 10) // narrower than sum of column widths
	tbl.Focus()

	// Should not panic and should produce output.
	view := tbl.View()
	if view == "" {
		t.Error("View() returned empty string under horizontal overflow")
	}
}

func TestTablePageNavigation(t *testing.T) {
	tbl := NewTable(sampleColumns(), sampleRows(50))
	tbl.SetSize(40, 12) // 10 visible rows
	tbl.Focus()

	tbl.Update(newKeyMsg("pgdown"))
	if got := tbl.SelectedRow(); got == 0 {
		t.Error("pgdown should move the selection down")
	}

	after := tbl.SelectedRow()
	tbl.Update(newKeyMsg("pgup"))
	if got := tbl.SelectedRow(); got >= after {
		t.Errorf("pgup should move selection up from %d, got %d", after, got)
	}
}
