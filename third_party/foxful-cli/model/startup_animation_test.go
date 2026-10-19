package model

import (
	"strings"
	"testing"
	"time"

	"github.com/anhoder/foxful-cli/util"
	"github.com/charmbracelet/colorprofile"
)

func TestStartupAnimationProgress(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		loaded   time.Duration
		want     float64
	}{
		{name: "zero duration is complete", want: 1},
		{name: "clamps below zero", duration: time.Second, loaded: -time.Second, want: 0},
		{name: "half complete", duration: time.Second, loaded: 500 * time.Millisecond, want: .5},
		{name: "clamps after completion", duration: time.Second, loaded: 2 * time.Second, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StartupPage{options: &StartupOptions{LoadingDuration: tt.duration}, loadedDuration: tt.loaded}
			if got := s.animationProgress(); got != tt.want {
				t.Fatalf("animationProgress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStartupSequenceUsesFadeRainbowAndGlitchStages(t *testing.T) {
	oldProfile := util.TermProfile
	util.TermProfile = colorprofile.TrueColor
	t.Cleanup(func() { util.TermProfile = oldProfile })

	for _, tt := range []struct {
		name   string
		loaded time.Duration
		want   startupLogoEffect
	}{
		{name: "starts with fade", loaded: 200 * time.Millisecond, want: logoFade},
		{name: "continues with rainbow", loaded: 600 * time.Millisecond, want: logoRainbow},
		{name: "hands off through glitch", loaded: 900 * time.Millisecond, want: logoGlitch},
		{name: "settles on static logo", loaded: 970 * time.Millisecond, want: logoStatic},
	} {
		t.Run(tt.name, func(t *testing.T) {
			startup := StartupPage{
				options: &StartupOptions{
					Animation:       StartupAnimationSequence,
					LoadingDuration: time.Second,
				},
				loadedDuration: tt.loaded,
			}
			got, _, _ := startup.logoEffect()
			if got != tt.want {
				t.Fatalf("logoEffect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStartupAnimationDisablesMotionWhenRequested(t *testing.T) {
	s := StartupPage{options: &StartupOptions{ReducedMotion: true}}
	if s.animationEnabled() {
		t.Fatal("ReducedMotion must disable startup animation")
	}
}

func TestRenderStartupLogoFadeStartsBlank(t *testing.T) {
	if got := renderStartupLogo("AB", logoFade, 0, 0); strings.Contains(got, "A") || strings.Contains(got, "B") {
		t.Fatalf("zero-progress fade rendered logo: %q", got)
	}
}

func TestSpecialForegroundIncludesProgress(t *testing.T) {
	s := StartupPage{
		options: &StartupOptions{
			LoadingDuration: time.Second,
			TickDuration:    50 * time.Millisecond,
			Animation:       StartupAnimationMatrixRain,
			Welcome:         "FOX",
		},
		loadedDuration: 500 * time.Millisecond,
	}
	a := &App{windowWidth: 80, windowHeight: 24, options: DefaultOptions()}
	if got := s.specialForeground(a, 1, 1); !strings.Contains(got, "50%") {
		t.Fatalf("special foreground omitted progress: %q", got)
	}
}

func TestIndentLines(t *testing.T) {
	if got, want := indentLines("A\nB", 2), "  A\n  B"; got != want {
		t.Fatalf("indentLines() = %q, want %q", got, want)
	}
}

func TestStartupStatusPercent(t *testing.T) {
	for _, tt := range []struct {
		progress float64
		want     string
	}{
		{0, "0%"},
		{.375, "38%"},
		{1, "100%"},
	} {
		if got := formatStartupPercent(tt.progress); got != tt.want {
			t.Errorf("formatStartupPercent(%v) = %q, want %q", tt.progress, got, tt.want)
		}
	}
}
