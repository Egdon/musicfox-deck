package model

import (
	"strings"
	"testing"

	"github.com/anhoder/foxful-cli/style"
)

// TestPopupCJKTitleWidthDoesNotExpandBorder verifies that CJK (wide) characters
// in the popup title do not push the top border wider than the rest of the frame.
// Regression test for the "detached top-right corner" artifact where CJK titles
// caused the top border to extend 2 columns per wide glyph beyond the correct width.
func TestPopupCJKTitleWidthDoesNotExpandBorder(t *testing.T) {
	tests := []struct {
		name  string
		title string
	}{
		{"CJK title", " 帮助 "},
		{"mixed ASCII+CJK", " Help 帮助 "},
		{"long CJK title", " 这是一个很长的中文标题用来测试截断 "},
		{"ASCII title", " Help "},
		{"empty title", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			popup, err := NewPopup(PopupSpec{
				Title:     tt.title,
				Content:   strings.Repeat("line\n", 10),
				MaxWidth:  80,
				MaxHeight: 15,
			})
			if err != nil {
				t.Fatal(err)
			}

			theme := style.DefaultDarkTheme()
			styles := style.NewStyleSet(theme).Popup
			rendered := popup.render(styles)

			screen := popupStyledScreen(rendered.content)
			if len(screen.Lines) < 2 {
				t.Fatal("popup has fewer than 2 lines")
			}

			topWidth := len(screen.Lines[0])
			bottomWidth := len(screen.Lines[len(screen.Lines)-1])

			// The top border (with title) and bottom border must have the same width.
			// If CJK handling is broken, the top border extends by the sum of wide glyphs' extra columns.
			if topWidth != bottomWidth {
				t.Errorf("top border width %d != bottom border width %d (title %q caused overflow)",
					topWidth, bottomWidth, tt.title)
			}

			// All content rows should also match the border width (the frame is consistent).
			for i := 1; i < len(screen.Lines)-1; i++ {
				if len(screen.Lines[i]) != topWidth {
					t.Errorf("content row %d width %d != frame width %d",
						i, len(screen.Lines[i]), topWidth)
				}
			}
		})
	}
}
