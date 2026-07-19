package gate

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

const sampleTiers = `# a comment
[tiers]
T0 = a
T1 = a b
T2 = a b c

[human-owned]
leitwerk/constitution.md
# comment inside a section
leitwerk/tiers.conf

[paths]
# glob = tier   (first match wins)
docs/**   = T0
src/**    = T1
**/*.go   = T2
*         = T1
`

func TestChecksForTier(t *testing.T) {
	tr := ParseTiers([]byte(sampleTiers))
	cases := []struct {
		tier string
		want []string
	}{
		{"T0", []string{"a"}},
		{"T1", []string{"a", "b"}},
		{"T2", []string{"a", "b", "c"}},
		{"T9", []string{}}, // unknown tier -> empty slice
	}
	for _, c := range cases {
		if got := tr.ChecksForTier(c.tier); !reflect.DeepEqual(got, c.want) {
			t.Errorf("ChecksForTier(%q) = %v, want %v", c.tier, got, c.want)
		}
	}
}

func TestTierForPath(t *testing.T) {
	tr := ParseTiers([]byte(sampleTiers))
	cases := []struct {
		path string
		want string
	}{
		{"docs/guide.md", "T0"},
		{"src/main.go", "T1"}, // src/** matches before **/*.go — first match wins
		{"pkg/util.go", "T2"}, // no src/ prefix, so **/*.go wins
		{"random.py", "T1"},   // catch-all
	}
	for _, c := range cases {
		got, ok := tr.TierForPath(c.path)
		if !ok || got != c.want {
			t.Errorf("TierForPath(%q) = (%q, %v), want (%q, true)", c.path, got, ok, c.want)
		}
	}
}

func TestTierForPathNoMatch(t *testing.T) {
	// A table without a catch-all leaves some paths unmatched (caller defaults T1).
	tr := ParseTiers([]byte("[paths]\ndocs/** = T0\n"))
	if got, ok := tr.TierForPath("src/app.py"); ok {
		t.Errorf("expected no match, got (%q, %v)", got, ok)
	}
}

func TestHumanOwnedMatch(t *testing.T) {
	tr := ParseTiers([]byte(sampleTiers))
	cases := []struct {
		path string
		want bool
	}{
		{"leitwerk/constitution.md", true},
		{"/repo/leitwerk/tiers.conf", true}, // absolute suffix
		{"src/app.py", false},
		{"leitwerk/roadmap.md", false}, // not in this sample's list
		// Filesystem-equivalent spellings must not bypass the guard (SEC-1):
		{"leitwerk//constitution.md", true},  // collapsed // separator
		{"leitwerk/./constitution.md", true}, // collapsed /./ component
		{"x/../leitwerk/tiers.conf", true},   // resolved ..
	}
	for _, c := range cases {
		_, ok := tr.HumanOwnedMatch(c.path)
		if ok != c.want {
			t.Errorf("HumanOwnedMatch(%q) ok = %v, want %v", c.path, ok, c.want)
		}
	}
}

// A differently-cased spelling of a human-owned file must be blocked on a
// case-insensitive filesystem (where it is the same inode) and only there.
func TestHumanOwnedMatchCaseFolding(t *testing.T) {
	tr := ParseTiers([]byte(sampleTiers))
	_, ok := tr.HumanOwnedMatch("leitwerk/Constitution.md")
	if ok != caseInsensitiveFS() {
		t.Errorf("case-folded guard match = %v, want %v (GOOS=%s)", ok, caseInsensitiveFS(), runtime.GOOS)
	}
}

func TestParseMalformedIsSafe(t *testing.T) {
	// Garbage input must not panic and must yield a usable (empty) table.
	for _, in := range []string{"", "no sections here\n===\n[unterminated", "[tiers]\ngarbage without equals\n"} {
		tr := ParseTiers([]byte(in))
		_ = tr.ChecksForTier("T2")
		_, _ = tr.TierForPath("x")
		_, _ = tr.HumanOwnedMatch("x")
	}
}

// The shipped default tiers file must carry the documented defaults, so a mutation
// to core/leitwerk.tiers is caught here as well as by the black-box selftest.
func TestShippedDefaults(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "leitwerk.tiers"))
	if err != nil {
		t.Fatalf("reading shipped tiers: %v", err)
	}
	tr := ParseTiers(data)
	if got := tr.ChecksForTier("T2"); !reflect.DeepEqual(got, []string{"lint", "types", "tests", "drift", "sast", "erosion"}) {
		t.Errorf("shipped T2 checks = %v", got)
	}
	for path, want := range map[string]string{
		"db/migrations/001.sql": "T2",
		"infra/main.tf":         "T2",
		"docs/guide.md":         "T0",
		"src/app.py":            "T1",
	} {
		if got, _ := tr.TierForPath(path); got != want {
			t.Errorf("shipped TierForPath(%q) = %q, want %q", path, got, want)
		}
	}
	if _, ok := tr.HumanOwnedMatch("leitwerk/constitution.md"); !ok {
		t.Errorf("shipped defaults should guard leitwerk/constitution.md")
	}
}
