package version

import "testing"

func TestCompareVersion(t *testing.T) {
	for _, c := range []struct {
		v1, v2 string
		equal  bool
		want   bool
	}{
		{"v1.10.0", "v1.9.0", false, true},
		{"v1.9.0", "v1.10.0", false, false},
		{"v2.0.0", "v1.99.99", false, true},
		{"v5.1.0", "v1.0.0", false, true},
		{"v1.0.0", "v5.1.0", false, false},
		{"v1.0.1", "v1.0.0", false, true},
		{"v1.0.0", "v1.0.0", false, false},
		{"v1.0.0", "v1.0.0", true, true},
		{"1.0.0", "v1.0.0", true, true},
		{"v1.0.0", "", false, true},
		// A build a few commits past a release is not older than it.
		{"v1.0.0", "v1.0.0-3-gabc1234", false, false},
		{"v1.0.0-dev", "v9.0.0", false, true},
		{"v9.0.0", "v1.0.0-dev", false, false},
	} {
		if got := CompareVersion(c.v1, c.v2, c.equal); got != c.want {
			t.Errorf("CompareVersion(%q, %q, %v) = %v, want %v", c.v1, c.v2, c.equal, got, c.want)
		}
	}
}
