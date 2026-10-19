package forkcheck

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	foxfulModule = "github.com/anhoder/foxful-cli"
	thirdParty   = "../../third_party/foxful-cli"
	vendored     = "../../vendor/github.com/anhoder/foxful-cli"
)

// The default build reads vendor/, so an edit in third_party/ has no effect
// until `make vendor` copies it over.
func TestVendoredFoxfulMatchesThirdParty(t *testing.T) {
	dirs := map[string]bool{}
	err := filepath.WalkDir(vendored, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(vendored, path)
		dirs[filepath.Dir(rel)] = true
		got, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		want, err := os.ReadFile(filepath.Join(thirdParty, rel))
		if err != nil || !bytes.Equal(got, want) {
			t.Errorf("vendor/%s/%s is stale; run make vendor", foxfulModule, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for dir := range dirs {
		entries, err := os.ReadDir(filepath.Join(thirdParty, dir))
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			if _, err := os.Stat(filepath.Join(vendored, dir, name)); err != nil {
				t.Errorf("third_party/foxful-cli/%s/%s is not vendored; run make vendor", dir, name)
			}
		}
	}
}

// Upstream go-musicfox bumps the foxful-cli version from time to time; such a
// merge must also move the in-tree base, or the replace would silently keep
// the old framework.
func TestFoxfulBaseVersion(t *testing.T) {
	gomod, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	upstream, err := os.ReadFile(filepath.Join(thirdParty, "UPSTREAM.md"))
	if err != nil {
		t.Fatal(err)
	}
	required := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(foxfulModule) + `\s+(v\S+)`).FindSubmatch(gomod)
	base := regexp.MustCompile(`(?m)^Base:\s*(v\S+)`).FindSubmatch(upstream)
	if required == nil || base == nil {
		t.Fatalf("cannot find the foxful-cli version in go.mod (%v) or UPSTREAM.md (%v)", required != nil, base != nil)
	}
	if !bytes.Equal(required[1], base[1]) {
		t.Errorf("go.mod requires %s %s but third_party/foxful-cli is based on %s; merge the upstream changes and update UPSTREAM.md",
			foxfulModule, required[1], base[1])
	}
}
