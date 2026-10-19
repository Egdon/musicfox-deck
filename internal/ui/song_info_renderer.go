package ui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/mattn/go-runewidth"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
)

// SongInfoRenderer is a dedicated UI component for rendering the current song's information.
type SongInfoRenderer struct {
	netease *Netease
	state   playerRendererState

	cachedView     string
	cachedLines    int
	cachedSongId   int64
	cachedState    types.State
	cachedVolume   int
	cachedMode     types.Mode
	cachedLike     bool
	cachedWidth    int
	cachedCentered bool
	cachedHover    PlaybarElement
	cachedStyleGen uint64

	spans []playbarSpan // where each element was last drawn
}

// playbarSegment is one piece of the playbar: unformatted text, its color and
// the element it belongs to.
type playbarSegment struct {
	text  string
	color color.Color
	bold  bool
	elem  PlaybarElement
}

// playbarSpan is the column range an element occupies.
type playbarSpan struct {
	elem   PlaybarElement
	x0, x1 int
}

// elementAt returns the playbar element drawn at column x.
func (r *SongInfoRenderer) elementAt(x int) PlaybarElement {
	for _, s := range r.spans {
		if x >= s.x0 && x < s.x1 {
			return s.elem
		}
	}
	return PlaybarElementNone
}

// NewSongInfoRenderer creates a new song info renderer component.
func NewSongInfoRenderer(netease *Netease, state playerRendererState) *SongInfoRenderer {
	return &SongInfoRenderer{
		netease: netease,
		state:   state,
	}
}

// Update handles UI messages.
func (r *SongInfoRenderer) Update(msg tea.Msg, a *model.App) {}

// View renders the song information component.
func (r *SongInfoRenderer) View(a *model.App, main *model.Main) (view string, lines int) {
	var (
		song           = r.state.CurSong()
		state          = r.state.State()
		volume         = r.state.Volume()
		mode           = r.state.Mode()
		width          = r.netease.WindowWidth()
		centered       = main.CenterEverything()
		hoveredElement = r.netease.playbarHoveredElement
	)

	var isLike bool
	if song.Id > 0 {
		isLike = likelist.IsLikeSong(song.Id)
	}

	// Output caching: skip full rebuild when nothing changed. styleGen guards
	// against replaying stale-colored output after a theme switch.
	styleGen := style.StyleGeneration()
	if song.Id == r.cachedSongId && state == r.cachedState && volume == r.cachedVolume &&
		mode == r.cachedMode && isLike == r.cachedLike && width == r.cachedWidth &&
		centered == r.cachedCentered && hoveredElement == r.cachedHover &&
		styleGen == r.cachedStyleGen {
		return r.cachedView, r.cachedLines
	}

	var (
		builder           strings.Builder
		segments          []playbarSegment
		appColors         = configs.GetCurrentAppColors()
		artistColor       = configs.SafeGetForeground(appColors.PlaybarArtist, configs.PlaybarArtistColor)
		modeColor         = configs.SafeGetForeground(appColors.PlaybarMode, configs.PlaybarModeColor)
		volumeColor       = configs.SafeGetForeground(appColors.PlaybarVolume, configs.PlaybarVolumeColor)
		playingColor      = configs.SafeGetForeground(appColors.PlaybarPlaying, configs.PlaybarPlayingColor)
		pausedColor       = configs.SafeGetForeground(appColors.PlaybarPaused, configs.PlaybarPausedColor)
		heartLikedColor   = configs.SafeGetForeground(appColors.PlaybarHeartLiked, configs.PlaybarHeartLikedColor)
		heartUnlikedColor = configs.SafeGetForeground(appColors.PlaybarHeartUnliked, configs.PlaybarHeartUnlikedColor)
		primary           = themePrimary()
	)

	// A hovered element is drawn bold in the primary color.
	add := func(text string, c color.Color, elem PlaybarElement) {
		hovered := elem != PlaybarElementNone && elem == hoveredElement
		if hovered {
			c = primary
		}
		segments = append(segments, playbarSegment{text: text, color: c, bold: hovered, elem: elem})
	}
	gap := func(n int) { add(strings.Repeat(" ", n), artistColor, PlaybarElementNone) }
	usedWidth := func() int {
		w := 0
		for _, seg := range segments {
			w += runewidth.StringWidth(seg.text)
		}
		return w
	}
	// Resolve the background once: the app background paints text glyph cells so
	// they never reveal content drawn beneath the TUI. nil leaves it transparent.
	segmentBg := configs.ResolveBackground(nil)
	renderSegment := func(segment playbarSegment) string {
		s := lipgloss.NewStyle().Foreground(segment.color).Bold(segment.bold)
		if segmentBg != nil {
			s = s.Background(segmentBg)
		}
		return s.Render(segment.text)
	}

	if main.MenuStartColumn()-MenuArrowWidth > 0 {
		if !main.CenterEverything() {
			// Plain spaces (not pre-rendered): renderSegment paints them with
			// segmentBg, and their width stays measurable.
			gap(main.MenuStartColumn() - MenuArrowWidth)
		}
		add(mode.Name(), modeColor, PlaybarElementMode)
		gap(3)
		add(fmt.Sprintf("VOL %d", volume), volumeColor, PlaybarElementNone)
		gap(2)
	}
	// Icons carry a space on each side so they make a usable click target.
	if state == types.Playing {
		add(" ▶ ", playingColor, PlaybarElementState)
	} else {
		add(" ‖ ", pausedColor, PlaybarElementState)
	}
	if song.Id > 0 {
		c := heartUnlikedColor
		if isLike {
			c = heartLikedColor
		}
		add(" ♥ ", c, PlaybarElementHeart)
	}
	gap(1)

	if r.state.CurSongIndex() < len(r.state.Playlist()) {
		// The song name keeps priority; the artist gets what is left.
		add(runewidth.Truncate(song.Name, max(0, width-usedWidth()-2), ""), primary, PlaybarElementSongName)
		gap(2)
		add(runewidth.Truncate(artistsString(song), max(0, width-usedWidth()), ""), artistColor, PlaybarElementArtist)
	}

	x := 0
	if main.CenterEverything() {
		totalWidth := 0
		widthLimit := width - SongInfoHorizontalPadding
		for index, segment := range segments {
			segmentWidth := runewidth.StringWidth(segment.text)
			if totalWidth+segmentWidth > widthLimit {
				segmentWidth = max(0, widthLimit-totalWidth)
				segments[index].text = runewidth.Truncate(segment.text, segmentWidth, "")
			}
			totalWidth += segmentWidth
		}
		x = (width - totalWidth) / 2
		builder.WriteString(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", x)))
	}
	// Record where every element lands, so mouse hit-testing uses exactly
	// what was drawn.
	r.spans = r.spans[:0]
	for _, segment := range segments {
		w := runewidth.StringWidth(segment.text)
		if segment.elem != PlaybarElementNone && w > 0 {
			r.spans = append(r.spans, playbarSpan{elem: segment.elem, x0: x, x1: x + w})
		}
		builder.WriteString(renderSegment(segment))
		x += w
	}
	if remainingWidth := width - x; remainingWidth > 0 {
		builder.WriteString(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", remainingWidth)))
	}

	// Return the actual number of lines: 1 content line + 1 blank line separator.
	// In the new JoinVertical model, the returned `lines` value is discarded (no longer
	// used to advance a `top` pointer). Only the actual string content matters for layout.
	// Output exactly 2 visual rows: content + single trailing newline = 1 blank separator.
	lines = SongInfoLines

	// Store output cache
	r.cachedView = builder.String() + "\n"
	r.cachedLines = lines
	r.cachedSongId = song.Id
	r.cachedState = state
	r.cachedVolume = volume
	r.cachedMode = mode
	r.cachedLike = isLike
	r.cachedWidth = width
	r.cachedCentered = centered
	r.cachedHover = hoveredElement
	r.cachedStyleGen = styleGen

	return r.cachedView, r.cachedLines
}

// artistsString joins the artists' names into a comma-separated string.
func artistsString(song structs.Song) string {
	if len(song.Artists) == 0 {
		return ""
	}
	var b strings.Builder
	for i, v := range song.Artists {
		if i != 0 {
			b.WriteString(",")
		}
		b.WriteString(v.Name)
	}
	return b.String()
}
