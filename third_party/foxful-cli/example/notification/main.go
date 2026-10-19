package main

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
)

// DemoMenu is the notification demo menu. Each item triggers a different
// notification variant via the Action callback.
type DemoMenu struct {
	model.DefaultMenu
	menus []model.MenuItem
}

func NewDemoMenu() *DemoMenu {
	return &DemoMenu{
		menus: []model.MenuItem{
			{Title: "Info", Subtitle: "Auto-dismiss in 4s"},
			{Title: "Success", Subtitle: "Auto-dismiss in 4s"},
			{Title: "Warning", Subtitle: "Manual dismiss"},
			{Title: "Error", Subtitle: "Manual dismiss"},
			{Title: "Long Message", Subtitle: "Truncated to 5 lines"},
			{Title: "Progress Update", Subtitle: "Shows update flow"},
			{Title: "Spam 10", Subtitle: "Test stacking limit"},
			{Title: "Action Buttons", Subtitle: "Retry or inspect a failed operation"},
			{Title: "Clear All", Subtitle: "Dismiss all notifications"},
		},
	}
}

func (m *DemoMenu) IsSearchable() bool                     { return true }
func (m *DemoMenu) GetMenuKey() string                     { return "notification_demo" }
func (m *DemoMenu) MenuViews() []model.MenuItem            { return m.menus }
func (m *DemoMenu) SubMenu(_ *model.App, _ int) model.Menu { return nil }

func (m *DemoMenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	switch index {
	case 0:
		app.Notify(model.NotificationSpec{
			Level:   model.NotificationInfo,
			Title:   "Info",
			Message: "This is an informational message that will auto-dismiss in 4 seconds.",
		})
	case 1:
		app.Notify(model.NotificationSpec{
			Level:   model.NotificationSuccess,
			Title:   "Success",
			Message: "Operation completed successfully! This will auto-dismiss.",
		})
	case 2:
		app.Notify(model.NotificationSpec{
			Level:   model.NotificationWarning,
			Title:   "Warning",
			Message: "This is a warning that requires manual dismissal. Click ✕ or clear all.",
			Timeout: 0,
		})
	case 3:
		app.Notify(model.NotificationSpec{
			Level:   model.NotificationError,
			Title:   "Error",
			Message: "An error occurred. This notification persists until dismissed.",
			Timeout: 0,
		})
	case 4:
		longMsg := `This is a very long message to demonstrate text wrapping and truncation.
Line 2: The notification system automatically wraps text to fit the max width.
Line 3: If the message exceeds the configured max lines (default 5), it will be truncated.
Line 4: Additional lines beyond the limit will not be visible.
Line 5: This should be the last visible line with an ellipsis.
Line 6: This line should be hidden.
Line 7: This line should also be hidden.`
		app.Notify(model.NotificationSpec{
			Level:   model.NotificationInfo,
			Title:   "Long Message",
			Message: longMsg,
			Timeout: 10 * time.Second,
		})
	case 5:
		// Progress update demo: shows how to live-update a notification.
		id := app.Notify(model.NotificationSpec{
			Level:   model.NotificationInfo,
			Title:   "Downloading",
			Message: "Progress: 0%",
			Timeout: 0,
		})
		go func() {
			for i := 1; i <= 10; i++ {
				time.Sleep(500 * time.Millisecond)
				app.UpdateNotification(id, model.NotificationSpec{
					Level:   model.NotificationInfo,
					Title:   "Downloading",
					Message: fmt.Sprintf("Progress: %d%%", i*10),
					Timeout: 0,
				})
			}
			time.Sleep(500 * time.Millisecond)
			app.UpdateNotification(id, model.NotificationSpec{
				Level:   model.NotificationSuccess,
				Title:   "Complete",
				Message: "Download finished successfully!",
				Timeout: 3 * time.Second,
			})
		}()
	case 6:
		for i := 1; i <= 10; i++ {
			app.Notify(model.NotificationSpec{
				Level:   model.NotificationInfo,
				Message: fmt.Sprintf("Notification #%d", i),
				Timeout: 0,
			})
		}
	case 7:
		app.Notify(model.NotificationSpec{
			Level:   model.NotificationError,
			Title:   "Deployment Failed",
			Message: "gateway-42 failed its health check.",
			Actions: []model.NotificationAction{
				{ID: "retry", Label: "Retry"},
				{ID: "details", Label: "Details"},
			},
			OnAction: func(result model.NotificationActionResult) {
				app.Notify(model.NotificationSpec{
					Level:   model.NotificationInfo,
					Title:   "Action Selected",
					Message: fmt.Sprintf("Notification %d: %s", result.NotificationID, result.ActionID),
					Timeout: 3 * time.Second,
				})
			},
		})
	case 8:
		app.ClearAllNotifications()
	}
	return nil, app.RerenderCmd(true)
}

func main() {
	opts := model.DefaultOptions()
	opts.AppName = "Notification Demo"
	opts.StatusBar = &model.DefaultStatusBar{}
	opts.NotificationOptions.DefaultTimeout = 4 * time.Second
	opts.NotificationOptions.MaxLines = 5

	// Show a welcome notification on startup.
	opts.InitHook = func(a *model.App) {
		a.Notify(model.NotificationSpec{
			Level:   model.NotificationInfo,
			Title:   "Welcome",
			Message: "Use arrow keys and Enter to trigger notifications from the menu.",
			Timeout: 5 * time.Second,
		})
	}

	app := model.NewApp(opts)
	app.With(model.WithMainMenu(NewDemoMenu(), nil))

	fmt.Println(app.Run())
}
