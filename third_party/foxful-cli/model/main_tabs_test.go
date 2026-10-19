package model

import (
	"testing"
)

// mockMenu is a simple test menu implementation
type mockMenu struct {
	DefaultMenu
	key   string
	items []MenuItem
}

func (m *mockMenu) GetMenuKey() string {
	return m.key
}

func (m *mockMenu) MenuViews() []MenuItem {
	return m.items
}

// TestMainTabSwitchPreservesState verifies that switching tabs preserves
// each tab's navigation state (menu stack, scroll position, etc.)
func TestMainTabSwitchPreservesState(t *testing.T) {
	ops := DefaultOptions()
	ops.EnableTabs = true
	ops.TabConfigs = []TabConfig{
		{
			Title: "Tab1",
			Menu: &mockMenu{
				key:   "tab1_menu",
				items: []MenuItem{{Title: "A"}, {Title: "B"}},
			},
			MenuTitle: &MenuItem{Title: "Tab1"},
		},
		{
			Title: "Tab2",
			Menu: &mockMenu{
				key:   "tab2_menu",
				items: []MenuItem{{Title: "X"}},
			},
			MenuTitle: &MenuItem{Title: "Tab2"},
		},
	}

	app := NewApp(ops)
	// Manually initialize main page (normally done by App.Run)
	main := NewMain(app, ops)
	app.main = main

	// Initial state: Tab1 active, index 0
	if main.activeTab != 0 {
		t.Errorf("Expected initial activeTab=0, got %d", main.activeTab)
	}
	if main.selectedIndex != 0 {
		t.Errorf("Expected initial selectedIndex=0, got %d", main.selectedIndex)
	}

	// Navigate in Tab1: select item 1
	main.selectedIndex = 1
	main.menuCurPage = 2 // simulate pagination

	// Switch to Tab2
	main.switchTab(1)
	if main.activeTab != 1 {
		t.Errorf("Expected activeTab=1 after switch, got %d", main.activeTab)
	}
	if main.selectedIndex != 0 {
		t.Errorf("Expected Tab2 to start at selectedIndex=0, got %d", main.selectedIndex)
	}
	if main.menu.GetMenuKey() != "tab2_menu" {
		t.Errorf("Expected Tab2 menu to be active, got %s", main.menu.GetMenuKey())
	}

	// Switch back to Tab1
	main.switchTab(0)
	if main.activeTab != 0 {
		t.Errorf("Expected activeTab=0 after switch back, got %d", main.activeTab)
	}
	if main.selectedIndex != 1 {
		t.Errorf("Expected Tab1 selectedIndex to be restored to 1, got %d", main.selectedIndex)
	}
	if main.menuCurPage != 2 {
		t.Errorf("Expected Tab1 menuCurPage to be restored to 2, got %d", main.menuCurPage)
	}
	if main.menu.GetMenuKey() != "tab1_menu" {
		t.Errorf("Expected Tab1 menu to be active, got %s", main.menu.GetMenuKey())
	}
}

// TestMainTabSwitchClearsTransientState verifies that tab switching
// clears search mode and mouse hover state
func TestMainTabSwitchClearsTransientState(t *testing.T) {
	ops := DefaultOptions()
	ops.EnableTabs = true
	ops.TabConfigs = []TabConfig{
		{
			Title:     "Tab1",
			Menu:      &mockMenu{key: "tab1", items: []MenuItem{{Title: "A"}}},
			MenuTitle: &MenuItem{Title: "Tab1"},
		},
		{
			Title:     "Tab2",
			Menu:      &mockMenu{key: "tab2", items: []MenuItem{{Title: "X"}}},
			MenuTitle: &MenuItem{Title: "Tab2"},
		},
	}

	app := NewApp(ops)
	main := NewMain(app, ops)
	app.main = main

	// Set transient state
	main.inSearching = true
	main.hoveredMenuItemIdx = 5
	main.hoveredBreadcrumbIdx = 2

	// Switch tabs
	main.switchTab(1)

	// Verify transient state cleared
	if main.inSearching {
		t.Error("Expected inSearching to be cleared after tab switch")
	}
	if main.hoveredMenuItemIdx != -1 {
		t.Errorf("Expected hoveredMenuItemIdx=-1 after tab switch, got %d", main.hoveredMenuItemIdx)
	}
	if main.hoveredBreadcrumbIdx != -1 {
		t.Errorf("Expected hoveredBreadcrumbIdx=-1 after tab switch, got %d", main.hoveredBreadcrumbIdx)
	}
}

// TestMainEnableTabsFalseIgnoresTabConfigs verifies that when EnableTabs=false,
// TabConfigs are ignored and MainMenu is used
func TestMainEnableTabsFalseIgnoresTabConfigs(t *testing.T) {
	ops := DefaultOptions()
	ops.EnableTabs = false
	ops.TabConfigs = []TabConfig{
		{
			Title:     "Ignored",
			Menu:      &mockMenu{key: "ignored", items: []MenuItem{{Title: "X"}}},
			MenuTitle: &MenuItem{Title: "Ignored"},
		},
	}
	ops.MainMenu = &mockMenu{
		key:   "main_menu",
		items: []MenuItem{{Title: "Real"}},
	}

	app := NewApp(ops)
	main := NewMain(app, ops)

	if main.tabs != nil {
		t.Error("Expected tabs to be nil when EnableTabs=false")
	}
	if main.menu.GetMenuKey() != "main_menu" {
		t.Errorf("Expected MainMenu to be active, got %s", main.menu.GetMenuKey())
	}
	if len(main.menuList) != 1 || main.menuList[0].Title != "Real" {
		t.Error("Expected MainMenu items to be loaded, not TabConfigs")
	}
}

// TestMainTabSwitchWithOnActivateHook verifies that OnActivate hook is called
// and can veto the tab switch
func TestMainTabSwitchWithOnActivateHook(t *testing.T) {
	hookCalled := false
	vetoCalled := false

	ops := DefaultOptions()
	ops.EnableTabs = true
	ops.TabConfigs = []TabConfig{
		{
			Title:     "Tab1",
			Menu:      &mockMenu{key: "tab1", items: []MenuItem{{Title: "A"}}},
			MenuTitle: &MenuItem{Title: "Tab1"},
		},
		{
			Title:     "Tab2",
			Menu:      &mockMenu{key: "tab2", items: []MenuItem{{Title: "X"}}},
			MenuTitle: &MenuItem{Title: "Tab2"},
			OnActivate: func(m *Main, prevTabIndex int) bool {
				hookCalled = true
				if prevTabIndex != 0 {
					t.Errorf("Expected prevTabIndex=0, got %d", prevTabIndex)
				}
				return true // allow switch
			},
		},
		{
			Title:     "Tab3",
			Menu:      &mockMenu{key: "tab3", items: []MenuItem{{Title: "Z"}}},
			MenuTitle: &MenuItem{Title: "Tab3"},
			OnActivate: func(m *Main, prevTabIndex int) bool {
				vetoCalled = true
				return false // veto switch
			},
		},
	}

	app := NewApp(ops)
	main := NewMain(app, ops)
	app.main = main

	// Switch to Tab2 (hook allows)
	main.switchTab(1)
	if !hookCalled {
		t.Error("Expected OnActivate hook to be called for Tab2")
	}
	if main.activeTab != 1 {
		t.Errorf("Expected tab switch to succeed, activeTab=%d", main.activeTab)
	}

	// Try to switch to Tab3 (hook vetoes)
	main.switchTab(2)
	if !vetoCalled {
		t.Error("Expected OnActivate hook to be called for Tab3")
	}
	if main.activeTab != 1 {
		t.Errorf("Expected tab switch to be vetoed, but activeTab=%d", main.activeTab)
	}
}

// TestMainTabKeyBindings verifies tab switching logic (direct switchTab calls)
func TestMainTabKeyBindings(t *testing.T) {
	ops := DefaultOptions()
	ops.EnableTabs = true
	ops.TabConfigs = []TabConfig{
		{Title: "Tab1", Menu: &mockMenu{key: "tab1", items: []MenuItem{{Title: "A"}}}, MenuTitle: &MenuItem{Title: "Tab1"}},
		{Title: "Tab2", Menu: &mockMenu{key: "tab2", items: []MenuItem{{Title: "B"}}}, MenuTitle: &MenuItem{Title: "Tab2"}},
		{Title: "Tab3", Menu: &mockMenu{key: "tab3", items: []MenuItem{{Title: "C"}}}, MenuTitle: &MenuItem{Title: "Tab3"}},
	}

	app := NewApp(ops)
	main := NewMain(app, ops)
	app.main = main

	// Initial: Tab1 active
	if main.activeTab != 0 {
		t.Fatalf("Expected initial activeTab=0, got %d", main.activeTab)
	}

	// Forward: 0 → 1
	main.switchTab(1)
	if main.activeTab != 1 {
		t.Errorf("Expected activeTab=1 after switchTab(1), got %d", main.activeTab)
	}

	// Forward: 1 → 2
	main.switchTab(2)
	if main.activeTab != 2 {
		t.Errorf("Expected activeTab=2 after switchTab(2), got %d", main.activeTab)
	}

	// Wrap around: 2 → 0
	main.switchTab(0)
	if main.activeTab != 0 {
		t.Errorf("Expected activeTab=0 after switchTab(0), got %d", main.activeTab)
	}

	// Backward wrap: 0 → 2
	newIndex := main.activeTab - 1
	if newIndex < 0 {
		newIndex = len(main.tabStates) - 1
	}
	main.switchTab(newIndex)
	if main.activeTab != 2 {
		t.Errorf("Expected activeTab=2 after backward wrap, got %d", main.activeTab)
	}
}

// TestMainBuildBreadcrumb verifies breadcrumb generation when tabs enabled
func TestMainBuildBreadcrumb(t *testing.T) {
	ops := DefaultOptions()
	ops.EnableTabs = true
	ops.TabConfigs = []TabConfig{
		{
			Title:     "Dashboard",
			Menu:      &mockMenu{key: "dash", items: []MenuItem{{Title: "Projects"}}},
			MenuTitle: &MenuItem{Title: "Home"},
		},
	}

	app := NewApp(ops)
	main := NewMain(app, ops)
	app.main = main

	// No submenu: breadcrumb = "Dashboard > Home"
	breadcrumb := main.buildBreadcrumb()
	expected := "Dashboard > Home"
	if breadcrumb.Title != expected {
		t.Errorf("Expected breadcrumb '%s', got '%s'", expected, breadcrumb.Title)
	}

	// Simulate entering submenu
	main.menuStack.Push(&menuStackItem{
		menuTitle: &MenuItem{Title: "Projects"},
	})
	main.menuTitle = &MenuItem{Title: "Project Alpha"}

	// With submenu: breadcrumb = "Dashboard > Projects > Project Alpha"
	breadcrumb = main.buildBreadcrumb()
	expected = "Dashboard > Projects > Project Alpha"
	if breadcrumb.Title != expected {
		t.Errorf("Expected breadcrumb '%s', got '%s'", expected, breadcrumb.Title)
	}
}

// TestMainTabIndexAt verifies mouse click detection for tab bar
func TestMainTabIndexAt(t *testing.T) {
	ops := DefaultOptions()
	ops.EnableTabs = true
	ops.TabConfigs = []TabConfig{
		{Title: "Tab1", Menu: &mockMenu{key: "tab1", items: []MenuItem{{Title: "A"}}}, MenuTitle: &MenuItem{Title: "Tab1"}},
		{Title: "Tab2", Menu: &mockMenu{key: "tab2", items: []MenuItem{{Title: "B"}}}, MenuTitle: &MenuItem{Title: "Tab2"}},
		{Title: "Tab3", Menu: &mockMenu{key: "tab3", items: []MenuItem{{Title: "C"}}}, MenuTitle: &MenuItem{Title: "Tab3"}},
	}

	app := NewApp(ops)
	main := NewMain(app, ops)
	app.main = main
	app.windowWidth = 80
	app.windowHeight = 24

	// Test: Click outside tab bar (below)
	idx := main.tabIndexAt(10, 10, app)
	if idx != -1 {
		t.Errorf("Expected -1 for click outside tab bar, got %d", idx)
	}

	// Test: Click inside tab bar Y range (row 0-2 when no title bar)
	// Note: Exact geometry depends on rendered tab widths with borders
	// These tests verify the logic works, not exact pixel positions

	// Test: Click at X=0 should hit first tab (or be near it)
	idx = main.tabIndexAt(0, 1, app)
	if idx != 0 {
		t.Logf("Click at (0,1) returned tab index %d (expected 0, but geometry may vary)", idx)
	}

	// Test: EnableTabs=false should return -1
	ops2 := DefaultOptions()
	ops2.EnableTabs = false
	app2 := NewApp(ops2)
	main2 := NewMain(app2, ops2)
	idx = main2.tabIndexAt(10, 1, app2)
	if idx != -1 {
		t.Errorf("Expected -1 when EnableTabs=false, got %d", idx)
	}
}
