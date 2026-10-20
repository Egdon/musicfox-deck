package model

import (
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
