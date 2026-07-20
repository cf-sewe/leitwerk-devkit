#!/usr/bin/env bash
# reference-app's own `tests` check — a repo-local override of the built-in
# (which calls bare `go`). It resolves Go via mise (the repo's pinned toolchain)
# and falls back to a working `go` on PATH (e.g. CI's setup-go); if neither
# resolves it skips honestly rather than fake a pass.
#   exit 0 = tests pass, 1 = a test failed, 2 = no working Go toolchain (skip).
set -euo pipefail

if command -v mise >/dev/null 2>&1 && mise exec -- go version >/dev/null 2>&1; then
  GO=(mise exec -- go)
elif command -v go >/dev/null 2>&1 && go version >/dev/null 2>&1; then
  GO=(go)
else
  echo "no working Go toolchain (install Go or mise)"; exit 2
fi

"${GO[@]}" test ./... && echo "go test green"
