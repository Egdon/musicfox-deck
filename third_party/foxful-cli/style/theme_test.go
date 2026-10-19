package style

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
)

func TestNotificationMessageStyleLeavesForegroundUnset(t *testing.T) {
	styles := NewStyleSet(DefaultDarkTheme())
	if got := styles.Notification.Message.GetForeground(); got != (lipgloss.NoColor{}) {
		t.Errorf("default notification message color = %v, want no explicit foreground", got)
	}
}

func TestNotificationActionStylesOwnTheSurfaceAndExposeHover(t *testing.T) {
	styles := NewStyleSet(DefaultDarkTheme()).Notification
	if got := styles.Action.GetBackground(); got == (lipgloss.NoColor{}) {
		t.Fatal("default notification action has no background")
	}
	if got := styles.Close.GetBackground(); got == (lipgloss.NoColor{}) {
		t.Fatal("notification close button does not own the notification surface")
	}
	if !styles.ActionHover.GetUnderline() {
		t.Fatal("notification action hover lacks a non-color cue")
	}
}

func TestDefaultSelectedItemBackgroundIsSubtler(t *testing.T) {
	tests := []struct {
		name  string
		theme Theme
		dark  bool
	}{
		{name: "dark", theme: DefaultDarkTheme(), dark: true},
		{name: "light", theme: DefaultLightTheme()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			styles := NewStyleSet(tt.theme)
			selected := mustColorful(t, styles.SelectedItem.GetBackground())
			background := mustColorful(t, tt.theme.Background)
			primary := mustColorful(t, tt.theme.Primary)

			previous := primary.BlendLab(background, 0.9).Clamped()
			if tt.dark {
				highlighted := primary.BlendLab(colorful.Color{R: 1, G: 1, B: 1}, 0.8)
				previous = highlighted.BlendLab(background, 0.7).Clamped()
			}
			if got, before := selected.DistanceLab(background), previous.DistanceLab(background); got >= before {
				t.Fatalf("selected background distance from terminal background = %.3f, want less than previous %.3f", got, before)
			}
			if selected.DistanceLab(background) == 0 {
				t.Fatal("selected background became indistinguishable from terminal background")
			}
		})
	}
}

func TestDefaultNotificationActionBackgroundIsSubtler(t *testing.T) {
	for _, theme := range []Theme{DefaultDarkTheme(), DefaultLightTheme()} {
		styles := NewStyleSet(theme)
		action := mustColorful(t, styles.Notification.Action.GetBackground())
		surface := mustColorful(t, styles.Notification.Surface)
		muted := mustColorful(t, theme.Muted)
		if got, before := action.DistanceLab(surface), muted.DistanceLab(surface); got >= before {
			t.Fatalf("notification action distance from surface = %.3f, want less than muted background %.3f", got, before)
		}
		if action.DistanceLab(surface) == 0 {
			t.Fatal("notification action background became indistinguishable from its surface")
		}
	}
}

func mustColorful(t testing.TB, value color.Color) colorful.Color {
	t.Helper()
	converted, ok := colorful.MakeColor(value)
	if !ok {
		t.Fatalf("cannot convert color %v", value)
	}
	return converted
}
