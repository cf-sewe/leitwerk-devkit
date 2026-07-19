package gate

import (
	"bufio"
	"bytes"
	pathpkg "path"
	"strings"
)

// pathRule is one "glob = tier" entry from the [paths] table, in file order.
type pathRule struct {
	glob string
	tier string
}

// Tiers is a parsed tiers file: the [tiers] table (tier → checks), the ordered
// [paths] table (glob → tier, first match wins), and the [human-owned] glob list.
type Tiers struct {
	tierChecks map[string]string
	paths      []pathRule
	humanOwned []string
}

// ParseTiers reads the INI-ish tiers format. It never errors: unknown sections,
// blank lines, and comments are ignored, so a malformed file yields an empty (but
// usable) table rather than a panic. Comments are lines whose first non-space rune
// is '#'.
func ParseTiers(data []byte) *Tiers {
	t := &Tiers{tierChecks: map[string]string{}}
	section := ""
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			section = strings.TrimSpace(trimmed[1 : len(trimmed)-1])
			continue
		}
		switch section {
		case "tiers":
			key, val, ok := splitKV(trimmed)
			if ok {
				t.tierChecks[key] = val
			}
		case "paths":
			glob, val, ok := splitKV(trimmed)
			if !ok || glob == "" {
				continue
			}
			// The tier is the first token of the value (tolerates trailing text).
			if tier := firstField(val); tier != "" {
				t.paths = append(t.paths, pathRule{glob: glob, tier: tier})
			}
		case "human-owned":
			if g := firstField(trimmed); g != "" {
				t.humanOwned = append(t.humanOwned, g)
			}
		}
	}
	return t
}

// splitKV splits "key = value" on the first '=', trimming surrounding space.
func splitKV(line string) (key, val string, ok bool) {
	i := strings.IndexByte(line, '=')
	if i < 0 {
		return "", "", false
	}
	return strings.TrimSpace(line[:i]), strings.TrimSpace(line[i+1:]), true
}

// firstField returns the first whitespace-delimited token of s, or "".
func firstField(s string) string {
	f := strings.Fields(s)
	if len(f) == 0 {
		return ""
	}
	return f[0]
}

// ChecksForTier returns the (space-separated) check names configured for tier, in
// file order. An unknown tier yields an empty slice.
func (t *Tiers) ChecksForTier(tier string) []string {
	return strings.Fields(t.tierChecks[tier])
}

// TierForPath returns the tier for path using the [paths] table (first matching
// glob wins). ok is false when no glob matches; callers default to T1.
func (t *Tiers) TierForPath(path string) (tier string, ok bool) {
	for _, r := range t.paths {
		if matchTier(r.glob, path) {
			return r.tier, true
		}
	}
	return "", false
}

// HumanOwnedMatch reports the first [human-owned] glob that path suffix-matches.
// The path is canonicalized first so filesystem-equivalent spellings
// (leitwerk//x, leitwerk/./x, a/../leitwerk/x) cannot slip past the textual match.
func (t *Tiers) HumanOwnedMatch(path string) (glob string, ok bool) {
	path = pathpkg.Clean(path)
	for _, g := range t.humanOwned {
		if matchGuard(g, path) {
			return g, true
		}
	}
	return "", false
}
