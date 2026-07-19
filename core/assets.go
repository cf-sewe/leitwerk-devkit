// Package assets bundles the gate's checks, templates, and default tiers file
// into the compiled binary via go:embed. This makes core/bin/leitwerk a real
// single static artifact: it can scaffold a repo (`init`) and fall back to its
// built-in checks even when installed with no sibling checks/ or templates/ on
// disk (e.g. via `go install`, which copies only the binary). When those
// directories DO sit next to the binary (a full checkout, or a release tarball),
// the on-disk copies win — the embedded copies are the fallback, so this repo's
// own runs and the golden selftest exercise the on-disk scripts unchanged.
package assets

import "embed"

//go:embed checks templates leitwerk.tiers
var FS embed.FS
