package model

import (
	"strings"
	"testing"
)

func sampleTree() *TreeNode {
	return &TreeNode{
		Label:    "Root",
		Expanded: true,
		Children: []*TreeNode{
			{Label: "Child1", Children: nil},
			{
				Label: "Child2",
				Children: []*TreeNode{
					{Label: "Grandchild1", Children: nil},
					{Label: "Grandchild2", Children: nil},
				},
			},
			{Label: "Child3", Children: nil},
		},
	}
}

func TestTreeNew(t *testing.T) {
	root := &TreeNode{Label: "Root"}
	tree := NewTree(root)

	if tree == nil {
		t.Fatal("NewTree returned nil")
	}

	if tree.SelectedNode() != root {
		t.Error("new tree should select root")
	}

	if tree.Focused() {
		t.Error("new tree should not be focused")
	}
}

func TestTreeFocusBlur(t *testing.T) {
	tree := NewTree(&TreeNode{Label: "Root"})

	if tree.Focused() {
		t.Error("tree should start unfocused")
	}

	tree.Focus()
	if !tree.Focused() {
		t.Error("tree should be focused after Focus()")
	}

	tree.Blur()
	if tree.Focused() {
		t.Error("tree should be unfocused after Blur()")
	}
}

func TestTreeNavigation(t *testing.T) {
	tree := NewTree(sampleTree())
	tree.Focus()
	tree.SetSize(80, 20)

	// Initial: Root selected
	if tree.SelectedNode().Label != "Root" {
		t.Errorf("initial selection = %q, want Root", tree.SelectedNode().Label)
	}

	// Down to Child1
	tree.Update(newKeyMsg("down"))
	if tree.SelectedNode().Label != "Child1" {
		t.Errorf("after down, selection = %q, want Child1", tree.SelectedNode().Label)
	}

	// Down to Child2
	tree.Update(newKeyMsg("j"))
	if tree.SelectedNode().Label != "Child2" {
		t.Errorf("after j, selection = %q, want Child2", tree.SelectedNode().Label)
	}

	// Up to Child1
	tree.Update(newKeyMsg("up"))
	if tree.SelectedNode().Label != "Child1" {
		t.Errorf("after up, selection = %q, want Child1", tree.SelectedNode().Label)
	}

	// Up to Root
	tree.Update(newKeyMsg("k"))
	if tree.SelectedNode().Label != "Root" {
		t.Errorf("after k, selection = %q, want Root", tree.SelectedNode().Label)
	}

	// Up wraps to last (Child3)
	tree.Update(newKeyMsg("up"))
	if tree.SelectedNode().Label != "Child3" {
		t.Errorf("after wrap up, selection = %q, want Child3", tree.SelectedNode().Label)
	}

	// Down wraps to first (Root)
	tree.Update(newKeyMsg("down"))
	if tree.SelectedNode().Label != "Root" {
		t.Errorf("after wrap down, selection = %q, want Root", tree.SelectedNode().Label)
	}
}

func TestTreeToggle(t *testing.T) {
	tree := NewTree(sampleTree())
	tree.Focus()
	tree.SetSize(80, 20)

	// Navigate to Child2 (has children)
	tree.Update(newKeyMsg("down")) // Child1
	tree.Update(newKeyMsg("down")) // Child2

	if tree.SelectedNode().Label != "Child2" {
		t.Fatalf("setup failed: selection = %q, want Child2", tree.SelectedNode().Label)
	}

	// Initially collapsed, no grandchildren visible
	view := tree.View()
	if strings.Contains(view, "Grandchild1") {
		t.Error("grandchildren should not be visible when collapsed")
	}

	// Toggle expand
	tree.Update(newKeyMsg("enter"))
	if !tree.SelectedNode().Expanded {
		t.Error("node should be expanded after toggle")
	}

	view = tree.View()
	if !strings.Contains(view, "Grandchild1") {
		t.Error("grandchildren should be visible when expanded")
	}

	// Toggle collapse
	tree.Update(newKeyMsg(" ")) // space
	if tree.SelectedNode().Expanded {
		t.Error("node should be collapsed after second toggle")
	}

	view = tree.View()
	if strings.Contains(view, "Grandchild1") {
		t.Error("grandchildren should not be visible when collapsed again")
	}
}

func TestTreeExpandCollapse(t *testing.T) {
	tree := NewTree(sampleTree())
	tree.Focus()
	tree.SetSize(80, 20)

	// Navigate to Child2
	tree.Update(newKeyMsg("down"))
	tree.Update(newKeyMsg("down"))

	// Expand with right/l
	tree.Update(newKeyMsg("right"))
	if !tree.SelectedNode().Expanded {
		t.Error("node should be expanded after right")
	}

	tree.Collapse()
	if tree.SelectedNode().Expanded {
		t.Error("node should be collapsed after Collapse()")
	}

	tree.Expand()
	if !tree.SelectedNode().Expanded {
		t.Error("node should be expanded after Expand()")
	}

	// Collapse with left/h
	tree.Update(newKeyMsg("left"))
	if tree.SelectedNode().Expanded {
		t.Error("node should be collapsed after left")
	}

	tree.Update(newKeyMsg("l"))
	if !tree.SelectedNode().Expanded {
		t.Error("node should be expanded after l")
	}

	tree.Update(newKeyMsg("h"))
	if tree.SelectedNode().Expanded {
		t.Error("node should be collapsed after h")
	}
}

func TestTreeLeafNode(t *testing.T) {
	tree := NewTree(sampleTree())
	tree.Focus()
	tree.SetSize(80, 20)

	// Navigate to Child1 (leaf node)
	tree.Update(newKeyMsg("down"))

	if tree.SelectedNode().Label != "Child1" {
		t.Fatalf("setup failed: selection = %q, want Child1", tree.SelectedNode().Label)
	}

	// Toggle on leaf should do nothing
	initialExpanded := tree.SelectedNode().Expanded
	tree.Toggle()
	if tree.SelectedNode().Expanded != initialExpanded {
		t.Error("toggle on leaf node should not change expanded state")
	}

	tree.Expand()
	tree.Collapse()
	// Should not panic
}

func TestTreeHomeEnd(t *testing.T) {
	tree := NewTree(sampleTree())
	tree.Focus()
	tree.SetSize(80, 20)

	// Move to middle
	tree.Update(newKeyMsg("down"))
	tree.Update(newKeyMsg("down"))

	// Home
	tree.Update(newKeyMsg("home"))
	if tree.SelectedNode().Label != "Root" {
		t.Errorf("after home, selection = %q, want Root", tree.SelectedNode().Label)
	}

	// End
	tree.Update(newKeyMsg("end"))
	if tree.SelectedNode().Label != "Child3" {
		t.Errorf("after end, selection = %q, want Child3", tree.SelectedNode().Label)
	}

	// g (home alternative)
	tree.Update(newKeyMsg("g"))
	if tree.SelectedNode().Label != "Root" {
		t.Errorf("after g, selection = %q, want Root", tree.SelectedNode().Label)
	}

	// G (end alternative)
	tree.Update(newKeyMsg("G"))
	if tree.SelectedNode().Label != "Child3" {
		t.Errorf("after G, selection = %q, want Child3", tree.SelectedNode().Label)
	}
}

func TestTreeExpandedNavigation(t *testing.T) {
	tree := NewTree(sampleTree())
	tree.Focus()
	tree.SetSize(80, 20)

	// Navigate to Child2 and expand
	tree.Update(newKeyMsg("down"))
	tree.Update(newKeyMsg("down"))
	tree.Update(newKeyMsg("enter")) // expand Child2

	// Now navigate through visible nodes including grandchildren
	tree.Update(newKeyMsg("down"))
	if tree.SelectedNode().Label != "Grandchild1" {
		t.Errorf("after expanding Child2 and moving down, selection = %q, want Grandchild1", tree.SelectedNode().Label)
	}

	tree.Update(newKeyMsg("down"))
	if tree.SelectedNode().Label != "Grandchild2" {
		t.Errorf("selection = %q, want Grandchild2", tree.SelectedNode().Label)
	}

	tree.Update(newKeyMsg("down"))
	if tree.SelectedNode().Label != "Child3" {
		t.Errorf("selection = %q, want Child3", tree.SelectedNode().Label)
	}
}

func TestTreeView(t *testing.T) {
	tree := NewTree(sampleTree())
	tree.SetSize(80, 20)

	view := tree.View()
	if view == "" {
		t.Error("view should not be empty")
	}

	// Should contain root
	if !strings.Contains(view, "Root") {
		t.Error("view should contain Root")
	}

	// Should contain children
	if !strings.Contains(view, "Child1") {
		t.Error("view should contain Child1")
	}

	// Should have expand/collapse indicators
	if !strings.Contains(view, "▸") && !strings.Contains(view, "▾") && !strings.Contains(view, "•") {
		t.Error("view should contain tree indicators")
	}
}

func TestTreeEmptyChildren(t *testing.T) {
	root := &TreeNode{
		Label:    "Root",
		Children: []*TreeNode{},
	}
	tree := NewTree(root)
	tree.Focus()

	// Should not panic
	tree.Toggle()
	tree.Expand()
	tree.Collapse()

	view := tree.View()
	if view == "" {
		t.Error("view should not be empty even with no children")
	}
}

func TestTreeUnfocusedNoUpdate(t *testing.T) {
	tree := NewTree(sampleTree())
	// Keep unfocused

	initialLabel := tree.SelectedNode().Label
	tree.Update(newKeyMsg("down"))

	if tree.SelectedNode().Label != initialLabel {
		t.Error("unfocused tree should not respond to keys")
	}
}

func TestTreeSelectedNodeTracking(t *testing.T) {
	tree := NewTree(sampleTree())
	tree.Focus()
	tree.SetSize(80, 20)

	// Track selection through navigation
	expected := []string{"Root", "Child1", "Child2", "Child3", "Root"}

	for i, want := range expected {
		if tree.SelectedNode().Label != want {
			t.Errorf("step %d: selection = %q, want %q", i, tree.SelectedNode().Label, want)
		}
		tree.Update(newKeyMsg("down"))
	}
}

func TestTreeScrollVisibility(t *testing.T) {
	// Create a tall tree
	root := &TreeNode{
		Label:    "Root",
		Children: make([]*TreeNode, 30),
	}
	for i := range root.Children {
		root.Children[i] = &TreeNode{Label: "Child"}
	}

	tree := NewTree(root)
	tree.Focus()
	tree.SetSize(80, 5) // Small viewport

	// Move to bottom
	tree.Update(newKeyMsg("end"))

	// Selection should be visible (scroll adjusted)
	view := tree.View()
	lines := strings.Split(view, "\n")
	if len(lines) > 5 {
		t.Errorf("view has %d lines, want at most 5", len(lines))
	}

	// Move to top
	tree.Update(newKeyMsg("home"))

	// Should scroll back to show top
	view = tree.View()
	if !strings.Contains(view, "Root") {
		t.Error("after home, view should show Root")
	}
}
