package model

import (
	"strings"
	"testing"
)

func TestTabsNew(t *testing.T) {
	tests := []struct {
		name       string
		titles     []string
		wantActive int
	}{
		{"empty", []string{}, 0},
		{"single", []string{"Tab1"}, 0},
		{"multiple", []string{"Tab1", "Tab2", "Tab3"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tabs := NewTabs(tt.titles)
			if tabs == nil {
				t.Fatal("NewTabs returned nil")
			}
			if tabs.Active() != tt.wantActive {
				t.Errorf("Active() = %d, want %d", tabs.Active(), tt.wantActive)
			}
			if tabs.Focused() {
				t.Error("new tabs should not be focused")
			}
		})
	}
}

func TestTabsFocusBlur(t *testing.T) {
	tabs := NewTabs([]string{"Tab1", "Tab2"})

	if tabs.Focused() {
		t.Error("tabs should start unfocused")
	}

	tabs.Focus()
	if !tabs.Focused() {
		t.Error("tabs should be focused after Focus()")
	}

	tabs.Blur()
	if tabs.Focused() {
		t.Error("tabs should be unfocused after Blur()")
	}
}

func TestTabsNavigation(t *testing.T) {
	tabs := NewTabs([]string{"Tab1", "Tab2", "Tab3"})
	tabs.Focus()

	// Initial state
	if tabs.Active() != 0 {
		t.Errorf("initial active = %d, want 0", tabs.Active())
	}

	// Next
	tabs.Update(newKeyMsg("right"))
	if tabs.Active() != 1 {
		t.Errorf("after right, active = %d, want 1", tabs.Active())
	}

	tabs.Update(newKeyMsg("l"))
	if tabs.Active() != 2 {
		t.Errorf("after l, active = %d, want 2", tabs.Active())
	}

	// Wrap forward
	tabs.Update(newKeyMsg("right"))
	if tabs.Active() != 0 {
		t.Errorf("after wrap forward, active = %d, want 0", tabs.Active())
	}

	// Prev
	tabs.Update(newKeyMsg("left"))
	if tabs.Active() != 2 {
		t.Errorf("after left (wrap back), active = %d, want 2", tabs.Active())
	}

	tabs.Update(newKeyMsg("h"))
	if tabs.Active() != 1 {
		t.Errorf("after h, active = %d, want 1", tabs.Active())
	}
}

func TestTabsHomeEnd(t *testing.T) {
	tabs := NewTabs([]string{"Tab1", "Tab2", "Tab3", "Tab4"})
	tabs.Focus()
	tabs.SetActive(2)

	// Home
	tabs.Update(newKeyMsg("home"))
	if tabs.Active() != 0 {
		t.Errorf("after home, active = %d, want 0", tabs.Active())
	}

	// End
	tabs.Update(newKeyMsg("end"))
	if tabs.Active() != 3 {
		t.Errorf("after end, active = %d, want 3", tabs.Active())
	}

	// g (home alternative)
	tabs.Update(newKeyMsg("g"))
	if tabs.Active() != 0 {
		t.Errorf("after g, active = %d, want 0", tabs.Active())
	}

	// G (end alternative)
	tabs.Update(newKeyMsg("G"))
	if tabs.Active() != 3 {
		t.Errorf("after G, active = %d, want 3", tabs.Active())
	}
}

func TestTabsNumericJump(t *testing.T) {
	tabs := NewTabs([]string{"Tab1", "Tab2", "Tab3", "Tab4", "Tab5"})
	tabs.Focus()

	tests := []struct {
		key  string
		want int
	}{
		{"1", 0},
		{"3", 2},
		{"5", 4},
		{"9", 4}, // out of range, should stay at 4
		{"2", 1},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			tabs.Update(newKeyMsg(tt.key))
			if tabs.Active() != tt.want {
				t.Errorf("after %s, active = %d, want %d", tt.key, tabs.Active(), tt.want)
			}
		})
	}
}

func TestTabsSetActive(t *testing.T) {
	tabs := NewTabs([]string{"Tab1", "Tab2", "Tab3"})

	tests := []struct {
		name  string
		index int
		want  int
	}{
		{"valid middle", 1, 1},
		{"valid last", 2, 2},
		{"clamp high", 10, 2},
		{"clamp low", -5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tabs.SetActive(tt.index)
			if tabs.Active() != tt.want {
				t.Errorf("SetActive(%d): got %d, want %d", tt.index, tabs.Active(), tt.want)
			}
		})
	}
}

func TestTabsZeroTabs(t *testing.T) {
	tabs := NewTabs([]string{})
	tabs.Focus()

	// Should not panic
	tabs.Next()
	tabs.Prev()
	tabs.SetActive(5)

	if tabs.Active() != 0 {
		t.Errorf("zero tabs active = %d, want 0", tabs.Active())
	}

	view := tabs.View()
	if view != "" {
		t.Errorf("zero tabs view = %q, want empty", view)
	}
}

func TestTabsView(t *testing.T) {
	tabs := NewTabs([]string{"Home", "Profile", "Settings"})
	tabs.SetSize(80, 1)
	tabs.SetActive(1)

	view := tabs.View()
	if view == "" {
		t.Error("view should not be empty")
	}

	// Should contain all tab titles
	if !strings.Contains(view, "Home") {
		t.Error("view missing 'Home'")
	}
	if !strings.Contains(view, "Profile") {
		t.Error("view missing 'Profile'")
	}
	if !strings.Contains(view, "Settings") {
		t.Error("view missing 'Settings'")
	}
}

func TestTabsUnfocusedNoUpdate(t *testing.T) {
	tabs := NewTabs([]string{"Tab1", "Tab2", "Tab3"})
	// Keep unfocused

	tabs.Update(newKeyMsg("right"))
	if tabs.Active() != 0 {
		t.Error("unfocused tabs should not respond to keys")
	}
}

func TestTabsNextPrev(t *testing.T) {
	tabs := NewTabs([]string{"A", "B", "C"})

	tabs.Next()
	if tabs.Active() != 1 {
		t.Errorf("Next: got %d, want 1", tabs.Active())
	}

	tabs.Next()
	if tabs.Active() != 2 {
		t.Errorf("Next: got %d, want 2", tabs.Active())
	}

	tabs.Next()
	if tabs.Active() != 0 {
		t.Errorf("Next wrap: got %d, want 0", tabs.Active())
	}

	tabs.Prev()
	if tabs.Active() != 2 {
		t.Errorf("Prev wrap: got %d, want 2", tabs.Active())
	}

	tabs.Prev()
	if tabs.Active() != 1 {
		t.Errorf("Prev: got %d, want 1", tabs.Active())
	}
}
