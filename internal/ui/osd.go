package ui

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// VCR-style on-screen display texts, shared by the now-playing page and the
// main-screen OSD toast.

func osdStateText(state types.State) string {
	switch state {
	case types.Playing:
		return "▶ PLAY"
	case types.Paused:
		return "▌▌ PAUSE"
	default:
		return "■ STOP"
	}
}

func osdVolumeText(volume int) string {
	lit := min(max((volume+9)/10, 0), 10)
	return fmt.Sprintf("VOL %s%s %d", strings.Repeat("▪", lit), strings.Repeat("▫", 10-lit), volume)
}

func osdSeekText(delta, at time.Duration) string {
	arrow := "▶▶"
	if delta < 0 {
		arrow = "◀◀"
	}
	return fmt.Sprintf("%s %+ds  %s", arrow, int(delta/time.Second), npClock(at))
}

func osdModeText(mode types.Mode) string {
	return "MODE · " + mode.String()
}

const osdToastTimeout = 1500 * time.Millisecond

// lastForeignNotify is the UnixNano time of the latest App.Notify call made for
// anything other than the OSD. App.Notify only predicts the ID it returns, so
// two notifications created close together may predict the same one.
var lastForeignNotify atomic.Int64

func markForeignNotify() {
	lastForeignNotify.Store(time.Now().UnixNano())
}

// osdNotifier keeps a single OSD toast on the main screen and updates it in
// place while it is visible. It is only used from the event loop.
type osdNotifier struct {
	id        model.NotificationID
	createdAt time.Time
	expireAt  time.Time
}

// osdCanUpdate reports whether the OSD toast created at createdAt is still
// visible and its predicted ID cannot have been taken by a notification
// created around the same time.
func osdCanUpdate(now, createdAt, expireAt, foreign time.Time) bool {
	if createdAt.IsZero() || !now.Before(expireAt) {
		return false
	}
	const guard = 500 * time.Millisecond
	return foreign.Before(createdAt.Add(-guard)) || foreign.After(createdAt.Add(guard))
}

func (o *osdNotifier) show(app *model.App, text string) {
	if !configs.AppConfig.Main.Notification.InApp || text == "" {
		return
	}
	now := time.Now()
	spec := model.NotificationSpec{Level: model.NotificationInfo, Message: text, Timeout: osdToastTimeout}
	if osdCanUpdate(now, o.createdAt, o.expireAt, time.Unix(0, lastForeignNotify.Load())) {
		app.UpdateNotification(o.id, spec)
	} else {
		o.id = app.Notify(spec)
		o.createdAt = now
	}
	o.expireAt = now.Add(osdToastTimeout)
}
