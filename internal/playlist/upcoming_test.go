package playlist

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func upcomingTestSongs(n int) []structs.Song {
	songs := make([]structs.Song, n)
	for i := range songs {
		songs[i] = structs.Song{Id: int64(i + 1)}
	}
	return songs
}

func songIDs(songs []structs.Song) []int64 {
	ids := make([]int64, len(songs))
	for i, s := range songs {
		ids[i] = s.Id
	}
	return ids
}

func equalIDs(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func newUpcomingManager(t *testing.T, mode types.Mode, index, total int) PlaylistManager {
	t.Helper()
	m := NewPlaylistManager()
	if err := m.SetPlayMode(mode); err != nil {
		t.Fatalf("SetPlayMode(%v): %v", mode, err)
	}
	if err := m.Initialize(index, upcomingTestSongs(total)); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	return m
}

func TestUpcomingDeterministicModes(t *testing.T) {
	tests := []struct {
		name  string
		mode  types.Mode
		index int
		n     int
		want  []int64
	}{
		{"ordered", types.PmOrdered, 1, 3, []int64{3, 4, 5}},
		{"ordered at end", types.PmOrdered, 4, 3, []int64{}},
		{"intelligent", types.PmIntelligent, 2, 5, []int64{4, 5}},
		{"list loop wraps", types.PmListLoop, 3, 10, []int64{5, 1, 2, 3}},
		{"single loop repeats current", types.PmSingleLoop, 2, 3, []int64{3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newUpcomingManager(t, tt.mode, tt.index, 5)
			if got := songIDs(m.Upcoming(tt.n)); !equalIDs(got, tt.want) {
				t.Errorf("Upcoming(%d) = %v, want %v", tt.n, got, tt.want)
			}
		})
	}
}

// The preview must match what automatic playback actually does next, up to
// the end of the shuffled order, where list random stops.
func TestUpcomingListRandomMatchesNextSong(t *testing.T) {
	m := newUpcomingManager(t, types.PmListRandom, 0, 8)
	for i, want := range songIDs(m.Upcoming(8)) {
		song, err := m.NextSong(false)
		if err != nil {
			t.Fatalf("NextSong #%d: %v", i, err)
		}
		if song.Id != want {
			t.Fatalf("NextSong #%d = %d, preview said %d", i, song.Id, want)
		}
	}
	if song, err := m.NextSong(false); err == nil {
		t.Fatalf("NextSong continued past the previewed order with %d", song.Id)
	}
}

func TestUpcomingNonDeterministicOrInvalid(t *testing.T) {
	if got := newUpcomingManager(t, types.PmInfRandom, 0, 5).Upcoming(3); got != nil {
		t.Errorf("infinite random: got %v, want nil", songIDs(got))
	}
	if got := NewPlaylistManager().Upcoming(3); got != nil {
		t.Errorf("empty playlist: got %v, want nil", songIDs(got))
	}
	if got := newUpcomingManager(t, types.PmOrdered, 0, 5).Upcoming(0); got != nil {
		t.Errorf("n=0: got %v, want nil", songIDs(got))
	}
}
