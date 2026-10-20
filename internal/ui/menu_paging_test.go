package ui

import "github.com/anhoder/foxful-cli/model"

// The menus that page load the next items in the background, and the one
// whose entry changes the playlist fetches before it applies.
var (
	_ model.BackgroundPager = (*AddToUserPlaylistMenu)(nil)
	_ model.BackgroundPager = (*AlbumNewMenu)(nil)
	_ model.BackgroundPager = (*AlbumSubscribeListMenu)(nil)
	_ model.BackgroundPager = (*AlbumTopMenu)(nil)
	_ model.BackgroundPager = (*ArtistsSubscribeListMenu)(nil)
	_ model.BackgroundPager = (*CloudMenu)(nil)
	_ model.BackgroundPager = (*CurPlaylist)(nil)
	_ model.BackgroundPager = (*DjRadioDetailMenu)(nil)
	_ model.BackgroundPager = (*DjSubListMenu)(nil)
	_ model.BackgroundPager = (*DjTodayRecommendMenu)(nil)
	_ model.BackgroundPager = (*PersonalFmMenu)(nil)
	_ model.BackgroundPager = (*SearchResultMenu)(nil)
	_ model.BackgroundPager = (*SimilarSongsMenu)(nil)
	_ model.BackgroundPager = (*UserPlaylistMenu)(nil)

	_ model.EnterFetcher = (*SimilarSongsMenu)(nil)
)
