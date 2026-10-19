// Form example — demonstrates interactive form with validation using the App framework.
//
// This example shows multiple form types (registration, contact, survey) accessible
// through a menu. Each form has custom validation, error handling, and submission flow.
//
// Navigation:
//
//	↑↓/jk         — navigate menu
//	Enter         — select form type
//	Tab/Down      — move to next field (in form)
//	Shift+Tab/Up  — move to previous field (in form)
//	Enter         — submit form (when valid)
//	b/esc         — back to menu
//	q/Ctrl+C      — quit
package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
)

var (
	mainMenu   = NewFormMenu()
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^\+?[0-9]{10,15}$`)
)

// ── FormMenu ────────────────────────────────────────────────────────

type FormMenu struct {
	model.DefaultMenu
	menus []model.MenuItem
}

func NewFormMenu() *FormMenu {
	return &FormMenu{
		menus: []model.MenuItem{
			{Title: "User Registration", Subtitle: "Sign up with name, email, age, username"},
			{Title: "Contact Form", Subtitle: "Get in touch: name, email, phone, message"},
			{Title: "Survey Form", Subtitle: "Quick feedback: rating, comments, recommend"},
		},
	}
}

func (m *FormMenu) GetMenuKey() string {
	return "form_menu"
}

func (m *FormMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *FormMenu) IsSearchable() bool {
	return true
}

func (m *FormMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

func (m *FormMenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	// Navigate to form page based on selection
	switch index {
	case 0:
		return NewFormPage(app, "registration"), nil
	case 1:
		return NewFormPage(app, "contact"), nil
	case 2:
		return NewFormPage(app, "survey"), nil
	}
	return nil, nil
}

// ── FormPage ────────────────────────────────────────────────────────

type FormPage struct {
	app      *model.App
	formType string
	form     *model.Form
	width    int
	height   int
}

func NewFormPage(app *model.App, formType string) *FormPage {
	p := &FormPage{
		app:      app,
		formType: formType,
	}

	// Create form based on type
	switch formType {
	case "registration":
		p.form = p.createRegistrationForm()
	case "contact":
		p.form = p.createContactForm()
	case "survey":
		p.form = p.createSurveyForm()
	}

	p.form.Focus()
	return p
}

func (p *FormPage) createRegistrationForm() *model.Form {
	return model.NewForm([]model.FormField{
		{
			Key:         "name",
			Label:       "Full Name",
			Placeholder: "John Doe",
			Required:    true,
		},
		{
			Key:         "email",
			Label:       "Email",
			Placeholder: "john@example.com",
			Required:    true,
			Validate: func(value string) error {
				if !emailRegex.MatchString(value) {
					return fmt.Errorf("invalid email format")
				}
				return nil
			},
		},
		{
			Key:         "age",
			Label:       "Age",
			Placeholder: "25",
			Required:    true,
			Validate: func(value string) error {
				age, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("must be a number")
				}
				if age < 18 || age > 120 {
					return fmt.Errorf("must be between 18 and 120")
				}
				return nil
			},
		},
		{
			Key:         "username",
			Label:       "Username",
			Placeholder: "johndoe",
			Required:    true,
			Validate: func(value string) error {
				if len(value) < 3 {
					return fmt.Errorf("must be at least 3 characters")
				}
				if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(value) {
					return fmt.Errorf("only letters, numbers, and underscore")
				}
				return nil
			},
		},
		{
			Key:         "bio",
			Label:       "Bio",
			Placeholder: "Tell us about yourself...",
			Required:    false,
		},
	})
}

func (p *FormPage) createContactForm() *model.Form {
	return model.NewForm([]model.FormField{
		{
			Key:         "name",
			Label:       "Name",
			Placeholder: "Your name",
			Required:    true,
		},
		{
			Key:         "email",
			Label:       "Email",
			Placeholder: "you@example.com",
			Required:    true,
			Validate: func(value string) error {
				if !emailRegex.MatchString(value) {
					return fmt.Errorf("invalid email format")
				}
				return nil
			},
		},
		{
			Key:         "phone",
			Label:       "Phone",
			Placeholder: "+1234567890",
			Required:    false,
			Validate: func(value string) error {
				if !phoneRegex.MatchString(value) {
					return fmt.Errorf("invalid phone number (10-15 digits)")
				}
				return nil
			},
		},
		{
			Key:         "subject",
			Label:       "Subject",
			Placeholder: "How can we help?",
			Required:    true,
		},
		{
			Key:         "message",
			Label:       "Message",
			Placeholder: "Your message here...",
			Required:    true,
			Validate: func(value string) error {
				if len(value) < 10 {
					return fmt.Errorf("message must be at least 10 characters")
				}
				return nil
			},
		},
	})
}

func (p *FormPage) createSurveyForm() *model.Form {
	return model.NewForm([]model.FormField{
		{
			Key:         "rating",
			Label:       "Rating",
			Placeholder: "1-5",
			Required:    true,
			Validate: func(value string) error {
				rating, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("must be a number")
				}
				if rating < 1 || rating > 5 {
					return fmt.Errorf("must be between 1 and 5")
				}
				return nil
			},
		},
		{
			Key:         "experience",
			Label:       "Experience",
			Placeholder: "excellent/good/fair/poor",
			Required:    true,
			Validate: func(value string) error {
				valid := map[string]bool{"excellent": true, "good": true, "fair": true, "poor": true}
				if !valid[strings.ToLower(value)] {
					return fmt.Errorf("must be: excellent, good, fair, or poor")
				}
				return nil
			},
		},
		{
			Key:         "recommend",
			Label:       "Recommend?",
			Placeholder: "yes/no",
			Required:    true,
			Validate: func(value string) error {
				lower := strings.ToLower(value)
				if lower != "yes" && lower != "no" {
					return fmt.Errorf("must be yes or no")
				}
				return nil
			},
		},
		{
			Key:         "comments",
			Label:       "Comments",
			Placeholder: "Additional feedback (optional)",
			Required:    false,
		},
	})
}

func (p *FormPage) Type() model.PageType {
	return model.PageType("form_" + p.formType)
}

func (p *FormPage) Msg() tea.Msg {
	return nil
}

func (p *FormPage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return false
}

func (p *FormPage) Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
	p.app = a

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "esc", "b", "B":
			if p.form.Submitted() {
				// Reset and go back
				p.form.Reset()
			}
			return a.Main(), nil
		}

		// Forward to form widget
		cmd := p.form.Update(msg)

		// Check if form was just submitted
		if p.form.Submitted() {
			// Trigger rerender to show success
			return p, a.RerenderCmd(true)
		}

		return p, cmd

	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.form.SetSize(p.width-4, p.height-12)
		return p, nil
	}

	return p, nil
}

func (p *FormPage) View(a *model.App) string {
	w, h := a.WindowWidth(), a.WindowHeight()
	if w <= 0 || h <= 0 {
		return ""
	}

	// Set widget size based on current window dimensions
	p.form.SetSize(w-4, h-12)

	ss := style.CurrentStyleSet()

	if p.form.Submitted() {
		return p.renderSuccess(ss)
	}
	return p.renderForm(ss)
}

func (p *FormPage) renderForm(ss style.StyleSet) string {
	// Header based on form type
	var header, subtitle string
	switch p.formType {
	case "registration":
		header = "📝 User Registration"
		subtitle = "Create your account. Fields marked with * are required."
	case "contact":
		header = "💬 Contact Us"
		subtitle = "We'd love to hear from you. Fill out the form below."
	case "survey":
		header = "📊 Quick Survey"
		subtitle = "Help us improve. Share your feedback."
	}

	title := ss.Title.Render(header)
	sub := ss.Muted.Render(subtitle)
	controls := ss.Muted.Render("Tab/↓: next • Shift+Tab/↑: previous • Enter: submit • b/esc: back")

	// Form content in a bordered box
	formView := p.form.View()
	formBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ss.Border.GetForeground()).
		Padding(1, 2).
		Width(p.width - 6).
		Render(formView)

	// Status footer
	var status string
	if p.form.IsValid() {
		status = ss.Success.Render("✓ Ready to submit — press Enter")
	} else {
		status = ss.Muted.Render("Complete all required fields to submit")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		sub,
		"",
		formBox,
		"",
		status,
		"",
		controls,
	)

	// Center if window is large enough
	if p.width > 80 {
		content = lipgloss.NewStyle().Width(p.width).Align(lipgloss.Center).Render(content)
	}

	return content
}

func (p *FormPage) renderSuccess(ss style.StyleSet) string {
	values := p.form.Values()

	header := ss.Success.Render("✓ Submission Successful!")
	divider := ss.Border.Render(strings.Repeat("─", 60))

	var details []string
	details = append(details, ss.Subtitle.Render("Submitted Information:"))
	details = append(details, "")

	// Format based on form type
	switch p.formType {
	case "registration":
		details = append(details, formatField(ss, "Full Name", values["name"]))
		details = append(details, formatField(ss, "Email", values["email"]))
		details = append(details, formatField(ss, "Age", values["age"]))
		details = append(details, formatField(ss, "Username", values["username"]))
		if values["bio"] != "" {
			details = append(details, formatField(ss, "Bio", values["bio"]))
		}
	case "contact":
		details = append(details, formatField(ss, "Name", values["name"]))
		details = append(details, formatField(ss, "Email", values["email"]))
		if values["phone"] != "" {
			details = append(details, formatField(ss, "Phone", values["phone"]))
		}
		details = append(details, formatField(ss, "Subject", values["subject"]))
		details = append(details, formatField(ss, "Message", values["message"]))
	case "survey":
		details = append(details, formatField(ss, "Rating", values["rating"]+"/5"))
		details = append(details, formatField(ss, "Experience", values["experience"]))
		details = append(details, formatField(ss, "Recommend", values["recommend"]))
		if values["comments"] != "" {
			details = append(details, formatField(ss, "Comments", values["comments"]))
		}
	}

	hint := ss.Muted.Render("Press b/esc to return to menu")

	content := lipgloss.JoinVertical(lipgloss.Left, details...)
	contentBox := lipgloss.NewStyle().
		Padding(1, 3).
		Width(p.width - 6).
		Render(content)

	result := lipgloss.JoinVertical(lipgloss.Left,
		"",
		header,
		divider,
		contentBox,
		"",
		hint,
	)

	if p.width > 80 {
		result = lipgloss.NewStyle().Width(p.width).Align(lipgloss.Center).Render(result)
	}

	return result
}

func formatField(ss style.StyleSet, label, value string) string {
	labelWidth := 12
	paddedLabel := lipgloss.NewStyle().Width(labelWidth).Align(lipgloss.Right).Render(label + ":")
	return ss.MenuItem.Render(paddedLabel) + "  " + ss.Normal.Render(value)
}

// ── main ────────────────────────────────────────────────────────────

func main() {
	ops := model.DefaultOptions()
	ops.EnableStartup = false
	ops.WhetherDisplayTitle = true
	ops.AppName = "Form Demo"
	ops.Ticker = model.DefaultTicker(500 * time.Millisecond)

	app := model.NewApp(ops)
	app.With(model.WithMainMenu(mainMenu, nil))

	fmt.Println(app.Run())
}
