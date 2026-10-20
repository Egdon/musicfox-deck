package model

import (
	tea "charm.land/bubbletea/v2"
)

// Background menu loading: a menu whose BeforeEnterMenuHook only fetches its
// own data says so, and Main runs that hook off the UI goroutine. The
// interface keeps drawing and taking keys while the data loads, the header
// already names the menu, and going back cancels the load.

// BackgroundLoader is implemented by menus whose BeforeEnterMenuHook only
// fetches data for themselves. Such a hook runs off the UI goroutine, so it
// must not touch the interface; a page it returns is shown, on the UI
// goroutine, once it is done.
type BackgroundLoader interface {
	LoadsInBackground() bool
}

func loadsInBackground(menu Menu) bool {
	l, ok := menu.(BackgroundLoader)
	return ok && l.LoadsInBackground()
}

// menuLoadedMsg reports a BeforeEnterMenuHook that ran in the background.
type menuLoadedMsg struct {
	load *enterMenuDeferred
	ok   bool
	page Page
}

// post delivers a message to the App from any goroutine, and reports
// whether it can.
func (a *App) post(msg tea.Msg) bool {
	switch {
	case a.postHook != nil:
		a.postHook(msg)
	case a.program != nil:
		go a.program.Send(msg)
	default:
		return false
	}
	return true
}

func (a *App) canPost() bool { return a.postHook != nil || a.program != nil }

// runPendingEnterMenu enters the menu whose loading state the last frame
// showed, running its hook in the background when the menu allows it.
func (m *Main) runPendingEnterMenu(a *App) (Page, tea.Cmd) {
	p := m.pendingEnterMenu
	if p.running {
		return m, nil
	}
	hook := p.newMenu.BeforeEnterMenuHook()
	if hook != nil && loadsInBackground(p.newMenu) && a.canPost() {
		p.running = true
		go func() {
			ok, page := false, Page(nil)
			defer func() {
				if recover() != nil { // a failing hook leaves the menu unopened
					ok, page = false, nil
				}
				a.post(menuLoadedMsg{load: p, ok: ok, page: page})
			}()
			ok, page = hook(m)
		}()
		return m, nil
	}
	m.pendingEnterMenu = nil
	ok, page := true, Page(nil)
	if hook != nil {
		ok, page = hook(m)
	}
	return m.finishEnterMenu(a, p, ok, page)
}

// menuLoaded finishes a load that ran in the background, unless it was
// cancelled meanwhile.
func (m *Main) menuLoaded(a *App, msg menuLoadedMsg) (Page, tea.Cmd) {
	if m.pendingEnterMenu != msg.load {
		return nil, nil
	}
	m.pendingEnterMenu = nil
	return m.finishEnterMenu(a, msg.load, msg.ok, msg.page)
}

// finishEnterMenu enters a menu once its hook has run, or stays where it is
// when the hook declined, showing the page the hook returned if any.
func (m *Main) finishEnterMenu(a *App, p *enterMenuDeferred, ok bool, newPage Page) (Page, tea.Cmd) {
	p.loading.Recover()
	if !ok {
		m.menuStack.Pop()
		if newPage != nil {
			return newPage, func() tea.Msg { return newPage.Msg() }
		}
		return m, a.RerenderCmd(true)
	}
	if p.newMenu != nil {
		p.newMenu.FormatMenuItem(p.newTitle)
	}
	m.hoveredMenuItemIdx = -1
	m.menu = p.newMenu
	m.menuList = p.newMenu.MenuViews()
	m.menuTitle = p.newTitle
	m.selectedIndex = 0
	m.menuCurPage = 1
	if newPage != nil {
		return newPage, func() tea.Msg { return newPage.Msg() }
	}
	return m, a.RerenderCmd(true)
}

// cancelMenuLoad drops a menu that is about to load or loading, as going
// back does, and reports whether there was one. A hook still running in the
// background finishes unseen.
func (m *Main) cancelMenuLoad() bool {
	p := m.pendingEnterMenu
	if p == nil {
		return false
	}
	m.pendingEnterMenu = nil
	if top, ok := m.menuStack.Peek().(*menuStackItem); ok && top == p.stackItem {
		m.menuStack.Pop()
	}
	p.loading.Recover()
	return true
}
