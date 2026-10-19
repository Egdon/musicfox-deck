// Tabs example — demonstrates the Tabs widget using the foxful-cli App framework.
//
// This example showcases the Tabs widget integrated with the Menu system.
// Select a menu item to display a Component that uses Tabs for navigation.
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
)

// MainMenu provides navigation to different tab demos
type MainMenu struct {
	model.DefaultMenu
	menus []model.MenuItem
}

func NewMainMenu() *MainMenu {
	return &MainMenu{
		menus: []model.MenuItem{
			{Title: "Dashboard Tabs", Subtitle: "System overview with multiple views"},
			{Title: "Settings Tabs", Subtitle: "Configuration panels"},
			{Title: "Analytics Tabs", Subtitle: "Data visualization"},
		},
	}
}

func (m *MainMenu) GetMenuKey() string {
	return "main_menu"
}

func (m *MainMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *MainMenu) SubMenu(app *model.App, index int) model.Menu {
	return nil // Using Action instead
}

func (m *MainMenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	// Navigate to the component page with the selected demo
	return NewTabsPage(app, index), nil
}

// TabsPage displays a Tabs widget as the main content
type TabsPage struct {
	app       *model.App
	demoIndex int
	tabs      *model.Tabs
	width     int
	height    int
}

func NewTabsPage(app *model.App, demoIndex int) *TabsPage {
	var tabs *model.Tabs

	switch demoIndex {
	case 0: // Dashboard
		tabs = model.NewTabs([]string{"Overview", "Performance", "Users", "System"})
	case 1: // Settings
		tabs = model.NewTabs([]string{"General", "Appearance", "Notifications", "Advanced"})
	case 2: // Analytics
		tabs = model.NewTabs([]string{"Traffic", "Engagement", "Conversion", "Reports"})
	default:
		tabs = model.NewTabs([]string{"Tab 1", "Tab 2", "Tab 3"})
	}

	tabs.Focus()

	return &TabsPage{
		app:       app,
		demoIndex: demoIndex,
		tabs:      tabs,
	}
}

func (p *TabsPage) IgnoreQuitKeyMsg(msg tea.KeyMsg) bool {
	return false
}

func (p *TabsPage) Type() model.PageType {
	return model.PtMain
}

func (p *TabsPage) Msg() tea.Msg {
	return nil
}

func (p *TabsPage) Update(msg tea.Msg, app *model.App) (model.Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "b":
			// Return to main menu
			return app.Main(), nil
		case "q", "ctrl+c":
			return p, tea.Quit
		}
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.tabs.SetSize(msg.Width-4, msg.Height-8)
		return p, nil
	}

	// Delegate to tabs widget
	p.tabs.Update(msg)

	// Update content based on active tab
	p.updateTabContent()

	return p, nil
}

func (p *TabsPage) updateTabContent() {
	active := p.tabs.Active()
	var content string

	switch p.demoIndex {
	case 0: // Dashboard
		content = p.renderDashboardContent(active)
	case 1: // Settings
		content = p.renderSettingsContent(active)
	case 2: // Analytics
		content = p.renderAnalyticsContent(active)
	}

	p.tabs.SetContent(content)
}

func (p *TabsPage) renderDashboardContent(tab int) string {
	ss := style.CurrentStyleSet()

	switch tab {
	case 0: // Overview
		return lipgloss.JoinVertical(
			lipgloss.Left,
			ss.Subtitle.Render("System Overview"),
			"",
			"Active Services: 12/12",
			"Uptime: 45 days, 3 hours",
			"Load Average: 0.45, 0.52, 0.48",
			"",
			ss.Success.Render("✓ All systems operational"),
		)
	case 1: // Performance
		return lipgloss.JoinVertical(
			lipgloss.Left,
			ss.Subtitle.Render("Performance Metrics"),
			"",
			"CPU: ████████░░░░░░░░░░░░ 40%",
			"RAM: ████████████░░░░░░░░ 60%",
			"Disk: ██████████████░░░░░░ 70%",
			"Network: ████░░░░░░░░░░░░░░ 20%",
		)
	case 2: // Users
		return lipgloss.JoinVertical(
			lipgloss.Left,
			ss.Subtitle.Render("User Statistics"),
			"",
			"Total Users: 1,234",
			"Online Now: 89",
			"New Today: 12",
			"Active Sessions: 156",
		)
	case 3: // System
		return lipgloss.JoinVertical(
			lipgloss.Left,
			ss.Subtitle.Render("System Information"),
			"",
			"Version: 1.0.0",
			"Build: 2026-07-24",
			"Platform: darwin/arm64",
			"Go Version: 1.22",
		)
	}
	return ""
}

func (p *TabsPage) renderSettingsContent(tab int) string {
	ss := style.CurrentStyleSet()

	switch tab {
	case 0: // General
		return lipgloss.JoinVertical(
			lipgloss.Left,
			ss.Subtitle.Render("General Settings"),
			"",
			"Language: English",
			"Timezone: UTC-8",
			"Auto-save: Enabled",
			"Check for updates: Daily",
		)
	case 1: // Appearance
		return lipgloss.JoinVertical(
			lipgloss.Left,
			ss.Subtitle.Render("Appearance"),
			"",
			"Theme: Dark Mode",
			"Font Size: 14px",
			"Line Height: 1.5",
			"Show Line Numbers: Yes",
		)
	case 2: // Notifications
		return lipgloss.JoinVertical(
			lipgloss.Left,
			ss.Subtitle.Render("Notifications"),
			"",
			"Desktop Notifications: Enabled",
			"Sound: Enabled",
			"Email Digest: Daily",
			"Quiet Hours: 22:00 - 08:00",
		)
	case 3: // Advanced
		return lipgloss.JoinVertical(
			lipgloss.Left,
			ss.Subtitle.Render("Advanced Settings"),
			"",
			"Debug Mode: Disabled",
			"Telemetry: Enabled",
			"Experimental Features: Disabled",
			"Developer Mode: Disabled",
		)
	}
	return ""
}

func (p *TabsPage) renderAnalyticsContent(tab int) string {
	ss := style.CurrentStyleSet()

	switch tab {
	case 0: // Traffic
		return lipgloss.JoinVertical(
			lipgloss.Left,
			ss.Subtitle.Render("Traffic Analytics"),
			"",
			"Daily Visitors",
			"  Mon: ████████████░░░░░░░░ 60%",
			"  Tue: ██████████████░░░░░░ 70%",
			"  Wed: ████████████████░░░░ 80%",
			"  Thu: ██████████████████░░ 90%",
			"  Fri: ████████████████████ 100%",
		)
	case 1: // Engagement
		return lipgloss.JoinVertical(
			lipgloss.Left,
			ss.Subtitle.Render("Engagement Metrics"),
			"",
			"Avg Session Duration: 12m 34s",
			"Pages per Session: 4.2",
			"Bounce Rate: 32%",
			"Return Visitor Rate: 68%",
		)
	case 2: // Conversion
		return lipgloss.JoinVertical(
			lipgloss.Left,
			ss.Subtitle.Render("Conversion Rates"),
			"",
			"Overall Conversion: 3.4%",
			"Cart Abandonment: 65%",
			"Email Signups: 12%",
			"Trial to Paid: 24%",
		)
	case 3: // Reports
		return lipgloss.JoinVertical(
			lipgloss.Left,
			ss.Subtitle.Render("Available Reports"),
			"",
			"• Weekly Summary Report",
			"• Monthly Analytics Report",
			"• Quarterly Business Review",
			"• Annual Performance Report",
		)
	}
	return ""
}

func (p *TabsPage) View(app *model.App) string {
	w, h := app.WindowWidth(), app.WindowHeight()
	if w <= 0 || h <= 0 {
		return ""
	}

	// Set widget size based on current window dimensions
	p.tabs.SetSize(w-4, h-8)

	ss := style.CurrentStyleSet()

	// Header
	var title string
	switch p.demoIndex {
	case 0:
		title = "Dashboard Tabs Demo"
	case 1:
		title = "Settings Tabs Demo"
	case 2:
		title = "Analytics Tabs Demo"
	}

	header := ss.Title.Render(title)
	instructions := ss.Muted.Render("left/right: switch tabs • 1-4: jump to tab • b/esc: back • q: quit")

	// Tabs with content (the widget handles the complete border system)
	tabsView := p.tabs.View()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		instructions,
		"",
		tabsView,
	)
}

func main() {
	ops := model.DefaultOptions()
	app := model.NewApp(ops)

	ops.MainMenu = NewMainMenu()

	fmt.Println(app.Run())
}
