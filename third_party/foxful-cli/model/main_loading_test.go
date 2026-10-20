package model

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// loaderMenu fetches in its hook until released.
type loaderMenu struct {
	mockMenu
	background bool
	release    chan struct{}
	page       Page // returned with false, as a hook that needs a login does
}

func (m *loaderMenu) LoadsInBackground() bool { return m.background }

func (m *loaderMenu) BeforeEnterMenuHook() Hook {
	return func(*Main) (bool, Page) {
		if m.release != nil {
			<-m.release
		}
		if m.page != nil {
			return false, m.page
		}
		return true, nil
	}
}

func newLoadingMain(t *testing.T) (*App, *Main, chan tea.Msg) {
	t.Helper()
	options := DefaultOptions()
	options.MainMenu = &mockMenu{key: "root", items: []MenuItem{{Title: "Alpha"}}}
	options.MainMenuTitle = &MenuItem{Title: "Root"}
	app := NewApp(options)
	app.windowWidth, app.windowHeight = 80, 24
	posted := make(chan tea.Msg, 4)
	app.postHook = func(msg tea.Msg) { posted <- msg }
	main := NewMain(app, options)
	app.main, app.page = main, main
	_, _ = main.Update(tea.WindowSizeMsg{Width: 80, Height: 24}, app)
	return app, main, posted
}

// tickAway runs the pending tick and fails if it waits for the hook.
func tickAway(t *testing.T, app *App, main *Main, sub *loaderMenu) {
	t.Helper()
	done := make(chan struct{})
	go func() { _, _ = main.Update(tickMainMsg{}, app); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		close(sub.release)
		t.Fatal("the interface waited for the menu's hook")
	}
}

func TestMenuLoadsInTheBackground(t *testing.T) {
	app, main, posted := newLoadingMain(t)
	sub := &loaderMenu{mockMenu: mockMenu{key: "sub", items: []MenuItem{{Title: "Inner"}}}, background: true, release: make(chan struct{})}
	main.enterMenuWithLoading(sub, &MenuItem{Title: "Sub"})
	tickAway(t, app, main, sub)
	if main.menu == Menu(sub) || main.pendingEnterMenu == nil || main.loadingTips == "" {
		t.Fatal("while it loads, the old menu shows with the loading tips")
	}
	_ = main.View(app)

	close(sub.release)
	app.Update(<-posted)
	if main.menu != Menu(sub) || main.pendingEnterMenu != nil || main.menuStack.Len() != 1 || main.loadingTips != "" {
		t.Fatalf("once loaded the menu is entered: menu %v, pending %v, depth %d, tips %q",
			main.menu.GetMenuKey(), main.pendingEnterMenu, main.menuStack.Len(), main.loadingTips)
	}
}

func TestBackCancelsALoadingMenu(t *testing.T) {
	app, main, posted := newLoadingMain(t)
	sub := &loaderMenu{mockMenu: mockMenu{key: "sub"}, background: true, release: make(chan struct{})}
	main.enterMenuWithLoading(sub, &MenuItem{Title: "Sub"})
	tickAway(t, app, main, sub)

	main.BackMenu()
	if main.pendingEnterMenu != nil || main.menuStack.Len() != 0 || main.loadingTips != "" {
		t.Fatalf("going back drops the load: depth %d, tips %q", main.menuStack.Len(), main.loadingTips)
	}
	close(sub.release)
	app.Update(<-posted)
	if main.menu == Menu(sub) || main.menuStack.Len() != 0 {
		t.Fatal("a cancelled load must not enter its menu")
	}
}

func TestForegroundMenuLoadsAtOnce(t *testing.T) {
	app, main, posted := newLoadingMain(t)
	sub := &loaderMenu{mockMenu: mockMenu{key: "sub"}}
	main.enterMenuWithLoading(sub, &MenuItem{Title: "Sub"})
	_, _ = main.Update(tickMainMsg{}, app)
	if main.menu != Menu(sub) || main.pendingEnterMenu != nil {
		t.Fatal("a menu that does not load in the background enters in the tick")
	}
	select {
	case msg := <-posted:
		t.Fatalf("unexpected %T", msg)
	default:
	}
}

func TestBackgroundLoadCanShowAPage(t *testing.T) {
	app, main, posted := newLoadingMain(t)
	login := &framePage{text: "LOGIN"}
	sub := &loaderMenu{mockMenu: mockMenu{key: "sub"}, background: true, release: make(chan struct{}), page: login}
	main.enterMenuWithLoading(sub, &MenuItem{Title: "Sub"})
	tickAway(t, app, main, sub)
	close(sub.release)
	app.Update(<-posted)
	if app.page != Page(login) || main.menuStack.Len() != 0 || main.loadingTips != "" {
		t.Fatalf("the hook's page shows and the menu stays: page %T, depth %d", app.page, main.menuStack.Len())
	}
}

// pagerMenu adds an item each time its list runs out, fetching until released.
type pagerMenu struct {
	mockMenu
	release chan struct{}
	fetches atomic.Int32
}

func (m *pagerMenu) BottomOutFetch() Fetch {
	next := MenuItem{Title: fmt.Sprintf("Item %d", len(m.items))}
	return func() Hook {
		m.fetches.Add(1)
		if m.release != nil {
			<-m.release
		}
		return func(*Main) (bool, Page) {
			m.items = append(m.items, next)
			return true, nil
		}
	}
}

func (m *pagerMenu) BottomOutHook() Hook { return HookOf(m.BottomOutFetch()) }

func newPager(main *Main, release chan struct{}) *pagerMenu {
	pager := &pagerMenu{mockMenu: mockMenu{key: "pager", items: []MenuItem{{Title: "Item 0"}, {Title: "Item 1"}}}, release: release}
	main.EnterMenu(pager, &MenuItem{Title: "Pager"})
	main.isDualColumn = false
	main.selectedIndex = 1 // the bottom
	return pager
}

// press sends a key and fails if handling it waits for a fetch.
func press(t *testing.T, app *App, main *Main, code rune, release chan struct{}) {
	t.Helper()
	done := make(chan struct{})
	go func() { _, _ = main.Update(tea.KeyPressMsg{Code: code}, app); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		close(release)
		t.Fatal("the interface waited for the fetch")
	}
}

func TestMovingPastTheBottomPagesInTheBackground(t *testing.T) {
	app, main, posted := newLoadingMain(t)
	release := make(chan struct{})
	pager := newPager(main, release)

	press(t, app, main, tea.KeyDown, release)
	if main.selectedIndex != 1 || main.pendingPage == nil || main.loadingTips == "" {
		t.Fatal("while more items load, the selection waits at the bottom with the loading tips")
	}
	_ = main.View(app)
	press(t, app, main, tea.KeyDown, release) // waits as well, without a second fetch

	close(release)
	app.Update(<-posted)
	if len(main.menuList) != 3 || main.selectedIndex != 2 || main.pendingPage != nil || main.loadingTips != "" {
		t.Fatalf("once loaded the move finishes: %d items, selected %d, tips %q", len(main.menuList), main.selectedIndex, main.loadingTips)
	}
	if n := pager.fetches.Load(); n != 1 {
		t.Fatalf("%d fetches, want 1", n)
	}
}

// Callers of MoveDown, such as a player going on to the next track, still
// get the next items at once.
func TestMoveDownStillPagesAtOnce(t *testing.T) {
	_, main, posted := newLoadingMain(t)
	newPager(main, nil)
	main.MoveDown()
	if len(main.menuList) != 3 || main.selectedIndex != 2 {
		t.Fatalf("%d items, selected %d", len(main.menuList), main.selectedIndex)
	}
	select {
	case msg := <-posted:
		t.Fatalf("unexpected %T", msg)
	default:
	}
}

func TestLeavingDropsAPageLoad(t *testing.T) {
	app, main, posted := newLoadingMain(t)
	release := make(chan struct{})
	pager := newPager(main, release)
	press(t, app, main, tea.KeyDown, release)

	main.BackMenu()
	if main.menu == Menu(pager) || main.pendingPage != nil || main.loadingTips != "" {
		t.Fatal("going back leaves the menu and drops the load")
	}
	close(release)
	app.Update(<-posted)
	if len(pager.items) != 2 || main.menu == Menu(pager) {
		t.Fatal("a dropped load must not add its items")
	}
}

func TestEnterMenuLoadsInTheBackground(t *testing.T) {
	app, main, posted := newLoadingMain(t)
	sub := &loaderMenu{mockMenu: mockMenu{key: "sub", items: []MenuItem{{Title: "Inner"}}}, background: true, release: make(chan struct{})}
	done := make(chan struct{})
	go func() { main.EnterMenu(sub, &MenuItem{Title: "Sub"}); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		close(sub.release)
		t.Fatal("EnterMenu waited for the menu's hook")
	}
	if main.menu == Menu(sub) || main.pendingEnterMenu == nil || main.loadingTips == "" {
		t.Fatal("while it loads, the old menu shows with the loading tips")
	}
	close(sub.release)
	app.Update(<-posted)
	if main.menu != Menu(sub) || main.menuStack.Len() != 1 || main.loadingTips != "" {
		t.Fatal("once loaded the menu is entered")
	}
}

func TestLoadInBackground(t *testing.T) {
	app, main, posted := newLoadingMain(t)
	release, applied := make(chan struct{}), false
	if !main.LoadInBackground(func() Hook {
		<-release
		return func(*Main) (bool, Page) { applied = true; return true, nil }
	}) {
		t.Fatal("LoadInBackground declined")
	}
	if main.loadingTips == "" {
		t.Fatal("the loading tips show while it runs")
	}
	close(release)
	app.Update(<-posted)
	if !applied || main.loadingTips != "" {
		t.Fatal("what it fetched applies on the UI goroutine")
	}

	login := &framePage{text: "LOGIN"}
	main.LoadInBackground(func() Hook { return func(*Main) (bool, Page) { return false, login } })
	app.Update(<-posted)
	if app.page != Page(login) {
		t.Fatalf("a page returned with false shows: %T", app.page)
	}
}

// parentMenu makes its submenu afresh each time.
type parentMenu struct {
	mockMenu
	sub func() Menu
}

func (m *parentMenu) SubMenu(*App, int) Menu { return m.sub() }

func TestRefreshLoadsAFreshCopyInTheBackground(t *testing.T) {
	app, main, posted := newLoadingMain(t)
	release := make(chan struct{})
	made := 0
	parent := &parentMenu{mockMenu: mockMenu{key: "parent", items: []MenuItem{{Title: "Sub"}}}, sub: func() Menu {
		made++
		items := []MenuItem{{Title: "One"}, {Title: "Two"}, {Title: fmt.Sprintf("Copy %d", made)}}
		return &loaderMenu{mockMenu: mockMenu{key: "sub", items: items}, background: true, release: release}
	}}
	main.menu, main.menuList = parent, parent.MenuViews()
	main.enterMenuWithLoading(nil, nil)
	_, _ = main.Update(tickMainMsg{}, app)
	close(release)
	app.Update(<-posted)
	shown := main.menu
	main.selectedIndex = 2

	release = make(chan struct{})
	main.RefreshMenuWithLoading()
	if main.menu != shown || main.loadingTips == "" {
		t.Fatal("while the copy loads, the menu shown stays with the loading tips")
	}
	close(release)
	app.Update(<-posted)
	if main.menu == shown || main.menu.GetMenuKey() != "sub" || main.menuList[2].Title != "Copy 2" || main.selectedIndex != 2 || main.loadingTips != "" {
		t.Fatalf("the copy replaces it and the selection stays: %v, selected %d", main.menuList, main.selectedIndex)
	}
}
