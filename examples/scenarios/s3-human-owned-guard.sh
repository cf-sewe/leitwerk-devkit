#!/usr/bin/env bash
# Scenario: human-owned files are guarded. `leitwerk guard` exits 3 on the
# scaffolded policy files (the Claude binding turns that into a blocked edit;
# open-code maps the same list to required review) and 0 on ordinary files.
set -euo pipefail

CLI="${1:?usage: $0 /path/to/leitwerk}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

"$CLI" init . >/dev/null

fail=0
expect_guard() { # <path> <exit-code>
  rc=0
  "$CLI" guard "$1" >/dev/null 2>&1 || rc=$?
  if [ "$rc" -ne "$2" ]; then echo "FAIL: guard $1 exit = $rc, want $2" >&2; fail=1; fi
}

expect_guard leitwerk/constitution.md          3
expect_guard leitwerk/tiers.conf               3
expect_guard "$tmp/leitwerk/tiers.conf"        3   # absolute path resolves too
expect_guard leitwerk//constitution.md         3   # equivalent spelling, no bypass
expect_guard src/app.py                        0

[ "$fail" -eq 0 ] && echo "PASS: human-owned guard (policy files exit 3, app code 0)"
exit "$fail"
