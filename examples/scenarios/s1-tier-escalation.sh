#!/usr/bin/env bash
# Scenario: paths map to blast-radius tiers under the default policy an adopter
# gets from `leitwerk init` — a migration escalates to T2, docs stay T0.
set -euo pipefail

CLI="${1:?usage: $0 /path/to/leitwerk}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

"$CLI" init . >/dev/null

fail=0
expect_tier() { # <path> <tier>
  got="$("$CLI" tier "$1")"
  if [ "$got" != "$2" ]; then echo "FAIL: tier $1 = $got, want $2" >&2; fail=1; fi
}

expect_tier db/migrations/001.sql T2
expect_tier infra/main.tf         T2
expect_tier docs/guide.md         T0
expect_tier src/app.py            T1

[ "$fail" -eq 0 ] && echo "PASS: tier escalation (migration/infra T2, docs T0, app T1)"
exit "$fail"
