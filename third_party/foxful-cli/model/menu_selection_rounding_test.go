package model

import (
	"strings"
	"testing"

	"github.com/anhoder/foxful-cli/style"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

func TestSelectedMenuEntryKeepsFinalContentCharacter(t *testing.T) {
	previous := style.CurrentStyleSet()
	style.SetStyleSet(style.NewStyleSet(style.DefaultDarkTheme()))
	t.Cleanup(func() { style.SetStyleSet(previous) })

	m := &Main{
		options:       &Options{Ticker: nil},
		menuList:      []MenuItem{{Title: "0123456789ABC"}},
		menuPageSize:  10,
		menuCurPage:   1,
		selectedIndex: 0,
	}

	plain := ansi.Strip(m.formatEntry(&m.menuList[0], 0, 20))
	rightSep := style.DefaultDarkTheme().MenuSelectedSepRight
	if !strings.Contains(plain, "0123456789ABC") {
		t.Fatalf("selected entry lost final content character: %q", plain)
	}
	if !strings.HasSuffix(plain, rightSep) {
		t.Fatalf("selected entry does not end with right separator: %q", plain)
	}
	wantWidth := 20 + runewidth.StringWidth(rightSep)
	if runewidth.StringWidth(plain) != wantWidth {
		t.Fatalf("selected entry width = %d, want %d: %q", runewidth.StringWidth(plain), wantWidth, plain)
	}
}

func TestSelectedMenuEntryUsesRightSeparatorForTrailingPadding(t *testing.T) {
	previous := style.CurrentStyleSet()
	style.SetStyleSet(style.NewStyleSet(style.DefaultDarkTheme()))
	t.Cleanup(func() { style.SetStyleSet(previous) })

	m := &Main{
		options:       &Options{Ticker: nil},
		menuList:      []MenuItem{{Title: "song"}},
		menuPageSize:  10,
		menuCurPage:   1,
		selectedIndex: 0,
	}

	plain := ansi.Strip(m.formatEntry(&m.menuList[0], 0, 20))
	if !strings.HasSuffix(plain, style.DefaultDarkTheme().MenuSelectedSepRight) {
		t.Fatalf("selected entry does not end with right separator: %q", plain)
	}
	if !strings.Contains(plain, "song") {
		t.Fatalf("selected entry lost content: %q", plain)
	}
	if runewidth.StringWidth(plain) != 20 {
		t.Fatalf("selected entry width = %d, want 20: %q", runewidth.StringWidth(plain), plain)
	}
}

func TestCenteredOverflowSelectionKeepsTitleColumn(t *testing.T) {
	previous := style.CurrentStyleSet()
	style.SetStyleSet(style.NewStyleSet(style.DefaultDarkTheme()))
	t.Cleanup(func() { style.SetStyleSet(previous) })

	app := &App{windowWidth: 30}
	m := &Main{
		app:           app,
		options:       &Options{Ticker: nil},
		menuList:      []MenuItem{{Title: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"}},
		menuPageSize:  10,
		menuCurPage:   1,
		selectedIndex: 0,
	}

	titleColumn := func(view string) int {
		plain := ansi.Strip(view)
		offset := strings.IndexByte(plain, 'A')
		if offset < 0 {
			t.Fatalf("rendered entry lost title: %q", plain)
		}
		return runewidth.StringWidth(plain[:offset])
	}

	selectedColumn := titleColumn(m.centeredMenuView(app, 1))
	m.selectedIndex = 1
	unselectedColumn := titleColumn(m.centeredMenuView(app, 1))
	if selectedColumn != unselectedColumn {
		t.Fatalf("selected title column = %d, unselected = %d", selectedColumn, unselectedColumn)
	}
}

func TestCenteredOverflowLeftSelectionKeepsRightColumn(t *testing.T) {
	previous := style.CurrentStyleSet()
	style.SetStyleSet(style.NewStyleSet(style.DefaultDarkTheme()))
	t.Cleanup(func() { style.SetStyleSet(previous) })

	app := &App{windowWidth: 60}
	m := &Main{
		app:           app,
		options:       &Options{Ticker: nil},
		isDualColumn:  true,
		menuList:      []MenuItem{{Title: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"}, {Title: "RIGHT"}},
		menuPageSize:  10,
		menuCurPage:   1,
		selectedIndex: 0,
	}

	rightColumn := func(view string) int {
		plain := ansi.Strip(view)
		offset := strings.Index(plain, "RIGHT")
		if offset < 0 {
			t.Fatalf("rendered entry lost right column title: %q", plain)
		}
		return runewidth.StringWidth(plain[:offset])
	}

	selectedColumn := rightColumn(m.centeredMenuView(app, 1))
	m.selectedIndex = 99
	unselectedColumn := rightColumn(m.centeredMenuView(app, 1))
	if selectedColumn != unselectedColumn {
		t.Fatalf("right column = %d when left selected, want %d", selectedColumn, unselectedColumn)
	}
}
