package menux

import (
	"testing"
	"time"
)

func TestSongDuration(t *testing.T) {
	for d, want := range map[time.Duration]string{
		0:                                     "",
		201 * time.Second:                     "3:21",
		59*time.Second + 600*time.Millisecond: "1:00",
		time.Hour + 2*time.Minute + 3*time.Second: "1:02:03",
	} {
		if got := songDuration(d); got != want {
			t.Errorf("songDuration(%v) = %q, want %q", d, got, want)
		}
	}
}
