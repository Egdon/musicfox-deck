package ui

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/anhoder/foxful-cli/model"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/internal/ui/deck"
)

// ---- framebuffer ----

// npScreen is the page framebuffer. Cells left without a background get the
// page background at the end, so widgets only set backgrounds they own.
type npScreen struct {
	buf  *uv.Buffer
	w, h int
	pal  npPalette
}

func (s *npScreen) reset(w, h int) {
	if s.buf == nil {
		s.buf = uv.NewBuffer(w, h)
	} else {
		s.buf.Resize(w, h)
	}
	s.w, s.h = w, h
	s.buf.Clear()
}

// render serializes the frame: exactly h lines, no trailing newline.
func (s *npScreen) render() string {
	return s.buf.Render()
}

func (s *npScreen) set(x, y int, content string, width int, st uv.Style) {
	if x < 0 || y < 0 || y >= s.h || x+width > s.w {
		return
	}
	s.buf.SetCell(x, y, &uv.Cell{Content: content, Width: width, Style: st})
}

func (s *npScreen) fill(x, y, w, h int, bg color.Color) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			s.set(xx, yy, " ", 1, uv.Style{Bg: bg})
		}
	}
}

func (s *npScreen) repeat(x, y, n int, ch string, st uv.Style) {
	for i := range max(n, 0) {
		s.set(x+i, y, ch, 1, st)
	}
}

// npNextCluster returns the first grapheme cluster of str, normalized so that
// grapheme and wcwidth measurements agree. The compositor measures graphemes
// while the renderer pads with wcwidth; a disagreement shifts the whole frame.
func npNextCluster(str string) (cluster string, width, size int) {
	cl, gw := ansi.FirstGraphemeCluster(str, ansi.GraphemeWidth)
	if len(cl) == 0 {
		_, size = utf8.DecodeRuneInString(str)
		return "", 0, size
	}
	if gw == ansi.WcWidth.StringWidth(cl) {
		return cl, gw, len(cl)
	}
	r, _ := utf8.DecodeRuneInString(cl)
	if rs := string(r); ansi.GraphemeWidth.StringWidth(rs) == ansi.WcWidth.StringWidth(rs) {
		return rs, ansi.WcWidth.StringWidth(rs), len(cl)
	}
	return "?", 1, len(cl)
}

func npSanitize(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, str)
}

func npWidth(str string) int {
	w := 0
	for len(str) > 0 {
		_, cw, n := npNextCluster(str)
		w += cw
		str = str[n:]
	}
	return w
}

// text draws str at (x, y) within maxW columns, ending with "…" when cut, and
// returns the number of columns used.
func (s *npScreen) text(x, y int, str string, st uv.Style, maxW int) int {
	if maxW <= 0 || str == "" {
		return 0
	}
	str = npSanitize(str)
	limit, cut := maxW, npWidth(str) > maxW
	if cut {
		limit = maxW - 1
	}
	col := 0
	for len(str) > 0 {
		cl, w, n := npNextCluster(str)
		str = str[n:]
		if w == 0 {
			continue
		}
		if col+w > limit {
			break
		}
		s.set(x+col, y, cl, w, st)
		col += w
	}
	if cut {
		s.set(x+col, y, "…", 1, st)
		col++
	}
	return col
}

// centered draws str centered on row y, within maxW columns.
func (s *npScreen) centered(y int, str string, st uv.Style, maxW int) {
	sw := min(npWidth(npSanitize(str)), maxW)
	s.text((s.w-sw)/2, y, str, st, maxW)
}

// ---- half-block pixels ----

// npPixels is a pixel canvas drawn with half blocks: every cell holds a top
// and a bottom pixel, which are roughly square in a terminal font.
type npPixels struct {
	w, h int
	px   []color.Color
}

func newNPPixels(w, h int) *npPixels {
	h += h % 2
	return &npPixels{w: w, h: h, px: make([]color.Color, w*h)}
}

func (p *npPixels) set(x, y int, c color.Color) {
	if x >= 0 && y >= 0 && x < p.w && y < p.h {
		p.px[y*p.w+x] = c
	}
}

func (p *npPixels) at(x, y int) color.Color {
	if x < 0 || y < 0 || x >= p.w || y >= p.h {
		return nil
	}
	return p.px[y*p.w+x]
}

func npSameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

// blit copies the canvas to the screen; transparent pixels keep the screen
// background.
func (s *npScreen) blit(p *npPixels, x0, y0 int) {
	for cy := 0; cy < p.h/2; cy++ {
		for cx := 0; cx < p.w; cx++ {
			top, bot := p.at(cx, cy*2), p.at(cx, cy*2+1)
			switch {
			case top == nil && bot == nil:
			case bot == nil:
				s.set(x0+cx, y0+cy, "▀", 1, uv.Style{Fg: top})
			case top == nil:
				s.set(x0+cx, y0+cy, "▄", 1, uv.Style{Fg: bot})
			case npSameColor(top, bot):
				s.set(x0+cx, y0+cy, "█", 1, uv.Style{Fg: top})
			default:
				s.set(x0+cx, y0+cy, "▀", 1, uv.Style{Fg: top, Bg: bot})
			}
		}
	}
}

// ---- seven-segment counter ----

// npCounterWidth and npCounterRows are the size of "mm:ss" in seven-segment
// digits of 4×7 pixels.
const (
	npCounterWidth = 21
	npCounterRows  = 4
)

// npDigitSegments holds the lit segments of each digit, bit 0 = a … bit 6 = g.
var npDigitSegments = [10]uint8{0x3F, 0x06, 0x5B, 0x4F, 0x66, 0x6D, 0x7D, 0x07, 0x7F, 0x6F}

// npSegmentPixels maps the segments a–g onto a 4×7 pixel grid.
var npSegmentPixels = [7][4][2]int{
	{{0, 0}, {1, 0}, {2, 0}, {3, 0}}, // a
	{{3, 0}, {3, 1}, {3, 2}, {3, 3}}, // b
	{{3, 3}, {3, 4}, {3, 5}, {3, 6}}, // c
	{{0, 6}, {1, 6}, {2, 6}, {3, 6}}, // d
	{{0, 3}, {0, 4}, {0, 5}, {0, 6}}, // e
	{{0, 0}, {0, 1}, {0, 2}, {0, 3}}, // f
	{{0, 3}, {1, 3}, {2, 3}, {3, 3}}, // g
}

func npDrawDigit(p *npPixels, x, digit int, lit, ghost color.Color) {
	for seg := range npSegmentPixels {
		for _, xy := range npSegmentPixels[seg] {
			p.set(x+xy[0], xy[1], ghost)
		}
	}
	for seg := range npSegmentPixels {
		if npDigitSegments[digit]&(1<<seg) == 0 {
			continue
		}
		for _, xy := range npSegmentPixels[seg] {
			p.set(x+xy[0], xy[1], lit)
		}
	}
}

// counter draws mm:ss in seven-segment digits whose unlit segments stay
// faintly visible, like a VFD tape counter.
func (s *npScreen) counter(x, y int, d time.Duration, lit, ghost color.Color) {
	total := min(max(int(d/time.Second), 0), 99*60+59)
	digits := [4]int{total / 600, total / 60 % 10, total % 60 / 10, total % 10}
	p := newNPPixels(npCounterWidth, npCounterRows*2)
	for i, dx := range [4]int{0, 5, 12, 17} {
		npDrawDigit(p, dx, digits[i], lit, ghost)
	}
	p.set(10, 2, lit)
	p.set(10, 4, lit)
	s.blit(p, x, y)
}

// ---- effects ----

func npLerp(from, to color.Color, t float64) color.Color {
	t = min(max(t, 0), 1)
	r1, g1, b1, _ := from.RGBA()
	r2, g2, b2, _ := to.RGBA()
	mix := func(a, b uint32) uint8 {
		return uint8((float64(a)*(1-t) + float64(b)*t) / 257)
	}
	return color.RGBA{mix(r1, r2), mix(g1, g2), mix(b1, b2), 0xFF}
}

// fillBackground gives every cell without a background the page background.
func (s *npScreen) fillBackground() {
	for y := 0; y < s.h; y++ {
		line := s.buf.Line(y)
		for x := range line {
			if line[x].Width == 0 && line[x].Content == "" {
				continue // wide-cell placeholder
			}
			if line[x].Style.Bg == nil {
				line[x].Style.Bg = s.pal.Bg
			}
		}
	}
}

// sweep is the tracking line of a track change: one row travelling from the
// bottom to the top as t runs 0→1, slightly lifting the background and
// washing out what it passes over.
func (s *npScreen) sweep(t float64) {
	y := int(math.Round((1 - t) * float64(s.h-1)))
	line := s.buf.Line(y)
	for x := range line {
		st := &line[x].Style
		if st.Bg != nil {
			st.Bg = npLerp(st.Bg, s.pal.Secondary, 0.10)
		}
		if st.Fg != nil && st.Bg != nil {
			st.Fg = npLerp(st.Fg, st.Bg, 0.35)
		}
	}
}

// fade blends every color towards the page background; k=1 is all background.
func (s *npScreen) fade(k float64) {
	if k <= 0 {
		return
	}
	for y := 0; y < s.h; y++ {
		line := s.buf.Line(y)
		for x := range line {
			st := &line[x].Style
			if st.Fg != nil {
				st.Fg = npLerp(st.Fg, s.pal.Bg, k)
			}
			if st.Bg != nil {
				st.Bg = npLerp(st.Bg, s.pal.Bg, k)
			}
		}
	}
}

// paint replaces the screen with a rendered frame, such as another page,
// faded towards the background by k.
func (s *npScreen) paint(frame string, k float64) {
	sb := uv.NewScreenBuffer(s.w, s.h) // measures cells as the renderer does
	uv.NewStyledString(frame).Draw(sb, sb.Bounds())
	s.buf.Clear()
	for y := range s.h {
		for x := range s.w {
			c := sb.CellAt(x, y)
			if c == nil || c.Width == 0 {
				continue
			}
			// Keep the frame's widths agreeing under both measures.
			if cl, w, _ := npNextCluster(c.Content); w != c.Width || cl != c.Content {
				if w < 1 || w > c.Width {
					cl, w = "?", 1
				}
				s.set(x, y, cl, w, c.Style)
				for i := w; i < c.Width; i++ {
					s.set(x+i, y, " ", 1, c.Style)
				}
				continue
			}
			s.buf.SetCell(x, y, c)
		}
	}
	s.fillBackground()
	s.fade(k)
}

// drawDeck copies the cells a cassette frame draws onto the screen at (x, y).
// Where only its shadow falls on something already drawn, that is darkened
// instead, so a cassette passing over text does not cut a hole in it.
func (s *npScreen) drawDeck(f *deck.Frame, x, y int) {
	if f == nil {
		return
	}
	for i, c := range f.Cells {
		if c.Ch == "" {
			continue
		}
		cx, cy := x+i%f.W, y+i/f.W
		if c.Shadow > 0 && s.shade(cx, cy, c.Shadow) {
			continue
		}
		st := uv.Style{Bg: c.Bg}
		if c.Ch != " " {
			st.Fg = c.Fg
		}
		s.set(cx, cy, c.Ch, 1, st)
	}
}

// shade darkens the cell at (x, y) by k and reports whether it did; blank
// cells on the page background are left for the shadow's own glyphs.
func (s *npScreen) shade(x, y int, k float64) bool {
	c := s.buf.CellAt(x, y)
	switch {
	case c == nil:
		return true
	case c.Width == 0: // the tail of a wide character: its head carries the color
		return true
	case (c.Content == " " || c.Content == "") && (c.Style.Bg == nil || npSameColor(c.Style.Bg, s.pal.Bg)):
		return false
	}
	black := color.RGBA{A: 0xff}
	if c.Style.Fg != nil {
		c.Style.Fg = npLerp(c.Style.Fg, black, k)
	}
	bg := c.Style.Bg
	if bg == nil {
		bg = s.pal.Bg
	}
	c.Style.Bg = npLerp(bg, black, k)
	return true
}

// ---- page layout ----

// npSnapshot is everything the page draws, captured once per frame so that
// the layout is a pure function of it.
type npSnapshot struct {
	Now        time.Time
	Pal        npPalette
	Song       structs.Song
	State      types.State
	BlueScreen bool
	Scanlines  bool // scanline texture on the idle screen
	Passed     time.Duration
	Volume     int
	Mode       types.Mode
	Liked      bool
	Quality    service.SongQualityLevel
	QueuePos   int // 1-based, 0 when unknown
	QueueLen   int
	TapeMins   int // total queue duration in minutes
	Upcoming   []structs.Song
	Lyrics     []npLyricLine
	LyricIndex int
	Sung       float64 // how much of the current line has been sung, 0..1
	LyricsMode bool
	OSD        string
	OSDFade    float64 // 0 shows the OSD fully, 1 hides it
	Keys       npKeys
	Angles     [2]float64
	Deck       *deck.Frame // the cassette on the stage; nil draws none
	Aspect     float64     // cell height over width; 0 means 2
}

type npLyricLine struct {
	Text, Trans string
}

// npKeys holds the display labels of the user's key bindings.
type npKeys struct {
	Prev, Toggle, Next, Like, Mode, VolDown, VolUp string
}

type npHitKind int

const (
	npHitNone npHitKind = iota
	npHitProgress
	npHitPrev
	npHitToggle
	npHitNext
	npHitLike
	npHitMode
	npHitVolDown
	npHitVolUp
	npHitLyrics
	npHitBack
	npHitUpNext
)

// npHit is a clickable region recorded while drawing.
type npHit struct {
	kind  npHitKind
	x, y  int
	w     int
	index int // position in the up-next list
}

func (h npHit) contains(x, y int) bool {
	return y == h.y && x >= h.x && x < h.x+h.w
}

func npProgressRatio(s *npSnapshot) float64 {
	if s.Song.Duration <= 0 {
		return 0
	}
	return min(max(float64(s.Passed)/float64(s.Song.Duration), 0), 1)
}

func npClock(d time.Duration) string {
	sec := max(int(d/time.Second), 0)
	return fmt.Sprintf("%02d:%02d", sec/60, sec%60)
}

// npSpaced letter-spaces a short Latin label: "UP NEXT" → "U P   N E X T".
func npSpaced(label string) string {
	var b strings.Builder
	for i, r := range label {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// npTapeType maps the configured quality to a cassette tape formulation.
func npTapeType(level service.SongQualityLevel) (string, string) {
	switch level {
	case service.Standard:
		return "TYPE I", "NORMAL"
	case service.Higher, service.Exhigh:
		return "TYPE II", "CrO₂"
	default:
		return "TYPE IV", "METAL"
	}
}

func npHighQuality(level service.SongQualityLevel) bool {
	switch level {
	case service.Standard, service.Higher, service.Exhigh:
		return false
	default:
		return true
	}
}

func npTapeTag(mins int) string {
	switch {
	case mins <= 46:
		return "C-46"
	case mins <= 60:
		return "C-60"
	case mins <= 90:
		return "C-90"
	default:
		return "C-120"
	}
}

// npRender lays out the page for the snapshot and returns the clickable regions.
func npRender(scr *npScreen, s *npSnapshot) []npHit {
	scr.pal = s.Pal
	if s.BlueScreen {
		return npRenderBlueScreen(scr, s)
	}
	switch npLayoutOf(scr.w, scr.h, s.LyricsMode) {
	case npLayoutWide:
		return npRenderWide(scr, s)
	case npLayoutLandscape:
		return npRenderLandscape(scr, s)
	case npLayoutPoster:
		return npRenderPoster(scr, s)
	case npLayoutLyrics:
		return npRenderLyrics(scr, s)
	default:
		return npRenderCompact(scr, s)
	}
}

// npHeader is a quiet surface band: the wordmark and page number on the left,
// status lamps on the right that are either lit or ghosted, never moved.
func npHeader(scr *npScreen, s *npSnapshot) {
	p := scr.pal
	scr.fill(0, 0, scr.w, 1, p.Surface)
	x := 2
	x += scr.text(x, 0, "■", uv.Style{Fg: p.Accent, Bg: p.Surface}, scr.w-4)
	x += 2
	x += scr.text(x, 0, "M U S I C F O X", uv.Style{Fg: p.Text, Bg: p.Surface, Attrs: uv.AttrBold}, scr.w-4-x)
	x += 4
	x += scr.text(x, 0, "01 / NOW PLAYING", uv.Style{Fg: p.Secondary, Bg: p.Surface}, scr.w-2-x)

	lamps := []struct {
		label string
		lit   bool
		color color.Color
	}{
		{"SHUF", s.Mode == types.PmListRandom || s.Mode == types.PmInfRandom, p.Accent2},
		{"LOOP", s.Mode == types.PmListLoop, p.Accent2},
		{"ONE", s.Mode == types.PmSingleLoop, p.Accent2},
		{"♥", s.Liked, p.Accent},
		{"LRC", len(s.Lyrics) > 0, p.Accent2},
		{"HQ", npHighQuality(s.Quality), p.Accent2},
	}
	total := -2
	for _, l := range lamps {
		total += npWidth(l.label) + 2
	}
	lx := scr.w - 2 - total
	if lx < x+4 {
		return
	}
	for _, l := range lamps {
		st := uv.Style{Fg: p.Ghost, Bg: p.Surface}
		if l.lit {
			st.Fg = l.color
		}
		lx += scr.text(lx, 0, l.label, st, scr.w-lx) + 2
	}
}

// npOSDRow is the VCR on-screen display: the transport state, then the tape
// speed, and a date stamp on the right.
func npOSDRow(scr *npScreen, s *npSnapshot, y int) {
	p := scr.pal
	if s.OSD != "" && s.OSDFade < 1 {
		osd := uv.Style{Fg: npLerp(p.Hero, p.Bg, s.OSDFade), Attrs: uv.AttrBold}
		n := scr.text(2, y, s.OSD, osd, scr.w/2)
		if s.Song.Id != 0 && s.OSD == osdStateText(s.State) {
			scr.text(2+n+3, y, "SP", uv.Style{Fg: npLerp(p.Secondary, p.Bg, s.OSDFade)}, 2)
		}
	}
	stamp := strings.ToUpper(s.Now.Format("Jan.02 2006  15:04"))
	if sw := npWidth(stamp); scr.w-2-sw > scr.w/2 {
		scr.text(scr.w-2-sw, y, stamp, uv.Style{Fg: p.Secondary}, sw)
	}
}

func npSection(scr *npScreen, x, y int, num, label string, w int) {
	n := scr.text(x, y, num, uv.Style{Fg: scr.pal.Text}, w)
	n += scr.text(x+n, y, " ── ", uv.Style{Fg: scr.pal.Rule}, w-n)
	scr.text(x+n, y, npSpaced(label), uv.Style{Fg: scr.pal.Secondary}, w-n)
}

func npArtistAlbum(song structs.Song) string {
	artist := artistsString(song)
	switch {
	case artist != "" && song.Album.Name != "":
		return artist + "  —  " + song.Album.Name
	case artist != "":
		return artist
	default:
		return song.Album.Name
	}
}

func npMeta(scr *npScreen, x, y, w int, s *npSnapshot) {
	p := scr.pal
	tape, formula := npTapeType(s.Quality)
	n := scr.text(x, y, " "+tape+" ", uv.Style{Fg: p.Text, Bg: p.Surface}, w)
	meta := "  " + formula + "  ·  " + qualityDisplayName(s.Quality) + "   " + s.Mode.String()
	if s.QueueLen > 0 && s.QueuePos > 0 {
		meta += fmt.Sprintf("   %02d / %02d", s.QueuePos, s.QueueLen)
	}
	scr.text(x+n, y, meta, uv.Style{Fg: p.Secondary}, w-n)
}

func npTimes(s *npSnapshot) string {
	return "−" + npClock(s.Song.Duration-s.Passed) + "   /   " + npClock(s.Song.Duration)
}

func npUpNext(scr *npScreen, x, y, w, rows int, s *npSnapshot) []npHit {
	if rows <= 0 {
		return nil
	}
	p := scr.pal
	dim := uv.Style{Fg: p.Secondary}
	switch {
	case s.Mode == types.PmInfRandom || (s.Mode == types.PmListRandom && s.Upcoming == nil):
		scr.text(x, y, model.T(MsgNowPlayingShuffleNext), dim, w)
		return nil
	case s.Mode == types.PmSingleLoop && len(s.Upcoming) > 0:
		n := scr.text(x, y, "∞  ", uv.Style{Fg: p.Accent2}, w)
		scr.text(x+n, y, s.Upcoming[0].Name, uv.Style{Fg: p.Text}, w-n)
		if rows > 1 {
			scr.text(x, y+1, model.T(MsgNowPlayingRepeatOne), dim, w)
		}
		return nil
	case s.QueueLen == 0:
		scr.text(x, y, model.T(MsgNowPlayingEmptyQueue), dim, w)
		return nil
	case len(s.Upcoming) == 0:
		scr.text(x, y, model.T(MsgNowPlayingEndOfList), dim, w)
		return nil
	}
	var hits []npHit
	for i, song := range s.Upcoming {
		if i >= rows {
			break
		}
		n := scr.text(x, y+i, fmt.Sprintf("%02d  ", i+1), uv.Style{Fg: p.Rule}, w)
		n += scr.text(x+n, y+i, song.Name, uv.Style{Fg: p.Text}, w-n)
		if artist := artistsString(song); artist != "" {
			scr.text(x+n, y+i, "  —  "+artist, dim, w-n)
		}
		hits = append(hits, npHit{kind: npHitUpNext, x: x, y: y + i, w: w, index: i})
	}
	return hits
}

// npLyricRow is a row of the lyrics: a line or its translation, and how
// many lines away from the current one it is.
type npLyricRow struct {
	text  string
	trans bool
	dist  int
}

// style dims a row with its distance from the current line.
func (r npLyricRow) style(p npPalette) uv.Style {
	switch {
	case r.dist == 0:
		return uv.Style{Fg: p.Text}
	case r.dist <= 2:
		return uv.Style{Fg: p.Secondary}
	default:
		return uv.Style{Fg: p.Rule}
	}
}

// npLyricWindow picks up to n rows around the current line; each
// translation sits under its line.
func npLyricWindow(s *npSnapshot, n int) []npLyricRow {
	var out []npLyricRow
	cur := 0
	for i, l := range s.Lyrics {
		if i == s.LyricIndex {
			cur = len(out)
		}
		d := i - s.LyricIndex
		if d < 0 {
			d = -d
		}
		out = append(out, npLyricRow{l.Text, false, d})
		if l.Trans != "" {
			out = append(out, npLyricRow{l.Trans, true, d})
		}
	}
	start := min(max(cur-(n-1)/2, 0), max(len(out)-n, 0))
	return out[start:min(start+n, len(out))]
}

// npLyrics shows the lyrics centered on the current line; each translation
// sits under its line, and lines fade with their distance from the current one.
func npLyrics(scr *npScreen, x, y, w, rows int, s *npSnapshot) {
	if rows <= 0 || w <= 2 {
		return
	}
	p := scr.pal
	if len(s.Lyrics) == 0 {
		scr.text(x+2, y+rows/2, model.T(MsgNowPlayingNoLyrics), uv.Style{Fg: p.Rule}, w-2)
		return
	}
	for r, ln := range npLyricWindow(s, rows) {
		if ln.dist == 0 && !ln.trans {
			scr.text(x, y+r, "▸", uv.Style{Fg: p.Accent}, 1)
			scr.karaoke(x+2, y+r, ln.text, s.Sung, w-2)
			continue
		}
		scr.text(x+2, y+r, ln.text, ln.style(p), w-2)
	}
}

// npLyricsCentered is npLyrics with every line centered on the screen.
func npLyricsCentered(scr *npScreen, y, rows, w int, s *npSnapshot) {
	if rows <= 0 || w <= 2 {
		return
	}
	p := scr.pal
	if len(s.Lyrics) == 0 {
		scr.centered(y+rows/2, model.T(MsgNowPlayingNoLyrics), uv.Style{Fg: p.Rule}, w)
		return
	}
	for r, ln := range npLyricWindow(s, rows) {
		if ln.dist == 0 && !ln.trans {
			lw := min(npWidth(npSanitize(ln.text)), w)
			scr.karaoke((scr.w-lw)/2, y+r, ln.text, s.Sung, w)
			continue
		}
		scr.centered(y+r, ln.text, ln.style(p), w)
	}
}

func npProgressBar(scr *npScreen, y int, s *npSnapshot) []npHit {
	x0, w := 2, scr.w-4
	if w < 4 {
		return nil
	}
	p := scr.pal
	head := -1
	if s.Song.Duration > 0 {
		head = int(math.Round(npProgressRatio(s) * float64(w-1)))
	}
	for i := range w {
		switch {
		case i < head:
			scr.set(x0+i, y, "━", 1, uv.Style{Fg: p.Secondary})
		case i == head:
			scr.set(x0+i, y, "●", 1, uv.Style{Fg: p.Accent})
		default:
			scr.set(x0+i, y, "─", 1, uv.Style{Fg: p.Ghost})
		}
	}
	return []npHit{{kind: npHitProgress, x: x0, y: y, w: w}}
}

type npLegendItem struct {
	kind   npHitKind
	key    string
	label  string
	volume bool
}

// npLegend draws the transport legend: key badges followed by their labels.
// Every item is clickable; items are shortened, then dropped from the right,
// until the row fits.
func npLegend(scr *npScreen, y int, s *npSnapshot, onBlue bool) []npHit {
	p := scr.pal
	badge := uv.Style{Fg: p.Text, Bg: p.Surface, Attrs: uv.AttrBold}
	label := uv.Style{Fg: p.Secondary}
	if onBlue {
		badge = uv.Style{Fg: p.Hero, Bg: npLerp(npNoSignal, p.Hero, 0.12), Attrs: uv.AttrBold}
		label = uv.Style{Fg: p.Text}
	}
	toggle := "▶ PLAY"
	if s.State == types.Playing {
		toggle = "▌▌ PAUSE"
	}
	lyrics := "LYRICS"
	if s.LyricsMode {
		lyrics = "DECK"
	}
	like := "♥ LIKE"
	if s.Liked {
		like = "♥ LIKED"
	}
	full := []npLegendItem{
		{npHitPrev, s.Keys.Prev, "◀◀ PREV", false},
		{npHitToggle, s.Keys.Toggle, toggle, false},
		{npHitNext, s.Keys.Next, "▶▶ NEXT", false},
		{npHitLike, s.Keys.Like, like, false},
		{npHitMode, s.Keys.Mode, "MODE", false},
		{npHitVolDown, "", "", true},
		{npHitLyrics, "L", lyrics, false},
		{npHitBack, "ESC", "BACK", false},
	}
	short := func(items []npLegendItem) []npLegendItem {
		out := make([]npLegendItem, len(items))
		copy(out, items)
		for i := range out {
			if f := strings.Fields(out[i].label); len(f) > 1 {
				out[i].label = f[0]
			}
		}
		return out
	}
	const volWidth = 18 // "VOL " + 10 segments + " 100"
	itemWidth := func(it npLegendItem) int {
		if it.volume {
			return npWidth(" "+s.Keys.VolDown+" ") + 1 + volWidth + 1 + npWidth(" "+s.Keys.VolUp+" ")
		}
		return npWidth(" "+it.key+" ") + 1 + npWidth(it.label)
	}
	fits := func(items []npLegendItem) bool {
		total := -3
		for _, it := range items {
			total += itemWidth(it) + 3
		}
		return total <= scr.w-4
	}
	items := full
	if !fits(items) {
		items = short(full)
	}
	for !fits(items) && len(items) > 1 {
		items = append(items[:len(items)-2:len(items)-2], items[len(items)-1]) // keep BACK
	}

	var hits []npHit
	x := 2
	for _, it := range items {
		start := x
		if it.volume {
			x += scr.text(x, y, " "+s.Keys.VolDown+" ", badge, scr.w-x)
			hits = append(hits, npHit{kind: npHitVolDown, x: start, y: y, w: x - start})
			x += 1 + scr.text(x+1, y, "VOL ", label, scr.w-x-1)
			lit := min(max((s.Volume+9)/10, 0), 10)
			for i := range 10 {
				seg := uv.Style{Fg: p.Ghost, Bg: label.Bg}
				if i < lit {
					seg.Fg = p.Text
				}
				if onBlue && i >= lit {
					seg.Fg = npLerp(npNoSignal, p.Hero, 0.3)
				}
				scr.set(x+i, y, "▪", 1, seg)
			}
			x += 10
			x += scr.text(x, y, fmt.Sprintf(" %3d", s.Volume), label, scr.w-x) + 1
			upStart := x
			x += scr.text(x, y, " "+s.Keys.VolUp+" ", badge, scr.w-x)
			hits = append(hits, npHit{kind: npHitVolUp, x: upStart, y: y, w: x - upStart})
		} else {
			st := label
			if it.kind == npHitLike && s.Liked && !onBlue {
				st = uv.Style{Fg: p.Accent}
			}
			x += scr.text(x, y, " "+it.key+" ", badge, scr.w-x)
			x += 1 + scr.text(x+1, y, it.label, st, scr.w-x-1)
			hits = append(hits, npHit{kind: it.kind, x: start, y: y, w: x - start})
		}
		x += 3
	}
	return hits
}

// npLayout is an arrangement of the page; the window size picks one.
type npLayout int

const (
	npLayoutCompact   npLayout = iota // text only, for tiny windows
	npLayoutLyrics                    // the lyrics view
	npLayoutWide                      // the cassette on a stage at the left, numbered columns at the right
	npLayoutLandscape                 // the wide layout for short windows, without the counter and the queue
	npLayoutPoster                    // narrow windows: the cassette with the track centered under it
)

func npLayoutOf(w, h int, lyricsMode bool) npLayout {
	switch {
	case w < 60 || h < 20:
		return npLayoutCompact
	case lyricsMode:
		return npLayoutLyrics
	case w >= 100 && h >= 28:
		return npLayoutWide
	case w >= 100:
		return npLayoutLandscape
	default:
		return npLayoutPoster
	}
}

// npStage is where the cassette rests in a layout, in cells; sh is 0 in
// the layouts without one.
func npStage(l npLayout, w, h int, aspect float64) (x, y, sw, sh int) {
	if aspect <= 0 {
		aspect = 2
	}
	switch l {
	case npLayoutWide: // between the OSD row and the counter
		return 2, 3, (w - 4) * 56 / 100, h - 10
	case npLayoutLandscape: // down to the progress bar
		sh = h - 6
		return 2, 3, min(int(float64(sh)*aspect*deck.FitAspect+0.5), (w-4)/2), sh
	case npLayoutPoster: // centered, as wide as the cassette
		pl := npPosterPlan(w, h, aspect)
		sw = min(int(float64(pl.stageH)*aspect*deck.FitAspect+0.5), w-4)
		return (w - sw) / 2, pl.stageY, sw, pl.stageH
	}
	return 0, 0, 0, 0
}

// npRenderWide puts the cassette on a stage at the left; the track and its
// lyrics run down a numbered column at the right, the counter and what plays
// next sit under them, then the progress bar and the legend.
func npRenderWide(scr *npScreen, s *npSnapshot) []npHit {
	w, h := scr.w, scr.h
	p := scr.pal
	var hits []npHit

	npHeader(scr, s)
	npOSDRow(scr, s, 1)

	sx, sy, sw, _ := npStage(npLayoutWide, w, h, s.Aspect)
	scr.drawDeck(s.Deck, sx, sy)

	rx := sx + sw + 3
	rw := w - 2 - rx
	npSection(scr, rx, sy, "02", "TRACK", rw)
	y := sy + 2
	for _, line := range npWrap(s.Song.Name, rw, 2) {
		scr.text(rx, y, line, uv.Style{Fg: p.Hero, Attrs: uv.AttrBold}, rw)
		y++
	}
	scr.text(rx, y, npArtistAlbum(s.Song), uv.Style{Fg: p.Text}, rw)
	npMeta(scr, rx, y+2, rw, s)

	bottom := h - 6 // the counter row
	lyricsY := y + 5
	npSection(scr, rx, lyricsY, "03", "LYRICS", rw)
	npLyrics(scr, rx, lyricsY+2, rw, bottom-1-lyricsY-2, s)

	scr.counter(sx, bottom, s.Passed, p.Text, p.Ghost)
	scr.text(sx+npCounterWidth+3, bottom+2, npTimes(s), uv.Style{Fg: p.Secondary}, sw-npCounterWidth-3)
	npSection(scr, rx, bottom, "04", "UP NEXT", rw)
	hits = append(hits, npUpNext(scr, rx, bottom+1, rw, 3, s)...)

	hits = append(hits, npProgressBar(scr, h-2, s)...)
	return append(hits, npLegend(scr, h-1, s, false)...)
}

// npWrap breaks a title into at most n lines of w columns, at spaces when it
// can; the last line ends with "…" when the title does not fit.
func npWrap(text string, w, n int) []string {
	text = npSanitize(text)
	var lines []string
	for len(lines) < n-1 && npWidth(text) > w {
		cut, col, lastSpace := 0, 0, -1
		for i, r := range text {
			cw := npWidth(string(r))
			if col+cw > w {
				break
			}
			if r == ' ' {
				lastSpace = i
			}
			col += cw
			cut = i + len(string(r))
		}
		if lastSpace > 0 {
			cut = lastSpace
		}
		lines = append(lines, strings.TrimRight(text[:cut], " "))
		text = strings.TrimLeft(text[cut:], " ")
	}
	return append(lines, text)
}

// karaoke draws the line being sung: the sung part in the hero color, the
// rest dimmer, and the character at the edge in between.
func (s *npScreen) karaoke(x, y int, text string, sung float64, maxW int) {
	if maxW <= 0 {
		return
	}
	text = npSanitize(text)
	total := npWidth(text)
	edge := sung * float64(total)
	cut := total > maxW
	limit := maxW
	if cut {
		limit = maxW - 1
	}
	col := 0
	for len(text) > 0 {
		cl, w, n := npNextCluster(text)
		text = text[n:]
		if w == 0 {
			continue
		}
		if col+w > limit {
			break
		}
		lit := min(max((edge-float64(col))/float64(w), 0), 1)
		s.set(x+col, y, cl, w, uv.Style{Fg: npLerp(s.pal.Secondary, s.pal.Hero, lit), Attrs: uv.AttrBold})
		col += w
	}
	if cut {
		s.set(x+col, y, "…", 1, uv.Style{Fg: s.pal.Secondary, Attrs: uv.AttrBold})
	}
}

// npRenderLandscape is the wide layout for short windows: the cassette at
// the left, and the track, the time and the lyrics at the right.
func npRenderLandscape(scr *npScreen, s *npSnapshot) []npHit {
	w, h := scr.w, scr.h
	p := scr.pal

	npHeader(scr, s)
	npOSDRow(scr, s, 1)

	sx, sy, sw, _ := npStage(npLayoutLandscape, w, h, s.Aspect)
	scr.drawDeck(s.Deck, sx, sy)

	rx := sx + sw + 3
	rw := w - 2 - rx
	npSection(scr, rx, sy, "02", "TRACK", rw)
	y := sy + 2
	for _, line := range npWrap(s.Song.Name, rw, 2) {
		scr.text(rx, y, line, uv.Style{Fg: p.Hero, Attrs: uv.AttrBold}, rw)
		y++
	}
	scr.text(rx, y, npArtistAlbum(s.Song), uv.Style{Fg: p.Text}, rw)
	y += 2
	n := scr.text(rx, y, npClock(s.Passed), uv.Style{Fg: p.Text, Attrs: uv.AttrBold}, rw)
	scr.text(rx+n+3, y, npTimes(s), uv.Style{Fg: p.Secondary}, rw-n-3)
	y += 2
	npSection(scr, rx, y, "03", "LYRICS", rw)
	npLyrics(scr, rx, y+2, rw, h-3-(y+2), s)

	hits := npProgressBar(scr, h-2, s)
	return append(hits, npLegend(scr, h-1, s, false)...)
}

// npPoster is the poster layout from top to bottom: the cassette, a gap,
// the title and the artist, a gap, then the lyrics. The block sits a little
// above the middle of the rows over the times.
type npPoster struct {
	stageY, stageH     int
	titleY             int
	lyricsY, lyricRows int
}

func npPosterPlan(w, h int, aspect float64) npPoster {
	const top = 3
	if aspect <= 0 {
		aspect = 2
	}
	avail := h - 7 // down to two rows over the times, the bar and the legend
	lyricRows := 1
	switch {
	case avail >= 22:
		lyricRows = 5
	case avail >= 15:
		lyricRows = 3
	}
	text := 3 + lyricRows
	stageH := min(avail-1-text, h*55/100, int(float64(w-8)/(aspect*deck.FitAspect)))
	y := top + (avail-stageH-1-text)*2/5
	return npPoster{stageY: y, stageH: stageH, titleY: y + stageH + 1, lyricsY: y + stageH + 4, lyricRows: lyricRows}
}

// npRenderPoster is the layout for narrow windows: the cassette, the track
// and the lines around the one being sung, all centered, then the times
// over the progress bar.
func npRenderPoster(scr *npScreen, s *npSnapshot) []npHit {
	w, h := scr.w, scr.h
	p := scr.pal
	tw := w - 8

	npHeader(scr, s)
	npOSDRow(scr, s, 1)

	pl := npPosterPlan(w, h, s.Aspect)
	sx, sy, _, _ := npStage(npLayoutPoster, w, h, s.Aspect)
	scr.drawDeck(s.Deck, sx, sy)
	scr.centered(pl.titleY, s.Song.Name, uv.Style{Fg: p.Hero, Attrs: uv.AttrBold}, tw)
	scr.centered(pl.titleY+1, npArtistAlbum(s.Song), uv.Style{Fg: p.Text}, tw)
	npLyricsCentered(scr, pl.lyricsY, pl.lyricRows, tw, s)

	scr.text(2, h-3, npClock(s.Passed), uv.Style{Fg: p.Text}, 6)
	left := "−" + npClock(s.Song.Duration-s.Passed)
	scr.text(w-2-npWidth(left), h-3, left, uv.Style{Fg: p.Secondary}, npWidth(left))
	hits := npProgressBar(scr, h-2, s)
	return append(hits, npLegend(scr, h-1, s, false)...)
}

// npRenderLyrics is the lyrics view: the track on one line, then as many
// lines of the lyrics as fit.
func npRenderLyrics(scr *npScreen, s *npSnapshot) []npHit {
	w, h := scr.w, scr.h
	p := scr.pal
	x, cw := 2, w-4

	npHeader(scr, s)
	npOSDRow(scr, s, 1)

	n := scr.text(x, 3, s.Song.Name, uv.Style{Fg: p.Hero, Attrs: uv.AttrBold}, cw)
	scr.text(x+n, 3, "   "+artistsString(s.Song), uv.Style{Fg: p.Secondary}, cw-n)
	npSection(scr, x, 5, "03", "LYRICS", cw)
	npLyrics(scr, x, 6, cw, h-9, s)

	hits := npProgressBar(scr, h-2, s)
	return append(hits, npLegend(scr, h-1, s, false)...)
}

// npRenderCompact keeps only text for small terminals.
func npRenderCompact(scr *npScreen, s *npSnapshot) []npHit {
	w, h := scr.w, scr.h
	p := scr.pal
	if h < 3 {
		scr.text(0, 0, s.Song.Name, uv.Style{Fg: p.Hero, Attrs: uv.AttrBold}, w)
		return nil
	}
	cw := w - 4
	if s.OSD != "" && s.OSDFade < 1 {
		scr.text(2, 0, s.OSD, uv.Style{Fg: npLerp(p.Hero, p.Bg, s.OSDFade), Attrs: uv.AttrBold}, cw)
	}
	var hits []npHit
	if h >= 6 {
		scr.text(2, 2, s.Song.Name, uv.Style{Fg: p.Hero, Attrs: uv.AttrBold}, cw)
		scr.text(2, 3, artistsString(s.Song), uv.Style{Fg: p.Text}, cw)
	}
	if h >= 8 {
		scr.text(2, 5, npClock(s.Passed)+" / "+npClock(s.Song.Duration), uv.Style{Fg: p.Secondary}, cw)
	}
	hits = append(hits, npProgressBar(scr, h-2, s)...)
	return append(hits, npLegend(scr, h-1, s, false)...)
}

// npRenderBlueScreen is the VCR "no signal" screen shown while nothing is
// loaded; it is the only screen that keeps a scanline texture.
func npRenderBlueScreen(scr *npScreen, s *npSnapshot) []npHit {
	w, h := scr.w, scr.h
	p := scr.pal
	osd := uv.Style{Fg: p.Hero, Attrs: uv.AttrBold}
	if h >= 2 {
		scr.text(2, 1, "■ STOP", osd, w-4)
		stamp := strings.ToUpper(s.Now.Format("Jan.02 2006  15:04"))
		if sw := npWidth(stamp); w-2-sw > w/2 {
			scr.text(w-2-sw, 1, stamp, uv.Style{Fg: p.Text}, sw)
		}
	}
	msg := npSpaced("NO SIGNAL")
	scr.text(max((w-npWidth(msg))/2, 0), h/2-1, msg, osd, w)
	hint := fmt.Sprintf(model.T(MsgNowPlayingStartHint), s.Keys.Toggle)
	scr.text(max((w-npWidth(hint))/2, 0), h/2+1, hint, uv.Style{Fg: p.Text}, w)
	var hits []npHit
	if h >= 4 {
		hits = npLegend(scr, h-1, s, true)
	}
	// Paint the blue behind everything drawn without a background.
	for y := range h {
		bg := npNoSignal
		if s.Scanlines && y%2 == 1 {
			bg = npNoSignalScan
		}
		line := scr.buf.Line(y)
		for x := range line {
			if line[x].Style.Bg == nil && !(line[x].Width == 0 && line[x].Content == "") {
				line[x].Style.Bg = bg
			}
		}
	}
	return hits
}
