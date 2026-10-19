package ui

import (
	"bytes"
	"image/color"
	"os"
	"path"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/filex"
)

// themeColor returns c unless the theme left it unset.
func themeColor(c, fallback color.Color) color.Color {
	if c == nil {
		return fallback
	}
	if _, none := c.(lipgloss.NoColor); none {
		return fallback
	}
	return c
}

// themePrimary is the active theme's primary color. Themes that leave it
// unset fall back to the legacy [theme].primaryColor.
func themePrimary() color.Color {
	return themeColor(style.CurrentStyleSet().Theme().Primary, util.GetPrimaryColor())
}

func themeSecondary() color.Color {
	return themeColor(style.CurrentStyleSet().Theme().Secondary, lipgloss.BrightBlack)
}

func themeAccent() color.Color {
	return themeColor(style.CurrentStyleSet().Theme().Accent, util.GetPrimaryColor())
}

// Tip colors of the form pages, set by a theme's [app] loginError,
// loginSuccess and authStatus; other themes keep the ANSI colors.
func tipError() color.Color {
	return configs.SafeGetForeground(configs.GetCurrentAppColors().LoginError, configs.LoginErrorColor)
}

func tipSuccess() color.Color {
	return configs.SafeGetForeground(configs.GetCurrentAppColors().LoginSuccess, configs.LoginSuccessColor)
}

func tipInfo() color.Color {
	return configs.SafeGetForeground(configs.GetCurrentAppColors().AuthStatus, configs.AuthStatusColor)
}

// builtinMarkdownPrefix marks a glamour style shipped in the binary, e.g.
// "builtin:warm_night" in a theme's [*.markdown] section.
const builtinMarkdownPrefix = "builtin:"

// markdownStyle returns the glamour style of the active theme. glamour reads
// custom styles only from files, so a built-in one is written to the cache
// directory on first use.
func markdownStyle(darkBackground bool) string {
	name := configs.GetCurrentMarkdownTheme(darkBackground)
	builtin, ok := strings.CutPrefix(name, builtinMarkdownPrefix)
	if !ok {
		return name
	}
	fallback := configs.MarkdownLightDefault
	if darkBackground {
		fallback = configs.MarkdownDarkDefault
	}
	data, err := filex.ReadFileFromEmbed(path.Join("embed/markdown", builtin+".json"))
	if err != nil {
		return fallback
	}
	file := filepath.Join(app.CacheDir(), "markdown", builtin+".json")
	if existing, err := os.ReadFile(file); err != nil || !bytes.Equal(existing, data) {
		if os.MkdirAll(filepath.Dir(file), 0o755) != nil || os.WriteFile(file, data, 0o644) != nil {
			return fallback
		}
	}
	return file
}
