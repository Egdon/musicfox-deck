// Tree example — demonstrates the Tree widget using the foxful-cli App framework.
//
// This example showcases the Tree widget integrated with the Menu system.
// Select a menu item to display different hierarchical tree structures.
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
)

// MainMenu provides navigation to different tree demos
type MainMenu struct {
	model.DefaultMenu
	menus []model.MenuItem
}

func NewMainMenu() *MainMenu {
	return &MainMenu{
		menus: []model.MenuItem{
			{Title: "File System Tree", Subtitle: "Browse project structure"},
			{Title: "Organization Chart", Subtitle: "Company hierarchy"},
			{Title: "Menu Structure", Subtitle: "Application menu tree"},
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
	// Navigate to the tree page with the selected demo
	return NewTreePage(app, index), nil
}

// TreePage displays a Tree widget as the main content
type TreePage struct {
	app       *model.App
	demoIndex int
	tree      *model.Tree
	width     int
	height    int
}

func NewTreePage(app *model.App, demoIndex int) *TreePage {
	var root *model.TreeNode

	switch demoIndex {
	case 0: // File System
		root = createFileSystemTree()
	case 1: // Organization
		root = createOrganizationTree()
	case 2: // Menu Structure
		root = createMenuTree()
	default:
		root = &model.TreeNode{Label: "Root"}
	}

	tree := model.NewTree(root)
	tree.Focus()

	return &TreePage{
		app:       app,
		demoIndex: demoIndex,
		tree:      tree,
	}
}

func createFileSystemTree() *model.TreeNode {
	return &model.TreeNode{
		Label:    "~/projects",
		Expanded: true,
		Children: []*model.TreeNode{
			{
				Label:    "foxful-cli",
				Expanded: true,
				Children: []*model.TreeNode{
					{
						Label:    "cmd",
						Expanded: false,
						Children: []*model.TreeNode{
							{Label: "main.go"},
							{Label: "version.go"},
							{Label: "config.go"},
						},
					},
					{
						Label:    "model",
						Expanded: true,
						Children: []*model.TreeNode{
							{Label: "app.go"},
							{Label: "tabs.go"},
							{Label: "tree.go"},
							{Label: "filepicker.go"},
							{Label: "table.go"},
							{Label: "form.go"},
						},
					},
					{
						Label:    "example",
						Expanded: false,
						Children: []*model.TreeNode{
							{
								Label: "tabs",
								Children: []*model.TreeNode{
									{Label: "main.go"},
								},
							},
							{
								Label: "tree",
								Children: []*model.TreeNode{
									{Label: "main.go"},
								},
							},
							{
								Label: "filepicker",
								Children: []*model.TreeNode{
									{Label: "main.go"},
								},
							},
							{
								Label: "widgets",
								Children: []*model.TreeNode{
									{Label: "main.go"},
								},
							},
						},
					},
					{
						Label:    "style",
						Expanded: false,
						Children: []*model.TreeNode{
							{Label: "theme.go"},
							{Label: "colors.go"},
						},
					},
					{Label: "README.md"},
					{Label: "go.mod"},
					{Label: "go.sum"},
					{Label: "LICENSE"},
				},
			},
			{
				Label:    "website",
				Expanded: false,
				Children: []*model.TreeNode{
					{
						Label: "src",
						Children: []*model.TreeNode{
							{Label: "index.html"},
							{Label: "styles.css"},
							{Label: "app.js"},
							{Label: "components.js"},
						},
					},
					{
						Label: "assets",
						Children: []*model.TreeNode{
							{Label: "logo.png"},
							{Label: "banner.jpg"},
							{Label: "favicon.ico"},
						},
					},
					{
						Label: "docs",
						Children: []*model.TreeNode{
							{Label: "guide.md"},
							{Label: "api.md"},
						},
					},
					{Label: "package.json"},
					{Label: "webpack.config.js"},
				},
			},
			{
				Label:    "scripts",
				Expanded: false,
				Children: []*model.TreeNode{
					{Label: "deploy.sh"},
					{Label: "backup.sh"},
					{Label: "cleanup.sh"},
					{Label: "test.sh"},
				},
			},
		},
	}
}

func createOrganizationTree() *model.TreeNode {
	return &model.TreeNode{
		Label:    "Acme Corporation",
		Expanded: true,
		Children: []*model.TreeNode{
			{
				Label:    "Executive Team",
				Expanded: true,
				Children: []*model.TreeNode{
					{Label: "CEO - Alice Johnson"},
					{Label: "CTO - Bob Smith"},
					{Label: "CFO - Carol Martinez"},
				},
			},
			{
				Label:    "Engineering",
				Expanded: true,
				Children: []*model.TreeNode{
					{
						Label:    "Frontend Team",
						Expanded: false,
						Children: []*model.TreeNode{
							{Label: "Lead - David Chen"},
							{Label: "Developer - Eva Rodriguez"},
							{Label: "Developer - Frank Wilson"},
						},
					},
					{
						Label:    "Backend Team",
						Expanded: false,
						Children: []*model.TreeNode{
							{Label: "Lead - Grace Kim"},
							{Label: "Developer - Henry Lee"},
							{Label: "Developer - Iris Brown"},
						},
					},
					{
						Label:    "DevOps Team",
						Expanded: false,
						Children: []*model.TreeNode{
							{Label: "Lead - Jack Davis"},
							{Label: "Engineer - Kate Miller"},
						},
					},
				},
			},
			{
				Label:    "Product",
				Expanded: false,
				Children: []*model.TreeNode{
					{Label: "Product Manager - Laura White"},
					{Label: "Designer - Mike Taylor"},
					{Label: "Analyst - Nina Anderson"},
				},
			},
			{
				Label:    "Sales & Marketing",
				Expanded: false,
				Children: []*model.TreeNode{
					{
						Label: "Sales",
						Children: []*model.TreeNode{
							{Label: "Sales Director - Oscar Brown"},
							{Label: "Account Manager - Paula Green"},
						},
					},
					{
						Label: "Marketing",
						Children: []*model.TreeNode{
							{Label: "Marketing Director - Quinn Hall"},
							{Label: "Content Writer - Rachel Moore"},
						},
					},
				},
			},
		},
	}
}

func createMenuTree() *model.TreeNode {
	return &model.TreeNode{
		Label:    "Application Menu",
		Expanded: true,
		Children: []*model.TreeNode{
			{
				Label:    "File",
				Expanded: true,
				Children: []*model.TreeNode{
					{Label: "New"},
					{Label: "Open..."},
					{Label: "Save"},
					{Label: "Save As..."},
					{Label: "---"},
					{
						Label:    "Recent Files",
						Expanded: false,
						Children: []*model.TreeNode{
							{Label: "project1.txt"},
							{Label: "document.md"},
							{Label: "notes.txt"},
						},
					},
					{Label: "---"},
					{Label: "Exit"},
				},
			},
			{
				Label:    "Edit",
				Expanded: false,
				Children: []*model.TreeNode{
					{Label: "Undo"},
					{Label: "Redo"},
					{Label: "---"},
					{Label: "Cut"},
					{Label: "Copy"},
					{Label: "Paste"},
					{Label: "---"},
					{Label: "Find..."},
					{Label: "Replace..."},
				},
			},
			{
				Label:    "View",
				Expanded: false,
				Children: []*model.TreeNode{
					{Label: "Zoom In"},
					{Label: "Zoom Out"},
					{Label: "Reset Zoom"},
					{Label: "---"},
					{Label: "Toggle Sidebar"},
					{Label: "Toggle Status Bar"},
					{
						Label: "Layout",
						Children: []*model.TreeNode{
							{Label: "Default"},
							{Label: "Compact"},
							{Label: "Wide"},
						},
					},
				},
			},
			{
				Label:    "Tools",
				Expanded: false,
				Children: []*model.TreeNode{
					{Label: "Settings..."},
					{Label: "Extensions..."},
					{Label: "Keyboard Shortcuts..."},
					{Label: "---"},
					{Label: "Terminal"},
					{Label: "Command Palette"},
				},
			},
			{
				Label:    "Help",
				Expanded: false,
				Children: []*model.TreeNode{
					{Label: "Documentation"},
					{Label: "Tutorials"},
					{Label: "Community Forum"},
					{Label: "---"},
					{Label: "Report Issue"},
					{Label: "Check for Updates"},
					{Label: "---"},
					{Label: "About"},
				},
			},
		},
	}
}

func (p *TreePage) IgnoreQuitKeyMsg(msg tea.KeyMsg) bool {
	return false
}

func (p *TreePage) Type() model.PageType {
	return model.PtMain
}

func (p *TreePage) Msg() tea.Msg {
	return nil
}

func (p *TreePage) Update(msg tea.Msg, app *model.App) (model.Page, tea.Cmd) {
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
		p.tree.SetSize(msg.Width-4, msg.Height-12)
		return p, nil
	}

	// Delegate to tree widget
	p.tree.Update(msg)

	return p, nil
}

func (p *TreePage) View(app *model.App) string {
	w, h := app.WindowWidth(), app.WindowHeight()
	if w <= 0 || h <= 0 {
		return ""
	}

	// Set widget size based on current window dimensions
	p.tree.SetSize(w-4, h-10)

	ss := style.CurrentStyleSet()

	// Header
	var title string
	switch p.demoIndex {
	case 0:
		title = "File System Tree Demo"
	case 1:
		title = "Organization Chart Demo"
	case 2:
		title = "Menu Structure Demo"
	}

	header := ss.Title.Render(title)
	instructions := lipgloss.JoinVertical(
		lipgloss.Left,
		ss.Muted.Render("up/down: navigate • enter/space: toggle • left/right: collapse/expand"),
		ss.Muted.Render("g/G: first/last • b/esc: back • q: quit"),
	)

	// Tree content
	treeView := p.tree.View()
	if treeView == "" {
		treeView = ss.Muted.Render("(empty tree)")
	}

	treeBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ss.SelectedItem.GetForeground()).
		Padding(1, 2).
		Width(p.width - 4).
		Height(p.height - 14).
		Render(treeView)

	// Selected node info
	selectedNode := p.tree.SelectedNode()
	var nodeInfo string
	if selectedNode != nil {
		childCount := len(selectedNode.Children)
		if childCount > 0 {
			expandStatus := "collapsed"
			if selectedNode.Expanded {
				expandStatus = "expanded"
			}
			nodeInfo = ss.Info.Render(fmt.Sprintf("Selected: %s (%d children, %s)", selectedNode.Label, childCount, expandStatus))
		} else {
			nodeInfo = ss.Info.Render(fmt.Sprintf("Selected: %s (leaf node)", selectedNode.Label))
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		instructions,
		"",
		treeBox,
		nodeInfo,
	)
}

func main() {
	ops := model.DefaultOptions()
	app := model.NewApp(ops)

	ops.MainMenu = NewMainMenu()

	fmt.Println(app.Run())
}
