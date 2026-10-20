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

// EnterFetcher is implemented by background loaders whose hook also changes
// state the interface reads: their BeforeEnterMenuHook is
// HookOf(BeforeEnterMenuFetch()), and only the fetch runs in the background
// while what it fetched applies on the UI goroutine.
type EnterFetcher interface {
	BeforeEnterMenuFetch() Fetch
}

// menuLoadedMsg reports a BeforeEnterMenuHook that ran in the background, or
// the fetch of an EnterFetcher with what applies it.
type menuLoadedMsg struct {
	load  *enterMenuDeferred
	ok    bool
	page  Page
	apply Hook
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
		var fetch Fetch
		if f, ok := p.newMenu.(EnterFetcher); ok {
			fetch = f.BeforeEnterMenuFetch()
		}
		go func() {
			msg := menuLoadedMsg{load: p}
			defer func() {
				if recover() != nil { // a failing hook leaves the menu unopened
					msg = menuLoadedMsg{load: p}
				}
				a.post(msg)
			}()
			if fetch != nil {
				msg.apply = fetch()
				return
			}
			msg.ok, msg.page = hook(m)
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
	if msg.apply != nil {
		msg.ok, msg.page = msg.apply(m)
	}
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
	m.cancelPageLoad()
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

// Background fetches: work that waits on the network runs off the UI
// goroutine and hands back a Hook that applies what it fetched on the UI
// goroutine. Moving past the bottom of a BackgroundPager loads its next items
// that way, LoadInBackground runs any other such work, and refreshing a menu
// that loads in the background loads a fresh copy of it.

// A Fetch runs off the UI goroutine. It may wait on the network but must not
// touch the interface or state the interface reads; it returns the Hook that
// applies what it fetched, on the UI goroutine.
type Fetch func() Hook

// HookOf is a Hook that runs fetch and applies what it fetched at once, for
// callers that load synchronously. It is nil when fetch is.
func HookOf(fetch Fetch) Hook {
	if fetch == nil {
		return nil
	}
	return func(m *Main) (bool, Page) {
		if apply := fetch(); apply != nil {
			return apply(m)
		}
		return true, nil
	}
}

// BackgroundPager is implemented by menus whose BottomOutHook is
// HookOf(BottomOutFetch()). Moving past the bottom with the keys or the wheel
// then loads the next items in the background, and the move finishes once
// they arrive. BottomOutHook itself still loads at once, for callers such as a
// player that needs the next items before it goes on.
type BackgroundPager interface {
	BottomOutFetch() Fetch
}

// pageLoad is a background load of more items for a move past the bottom.
type pageLoad struct {
	menu     Menu
	loading  *Loading
	selected int  // where the move started
	right    bool // a move right in two columns, else down
}

// pageLoadedMsg reports a page load that ran in the background.
type pageLoadedMsg struct {
	load  *pageLoad
	apply Hook
}

// moveDown is MoveDown for the keys and the wheel: past the bottom of a
// BackgroundPager it loads the next items in the background.
func (m *Main) moveDown(a *App) Page {
	step := 1
	if m.isDualColumn {
		step = 2
	}
	if m.selectedIndex >= 0 && m.selectedIndex+step > len(m.menuList)-1 && m.pageInBackground(a, false) {
		return nil
	}
	return m.MoveDown()
}

// moveRight is MoveRight for the keys, loading in the background as moveDown.
func (m *Main) moveRight(a *App) Page {
	if m.isDualColumn && m.selectedIndex%2 == 0 && m.selectedIndex >= len(m.menuList)-1 && m.pageInBackground(a, true) {
		return nil
	}
	return m.MoveRight()
}

// pageInBackground starts loading the next items of a BackgroundPager for a
// move past the bottom, and reports whether the move waits for them. A move
// past the bottom while they load waits as well.
func (m *Main) pageInBackground(a *App, right bool) bool {
	if m.pendingPage != nil {
		return true
	}
	pager, ok := m.menu.(BackgroundPager)
	if !ok || !a.canPost() {
		return false
	}
	fetch := pager.BottomOutFetch()
	if fetch == nil {
		return false
	}
	p := &pageLoad{menu: m.menu, loading: NewLoading(m), selected: m.selectedIndex, right: right}
	p.loading.Start()
	m.pendingPage = p
	go func() {
		var apply Hook
		defer func() {
			if recover() != nil { // a failing fetch loads nothing
				apply = nil
			}
			a.post(pageLoadedMsg{load: p, apply: apply})
		}()
		apply = fetch()
	}()
	return true
}

// pageLoaded adds the items a page load fetched and finishes the move that
// waited for them, unless the menu changed meanwhile.
func (m *Main) pageLoaded(a *App, msg pageLoadedMsg) (Page, tea.Cmd) {
	p := msg.load
	if m.pendingPage != p {
		return nil, nil
	}
	m.pendingPage = nil
	p.loading.Complete()
	if msg.apply == nil || m.menu != p.menu {
		return m, a.RerenderCmd(true)
	}
	ok, newPage := msg.apply(m)
	if ok {
		m.menuList = m.menu.MenuViews()
		if m.selectedIndex == p.selected { // the selection waited at the bottom
			if p.right {
				newPage = m.stepRight(newPage)
			} else {
				newPage = m.stepDown(newPage)
			}
		}
	}
	if newPage != nil {
		return newPage, func() tea.Msg { return newPage.Msg() }
	}
	return m, a.RerenderCmd(true)
}

// cancelPageLoad drops a page load in flight, as leaving its menu does; the
// fetch finishes unseen.
func (m *Main) cancelPageLoad() {
	if p := m.pendingPage; p != nil {
		m.pendingPage = nil
		p.loading.Recover()
	}
}

// backgroundLoad is work LoadInBackground runs.
type backgroundLoad struct {
	loading *Loading
}

// backgroundLoadedMsg reports work LoadInBackground ran.
type backgroundLoadedMsg struct {
	load  *backgroundLoad
	apply Hook
}

// LoadInBackground shows the loading state, runs fetch off the UI goroutine
// and applies what it returns on the UI goroutine; a page returned with false
// is shown. It reports false, running nothing, when the App cannot deliver
// messages, as in tests; HookOf(fetch) then runs it at once.
func (m *Main) LoadInBackground(fetch Fetch) bool {
	a := m.app
	if fetch == nil || a == nil || !a.canPost() {
		return false
	}
	l := &backgroundLoad{loading: NewLoading(m)}
	l.loading.Start()
	go func() {
		var apply Hook
		defer func() {
			if recover() != nil { // a failing fetch applies nothing
				apply = nil
			}
			a.post(backgroundLoadedMsg{load: l, apply: apply})
		}()
		apply = fetch()
	}()
	return true
}

// backgroundLoaded applies what LoadInBackground fetched.
func (m *Main) backgroundLoaded(a *App, msg backgroundLoadedMsg) (Page, tea.Cmd) {
	msg.load.loading.Complete()
	if msg.apply != nil {
		if ok, page := msg.apply(m); !ok && page != nil {
			return page, func() tea.Msg { return page.Msg() }
		}
	}
	return m, a.RerenderCmd(true)
}

// refreshInBackground reloads the current menu as a fresh copy its parent
// makes again, in the background, and reports whether it could: the parent
// must make a menu with the same key that loads in the background. The copy
// replaces the menu shown, keeping the selection, once it has loaded.
func (m *Main) refreshInBackground() bool {
	a := m.app
	parent, ok := m.menuStack.Peek().(*menuStackItem)
	if !ok || a == nil || !a.canPost() {
		return false
	}
	fresh := parent.menu.SubMenu(a, parent.selectedIndex)
	if fresh == nil || fresh.GetMenuKey() != m.menu.GetMenuKey() || !loadsInBackground(fresh) {
		return false
	}
	shown := m.menu
	hook := fresh.BeforeEnterMenuHook()
	return m.LoadInBackground(func() Hook {
		ok, page := true, Page(nil)
		if hook != nil {
			ok, page = hook(m)
		}
		return func(m *Main) (bool, Page) {
			if !ok || m.menu != shown {
				return ok, page
			}
			m.menu = fresh
			m.menuList = fresh.MenuViews()
			if m.selectedIndex >= len(m.menuList) {
				m.selectedIndex = max(len(m.menuList)-1, 0)
			}
			if m.menuPageSize > 0 {
				if pages := max((len(m.menuList)+m.menuPageSize-1)/m.menuPageSize, 1); m.menuCurPage > pages {
					m.menuCurPage = pages
				}
			}
			return true, page
		}
	})
}
