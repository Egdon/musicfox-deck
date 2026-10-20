package ui

import (
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/menux"
)

const CurPlaylistKey = "cur_playlist"

type CurPlaylist struct {
	baseMenu
	menus []model.MenuItem
	songs []structs.Song
}

func NewCurPlaylist(base baseMenu, songs []structs.Song) *CurPlaylist {
	return &CurPlaylist{
		baseMenu: base,
		songs:    songs,
		menus:    menux.GetViewFromSongs(songs),
	}
}

func (m *CurPlaylist) IsSearchable() bool {
	return true
}

func (m *CurPlaylist) IsPlayable() bool {
	return true
}

func (m *CurPlaylist) GetMenuKey() string {
	return CurPlaylistKey
}

func (m *CurPlaylist) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *CurPlaylist) Songs() []structs.Song {
	return m.songs
}

func (m *CurPlaylist) BottomOutHook() model.Hook { return model.HookOf(m.BottomOutFetch()) }

// BottomOutFetch loads more of the menu the playlist came from, then shows
// the playlist again.
func (m *CurPlaylist) BottomOutFetch() model.Fetch {
	playing := m.netease.player.playingMenu
	if playing == nil || playing.GetMenuKey() == CurPlaylistKey {
		return nil
	}
	var fetch model.Fetch
	if pager, ok := playing.(model.BackgroundPager); ok {
		fetch = pager.BottomOutFetch()
	} else if hook := playing.BottomOutHook(); hook != nil {
		fetch = func() model.Hook { return hook } // it loads on the UI goroutine
	}
	if fetch == nil {
		return nil
	}
	return func() model.Hook {
		apply := fetch()
		return func(main *model.Main) (bool, model.Page) {
			res, page := true, model.Page(nil)
			if apply != nil {
				res, page = apply(main)
			}
			m.songs = m.netease.player.Playlist()
			m.menus = menux.GetViewFromSongs(m.songs)
			return res, page
		}
	}
}
