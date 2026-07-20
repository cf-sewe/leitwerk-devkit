#!/usr/bin/env bash
# reference-app's own `types` check — a repo-local override of the built-in.
# Runs `go vet`, resolving Go via mise then a working PATH `go`, else skips.
#   exit 0 = clean, 1 = vet problems, 2 = no working Go toolchain (skip).
set -euo pipefail

if command -v mise >/dev/null 2>&1 && mise exec -- go version >/dev/null 2>&1; then
  GO=(mise exec -- go)
elif command -v go >/dev/null 2>&1 && go version >/dev/null 2>&1; then
  GO=(go)
else
  echo "no working Go toolchain (install Go or mise)"; exit 2
fi

"${GO[@]}" vet ./... && echo "go vet clean"
