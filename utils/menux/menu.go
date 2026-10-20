package menux

import (
	"fmt"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/model"

	ds "github.com/go-musicfox/go-musicfox/internal/structs"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// GetViewFromSongs 从歌曲列表获取View
func GetViewFromSongs(songs []ds.Song) []model.MenuItem {
	var menus []model.MenuItem
	for _, song := range songs {
		var artists []string
		for _, artist := range song.Artists {
			artists = append(artists, artist.Name)
		}
		menus = append(menus, model.MenuItem{
			Title:    _struct.ReplaceSpecialStr(song.Name),
			Subtitle: _struct.ReplaceSpecialStr(strings.Join(artists, ",")),
			Meta:     songDuration(song.Duration),
		})
	}

	return menus
}

// songDuration formats a song length as m:ss, or h:mm:ss for long ones.
func songDuration(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	s := int(d.Round(time.Second).Seconds())
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
	}
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}

// GetViewFromAlbums 从歌曲列表获取View
func GetViewFromAlbums(albums []ds.Album) []model.MenuItem {
	var menus []model.MenuItem
	for _, album := range albums {
		var artists []string
		for _, artist := range album.Artists {
			artists = append(artists, artist.Name)
		}
		artistsStr := fmt.Sprintf("[%s]", strings.Join(artists, ","))
		menus = append(menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(album.Name), Subtitle: _struct.ReplaceSpecialStr(artistsStr)})
	}

	return menus
}

// GetViewFromPlaylists 从歌单列表获取View
func GetViewFromPlaylists(playlists []ds.Playlist) []model.MenuItem {
	var menus []model.MenuItem
	for _, playlist := range playlists {
		menus = append(menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(playlist.Name)})
	}

	return menus
}

// GetViewFromArtists 从歌手列表获取View
func GetViewFromArtists(artists []ds.Artist) []model.MenuItem {
	var menus []model.MenuItem
	for _, artist := range artists {
		menus = append(menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(artist.Name)})
	}

	return menus
}

// GetViewFromUsers 用户列表获取View
func GetViewFromUsers(users []ds.User) []model.MenuItem {
	var menus []model.MenuItem
	for _, user := range users {
		menus = append(menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(user.Nickname)})
	}

	return menus
}

// GetViewFromDjRadios DjRadio列表获取View
func GetViewFromDjRadios(radios []ds.DjRadio) []model.MenuItem {
	var menus []model.MenuItem
	for _, radio := range radios {
		var dj string
		if radio.Dj.Nickname != "" {
			dj = fmt.Sprintf("[%s]", radio.Dj.Nickname)
		}
		menus = append(menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(radio.Name), Subtitle: _struct.ReplaceSpecialStr(dj)})
	}

	return menus
}

// GetViewFromDjCate 分类列表获取View
func GetViewFromDjCate(categories []ds.DjCategory) []model.MenuItem {
	var menus []model.MenuItem
	for _, category := range categories {
		menus = append(menus, model.MenuItem{Title: _struct.ReplaceSpecialStr(category.Name)})
	}

	return menus
}
