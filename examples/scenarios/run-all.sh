#!/usr/bin/env bash
# Runs every scenario against a built CLI. Used directly by humans and by the
# repo's `selftest` check, so the scenarios are CI-verified documentation.
# Exit 0 = all scenarios hold, 1 = at least one regressed.
set -euo pipefail

CLI="${1:?usage: $0 /path/to/leitwerk}"
case "$CLI" in /*) : ;; *) CLI="$PWD/$CLI" ;; esac
[ -x "$CLI" ] || { echo "not an executable CLI: $CLI" >&2; exit 1; }

SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

fail=0
for s in "$SELF_DIR"/s[0-9]*-*.sh; do
  if ! "$s" "$CLI"; then
    echo "FAIL: scenario $(basename "$s")" >&2
    fail=1
  fi
done

[ "$fail" -eq 0 ] && echo "all scenarios hold"
exit "$fail"
