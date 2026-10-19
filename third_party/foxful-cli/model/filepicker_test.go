package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// buildFixtureTree creates a temporary directory tree for testing.
// Layout:
//
//	root/
//	  alpha/        (dir)
//	  zeta/         (dir)
//	  .hidden/      (hidden dir)
//	  banana.txt    (file)
//	  apple.txt     (file)
//	  .secret.txt   (hidden file)
func buildFixtureTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	dirs := []string{"alpha", "zeta", ".hidden"}
	for _, d := range dirs {
		if err := os.Mkdir(filepath.Join(root, d), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	files := []string{"banana.txt", "apple.txt", ".secret.txt"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(root, f), []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", f, err)
		}
	}

	// A nested file inside alpha for navigation tests
	if err := os.WriteFile(filepath.Join(root, "alpha", "nested.txt"), []byte("y"), 0o644); err != nil {
		t.Fatalf("write nested: %v", err)
	}

	return root
}

func keyMsg(s string) tea.KeyMsg {
	// Map string names to Key codes
	var code rune
	var text string
	switch s {
	case "up":
		code = tea.KeyUp
	case "down":
		code = tea.KeyDown
	case "left":
		code = tea.KeyLeft
	case "right":
		code = tea.KeyRight
	case "enter":
		code = tea.KeyEnter
	case "backspace":
		code = tea.KeyBackspace
	case "home":
		code = tea.KeyHome
	case "end":
		code = tea.KeyEnd
	case "pgup":
		code = tea.KeyPgUp
	case "pgdown":
		code = tea.KeyPgDown
	default:
		// Single character keys like 'j', 'k', 'g', 'G', 'h', 'l'
		for _, r := range s {
			code = r
			text = s
			break
		}
	}
	return tea.KeyPressMsg(tea.Key{Code: code, Text: text})
}

func TestFilePickerListingSortsDirsBeforeFiles(t *testing.T) {
	root := buildFixtureTree(t)
	fp := NewFilePicker(root)

	// With hidden files off, we expect: alpha/, zeta/ (dirs), then apple.txt, banana.txt (files)
	if len(fp.entries) != 4 {
		t.Fatalf("expected 4 visible entries, got %d: %+v", len(fp.entries), fp.entries)
	}

	want := []struct {
		name  string
		isDir bool
	}{
		{"alpha", true},
		{"zeta", true},
		{"apple.txt", false},
		{"banana.txt", false},
	}
	for i, w := range want {
		if fp.entries[i].name != w.name || fp.entries[i].isDir != w.isDir {
			t.Errorf("entry %d = {%q, dir=%v}, want {%q, dir=%v}",
				i, fp.entries[i].name, fp.entries[i].isDir, w.name, w.isDir)
		}
	}
}

func TestFilePickerHiddenToggle(t *testing.T) {
	root := buildFixtureTree(t)
	fp := NewFilePicker(root)

	// Default: hidden files not shown (4 entries)
	if len(fp.entries) != 4 {
		t.Fatalf("expected 4 entries with hidden off, got %d", len(fp.entries))
	}

	// Enable hidden files: adds .hidden/ and .secret.txt → 6 entries
	fp.SetShowHidden(true)
	if len(fp.entries) != 6 {
		t.Fatalf("expected 6 entries with hidden on, got %d", len(fp.entries))
	}

	// Verify a hidden dir and hidden file are present
	var haveHiddenDir, haveHiddenFile bool
	for _, e := range fp.entries {
		if e.name == ".hidden" && e.isDir {
			haveHiddenDir = true
		}
		if e.name == ".secret.txt" && !e.isDir {
			haveHiddenFile = true
		}
	}
	if !haveHiddenDir || !haveHiddenFile {
		t.Errorf("hidden entries missing: dir=%v file=%v", haveHiddenDir, haveHiddenFile)
	}

	// Toggle back off
	fp.SetShowHidden(false)
	if len(fp.entries) != 4 {
		t.Fatalf("expected 4 entries after hiding again, got %d", len(fp.entries))
	}
}

func TestFilePickerNavigateIntoDirectory(t *testing.T) {
	root := buildFixtureTree(t)
	fp := NewFilePicker(root)
	fp.SetSize(40, 10)

	// alpha/ is the first entry (index 0). Enter it.
	fp.Update(keyMsg("enter"))

	if fp.CurrentDir() != filepath.Join(root, "alpha") {
		t.Fatalf("expected to be in alpha, got %q", fp.CurrentDir())
	}
	// alpha contains nested.txt
	if len(fp.entries) != 1 || fp.entries[0].name != "nested.txt" {
		t.Fatalf("expected nested.txt inside alpha, got %+v", fp.entries)
	}
}

func TestFilePickerNavigateOutOfDirectory(t *testing.T) {
	root := buildFixtureTree(t)
	fp := NewFilePicker(root)
	fp.SetSize(40, 10)

	// Enter alpha, then go back to parent
	fp.Update(keyMsg("enter"))
	if fp.CurrentDir() != filepath.Join(root, "alpha") {
		t.Fatalf("setup: expected alpha, got %q", fp.CurrentDir())
	}

	fp.Update(keyMsg("backspace"))
	if fp.CurrentDir() != root {
		t.Fatalf("expected back at root, got %q", fp.CurrentDir())
	}
	if len(fp.entries) != 4 {
		t.Fatalf("expected 4 entries back at root, got %d", len(fp.entries))
	}
}

func TestFilePickerSelectionReturnsCorrectPath(t *testing.T) {
	root := buildFixtureTree(t)
	fp := NewFilePicker(root)
	fp.SetSize(40, 10)

	// Move down to a file (index 2: apple.txt)
	fp.Update(keyMsg("down"))
	fp.Update(keyMsg("down"))

	path, isDir := fp.Selected()
	wantPath := filepath.Join(root, "apple.txt")
	if path != wantPath {
		t.Errorf("Selected() path = %q, want %q", path, wantPath)
	}
	if isDir {
		t.Errorf("Selected() isDir = true, want false for apple.txt")
	}
	if fp.SelectedPath() != wantPath {
		t.Errorf("SelectedPath() = %q, want %q", fp.SelectedPath(), wantPath)
	}

	// First entry is a dir
	fp.Update(keyMsg("home"))
	path, isDir = fp.Selected()
	if !isDir {
		t.Errorf("Selected() isDir = false, want true for alpha")
	}
	if path != filepath.Join(root, "alpha") {
		t.Errorf("Selected() path = %q, want alpha path", path)
	}
}

func TestFilePickerNavigationClamps(t *testing.T) {
	root := buildFixtureTree(t)
	fp := NewFilePicker(root)
	fp.SetSize(40, 10)

	// Up at top stays at 0
	fp.Update(keyMsg("up"))
	if fp.selected != 0 {
		t.Errorf("up at top: selected = %d, want 0", fp.selected)
	}

	// End goes to last, further down stays clamped
	fp.Update(keyMsg("end"))
	last := len(fp.entries) - 1
	if fp.selected != last {
		t.Errorf("end: selected = %d, want %d", fp.selected, last)
	}
	fp.Update(keyMsg("down"))
	if fp.selected != last {
		t.Errorf("down at bottom: selected = %d, want %d", fp.selected, last)
	}
}

func TestFilePickerNeverAboveRoot(t *testing.T) {
	fp := NewFilePicker("/")
	fp.SetSize(40, 10)

	before := fp.CurrentDir()
	fp.Update(keyMsg("backspace"))
	if fp.CurrentDir() != before {
		t.Errorf("navigating above root changed dir: %q -> %q", before, fp.CurrentDir())
	}
}

func TestFilePickerEmptyDirRendersEmptyState(t *testing.T) {
	empty := t.TempDir()
	fp := NewFilePicker(empty)
	fp.SetSize(40, 10)

	if len(fp.entries) != 0 {
		t.Fatalf("expected empty dir, got %d entries", len(fp.entries))
	}

	view := fp.View()
	if !strings.Contains(view, "empty") {
		t.Errorf("expected empty-state text in view, got:\n%s", view)
	}

	// Selection on empty dir returns nothing
	if p := fp.SelectedPath(); p != "" {
		t.Errorf("SelectedPath on empty dir = %q, want empty", p)
	}
	if _, isDir := fp.Selected(); isDir {
		t.Errorf("Selected on empty dir returned isDir=true")
	}
}

func TestFilePickerReadErrorRendersGracefully(t *testing.T) {
	// Point at a non-existent path — construction must not panic.
	fp := NewFilePicker("/this/path/should/not/exist/xyz123")
	fp.SetSize(40, 10)

	if fp.readError == nil {
		t.Fatalf("expected a read error for non-existent path")
	}

	view := fp.View()
	if !strings.Contains(view, "Error") {
		t.Errorf("expected error text in view, got:\n%s", view)
	}
}

func TestFilePickerFocusBlur(t *testing.T) {
	root := buildFixtureTree(t)
	fp := NewFilePicker(root)
	fp.SetSize(40, 10)

	if !fp.Focused() {
		t.Errorf("expected focused by default")
	}

	// Blurred: input ignored
	fp.Blur()
	if fp.Focused() {
		t.Errorf("expected blurred after Blur()")
	}
	before := fp.selected
	fp.Update(keyMsg("down"))
	if fp.selected != before {
		t.Errorf("blurred picker should ignore input: %d -> %d", before, fp.selected)
	}

	// Refocus: input handled
	fp.Focus()
	fp.Update(keyMsg("down"))
	if fp.selected == before {
		t.Errorf("focused picker should handle input")
	}
}

func TestFilePickerScrollKeepsSelectionVisible(t *testing.T) {
	// Build a directory with many files to force scrolling.
	root := t.TempDir()
	for i := 0; i < 30; i++ {
		name := filepath.Join(root, "file"+string(rune('a'+i%26))+string(rune('0'+i/26))+".txt")
		if err := os.WriteFile(name, []byte("x"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	fp := NewFilePicker(root)
	fp.SetSize(40, 6) // small height forces scrolling

	fp.Update(keyMsg("end"))
	last := len(fp.entries) - 1
	if fp.selected != last {
		t.Fatalf("end: selected = %d, want %d", fp.selected, last)
	}
	// Selected must be within the visible window.
	visibleHeight := fp.height - 1
	if fp.selected < fp.scrollOffset || fp.selected >= fp.scrollOffset+visibleHeight {
		t.Errorf("selected %d not visible in window [%d, %d)",
			fp.selected, fp.scrollOffset, fp.scrollOffset+visibleHeight)
	}
}
