package model

import (
	"sync"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestI18nLookup(t *testing.T) {
	const id MessageID = "greeting"

	tests := []struct {
		name     string
		locale   string
		fallback string
		messages map[string]map[MessageID]string
		want     string
	}{
		{
			name:   "exact locale",
			locale: "fr-CA",
			messages: map[string]map[MessageID]string{
				"fr-CA": {id: "Bonjour Canada"},
				"en":    {id: "Hello"},
			},
			want: "Bonjour Canada",
		},
		{
			name:   "fallback locale",
			locale: "fr",
			messages: map[string]map[MessageID]string{
				"en": {id: "Hello"},
			},
			want: "Hello",
		},
		{
			name:   "language only fallback",
			locale: "zh-CN",
			messages: map[string]map[MessageID]string{
				"zh": {id: "Ni hao"},
				"en": {id: "Hello"},
			},
			want: "Ni hao",
		},
		{
			name:   "missing message returns id",
			locale: "fr",
			want:   string(id),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog := NewCatalog()
			for locale, messages := range test.messages {
				catalog.Register(locale, messages)
			}
			catalog.SetLocale(test.locale)
			if test.fallback != "" {
				catalog.SetFallbackLocale(test.fallback)
			}

			if got := catalog.T(id); got != test.want {
				t.Fatalf("T(%q) = %q, want %q", id, got, test.want)
			}
		})
	}
}

func TestI18nTf(t *testing.T) {
	catalog := NewCatalog()
	catalog.Register("en", map[MessageID]string{"count": "%d items"})

	if got := catalog.Tf("count", 3); got != "3 items" {
		t.Fatalf("Tf() = %q, want %q", got, "3 items")
	}
}

func TestI18nRegisterOverrides(t *testing.T) {
	catalog := NewCatalog()
	catalog.Register("en", map[MessageID]string{"label": "First"})
	catalog.Register("en", map[MessageID]string{"label": "Second"})

	if got := catalog.T("label"); got != "Second" {
		t.Fatalf("T() = %q, want %q", got, "Second")
	}
}

func TestI18nLocaleRoundTrip(t *testing.T) {
	catalog := NewCatalog()
	catalog.SetLocale("pt-BR")

	if got := catalog.Locale(); got != "pt-BR" {
		t.Fatalf("Locale() = %q, want %q", got, "pt-BR")
	}
}

func TestI18nConcurrentAccess(t *testing.T) {
	catalog := NewCatalog()
	catalog.Register("en", map[MessageID]string{"message": "initial"})

	const workers = 8
	const iterations = 100
	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for iteration := 0; iteration < iterations; iteration++ {
				if worker%2 == 0 {
					catalog.Register("en", map[MessageID]string{"message": "updated"})
					catalog.SetLocale("en-US")
				} else {
					_ = catalog.T("message")
					_ = catalog.Tf("message")
					_ = catalog.Locale()
				}
			}
		}(worker)
	}
	wg.Wait()
}

func TestI18nWidgetIntegration(t *testing.T) {
	// Prove that widgets actually use i18n rather than hardcoded strings.
	// Register a Chinese translation, set locale, render a widget, confirm
	// the translated text appears.
	DefaultCatalog().Register("zh-CN", map[MessageID]string{
		MsgNoData: "无数据",
	})
	defer SetLocale("en") // reset to English after test

	SetLocale("zh-CN")

	// Table with no rows should render the translated "No data" message
	table := NewTable([]Column{{Title: "ID"}}, nil)
	table.SetSize(20, 5)
	view := table.View()
	plain := ansi.Strip(view)

	if plain != "无数据" {
		t.Errorf("Table empty state (stripped) = %q, want %q (translated MsgNoData)", plain, "无数据")
	}
}
