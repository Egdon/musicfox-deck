package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
)

func main() {
	ops := model.DefaultOptions()
	ops.AppName = "Multi-Tab Navigation Demo"

	// Enable multi-tab mode
	ops.EnableTabs = true
	ops.TabConfigs = []model.TabConfig{
		{
			Title:     "Dashboard",
			Menu:      NewDashboardMenu(),
			MenuTitle: &model.MenuItem{Title: "Dashboard"},
		},
		{
			Title:     "Settings",
			Menu:      NewSettingsMenu(),
			MenuTitle: &model.MenuItem{Title: "Settings"},
		},
		{
			Title:     "Logs",
			Menu:      NewLogsMenu(),
			MenuTitle: &model.MenuItem{Title: "System Logs"},
		},
	}

	app := model.NewApp(ops)
	if err := app.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}

// ── Dashboard Tab ──

type DashboardMenu struct {
	model.DefaultMenu
	items []model.MenuItem
}

func NewDashboardMenu() *DashboardMenu {
	return &DashboardMenu{
		items: []model.MenuItem{
			{Title: "Overview", Subtitle: "System summary"},
			{Title: "Projects", Subtitle: "12 active"},
			{Title: "Users", Subtitle: "245 registered"},
			{Title: "Analytics", Subtitle: "View reports"},
			{Title: "Notifications", Subtitle: "3 new"},
		},
	}
}

func (m *DashboardMenu) GetMenuKey() string {
	return "dashboard_menu"
}

func (m *DashboardMenu) MenuViews() []model.MenuItem {
	return m.items
}

func (m *DashboardMenu) SubMenu(_ *model.App, index int) model.Menu {
	switch index {
	case 1: // Projects
		return NewProjectsSubmenu()
	case 2: // Users
		return NewUsersSubmenu()
	default:
		return nil
	}
}

func (m *DashboardMenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	// Show popup for non-submenu items
	if index == 0 || index == 3 || index == 4 {
		itemTitle := m.items[index].Title
		popup, _ := model.NewPopup(model.PopupSpec{
			Title:   itemTitle,
			Content: fmt.Sprintf("Viewing %s details...", itemTitle),
			Actions: []model.PopupAction{
				{ID: "ok", Label: "OK", IsCancel: true},
			},
		})
		app.ShowPopup(popup)
		return nil, app.RerenderCmd(true)
	}
	return nil, nil
}

// ── Projects Submenu ──

type ProjectsSubmenu struct {
	model.DefaultMenu
	items []model.MenuItem
}

func NewProjectsSubmenu() *ProjectsSubmenu {
	return &ProjectsSubmenu{
		items: []model.MenuItem{
			{Title: "Project Alpha", Subtitle: "Active"},
			{Title: "Project Beta", Subtitle: "In Review"},
			{Title: "Project Gamma", Subtitle: "Completed"},
			{Title: "Project Delta", Subtitle: "Planning"},
		},
	}
}

func (m *ProjectsSubmenu) GetMenuKey() string {
	return "projects_submenu"
}

func (m *ProjectsSubmenu) MenuViews() []model.MenuItem {
	return m.items
}

func (m *ProjectsSubmenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	itemTitle := m.items[index].Title
	popup, _ := model.NewPopup(model.PopupSpec{
		Title:   "Project Details",
		Content: fmt.Sprintf("Details for %s\n\nStatus: %s\nTeam: 5 members\nDeadline: Q2 2025", itemTitle, m.items[index].Subtitle),
		Actions: []model.PopupAction{
			{ID: "ok", Label: "OK", IsCancel: true},
		},
	})
	app.ShowPopup(popup)
	return nil, app.RerenderCmd(true)
}

// ── Users Submenu ──

type UsersSubmenu struct {
	model.DefaultMenu
	items []model.MenuItem
}

func NewUsersSubmenu() *UsersSubmenu {
	return &UsersSubmenu{
		items: []model.MenuItem{
			{Title: "Alice Johnson", Subtitle: "Admin"},
			{Title: "Bob Smith", Subtitle: "Developer"},
			{Title: "Charlie Brown", Subtitle: "Designer"},
			{Title: "Diana Prince", Subtitle: "Manager"},
		},
	}
}

func (m *UsersSubmenu) GetMenuKey() string {
	return "users_submenu"
}

func (m *UsersSubmenu) MenuViews() []model.MenuItem {
	return m.items
}

func (m *UsersSubmenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	itemTitle := m.items[index].Title
	popup, _ := model.NewPopup(model.PopupSpec{
		Title:   "User Profile",
		Content: fmt.Sprintf("%s\n\nRole: %s\nEmail: %s@example.com\nStatus: Active", itemTitle, m.items[index].Subtitle, itemTitle[:3]),
		Actions: []model.PopupAction{
			{ID: "ok", Label: "OK", IsCancel: true},
		},
	})
	app.ShowPopup(popup)
	return nil, app.RerenderCmd(true)
}

// ── Settings Tab ──

type SettingsMenu struct {
	model.DefaultMenu
	items []model.MenuItem
}

func NewSettingsMenu() *SettingsMenu {
	return &SettingsMenu{
		items: []model.MenuItem{
			{Title: "Account", Subtitle: "Profile & security"},
			{Title: "Appearance", Subtitle: "Theme & layout"},
			{Title: "Notifications", Subtitle: "Email & alerts"},
			{Title: "Privacy", Subtitle: "Data & permissions"},
			{Title: "Advanced", Subtitle: "Developer options"},
		},
	}
}

func (m *SettingsMenu) GetMenuKey() string {
	return "settings_menu"
}

func (m *SettingsMenu) MenuViews() []model.MenuItem {
	return m.items
}

func (m *SettingsMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index == 1 { // Appearance
		return NewAppearanceSubmenu()
	}
	return nil
}

func (m *SettingsMenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	// Show popup for non-submenu items
	if index != 1 {
		itemTitle := m.items[index].Title
		popup, _ := model.NewPopup(model.PopupSpec{
			Title:   itemTitle + " Settings",
			Content: fmt.Sprintf("Configure %s settings here...", itemTitle),
			Actions: []model.PopupAction{
				{ID: "ok", Label: "OK", IsCancel: true},
			},
		})
		app.ShowPopup(popup)
		return nil, app.RerenderCmd(true)
	}
	return nil, nil
}

// ── Appearance Submenu ──

type AppearanceSubmenu struct {
	model.DefaultMenu
	items []model.MenuItem
}

func NewAppearanceSubmenu() *AppearanceSubmenu {
	return &AppearanceSubmenu{
		items: []model.MenuItem{
			{Title: "Theme", Subtitle: "Dark mode"},
			{Title: "Font Size", Subtitle: "Medium"},
			{Title: "Layout", Subtitle: "Dual column"},
		},
	}
}

func (m *AppearanceSubmenu) GetMenuKey() string {
	return "appearance_submenu"
}

func (m *AppearanceSubmenu) MenuViews() []model.MenuItem {
	return m.items
}

func (m *AppearanceSubmenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	itemTitle := m.items[index].Title
	popup, _ := model.NewPopup(model.PopupSpec{
		Title:   itemTitle,
		Content: fmt.Sprintf("Current: %s\n\nClick to change...", m.items[index].Subtitle),
		Actions: []model.PopupAction{
			{ID: "ok", Label: "OK", IsCancel: true},
		},
	})
	app.ShowPopup(popup)
	return nil, app.RerenderCmd(true)
}

// ── Logs Tab ──

type LogsMenu struct {
	model.DefaultMenu
	items []model.MenuItem
}

func NewLogsMenu() *LogsMenu {
	return &LogsMenu{
		items: []model.MenuItem{
			{Title: "Application Logs", Subtitle: "Last updated: 2m ago"},
			{Title: "Error Logs", Subtitle: "3 errors today"},
			{Title: "Access Logs", Subtitle: "1.2k requests"},
			{Title: "Audit Trail", Subtitle: "Security events"},
			{Title: "Debug Logs", Subtitle: "Developer mode"},
		},
	}
}

func (m *LogsMenu) GetMenuKey() string {
	return "logs_menu"
}

func (m *LogsMenu) MenuViews() []model.MenuItem {
	return m.items
}

func (m *LogsMenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	itemTitle := m.items[index].Title
	popup, _ := model.NewPopup(model.PopupSpec{
		Title:   itemTitle,
		Content: fmt.Sprintf("Viewing %s\n\n[12:34:56] INFO  System started\n[12:35:10] WARN  Cache miss\n[12:35:45] ERROR Connection timeout\n\nShowing last 100 entries...", itemTitle),
		Actions: []model.PopupAction{
			{ID: "ok", Label: "OK", IsCancel: true},
		},
	})
	app.ShowPopup(popup)
	return nil, app.RerenderCmd(true)
}
