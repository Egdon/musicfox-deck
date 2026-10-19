package model

import (
	"strings"
	"testing"
)

func TestMarkdownRenderToString(t *testing.T) {
	tests := []struct {
		name    string
		content string
		width   int
		wantErr bool
	}{
		{
			name:    "simple_markdown",
			content: "# Hello\n\nThis is **bold** text.",
			width:   80,
			wantErr: false,
		},
		{
			name:    "empty_content",
			content: "",
			width:   80,
			wantErr: false,
		},
		{
			name:    "zero_width_uses_default",
			content: "Some text",
			width:   0,
			wantErr: false,
		},
		{
			name:    "with_emoji_codes",
			content: "# Test :rocket:",
			width:   80,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := NewMarkdown(tt.content, WithMarkdownDarkStyle("dark"))
			rendered, err := md.RenderToString(tt.width)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.content == "" && rendered != "" {
				t.Errorf("Expected empty result for empty content, got %q", rendered)
			}
			if tt.content != "" && rendered == "" && !tt.wantErr {
				t.Errorf("Expected non-empty result, got empty string")
			}
		})
	}
}

func TestMarkdownRenderToStringWithEmoji(t *testing.T) {
	content := "# Test :rocket: emoji"
	md := NewMarkdown(content,
		WithMarkdownDarkStyle("dark"),
		WithMarkdownEmoji(true),
	)

	rendered, err := md.RenderToString(80)
	if err != nil {
		t.Fatalf("RenderToString() error = %v", err)
	}

	if rendered == "" {
		t.Error("Expected non-empty rendered output")
	}
}

func TestNewMarkdownPopup_DefaultCloseButton(t *testing.T) {
	popup, err := NewMarkdownPopup(MarkdownPopupSpec{
		Title:           "Test",
		MarkdownContent: "# Hello",
		Actions:         nil, // nil should add default Close button
	})

	if err != nil {
		t.Fatalf("NewMarkdownPopup() error = %v", err)
	}

	if popup == nil {
		t.Fatal("Expected non-nil popup")
	}

	// Check that a Close button was added
	if len(popup.actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(popup.actions))
	}

	if len(popup.actions) > 0 {
		if popup.actions[0].ID != "close" {
			t.Errorf("Expected action ID 'close', got %q", popup.actions[0].ID)
		}
		if popup.actions[0].Label != "Close" {
			t.Errorf("Expected action label 'Close', got %q", popup.actions[0].Label)
		}
		if !popup.actions[0].IsCancel {
			t.Error("Expected IsCancel to be true")
		}
	}
}

func TestNewMarkdownPopup_CustomButtons(t *testing.T) {
	popup, err := NewMarkdownPopup(MarkdownPopupSpec{
		Title:           "Confirm",
		MarkdownContent: "Are you sure?",
		Actions: []PopupAction{
			{ID: "yes", Label: "Yes"},
			{ID: "no", Label: "No", IsCancel: true},
		},
	})

	if err != nil {
		t.Fatalf("NewMarkdownPopup() error = %v", err)
	}

	if len(popup.actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(popup.actions))
	}

	if len(popup.actions) >= 2 {
		if popup.actions[0].ID != "yes" {
			t.Errorf("Expected first action ID 'yes', got %q", popup.actions[0].ID)
		}
		if popup.actions[1].ID != "no" {
			t.Errorf("Expected second action ID 'no', got %q", popup.actions[1].ID)
		}
	}
}

func TestNewMarkdownPopup_NoButtons(t *testing.T) {
	popup, err := NewMarkdownPopup(MarkdownPopupSpec{
		Title:           "Info",
		MarkdownContent: "Read only",
		Actions:         []PopupAction{}, // Empty slice = no buttons
	})

	if err != nil {
		t.Fatalf("NewMarkdownPopup() error = %v", err)
	}

	if len(popup.actions) != 0 {
		t.Errorf("Expected 0 actions, got %d", len(popup.actions))
	}
}

func TestNewMarkdownPopup_RendersMarkdown(t *testing.T) {
	content := "# Title\n\nThis is **bold** text."
	popup, err := NewMarkdownPopup(MarkdownPopupSpec{
		Title:           "Test",
		MarkdownContent: content,
		MaxWidth:        80,
	})

	if err != nil {
		t.Fatalf("NewMarkdownPopup() error = %v", err)
	}

	// Check that content was rendered (not raw markdown)
	if popup.content == content {
		t.Error("Expected rendered content, got raw markdown")
	}

	// Check that content is not empty
	if strings.TrimSpace(popup.content) == "" {
		t.Error("Expected non-empty rendered content")
	}
}

func TestNewMarkdownPopup_MaxWidthAccounting(t *testing.T) {
	// MaxWidth should account for popup frame overhead
	popup, err := NewMarkdownPopup(MarkdownPopupSpec{
		Title:           "Test",
		MarkdownContent: "Some content",
		MaxWidth:        60,
	})

	if err != nil {
		t.Fatalf("NewMarkdownPopup() error = %v", err)
	}

	if popup.maxWidth != 60 {
		t.Errorf("Expected maxWidth to be 60, got %d", popup.maxWidth)
	}
}

func TestNewMarkdownPopup_WithEmoji(t *testing.T) {
	popup, err := NewMarkdownPopup(MarkdownPopupSpec{
		Title:           "Emoji",
		MarkdownContent: ":rocket: Test",
		MarkdownEmoji:   true,
	})

	if err != nil {
		t.Fatalf("NewMarkdownPopup() error = %v", err)
	}

	if popup == nil {
		t.Fatal("Expected non-nil popup")
	}
}

func TestNewMarkdownPopup_EmptyContent(t *testing.T) {
	popup, err := NewMarkdownPopup(MarkdownPopupSpec{
		Title:           "Empty",
		MarkdownContent: "",
	})

	if err != nil {
		t.Fatalf("NewMarkdownPopup() error = %v", err)
	}

	if popup == nil {
		t.Fatal("Expected non-nil popup")
	}
}
