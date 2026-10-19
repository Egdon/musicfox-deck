// FilePicker example — demonstrates the FilePicker widget using the foxful-cli App framework.
//
// This example showcases the FilePicker widget integrated with the Menu system.
// Select a menu item to browse different starting directories.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
)

// MainMenu provides navigation to different filepicker demos
type MainMenu struct {
	model.DefaultMenu
	menus []model.MenuItem
}

func NewMainMenu() *MainMenu {
	return &MainMenu{
		menus: []model.MenuItem{
			{Title: "Browse Current Directory", Subtitle: "Start in working directory"},
			{Title: "Browse Home Directory", Subtitle: "Start in home folder"},
			{Title: "Browse Temp Directory", Subtitle: "Start in temporary files"},
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
	// Navigate to the filepicker page with the selected starting directory
	return NewFilePickerPage(app, index), nil
}

// FilePickerPage displays a FilePicker widget as the main content
type FilePickerPage struct {
	app         *model.App
	demoIndex   int
	picker      *model.FilePicker
	width       int
	height      int
	showHidden  bool
	lastMessage string
}

func NewFilePickerPage(app *model.App, demoIndex int) *FilePickerPage {
	var startDir string

	switch demoIndex {
	case 0: // Current directory
		startDir, _ = os.Getwd()
	case 1: // Home directory
		startDir, _ = os.UserHomeDir()
	case 2: // Temp directory
		startDir = os.TempDir()
	default:
		startDir, _ = os.Getwd()
	}

	if startDir == "" {
		startDir = "."
	}

	picker := model.NewFilePicker(startDir)
	picker.Focus()

	return &FilePickerPage{
		app:       app,
		demoIndex: demoIndex,
		picker:    picker,
	}
}

func (p *FilePickerPage) IgnoreQuitKeyMsg(msg tea.KeyMsg) bool {
	return false
}

func (p *FilePickerPage) Type() model.PageType {
	return model.PtMain
}

func (p *FilePickerPage) Msg() tea.Msg {
	return nil
}

func (p *FilePickerPage) Update(msg tea.Msg, app *model.App) (model.Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "b":
			// Return to main menu
			return app.Main(), nil
		case "q", "ctrl+c":
			return p, tea.Quit

		case "H": // capital H to toggle hidden (shift+h)
			p.showHidden = !p.showHidden
			p.picker.SetShowHidden(p.showHidden)
			if p.showHidden {
				p.lastMessage = "Hidden files shown"
			} else {
				p.lastMessage = "Hidden files hidden"
			}
			return p, nil

		case "enter":
			// Check if we selected a directory or file
			path, isDir := p.picker.Selected()
			if !isDir && path != "" {
				p.lastMessage = fmt.Sprintf("Selected: %s", filepath.Base(path))
			} else {
				// Clear message when entering a directory
				p.lastMessage = ""
			}
		}

	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.picker.SetSize(msg.Width-4, msg.Height-15)
		return p, nil
	}

	// Delegate to filepicker widget
	p.picker.Update(msg)

	return p, nil
}

func (p *FilePickerPage) View(app *model.App) string {
	w, h := app.WindowWidth(), app.WindowHeight()
	if w <= 0 || h <= 0 {
		return ""
	}

	// Set widget size based on current window dimensions
	p.picker.SetSize(w-4, h-12)

	ss := style.CurrentStyleSet()

	// Header
	var title string
	switch p.demoIndex {
	case 0:
		title = "FilePicker Demo - Current Directory"
	case 1:
		title = "FilePicker Demo - Home Directory"
	case 2:
		title = "FilePicker Demo - Temp Directory"
	}

	header := ss.Title.Render(title)
	instructions := lipgloss.JoinVertical(
		lipgloss.Left,
		ss.Muted.Render("up/down: navigate • enter: select/enter directory • backspace/left: parent"),
		ss.Muted.Render("shift+h: toggle hidden files • b/esc: back • q: quit"),
	)

	// Current directory path
	currentDir := p.picker.CurrentDir()
	pathDisplay := ss.Subtitle.Render(fmt.Sprintf("Current: %s", currentDir))

	// FilePicker content
	pickerView := p.picker.View()
	if pickerView == "" {
		pickerView = ss.Muted.Render("(empty directory)")
	}

	pickerBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ss.SelectedItem.GetForeground()).
		Padding(1, 2).
		Width(p.width - 4).
		Height(p.height - 16).
		Render(pickerView)

	// Status line
	var statusLine string
	if p.lastMessage != "" {
		statusLine = ss.Success.Render(p.lastMessage)
	} else {
		// Show selected entry info
		path, isDir := p.picker.Selected()
		if path != "" {
			if isDir {
				statusLine = ss.Info.Render(fmt.Sprintf("📁 %s", filepath.Base(path)))
			} else {
				// Try to get file size
				info, err := os.Stat(path)
				if err == nil {
					statusLine = ss.Info.Render(fmt.Sprintf("📄 %s (%s)", filepath.Base(path), formatSize(info.Size())))
				} else {
					statusLine = ss.Info.Render(fmt.Sprintf("📄 %s", filepath.Base(path)))
				}
			}
		}
	}

	// Hidden files indicator
	hiddenStatus := ""
	if p.showHidden {
		hiddenStatus = ss.Muted.Render(" • showing hidden files")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		instructions,
		"",
		pathDisplay,
		pickerBox,
		lipgloss.JoinHorizontal(lipgloss.Left, statusLine, hiddenStatus),
	)
}

// formatSize formats a file size in bytes to a human-readable string
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func main() {
	ops := model.DefaultOptions()
	app := model.NewApp(ops)

	ops.MainMenu = NewMainMenu()

	fmt.Println(app.Run())
}
