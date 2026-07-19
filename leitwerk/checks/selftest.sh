#!/usr/bin/env bash
# Repo check: the golden behaviour of the CLI itself — the executable oracle the
# whole framework rests on. If tier selection or the gate regresses, this fails.
# Exit 0 = behaviour intact, 1 = a regression.
set -euo pipefail

CLI="$PWD/core/bin/leitwerk"
[ -x "$CLI" ] || { echo "CLI not found at $CLI" >&2; exit 1; }

fail=0
assert() { # <description> <expected> <actual>
  if [ "$2" != "$3" ]; then echo "FAIL: $1 (expected '$2', got '$3')" >&2; fail=1; fi
}

# 1. Tier mapping is the safety-critical logic. Test the glob ENGINE against the
#    shipped default tiers (not this repo's policy), so the assertions are
#    deterministic regardless of leitwerk/tiers.conf.
export LEITWERK_TIERS="$PWD/core/leitwerk.tiers"
assert "migration is T2"  T2 "$("$CLI" tier db/migrations/001.sql)"
assert "infra is T2"      T2 "$("$CLI" tier infra/main.tf)"
assert "markdown is T0"   T0 "$("$CLI" tier docs/guide.md)"
assert "app code is T1"   T1 "$("$CLI" tier src/app.py)"
unset LEITWERK_TIERS

# 2. The gate runs green on the bundled example (proves end-to-end execution).
if ( cd examples/reference-app && "$CLI" verify --tier T0 >/dev/null 2>&1 ); then
  :
else
  echo "FAIL: gate not green on reference-app" >&2; fail=1
fi

[ "$fail" -eq 0 ] && echo "CLI golden behaviour intact (tiers + gate)"
exit "$fail"
