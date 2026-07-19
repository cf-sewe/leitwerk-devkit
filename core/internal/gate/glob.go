// Package gate is the tool-agnostic verification gate: glob→regex translation,
// tiers parsing, tier/human-owned matching, check resolution, and the verify
// runner. It is deliberately I/O-light and dependency-free (stdlib only) so every
// piece is unit-testable and the resulting binary is a single static artifact.
package gate

import (
	"regexp"
	"runtime"
	"strings"
)

// Sentinels used while translating a glob so the single-star pass does not
// re-mangle the regex the double-star passes just produced. NUL-wrapped so they
// can never collide with any real glob text, and all are removed before return.
const (
	sentDblSlash = "\x00DBLSLASH\x00"
	sentDblStar  = "\x00DBLSTAR\x00"
)

// GlobToRegex translates a path glob into the body of an anchored regular
// expression, matching the historical awk engine exactly:
//
//	.     -> \.          (only dots are escaped)
//	**/   -> (.*/)?      (optional leading path segments)
//	**    -> .*          (anything, across separators)
//	*     -> [^/]*       (one path segment)
//	bare * -> .*         (a lone "*" is the catch-all)
//
// The caller anchors the result: "^"+body+"$" for tier lookup, "(^|/)"+body+"$"
// for the human-owned suffix match.
func GlobToRegex(glob string) string {
	g := strings.ReplaceAll(glob, ".", `\.`)
	g = strings.ReplaceAll(g, "**/", sentDblSlash)
	g = strings.ReplaceAll(g, "**", sentDblStar)
	g = strings.ReplaceAll(g, "*", "[^/]*")
	g = strings.ReplaceAll(g, sentDblSlash, "(.*/)?")
	g = strings.ReplaceAll(g, sentDblStar, ".*")
	if g == "[^/]*" { // the original glob was a lone "*"
		g = ".*"
	}
	return g
}

// globMatches reports whether path matches glob once anchored with prefix.
// A glob that translates to an invalid regex (a malformed tiers file) is treated
// as a non-match rather than a panic, keeping the gate robust on bad input.
func globMatches(glob, path, prefix string) bool {
	re, err := regexp.Compile(prefix + GlobToRegex(glob) + "$")
	if err != nil {
		return false
	}
	return re.MatchString(path)
}

// matchTier anchors with ^…$ (a path is classified by a full match).
func matchTier(glob, path string) bool { return globMatches(glob, path, "^") }

// matchGuard anchors with (^|/)…$ so an absolute path from a hook payload still
// suffix-matches a human-owned glob. On a case-insensitive filesystem it also
// matches case-insensitively, so a differently-cased spelling of a human-owned
// file (which is the same inode there) cannot slip past the guard. The path is
// expected to be pre-canonicalized by the caller (see normalizeGuardPath).
func matchGuard(glob, path string) bool {
	prefix := "(^|/)"
	if caseInsensitiveFS() {
		prefix = "(?i)" + prefix
	}
	return globMatches(glob, path, prefix)
}

// caseInsensitiveFS reports whether the default filesystem treats paths
// case-insensitively (macOS APFS/HFS+ and Windows). On case-sensitive systems
// (typical Linux) a differently-cased path is a genuinely different file, so the
// guard stays case-sensitive there.
func caseInsensitiveFS() bool {
	return runtime.GOOS == "darwin" || runtime.GOOS == "windows"
}
