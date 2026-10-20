package ui

import (
	"testing"

	"github.com/anhoder/foxful-cli/model"
)

// Menus that only fetch their own data load in the background; the ones
// whose hooks touch the interface load on the UI goroutine.
func TestWhichMenusLoadInTheBackground(t *testing.T) {
	for _, c := range []struct {
		name string
		menu model.Menu
		want bool
	}{
		{"playlist", &PlaylistDetailMenu{}, true},
		{"ranks", &RanksMenu{}, true},
		{"user playlists", &UserPlaylistMenu{}, true},
		{"to page", &MenuToPage{}, false},
		{"search result", &SearchResultMenu{}, false},
	} {
		l, ok := c.menu.(model.BackgroundLoader)
		if !ok || l.LoadsInBackground() != c.want {
			t.Errorf("%s: loads in the background %v, want %v", c.name, ok && l.LoadsInBackground(), c.want)
		}
	}
}
