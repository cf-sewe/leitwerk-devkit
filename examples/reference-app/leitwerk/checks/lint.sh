#!/usr/bin/env bash
# reference-app's own `lint` check — a repo-local override of the built-in. Uses
# gofmt (part of the Go toolchain) so the result is stable across environments,
# rather than depending on whether golangci-lint happens to be installed.
# Resolves Go via mise (the repo's pinned toolchain) then a working PATH gofmt.
#   exit 0 = formatted, 1 = unformatted files, 2 = no Go toolchain (skip).
set -euo pipefail

if command -v mise >/dev/null 2>&1 && mise exec -- go version >/dev/null 2>&1; then
  GOFMT=(mise exec -- gofmt)
elif command -v gofmt >/dev/null 2>&1 && gofmt -h >/dev/null 2>&1; then
  GOFMT=(gofmt)
else
  echo "no Go toolchain for gofmt"; exit 2
fi

# Capture gofmt's exit explicitly: it exits non-zero (e.g. 2) on an unparseable
# file. Under `set -e` an uncaptured non-zero would propagate and be read as a
# skip (exit 2) — a check faking a pass. Map any gofmt error to a red (exit 1);
# exit 2 is reserved for the "no toolchain" branch above.
rc=0
unformatted="$("${GOFMT[@]}" -l . 2>&1)" || rc=$?
if [ "$rc" -ne 0 ]; then
  echo "gofmt could not process the tree (parse error?):"; echo "$unformatted"; exit 1
fi
if [ -n "$unformatted" ]; then
  echo "gofmt: these files need formatting:"; echo "$unformatted"; exit 1
fi
echo "gofmt clean"
