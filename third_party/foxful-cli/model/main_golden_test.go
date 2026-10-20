package model

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
)

// The golden frames pin Main's layout so that refactors can prove they render
// byte for byte the same. Update them with
//
//	go test -mod=readonly github.com/anhoder/foxful-cli/model -run TestMainGoldenFrames -update
//
// and set GOLDEN_DUMP to a directory to inspect the frames.
var updateGolden = flag.Bool("update", false, "rewrite the golden frame hashes")

const goldenFile = "testdata/main_frames.golden"

type goldenStatusBar struct{}

func (goldenStatusBar) View(a *App, _ *Main) string {
	return lipgloss.NewStyle().Width(a.WindowWidth()).Render(" » 网易云音乐 / 排行榜 / 飙升榜")
}

type goldenComponent struct{}

func (goldenComponent) Update(tea.Msg, *App) {}

func (goldenComponent) View(_ *App, _ *Main) (string, int) {
	return "\n  顺序播放  VOL 55  ▶ ♥  王子系变装  圈圈菌", 2
}

var goldenItems = []MenuItem{
	{Title: "每日推荐歌曲"}, {Title: "每日推荐歌单"}, {Title: "我的歌单"}, {Title: "我的收藏"},
	{Title: "私人FM"}, {Title: "专辑列表"}, {Title: "搜索"}, {Title: "排行榜"},
	{Title: "精选歌单"}, {Title: "热门歌手"}, {Title: "最近播放歌曲"}, {Title: "云盘"},
}

var goldenSongs = []MenuItem{
	{Title: "遗憾总是贯穿人生始终", Subtitle: "边缘,粘了咕叽"}, {Title: "学功夫练武术", Subtitle: "概念磨"},
	{Title: "王子系变装", Subtitle: "圈圈菌"}, {Title: "Odd Dice", Subtitle: "millsage"},
	{Title: "与天齐", Subtitle: "苏也莉"}, {Title: "Stay", Subtitle: "尹浩宇Patrick"},
	{Title: "琥珀色的情话", Subtitle: "周传雄"}, {Title: "淡想 (Memories)", Subtitle: "周童"},
	{Title: "A very long English title that will certainly not fit in a narrow column", Subtitle: "Somebody, Someone Else"},
	{Title: "伦敦雨季", Subtitle: "藤竹京,DY,鲸,风岸"}, {Title: "99", Subtitle: "TizzyT"},
	{Title: "海边的卡带", Subtitle: "Tape Unit"}, {Title: "Night Drive", Subtitle: "Tape Unit"},
}

func warmNightTestTheme() style.Theme {
	t := style.DefaultDarkTheme()
	t.Primary = lipgloss.Color("#DDD6C9")
	t.Secondary = lipgloss.Color("#9E978A")
	t.Accent = lipgloss.Color("#D2714F")
	t.Background = lipgloss.Color("#151412")
	t.Foreground = lipgloss.Color("#CBC4B6")
	t.Surface = lipgloss.Color("#1D1C19")
	t.AppBackground = style.Highlight{Bg: lipgloss.Color("#151412")}
	t.SelectedItem = style.Highlight{Fg: lipgloss.Color("#DDD6C9"), Bg: lipgloss.Color("#24221F")}
	t.MenuItem = style.Highlight{Fg: lipgloss.Color("#CBC4B6")}
	t.MenuSelectedArrow = "▎  "
	t.MenuArrow = style.Highlight{Fg: lipgloss.Color("#D2714F")}
	t.MenuIndex = style.Highlight{Fg: lipgloss.Color("#7A7368")}
	t.MenuSelectedSepLeft, t.MenuSelectedSepRight = "▐", "▌"
	return t
}

func transparentTestTheme() style.Theme {
	t := style.DefaultDarkTheme()
	t.Background = lipgloss.NoColor{}
	t.Surface = lipgloss.NoColor{}
	t.AppBackground = style.Highlight{Bg: lipgloss.NoColor{}}
	t.SelectedItem = style.Highlight{Bg: lipgloss.NoColor{}}
	t.MenuSelectedSepLeft, t.MenuSelectedSepRight = " ", " "
	return t
}

type goldenCase struct {
	w, h      int
	theme     string
	dual      bool
	centered  bool
	statusTop bool
	sub       bool
	dynamic   bool
}

func (c goldenCase) name() string {
	return fmt.Sprintf("%dx%d/%s/dual=%v/centered=%v/status=%v/sub=%v/dynamic=%v",
		c.w, c.h, c.theme, c.dual, c.centered, c.statusTop, c.sub, c.dynamic)
}

func goldenCases() []goldenCase {
	var cs []goldenCase
	for _, size := range [][2]int{{80, 24}, {120, 40}, {200, 50}} {
		for _, theme := range []string{"default", "warm", "transparent"} {
			for _, layout := range [][2]bool{{true, false}, {false, false}, {true, true}, {false, true}} {
				cs = append(cs, goldenCase{w: size[0], h: size[1], theme: theme, dual: layout[0], centered: layout[1]})
			}
			cs = append(cs,
				goldenCase{w: size[0], h: size[1], theme: theme, dual: true, statusTop: true, sub: true},
				goldenCase{w: size[0], h: size[1], theme: theme, dual: false, statusTop: true, sub: true, dynamic: true},
				goldenCase{w: size[0], h: size[1], theme: theme, dual: true, centered: true, sub: true, dynamic: true},
			)
		}
	}
	return cs
}

func renderGoldenFrame(t *testing.T, c goldenCase) string {
	t.Helper()
	themes := map[string]func() style.Theme{
		"default":     style.DefaultDarkTheme,
		"warm":        warmNightTestTheme,
		"transparent": transparentTestTheme,
	}
	style.SetStyleSet(style.NewStyleSet(themes[c.theme]()))

	options := DefaultOptions()
	options.WhetherDisplayTitle = true
	options.DualColumn = c.dual
	options.CenterEverything = c.centered
	options.DynamicRowCount = c.dynamic
	options.MainMenu = &mockMenu{key: "main", items: goldenItems}
	options.MainMenuTitle = &MenuItem{Title: "网易云音乐", Subtitle: "[anhoder]"}
	options.Components = []Component{goldenComponent{}}
	if c.statusTop {
		options.StatusBar = goldenStatusBar{}
		options.StatusBarPosition = StatusBarTop
	}

	app := NewApp(options)
	app.windowWidth, app.windowHeight = c.w, c.h
	main := NewMain(app, options)
	app.main = main
	main.Update(tea.WindowSizeMsg{Width: c.w, Height: c.h}, app)
	main.selectedIndex = 3
	if c.sub {
		main.EnterMenu(&mockMenu{key: "songs", items: goldenSongs}, &MenuItem{Title: "飙升榜", Subtitle: "[更新27首]"})
		main.selectedIndex = 8
	}
	return main.View(app)
}

func TestMainGoldenFrames(t *testing.T) {
	prev := style.CurrentStyleSet()
	t.Cleanup(func() { style.SetStyleSet(prev) })

	want := map[string]string{}
	if f, err := os.Open(goldenFile); err == nil {
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if name, sum, ok := strings.Cut(sc.Text(), " "); ok {
				want[name] = sum
			}
		}
		f.Close()
	} else if !*updateGolden {
		t.Fatalf("missing %s; run with -update", goldenFile)
	}

	dump := os.Getenv("GOLDEN_DUMP")
	var out strings.Builder
	for i, c := range goldenCases() {
		frame := renderGoldenFrame(t, c)
		h := sha256.Sum256([]byte(frame))
		sum := hex.EncodeToString(h[:])
		fmt.Fprintf(&out, "%s %s\n", c.name(), sum)
		if dump != "" {
			if err := os.WriteFile(filepath.Join(dump, fmt.Sprintf("%02d.ans", i)), []byte(c.name()+"\n"+frame), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if !*updateGolden && want[c.name()] != sum {
			t.Errorf("%s: frame changed (set GOLDEN_DUMP to inspect)", c.name())
		}
	}
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(goldenFile), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenFile, []byte(out.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
