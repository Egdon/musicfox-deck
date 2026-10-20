package ui

import (
	"fmt"
	"maps"
	"strconv"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

type SimilarSongsMenu struct {
	baseMenu
	menus        []model.MenuItem
	songs        []structs.Song
	existSongIds map[int64]struct{}
	relateSongId int64
}

var _ SongsMenu = (*SimilarSongsMenu)(nil)

func NewSimilarSongsMenu(base baseMenu, song structs.Song) *SimilarSongsMenu {
	songs := []structs.Song{song}
	return &SimilarSongsMenu{
		baseMenu:     base,
		menus:        menux.GetViewFromSongs(songs),
		songs:        songs,
		existSongIds: map[int64]struct{}{song.Id: {}},
		relateSongId: song.Id,
	}
}

func (m *SimilarSongsMenu) IsSearchable() bool {
	return true
}

func (m *SimilarSongsMenu) IsPlayable() bool {
	return true
}

func (m *SimilarSongsMenu) GetMenuKey() string {
	return fmt.Sprintf("simi_songs_%d", m.relateSongId)
}

func (m *SimilarSongsMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *SimilarSongsMenu) BeforeEnterMenuHook() model.Hook {
	return model.HookOf(m.BeforeEnterMenuFetch())
}

// BeforeEnterMenuFetch fetches in the background; the songs join the
// playlist on the UI goroutine.
func (m *SimilarSongsMenu) BeforeEnterMenuFetch() model.Fetch {
	return m.similarFetch(m.relateSongId)
}

func (m *SimilarSongsMenu) Songs() []structs.Song {
	return m.songs
}

func (m *SimilarSongsMenu) BottomOutHook() model.Hook { return model.HookOf(m.BottomOutFetch()) }

func (m *SimilarSongsMenu) BottomOutFetch() model.Fetch {
	if len(m.songs) == 0 {
		return nil
	}
	return m.similarFetch(m.songs[len(m.songs)-1].Id)
}

// similarFetch fetches songs similar to songId that the menu does not have
// yet, and adds them to the menu and the playlist on the UI goroutine.
func (m *SimilarSongsMenu) similarFetch(songId int64) model.Fetch {
	known := maps.Clone(m.existSongIds)
	return func() model.Hook {
		songs := similarSongs(songId, known, 2)
		return func(*model.Main) (bool, model.Page) {
			if len(songs) == 0 {
				return false, nil
			}
			for _, song := range songs {
				m.existSongIds[song.Id] = struct{}{}
			}
			m.songs = append(m.songs, songs...)
			m.menus = menux.GetViewFromSongs(m.songs)
			_ = m.netease.player.playlistManager.Initialize(m.netease.player.CurSongIndex(), m.songs)
			m.netease.player.playlistUpdateAt = time.Now()
			return true, nil
		}
	}
}

// similarSongs fetches songs similar to songId that known lacks and adds them
// to known, asking again up to maxTry times while there are none.
func similarSongs(songId int64, known map[int64]struct{}, maxTry int) []structs.Song {
	simiSongService := service.SimiSongService{
		ID: strconv.FormatInt(songId, 10),
	}
	code, response := simiSongService.SimiSong()
	codeType := _struct.CheckCode(code)
	if codeType != _struct.Success {
		return nil
	}

	var songs []structs.Song
	for _, song := range _struct.GetSimiSongs(response) {
		if _, ok := known[song.Id]; !ok {
			known[song.Id] = struct{}{}
			songs = append(songs, song)
		}
	}

	if len(songs) == 0 && maxTry > 0 {
		return similarSongs(songId, known, maxTry-1)
	}
	return songs
}
