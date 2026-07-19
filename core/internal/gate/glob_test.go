package gate

import "testing"

func TestGlobToRegex(t *testing.T) {
	cases := []struct {
		glob string
		want string
	}{
		{"core/bin/**", "core/bin/.*"},
		{"core/checks/**", "core/checks/.*"},
		{"**/*.sh", `(.*/)?[^/]*\.sh`},
		{"**/*.md", `(.*/)?[^/]*\.md`},
		{"*.md", `[^/]*\.md`},
		{"*", ".*"}, // bare star is the catch-all
		{"**", ".*"},
		{".github/**", `\.github/.*`},
		{"**/db/migrations/**", `(.*/)?db/migrations/.*`},
		{"infra/**", "infra/.*"},
		{"docs/**", "docs/.*"},
	}
	for _, c := range cases {
		if got := GlobToRegex(c.glob); got != c.want {
			t.Errorf("GlobToRegex(%q) = %q, want %q", c.glob, got, c.want)
		}
	}
}

func TestMatchTier(t *testing.T) {
	cases := []struct {
		glob, path string
		want       bool
	}{
		{"docs/**", "docs/guide.md", true},
		{"docs/**", "docsx/guide.md", false},       // segment boundary
		{"**/*.md", "docs/guide.md", true},         // optional leading segments
		{"**/*.md", "README.md", true},             // zero leading segments
		{"**/*.md", "docs/deep/nested/x.md", true}, // (.*/)? spans separators
		{"*.md", "guide.md", true},                 // single segment
		{"*.md", "docs/guide.md", false},           // * does not cross '/'
		{"**/db/migrations/**", "db/migrations/001.sql", true},
		{"**/db/migrations/**", "app/db/migrations/001.sql", true},
		{"**/db/migrations/**", "db/other/001.sql", false},
		{"*", "anything/at/all", true}, // catch-all
		{"core/bin/**", "core/bin/leitwerk", true},
		{"core/bin/**", "core/checks/x.sh", false},
	}
	for _, c := range cases {
		if got := matchTier(c.glob, c.path); got != c.want {
			t.Errorf("matchTier(%q, %q) = %v, want %v", c.glob, c.path, got, c.want)
		}
	}
}

func TestMatchGuard(t *testing.T) {
	cases := []struct {
		glob, path string
		want       bool
	}{
		{"leitwerk/constitution.md", "leitwerk/constitution.md", true},
		{"leitwerk/tiers.conf", "/abs/path/to/leitwerk/tiers.conf", true}, // absolute-path suffix
		{"leitwerk/constitution.md", "src/app.py", false},
		{"leitwerk/constitution.md", "notleitwerk/constitution.md", false}, // needs ^ or / before
		{"leitwerk/roadmap.md", "a/b/leitwerk/roadmap.md", true},
	}
	for _, c := range cases {
		if got := matchGuard(c.glob, c.path); got != c.want {
			t.Errorf("matchGuard(%q, %q) = %v, want %v", c.glob, c.path, got, c.want)
		}
	}
}

// A malformed glob that translates to an invalid regex must be a non-match, not a
// panic (robustness on a corrupted tiers file).
func TestMatchInvalidGlobIsNotAPanic(t *testing.T) {
	if matchTier("a(b", "a(b") {
		t.Errorf("expected invalid-regex glob to not match")
	}
}
