package deck

import (
	"regexp"
	"strconv"
	"strings"
)

// Terminal describes the terminal for picking a tier: its environment and,
// once it has answered, its XTVERSION reply such as "ghostty 1.3.1" or
// "kitty(0.40.1)".
type Terminal struct {
	Getenv  func(string) string
	Version string
}

// DetectTier returns the finest tier the terminal draws itself, independent
// of the font. Unknown terminals get quadrants, which every font has.
func DetectTier(t Terminal) Tier {
	if t.Version != "" {
		if tier, ok := tierFromVersion(t.Version); ok {
			return tier
		}
	}
	getenv := t.Getenv
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	return tierFromEnv(getenv)
}

var versionRe = regexp.MustCompile(`^\s*([A-Za-z][A-Za-z0-9._-]*)[\s(]*v?([0-9][0-9.]*)?`)

// Native drawing support per terminal: the first version that draws octants
// and the first that draws sextants; "" means none does.
// Sources: arewelegacycomputingyet.com and the terminals' release notes.
func tierFromVersion(reply string) (Tier, bool) {
	m := versionRe.FindStringSubmatch(reply)
	if m == nil {
		return 0, false
	}
	ver := m[2]
	switch strings.ToLower(m[1]) {
	case "tmux", "screen":
		// The multiplexer answers for itself; the outer terminal is unknown.
		return Quadrant, true
	case "ghostty":
		return tierByVersion(ver, "1.1.1", "0"), true
	case "kitty":
		return tierByVersion(ver, "0.40", "0.19"), true
	case "foot":
		return tierByVersion(ver, "1.20", "1.7"), true
	case "wezterm":
		return tierByVersion(ver, "", "0"), true
	case "iterm2":
		return tierByVersion(ver, "", "3.6.7"), true
	case "vte":
		return tierByVersion(ver, "7800", "6000"), true
	case "konsole":
		return tierByVersion(ver, "", "0"), true
	case "rio":
		return tierByVersion(ver, "0.2.13", "0.2.13"), true
	}
	return 0, false
}

// tierFromEnv guesses from the environment. Every terminal sets TERM, so it
// comes first; TERM_PROGRAM and the terminal-specific variables also leak
// into terminals launched from inside another one.
func tierFromEnv(getenv func(string) string) Tier {
	term, program := getenv("TERM"), getenv("TERM_PROGRAM")
	switch {
	case getenv("TMUX") != "" || program == "tmux" || strings.HasPrefix(term, "screen") || strings.HasPrefix(term, "tmux"):
		return Quadrant
	case term == "xterm-ghostty":
		return tierByVersion(getenv("TERM_PROGRAM_VERSION"), "1.1.1", "0")
	case term == "xterm-kitty":
		return Octant
	case strings.HasPrefix(term, "foot"):
		return Octant
	case term == "alacritty":
		return Sextant
	case term == "wezterm":
		return Sextant
	case term == "xterm-rio" || term == "rio":
		return Octant
	case program == "ghostty":
		return tierByVersion(getenv("TERM_PROGRAM_VERSION"), "1.1.1", "0")
	case program == "WezTerm":
		return Sextant
	case program == "iTerm.app":
		return tierByVersion(getenv("TERM_PROGRAM_VERSION"), "", "3.6.7")
	case program == "rio":
		return tierByVersion(getenv("TERM_PROGRAM_VERSION"), "0.2.13", "0.2.13")
	case getenv("VTE_VERSION") != "":
		return tierByVersion(getenv("VTE_VERSION"), "7800", "6000")
	case getenv("KITTY_WINDOW_ID") != "":
		return Octant
	case getenv("ALACRITTY_WINDOW_ID") != "":
		return Sextant
	case getenv("KONSOLE_VERSION") != "":
		return tierByVersion(getenv("KONSOLE_VERSION"), "", "220400")
	}
	return Quadrant
}

// tierByVersion returns Octant from octantMin on, Sextant from sextantMin on
// and Quadrant below. An empty minimum means never; an unknown version counts
// as recent.
func tierByVersion(version, octantMin, sextantMin string) Tier {
	atLeast := func(min string) bool { return min != "" && (version == "" || versionAtLeast(version, min)) }
	switch {
	case atLeast(octantMin):
		return Octant
	case atLeast(sextantMin):
		return Sextant
	}
	return Quadrant
}

// versionAtLeast compares dotted numeric versions; a single number such as a
// VTE build is compared as a whole.
func versionAtLeast(v, min string) bool {
	vs, ms := strings.Split(v, "."), strings.Split(min, ".")
	for i := range max(len(vs), len(ms)) {
		a, b := 0, 0
		if i < len(vs) {
			a, _ = strconv.Atoi(strings.TrimFunc(vs[i], func(r rune) bool { return r < '0' || r > '9' }))
		}
		if i < len(ms) {
			b, _ = strconv.Atoi(ms[i])
		}
		if a != b {
			return a > b
		}
	}
	return true
}
