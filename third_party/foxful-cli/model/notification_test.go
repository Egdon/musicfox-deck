package model

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/style"
)

func renderNotificationForTest(spec NotificationSpec) (*App, *Notification, notificationRender) {
	app := NewApp(DefaultOptions())
	notification := &Notification{id: 7, spec: cloneNotificationSpec(spec), hoveredAction: -1}
	rendered := app.renderNotification(
		notification,
		style.NewStyleSet(style.DefaultDarkTheme()).Notification,
		40,
		5,
		20,
	)
	notification.setBounds(3, 2, 40, strings.Count(rendered.content, "\n")+1, rendered.actionBounds, rendered.actionArea)
	return app, notification, rendered
}

func TestNotificationActionsRenderWithAccurateGeometry(t *testing.T) {
	_, notification, rendered := renderNotificationForTest(NotificationSpec{
		Level:   NotificationError,
		Title:   "Failed",
		Message: "Try the operation again.",
		Actions: []NotificationAction{
			{ID: "retry", Label: "Retry"},
			{ID: "details", Label: "Details"},
		},
	})

	if !strings.Contains(rendered.content, "Retry") || !strings.Contains(rendered.content, "Details") {
		t.Fatalf("rendered notification does not contain both actions: %q", rendered.content)
	}
	if got := len(notification.contentLines); got != 2 {
		t.Fatalf("selectable content lines = %d, want title and body exactly once", got)
	}
	if got := len(notification.actionBounds); got != 2 {
		t.Fatalf("action bounds = %d, want 2", got)
	}
	for index, bound := range notification.actionBounds {
		if got := notification.actionAt(bound.x, bound.y); got != index {
			t.Fatalf("actionAt for bound %d = %d", index, got)
		}
		if notification.pointInContent(bound.x, bound.y) {
			t.Fatalf("action %d overlaps selectable content", index)
		}
	}
}

func TestNotificationActionsWrapAndTruncateWithinContentWidth(t *testing.T) {
	app := NewApp(DefaultOptions())
	styles := style.NewStyleSet(style.DefaultDarkTheme()).Notification
	notification := &Notification{
		spec: NotificationSpec{Actions: []NotificationAction{
			{ID: "first", Label: "First action"},
			{ID: "second", Label: "Second action"},
		}},
		hoveredAction: -1,
	}
	rendered := app.renderNotification(notification, styles, 24, 5, 10)
	if len(rendered.actionBounds) != 2 || rendered.actionBounds[1].y <= rendered.actionBounds[0].y {
		t.Fatalf("narrow action bounds did not wrap: %+v", rendered.actionBounds)
	}

	notification.spec.Actions = []NotificationAction{{ID: "long", Label: strings.Repeat("x", 60)}}
	rendered = app.renderNotification(notification, styles, 20, 5, 10)
	if got := rendered.actionBounds[0].w; got > 16 {
		t.Fatalf("truncated action width = %d, want <= 16", got)
	}
}

func TestNotificationActionClickDismissesThenCallsBackOnce(t *testing.T) {
	app, notification, _ := renderNotificationForTest(NotificationSpec{
		Title:   "Failed",
		Message: "Try again.",
		Actions: []NotificationAction{{ID: "retry", Label: "Retry"}},
	})
	app.notifications = []*Notification{notification}

	calls := 0
	notification.spec.OnAction = func(result NotificationActionResult) {
		calls++
		if len(app.notifications) != 0 {
			t.Fatal("action callback ran before notification removal")
		}
		if result.NotificationID != notification.id || result.ActionID != "retry" {
			t.Fatalf("action result = %+v", result)
		}
	}

	bound := notification.actionBounds[0]
	_, cmd := app.Update(tea.MouseClickMsg(tea.Mouse{X: bound.x, Y: bound.y, Button: tea.MouseLeft}))
	if cmd == nil {
		t.Fatal("action click did not request a rerender")
	}
	if calls != 1 {
		t.Fatalf("action callback calls = %d, want 1", calls)
	}
	if len(app.notifications) != 0 {
		t.Fatal("action click did not dismiss notification")
	}
}

func TestNotificationActionHoverAndBlankArea(t *testing.T) {
	_, notification, _ := renderNotificationForTest(NotificationSpec{
		Message: "Choose an action.",
		Actions: []NotificationAction{{ID: "retry", Label: "Retry"}},
	})

	bound := notification.actionBounds[0]
	hover := notification.handleMouse(tea.MouseMotionMsg(tea.Mouse{X: bound.x, Y: bound.y}))
	if !hover.consumed || !hover.rerender || notification.hoveredAction != 0 {
		t.Fatalf("hover result = %+v, hovered action = %d", hover, notification.hoveredAction)
	}

	blank := notification.handleMouse(tea.MouseClickMsg(tea.Mouse{
		X:      notification.actionArea.x,
		Y:      notification.actionArea.y,
		Button: tea.MouseLeft,
	}))
	if !blank.consumed || blank.dismiss || blank.action != nil {
		t.Fatalf("blank action-area click = %+v, want inert and consumed", blank)
	}
}

func TestNotificationActionsPersistByDefaultAndSpecsAreCopied(t *testing.T) {
	app := NewApp(DefaultOptions())
	actions := []NotificationAction{{ID: "retry", Label: "Retry"}}
	app.handleShowNotification(NotificationSpec{
		Level:   NotificationInfo,
		Message: "Failed",
		Actions: actions,
	})

	notification := app.notifications[0]
	if !notification.expireAt.IsZero() {
		t.Fatal("interactive notification received an implicit timeout")
	}
	actions[0].ID = "mutated"
	if got := notification.spec.Actions[0].ID; got != "retry" {
		t.Fatalf("stored action ID changed to %q after caller mutation", got)
	}

	notification.boundsSet = true
	notification.actionBounds = []notificationRect{{x: 1, y: 1, w: 1, h: 1}}
	notification.hoveredAction = 0
	notification.hasSelection = true
	updatedActions := []NotificationAction{{ID: "details", Label: "Details"}}
	app.updateNotificationContent(notification.id, NotificationSpec{Actions: updatedActions})
	updatedActions[0].ID = "mutated-again"

	if got := notification.spec.Actions[0].ID; got != "details" {
		t.Fatalf("updated action ID changed to %q after caller mutation", got)
	}
	if notification.boundsSet || len(notification.actionBounds) != 0 || notification.hoveredAction != -1 || notification.hasSelection {
		t.Fatal("notification update retained stale interaction geometry")
	}
}

func TestCompositeNotificationsClearsHiddenBounds(t *testing.T) {
	app := NewApp(DefaultOptions())
	app.windowWidth = 120
	app.windowHeight = 12
	app.options.NotificationOptions.MaxWidth = 40
	old := &Notification{
		id:            1,
		spec:          NotificationSpec{Message: "old"},
		hoveredAction: -1,
	}
	app.notifications = []*Notification{old}
	base := strings.TrimSuffix(strings.Repeat(strings.Repeat(" ", 120)+"\n", 12), "\n")
	app.compositeNotifications(base)
	if !old.boundsSet {
		t.Fatal("old notification was not initially visible")
	}

	newest := &Notification{
		id: 2,
		spec: NotificationSpec{
			Title:   "newest",
			Message: "one\ntwo\nthree\nfour\nfive",
		},
		hoveredAction: -1,
	}
	app.notifications = append(app.notifications, newest)
	app.compositeNotifications(base)

	if !newest.boundsSet {
		t.Fatal("newest notification was not visible")
	}
	if old.boundsSet {
		t.Fatal("hidden notification retained stale bounds")
	}
}

type notificationMouseSpyPage struct {
	motionEvents int
}

func (p *notificationMouseSpyPage) IgnoreQuitKeyMsg(tea.KeyMsg) bool { return false }
func (p *notificationMouseSpyPage) Type() PageType                   { return PageType("notification-test") }
func (p *notificationMouseSpyPage) View(*App) string                 { return "" }
func (p *notificationMouseSpyPage) Msg() tea.Msg                     { return nil }
func (p *notificationMouseSpyPage) Update(msg tea.Msg, _ *App) (Page, tea.Cmd) {
	if _, ok := msg.(tea.MouseMotionMsg); ok {
		p.motionEvents++
	}
	return p, nil
}

func TestNotificationActionHoverClearsOutsideWithoutBlockingPage(t *testing.T) {
	app, notification, _ := renderNotificationForTest(NotificationSpec{
		Message: "Choose an action.",
		Actions: []NotificationAction{{ID: "retry", Label: "Retry"}},
	})
	page := &notificationMouseSpyPage{}
	app.page = page
	app.notifications = []*Notification{notification}

	bound := notification.actionBounds[0]
	app.Update(tea.MouseMotionMsg(tea.Mouse{X: bound.x, Y: bound.y}))
	if notification.hoveredAction != 0 || page.motionEvents != 0 {
		t.Fatalf("inside hover: action=%d page events=%d", notification.hoveredAction, page.motionEvents)
	}

	_, cmd := app.Update(tea.MouseMotionMsg(tea.Mouse{X: 0, Y: 0}))
	if notification.hoveredAction != -1 {
		t.Fatalf("hovered action after leaving = %d", notification.hoveredAction)
	}
	if page.motionEvents != 1 {
		t.Fatalf("outside motion events delivered to page = %d, want 1", page.motionEvents)
	}
	if cmd == nil {
		t.Fatal("leaving hovered action did not request a rerender")
	}
}
