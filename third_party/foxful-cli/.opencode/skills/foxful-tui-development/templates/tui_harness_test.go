package model

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// skillTUIHarness 通过真实 App.Update/App.View 驱动 Page。
// 将本文件复制到 model/ 后，在同包测试中按目标场景改写示例 Page 或 buildPage。
type skillTUIHarness struct {
	t       *testing.T
	App     *App
	Width   int
	Height  int
	LastCmd tea.Cmd
}

func newSkillTUIHarness(
	t *testing.T,
	width, height int,
	configure func(*Options),
	buildPage func(*App, *Options) Page,
) *skillTUIHarness {
	t.Helper()
	if width <= 0 || height <= 0 {
		t.Fatalf("terminal size must be positive, got %dx%d", width, height)
	}

	options := DefaultOptions()
	options.EnableStartup = false
	if configure != nil {
		configure(options)
	}

	app := NewApp(options)
	page := buildPage(app, options)
	if page == nil {
		t.Fatal("buildPage returned nil")
	}
	app.setPage(page)

	h := &skillTUIHarness{t: t, App: app, Width: width, Height: height}
	h.LastCmd = h.Step(tea.WindowSizeMsg{Width: width, Height: height})
	return h
}

// Step 只发送一条消息，不隐式执行返回的 Cmd。
func (h *skillTUIHarness) Step(msg tea.Msg) tea.Cmd {
	h.t.Helper()
	_, cmd := h.App.Update(msg)
	h.LastCmd = cmd
	return cmd
}

// Drain 显式执行有限 Cmd 链。它展开公开的 tea.BatchMsg，并给每条 Cmd 设置超时。
// ticker、无限命令链和需要并发顺序语义的场景应由测试直接控制消息。
func (h *skillTUIHarness) Drain(cmd tea.Cmd, maxSteps int, perCmdTimeout time.Duration) []tea.Msg {
	h.t.Helper()
	if cmd == nil {
		return nil
	}
	if maxSteps <= 0 || perCmdTimeout <= 0 {
		h.t.Fatalf("Drain limits must be positive, got maxSteps=%d timeout=%s", maxSteps, perCmdTimeout)
	}

	queue := []tea.Cmd{cmd}
	messages := make([]tea.Msg, 0, maxSteps)
	for len(queue) > 0 && len(messages) < maxSteps {
		current := queue[0]
		queue = queue[1:]
		if current == nil {
			continue
		}

		result := make(chan tea.Msg, 1)
		go func() { result <- current() }()

		var msg tea.Msg
		select {
		case msg = <-result:
		case <-time.After(perCmdTimeout):
			h.t.Fatalf("Cmd exceeded timeout %s after %d message(s)", perCmdTimeout, len(messages))
		}

		switch typed := msg.(type) {
		case nil:
			continue
		case tea.BatchMsg:
			queue = append(queue, typed...)
		default:
			messages = append(messages, typed)
			if next := h.Step(typed); next != nil {
				queue = append(queue, next)
			}
		}
	}
	if len(queue) > 0 {
		h.t.Fatalf("Cmd chain exceeded maxSteps=%d", maxSteps)
	}
	return messages
}

func (h *skillTUIHarness) View() string {
	h.t.Helper()
	return h.App.View().Content
}

func (h *skillTUIHarness) PlainView() string {
	h.t.Helper()
	return ansi.Strip(h.View())
}

func (h *skillTUIHarness) ViewSize() (width, height int) {
	h.t.Helper()
	view := h.View()
	return lipgloss.Width(view), lipgloss.Height(view)
}

func skillTextKey(text string) tea.KeyPressMsg {
	runes := []rune(text)
	if len(runes) == 0 {
		return tea.KeyPressMsg{}
	}
	return tea.KeyPressMsg(tea.Key{Code: runes[0], Text: text})
}

func skillSpecialKey(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}

func skillClick(x, y int, button tea.MouseButton) tea.MouseClickMsg {
	return tea.MouseClickMsg(tea.Mouse{X: x, Y: y, Button: button})
}

func skillWheel(x, y int, button tea.MouseButton) tea.MouseWheelMsg {
	return tea.MouseWheelMsg(tea.Mouse{X: x, Y: y, Button: button})
}

// 以下 Page 与测试证明模板复制后能真实经过 App 的消息路由和渲染出口。
type skillHarnessPage struct {
	width   int
	height  int
	lastKey string
}

func (p *skillHarnessPage) IgnoreQuitKeyMsg(tea.KeyMsg) bool { return false }
func (p *skillHarnessPage) Type() PageType                   { return PageType("skill-harness") }
func (p *skillHarnessPage) Msg() tea.Msg                     { return nil }

func (p *skillHarnessPage) Update(msg tea.Msg, _ *App) (Page, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		p.width, p.height = typed.Width, typed.Height
	case tea.KeyPressMsg:
		p.lastKey = typed.String()
	}
	return p, nil
}

func (p *skillHarnessPage) View(_ *App) string {
	return fmt.Sprintf("size=%dx%d key=%s", p.width, p.height, p.lastKey)
}

func TestSkillTUIHarnessRoutesAppMessages(t *testing.T) {
	h := newSkillTUIHarness(t, 40, 12, nil, func(_ *App, _ *Options) Page {
		return &skillHarnessPage{}
	})
	if got := h.PlainView(); !strings.Contains(got, "size=40x12") {
		t.Fatalf("initial App view = %q, want routed WindowSize", got)
	}

	cmd := func() tea.Msg { return skillTextKey("x") }
	messages := h.Drain(cmd, 4, 250*time.Millisecond)
	if len(messages) != 1 {
		t.Fatalf("drained messages = %d, want 1", len(messages))
	}
	if got := h.PlainView(); !strings.Contains(got, "key=x") {
		t.Fatalf("App view after drained key = %q, want key=x", got)
	}

	width, height := h.ViewSize()
	if width == 0 || height == 0 {
		t.Fatalf("rendered size = %dx%d, want non-zero terminal cells", width, height)
	}
}
