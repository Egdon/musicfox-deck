package model

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNewForm(t *testing.T) {
	tests := []struct {
		name      string
		fields    []FormField
		wantCount int
	}{
		{
			name:      "empty form",
			fields:    []FormField{},
			wantCount: 0,
		},
		{
			name: "single field",
			fields: []FormField{
				{Key: "name", Label: "Name", Placeholder: "Enter name"},
			},
			wantCount: 1,
		},
		{
			name: "multiple fields",
			fields: []FormField{
				{Key: "name", Label: "Name"},
				{Key: "email", Label: "Email"},
				{Key: "age", Label: "Age"},
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := NewForm(tt.fields)
			if form == nil {
				t.Fatal("NewForm returned nil")
			}
			if len(form.fields) != tt.wantCount {
				t.Errorf("field count = %d, want %d", len(form.fields), tt.wantCount)
			}
			if len(form.inputs) != tt.wantCount {
				t.Errorf("input count = %d, want %d", len(form.inputs), tt.wantCount)
			}
			if form.focused {
				t.Error("new form should not be focused")
			}
			if form.submitted {
				t.Error("new form should not be submitted")
			}
		})
	}
}

func TestFormFocusBlur(t *testing.T) {
	fields := []FormField{
		{Key: "field1", Label: "Field 1"},
		{Key: "field2", Label: "Field 2"},
	}
	form := NewForm(fields)

	// Initially not focused
	if form.Focused() {
		t.Error("new form should not be focused")
	}

	// Focus
	form.Focus()
	if !form.Focused() {
		t.Error("form should be focused after Focus()")
	}
	if !form.inputs[0].Focused() {
		t.Error("first input should be focused")
	}

	// Blur
	form.Blur()
	if form.Focused() {
		t.Error("form should not be focused after Blur()")
	}
	for i, input := range form.inputs {
		if input.Focused() {
			t.Errorf("input %d should not be focused after Blur()", i)
		}
	}
}

func TestFormFocusNavigation(t *testing.T) {
	fields := []FormField{
		{Key: "field1", Label: "Field 1"},
		{Key: "field2", Label: "Field 2"},
		{Key: "field3", Label: "Field 3"},
	}
	form := NewForm(fields)
	form.Focus()

	tests := []struct {
		name       string
		keyCode    rune
		wantFocIdx int
	}{
		{"initial focus", 0, 0},
		{"tab to next", tea.KeyTab, 1},
		{"tab to next again", tea.KeyTab, 2},
		{"tab wraps to first", tea.KeyTab, 0},
		{"shift+tab to prev", tea.KeyTab, 2}, // will be modified below
		{"shift+tab to prev again", tea.KeyTab, 1},
		{"down acts like tab", tea.KeyDown, 2},
		{"up acts like shift+tab", tea.KeyUp, 1},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.keyCode != 0 {
				key := tea.Key{Code: tt.keyCode}
				// Handle shift+tab cases
				if i == 4 || i == 5 {
					key.Mod = tea.ModShift
				}
				form.Update(tea.KeyPressMsg(key))
			}
			if form.focusedIdx != tt.wantFocIdx {
				t.Errorf("focusedIdx = %d, want %d", form.focusedIdx, tt.wantFocIdx)
			}
		})
	}
}

func TestFormFocusWrapping(t *testing.T) {
	fields := []FormField{
		{Key: "a", Label: "A"},
		{Key: "b", Label: "B"},
	}
	form := NewForm(fields)
	form.Focus()

	// Start at field 0
	if form.focusedIdx != 0 {
		t.Fatalf("expected initial focus at 0, got %d", form.focusedIdx)
	}

	// Tab forward: 0 -> 1
	form.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	if form.focusedIdx != 1 {
		t.Errorf("after tab, focusedIdx = %d, want 1", form.focusedIdx)
	}

	// Tab forward again: 1 -> 0 (wrap)
	form.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	if form.focusedIdx != 0 {
		t.Errorf("after tab (wrap), focusedIdx = %d, want 0", form.focusedIdx)
	}

	// Shift+tab backward: 0 -> 1 (wrap backward)
	form.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift}))
	if form.focusedIdx != 1 {
		t.Errorf("after shift+tab (wrap), focusedIdx = %d, want 1", form.focusedIdx)
	}

	// Shift+tab backward: 1 -> 0
	form.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift}))
	if form.focusedIdx != 0 {
		t.Errorf("after shift+tab, focusedIdx = %d, want 0", form.focusedIdx)
	}
}

func TestFormValidationRequired(t *testing.T) {
	fields := []FormField{
		{Key: "name", Label: "Name", Required: true},
		{Key: "email", Label: "Email", Required: false},
	}
	form := NewForm(fields)
	form.Focus()

	// Initially valid (no validation run yet)
	if form.errors[0] != nil {
		t.Error("error should be nil before validation")
	}

	// Validate with empty required field
	form.validateField(0)
	if form.errors[0] == nil {
		t.Error("required field should have error when empty")
	}
	if form.IsValid() {
		t.Error("form should not be valid with empty required field")
	}

	// Set value and validate again
	form.inputs[0].SetValue("John")
	form.validateField(0)
	if form.errors[0] != nil {
		t.Errorf("error should be nil after setting value, got %v", form.errors[0])
	}

	// Non-required field should be valid when empty
	form.validateField(1)
	if form.errors[1] != nil {
		t.Errorf("non-required field should be valid when empty, got %v", form.errors[1])
	}
}

func TestFormValidationCustom(t *testing.T) {
	emailValidator := func(s string) error {
		if !strings.Contains(s, "@") {
			return errors.New("invalid email")
		}
		return nil
	}

	fields := []FormField{
		{Key: "email", Label: "Email", Validate: emailValidator},
	}
	form := NewForm(fields)
	form.Focus()

	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"empty passes custom validation", "", false},
		{"invalid email", "notanemail", true},
		{"valid email", "test@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form.inputs[0].SetValue(tt.value)
			form.validateField(0)
			hasError := form.errors[0] != nil
			if hasError != tt.wantError {
				t.Errorf("hasError = %v, want %v (value=%q, error=%v)",
					hasError, tt.wantError, tt.value, form.errors[0])
			}
		})
	}
}

func TestFormValidationOnBlur(t *testing.T) {
	fields := []FormField{
		{Key: "name", Label: "Name", Required: true},
		{Key: "email", Label: "Email", Required: true},
	}
	form := NewForm(fields)
	form.Focus()

	// Tab away from first field (should trigger validation)
	form.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	if form.errors[0] == nil {
		t.Error("validation should run on blur (tab away)")
	}

	// Set value and tab away again
	form.inputs[1].SetValue("test")
	form.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	if form.errors[1] != nil {
		t.Error("field with value should not have error after blur")
	}
}

func TestFormSubmit(t *testing.T) {
	fields := []FormField{
		{Key: "name", Label: "Name", Required: true},
		{Key: "age", Label: "Age", Required: false},
	}
	form := NewForm(fields)
	form.Focus()

	// Submit with invalid data
	form.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if form.Submitted() {
		t.Error("form should not submit when invalid")
	}

	// Fill required field and submit
	form.inputs[0].SetValue("Alice")
	form.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if !form.Submitted() {
		t.Error("form should submit when valid")
	}
}

func TestFormValues(t *testing.T) {
	fields := []FormField{
		{Key: "name", Label: "Name"},
		{Key: "email", Label: "Email"},
		{Key: "age", Label: "Age"},
	}
	form := NewForm(fields)
	form.Focus()

	// Set values
	form.inputs[0].SetValue("Alice")
	form.inputs[1].SetValue("alice@example.com")
	form.inputs[2].SetValue("30")

	values := form.Values()
	if len(values) != 3 {
		t.Errorf("Values() length = %d, want 3", len(values))
	}
	if values["name"] != "Alice" {
		t.Errorf("Values()[name] = %q, want %q", values["name"], "Alice")
	}
	if values["email"] != "alice@example.com" {
		t.Errorf("Values()[email] = %q, want %q", values["email"], "alice@example.com")
	}
	if values["age"] != "30" {
		t.Errorf("Values()[age] = %q, want %q", values["age"], "30")
	}
}

func TestFormReset(t *testing.T) {
	fields := []FormField{
		{Key: "name", Label: "Name", Required: true},
	}
	form := NewForm(fields)
	form.Focus()

	// Set value and submit
	form.inputs[0].SetValue("Alice")
	form.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	if !form.Submitted() {
		t.Fatal("form should be submitted")
	}

	// Reset
	form.Reset()
	if form.Submitted() {
		t.Error("submitted should be false after Reset()")
	}
	if form.inputs[0].Value() != "" {
		t.Errorf("field value should be empty after Reset(), got %q", form.inputs[0].Value())
	}
	if form.errors[0] != nil {
		t.Errorf("errors should be nil after Reset(), got %v", form.errors[0])
	}
}

func TestFormViewRendersNonEmpty(t *testing.T) {
	fields := []FormField{
		{Key: "name", Label: "Name", Placeholder: "Your name"},
		{Key: "email", Label: "Email", Placeholder: "your@email.com"},
	}
	form := NewForm(fields)
	form.Focus()
	form.SetSize(80, 25)

	view := form.View()
	if view == "" {
		t.Error("View() should not be empty")
	}
	if !strings.Contains(view, "Name") {
		t.Error("View() should contain label 'Name'")
	}
	if !strings.Contains(view, "Email") {
		t.Error("View() should contain label 'Email'")
	}
}

func TestFormViewEmptyFields(t *testing.T) {
	form := NewForm([]FormField{})
	view := form.View()
	if view != "" {
		t.Error("View() should be empty for form with no fields")
	}
}

func TestFormViewShowsErrors(t *testing.T) {
	fields := []FormField{
		{Key: "email", Label: "Email", Validate: func(s string) error {
			if s == "invalid" {
				return errors.New("bad email")
			}
			return nil
		}},
	}
	form := NewForm(fields)
	form.Focus()
	form.SetSize(80, 25)

	// Set invalid value and validate
	form.inputs[0].SetValue("invalid")
	form.validateField(0)

	view := form.View()
	if !strings.Contains(view, "bad email") {
		t.Error("View() should show validation error message")
	}
}

func TestFormSetSize(t *testing.T) {
	form := NewForm([]FormField{{Key: "test", Label: "Test"}})
	form.SetSize(100, 50)

	if form.width != 100 {
		t.Errorf("width = %d, want 100", form.width)
	}
	if form.height != 50 {
		t.Errorf("height = %d, want 50", form.height)
	}
}

func TestFormIsValid(t *testing.T) {
	fields := []FormField{
		{Key: "field1", Label: "Field 1", Required: true},
		{Key: "field2", Label: "Field 2", Required: true},
	}
	form := NewForm(fields)
	form.Focus()

	// All fields empty - invalid
	form.validateField(0)
	form.validateField(1)
	if form.IsValid() {
		t.Error("form should be invalid when required fields are empty")
	}

	// One field filled
	form.inputs[0].SetValue("test")
	form.validateField(0)
	if form.IsValid() {
		t.Error("form should be invalid when some required fields are empty")
	}

	// All fields filled
	form.inputs[1].SetValue("test2")
	form.validateField(1)
	if !form.IsValid() {
		t.Error("form should be valid when all required fields are filled")
	}
}

func TestFormUpdateUnfocused(t *testing.T) {
	form := NewForm([]FormField{{Key: "test", Label: "Test"}})
	// Don't focus the form
	cmd := form.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	if cmd != nil {
		t.Error("Update should return nil when form is not focused")
	}
}
