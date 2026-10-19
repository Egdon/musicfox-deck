package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"github.com/anhoder/foxful-cli/model"
)

var (
	mainMenu = NewMainMenu()
	subMenu  = NewSubMenu()
)

type MainMenu struct {
	model.DefaultMenu
	menus []model.MenuItem
}

func NewMainMenu() *MainMenu {
	m := &MainMenu{}
	m.menus = []model.MenuItem{
		{Title: "Main Menu 1", Subtitle: "Sub Title"},
		{Title: "Main Menu 2", Subtitle: "very long long long long long long long long long long long long subtitle"},
		{Title: "Main Menu 3"},
		{Title: "Main Menu 4"},
		{Title: "Main Menu 5"},
		{Title: "Press 'm' for Markdown preview"},
		{Title: "Press 'r' for Resizable Markdown demo"},
	}
	return m
}

func (m *MainMenu) IsSearchable() bool          { return true }
func (m *MainMenu) GetMenuKey() string          { return "main_menu" }
func (m *MainMenu) MenuViews() []model.MenuItem { return m.menus }

func (m *MainMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.menus) {
		return nil
	}
	return subMenu
}

type SubMenu struct {
	model.DefaultMenu
	menus []model.MenuItem
}

func NewSubMenu() *SubMenu {
	return &SubMenu{
		menus: []model.MenuItem{
			{Title: "Sub Menu 1"},
			{Title: "Sub Menu 2"},
			{Title: "Sub Menu 3"},
		},
	}
}

func (m *SubMenu) GetMenuKey() string                     { return "sub_menu" }
func (m *SubMenu) MenuViews() []model.MenuItem            { return m.menus }
func (m *SubMenu) SubMenu(_ *model.App, _ int) model.Menu { return nil }

const markdownContent = `
# Markdown Component Demo

This is a **markdown** rendering demo powered by [Glamour](https://github.com/charmbracelet/glamour).

---

## Text Formatting

You can use **bold**, *italic*, ~~strikethrough~~, and ` + "`" + `inline code` + "`" + `.

---

## Code Block

` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("Hello, Foxful!")
}
` + "```" + `

---

## Lists

### Unordered

- Item one
- Item two
  - Nested item
  - Another nested item
- Item three

### Ordered

1. First item
2. Second item
3. Third item

---

## Task List

- [x] Implement MarkdownComponent
- [x] Add Glamour dependency
- [ ] Publish to go-musicfox

---

## Blockquote

> Foxful CLI is a modular TUI component library for building interactive terminal applications.

---

## Table

| Feature  | Status | Priority |
|----------|--------|----------|
| Heading  | done   | high     |
| Code     | done   | high     |
| Table    | done   | medium   |
| Emoji    | todo   | low      |

---

> Tip: Press Esc or click outside the popup to dismiss.
`

const resizeInstructions = `

---

## 🎉 New Feature: Popup Resize!

This popup is **resizable**! Try the following:

### How to Resize
1. Look at the **bottom-right corner** of this popup
2. You'll see a **◢** indicator (resize handle)
3. **Click and drag** the ◢ to resize the popup
4. Your mouse cursor will change to **↘** (diagonal arrow)

### Features
- Resize to see more or less content at once
- Works seamlessly with scrolling
- Minimum size protection ensures buttons stay visible
- All content remains readable during resize

### Keyboard Shortcuts
- **m** - Open standard Markdown preview
- **r** - Open this resizable demo (you're here!)
- **Esc** - Close popup
- **Mouse wheel** - Scroll content

### Try It Now!
**Drag the bottom-right corner ◢ to make this popup bigger or smaller.**

*This is a Phase 1 MVP feature - more corner handles coming in Phase 2!*
`

// MarkdownController opens a markdown preview popup when the user presses 'm'.
type MarkdownController struct {
	renderer  *glamour.TermRenderer
	lastWidth int
}

func NewMarkdownController() *MarkdownController {
	return &MarkdownController{}
}

func (c *MarkdownController) KeyMsgHandle(msg tea.KeyMsg, a *model.App) (bool, model.Page, tea.Cmd) {
	if msg.String() != "m" && msg.String() != "r" {
		return false, nil, nil
	}

	w := a.WindowWidth()
	h := a.WindowHeight()
	popupWidth := w * 70 / 100
	if popupWidth < 40 {
		popupWidth = 40
	}
	if popupWidth > 120 {
		popupWidth = 120
	}

	maxH := h * 80 / 100
	if maxH < 10 {
		maxH = 10
	}

	// Use new NewMarkdownPopup API for cleaner code
	var popup *model.Popup
	var err error

	if msg.String() == "r" {
		// Demo: Resizable Markdown popup with auto-detected style
		popup, err = model.NewMarkdownPopup(model.MarkdownPopupSpec{
			Title:              "Resizable Markdown Preview",
			MarkdownContent:    markdownContent + resizeInstructions,
			MarkdownDarkStyle:  "dracula",
			MarkdownLightStyle: "light",
			MarkdownEmoji:      true,
			MaxWidth:           popupWidth,
			MaxHeight:          maxH,
			CloseKeys:          []string{"esc", "q"},
			Actions: []model.PopupAction{
				{ID: "ok", Label: "OK"},
			},
		})
	} else {
		// Original: Standard Markdown popup with auto-detected style
		popup, err = model.NewMarkdownPopup(model.MarkdownPopupSpec{
			Title:              "Markdown Preview",
			MarkdownContent:    markdownContent,
			MarkdownDarkStyle:  "dracula",
			MarkdownLightStyle: "light",
			MarkdownEmoji:      true,
			MaxWidth:           popupWidth,
			MaxHeight:          maxH,
			DisableResize:      true,
			CloseKeys:          []string{"esc", "q"},
			Actions: []model.PopupAction{
				{ID: "ok", Label: "OK"},
			},
		})
	}

	if err != nil {
		return false, nil, nil
	}
	a.ShowPopup(popup)
	return true, nil, a.RerenderCmd(true)
}

func main() {
	ops := model.DefaultOptions()
	ops.MainMenu = mainMenu
	ops.AppName = "Markdown Demo"
	ops.KBControllers = []model.KeyboardController{NewMarkdownController()}

	app := model.NewApp(ops)
	fmt.Println(app.Run())
}
