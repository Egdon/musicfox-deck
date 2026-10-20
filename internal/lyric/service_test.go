package lyric

import (
	"context"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

func TestStateIncludesLyricOffset(t *testing.T) {
	service := NewService(nil, false, 250*time.Millisecond, false)
	if got, want := service.State().OffsetMs, int64(250); got != want {
		t.Errorf("initial offset = %dms, want %dms", got, want)
	}

	service.SetOffset(-100 * time.Millisecond)
	if got, want := service.State().OffsetMs, int64(-100); got != want {
		t.Errorf("updated offset = %dms, want %dms", got, want)
	}
}

// blockingFetcher holds each fetch until it is released.
type blockingFetcher struct {
	release chan structs.LRCData
}

func (f *blockingFetcher) GetLyric(context.Context, structs.Song) (structs.LRCData, error) {
	return <-f.release, nil
}

// State must not wait for a fetch: the interface reads it on every frame.
func TestStateDoesNotWaitForFetch(t *testing.T) {
	f := &blockingFetcher{release: make(chan structs.LRCData)}
	service := NewService(f, false, 0, false)
	done := make(chan error)
	go func() { done <- service.SetSong(context.Background(), structs.Song{Id: 1}) }()

	read := make(chan struct{})
	go func() { service.State(); close(read) }()
	select {
	case <-read:
	case <-time.After(time.Second):
		t.Fatal("State blocked while the lyrics were being fetched")
	}

	f.release <- structs.LRCData{Original: "[00:01.00]hello"}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if st := service.State(); len(st.Fragments) != 1 || st.Fragments[0].Content != "hello" {
		t.Fatalf("lyrics after the fetch: %+v", st.Fragments)
	}
}

// A fetch that ends after a later song started loading is dropped.
func TestStaleFetchIsDropped(t *testing.T) {
	f := &blockingFetcher{release: make(chan structs.LRCData)}
	service := NewService(f, false, 0, false)
	first := make(chan error)
	go func() { first <- service.SetSong(context.Background(), structs.Song{Id: 1}) }()
	time.Sleep(20 * time.Millisecond) // let the first fetch start
	second := make(chan error)
	go func() { second <- service.SetSong(context.Background(), structs.Song{Id: 2}) }()
	time.Sleep(20 * time.Millisecond)

	f.release <- structs.LRCData{Original: "[00:01.00]old song"}
	<-first
	f.release <- structs.LRCData{Original: "[00:01.00]new song"}
	if err := <-second; err != nil {
		t.Fatal(err)
	}
	if st := service.State(); len(st.Fragments) != 1 || st.Fragments[0].Content != "new song" {
		t.Fatalf("lyrics: %+v", st.Fragments)
	}
}
