package configs

import (
	"image/color"
	"testing"

	"github.com/anhoder/foxful-cli/style"
)

func TestThemeMenuItemsUseAppBackground(t *testing.T) {
	theme, err := parseThemeFile(`
name = "menu-background"

[dark.highlights]
selectedItemBg = "#121212"

[dark.app]
background = "#E1E2E3"
`)
	if err != nil {
		t.Fatalf("parse theme: %v", err)
	}

	styles := style.NewStyleSet(theme.Dark.toTheme())
	if !sameColor(styles.MenuItem.GetBackground(), styles.AppBackground.GetBackground()) {
		t.Fatalf("menu item background = %v, want app background %v", styles.MenuItem.GetBackground(), styles.AppBackground.GetBackground())
	}
}

func sameColor(left, right color.Color) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	leftR, leftG, leftB, leftA := left.RGBA()
	rightR, rightG, rightB, rightA := right.RGBA()
	return leftR == rightR && leftG == rightG && leftB == rightB && leftA == rightA
}

func TestThemeExtraKeysMapToStyles(t *testing.T) {
	theme, err := parseThemeFile(`
name = "extra-keys"

[dark.highlights]
selectedItem = "#DDDDDD"
selectedItemBg = "#222222"

[dark.popup]
actionFocusedFg = "#111111"
actionFocusedBg = "#CCCCCC"
contextMenuItem = "#BBBBBB"

[dark.notification]
title = "#ABABAB"
message = "#9A9A9A"
infoIcon = "○ "

[dark.statusBar]
barBg = "#101010"
breadcrumbBg = "#202020"
timeBg = "#303030"

[dark.app]
contextMenuBorder = "#404040"
`)
	if err != nil {
		t.Fatalf("parse theme: %v", err)
	}

	ss := style.NewStyleSet(theme.Dark.toTheme())
	checks := []struct {
		name      string
		got, want color.Color
	}{
		{"selected item fg", ss.SelectedItem.GetForeground(), lipglossColor("#DDDDDD")},
		{"selected item bg", ss.SelectedItem.GetBackground(), lipglossColor("#222222")},
		{"focused action fg", ss.Popup.ActionFocused.GetForeground(), lipglossColor("#111111")},
		{"focused action bg", ss.Popup.ActionFocused.GetBackground(), lipglossColor("#CCCCCC")},
		{"context menu item fg", ss.Popup.ContextMenuItem.GetForeground(), lipglossColor("#BBBBBB")},
		{"context menu border", ss.Popup.ContextMenuFrame.GetBorderTopForeground(), lipglossColor("#404040")},
		{"notification title", ss.Notification.Title.GetForeground(), lipglossColor("#ABABAB")},
		{"notification message", ss.Notification.Message.GetForeground(), lipglossColor("#9A9A9A")},
		{"status bar bg", ss.StatusBar.GetBackground(), lipglossColor("#101010")},
		{"breadcrumb bg", ss.StatusBarBreadcrumb.GetBackground(), lipglossColor("#202020")},
		{"time bg", ss.StatusBarTime.GetBackground(), lipglossColor("#303030")},
	}
	for _, c := range checks {
		if !sameColor(c.got, c.want) {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
	if ss.Notification.InfoIcon != "○ " {
		t.Errorf("info icon = %q, want %q", ss.Notification.InfoIcon, "○ ")
	}
}

func lipglossColor(hex string) color.Color {
	return parseColor(hex)
}
