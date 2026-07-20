package gate

import "testing"

func TestRankTier(t *testing.T) {
	if !(RankTier("T0") < RankTier("T1") && RankTier("T1") < RankTier("T2")) {
		t.Fatalf("ladder must be T0<T1<T2, got %d %d %d", RankTier("T0"), RankTier("T1"), RankTier("T2"))
	}
	// An unrecognized tier ranks above T2 so --auto never under-verifies.
	if RankTier("TX") <= RankTier("T2") {
		t.Errorf("unknown tier rank = %d, want > T2 (%d)", RankTier("TX"), RankTier("T2"))
	}
}

func TestHighestTier(t *testing.T) {
	// sampleTiers (tiers_test.go): docs/**=T0, src/**=T1, **/*.go=T2, *=T1.
	tr := ParseTiers([]byte(sampleTiers))
	cases := []struct {
		name  string
		paths []string
		want  string
	}{
		{"empty set is T0", nil, "T0"},
		{"docs only is T0", []string{"docs/a.md", "docs/b.md"}, "T0"},
		{"app code is T1", []string{"docs/a.md", "src/app.py"}, "T1"},
		{"mixed picks the max (a .go file is T2)", []string{"docs/a.md", "src/app.py", "pkg/x.go"}, "T2"},
	}
	for _, c := range cases {
		if got, _ := HighestTier(tr, c.paths); got != c.want {
			t.Errorf("%s: HighestTier = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestHighestTierNamesDecidingPath(t *testing.T) {
	tr := ParseTiers([]byte(sampleTiers))
	got, deciding := HighestTier(tr, []string{"docs/a.md", "pkg/x.go", "src/app.py"})
	if got != "T2" || deciding != "pkg/x.go" {
		t.Errorf("HighestTier = (%q, %q), want (T2, pkg/x.go)", got, deciding)
	}
	// A T0-only set names no deciding path.
	if _, d := HighestTier(tr, []string{"docs/a.md"}); d != "" {
		t.Errorf("deciding path for T0-only set = %q, want empty", d)
	}
}

// With no catch-all rule, an unmatched path must fall back to T1 (like the `tier`
// command), NOT to the unknown-tier rank. This guards the `if !ok { pt = "T1" }`
// branch in HighestTier: a regression there gives pt="" → RankTier("")=99, which
// would out-rank a real T2 sibling and make HighestTier return "" — silently
// downgrading the run to T1 in cmdVerify. Every other tiers table in the suite
// has a `*` catch-all, so this is the only test that exercises the fallback.
func TestHighestTierNoCatchAllFallback(t *testing.T) {
	nc := ParseTiers([]byte("[paths]\ndocs/** = T0\n"))
	if got, _ := HighestTier(nc, []string{"docs/a.md", "unmatched.py"}); got != "T1" {
		t.Errorf("unmatched path, no catch-all: HighestTier = %q, want T1", got)
	}
	sql := ParseTiers([]byte("[paths]\n**/*.sql = T2\n"))
	if got, _ := HighestTier(sql, []string{"unmatched.py", "db/x.sql"}); got != "T2" {
		t.Errorf("unmatched + T2 sibling: HighestTier = %q, want T2 (unmatched must not out-rank)", got)
	}
}

// A path mapping to a non-standard tier name ranks above T2, so --auto errs
// toward more verification on an off-ladder tiers file (end-to-end, not just RankTier).
func TestHighestTierUnknownTierName(t *testing.T) {
	tr := ParseTiers([]byte("[paths]\nweird/** = TX\n*.md = T0\n"))
	if got, _ := HighestTier(tr, []string{"a.md", "weird/x"}); got != "TX" {
		t.Errorf("path on a non-standard tier: HighestTier = %q, want TX", got)
	}
}
