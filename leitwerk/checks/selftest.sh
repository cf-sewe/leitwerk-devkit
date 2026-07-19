#!/usr/bin/env bash
# Repo check: the golden behaviour of the CLI itself — the executable oracle the
# whole framework rests on. The CLI is now a compiled Go binary, so this check
# (1) builds it from source, (2) runs its Go unit + integration tests, and
# (3) re-asserts the external contract against the built binary. If tier selection,
# the guard, or the gate regresses, this fails. Exit 0 = intact, 1 = a regression.
set -euo pipefail

# Resolve a `go` command. Prefer mise (the repo's declared toolchain manager, per
# mise.toml) so we get the pinned Go and never a stale shim from another version
# manager on PATH. CI has no mise (it uses setup-go), so fall back to a bare `go`.
if command -v mise >/dev/null 2>&1; then
  GO=(mise exec -- go)
elif command -v go >/dev/null 2>&1; then
  GO=(go)
else
  echo "go toolchain not found (install it; the repo pins it in mise.toml)" >&2
  exit 1
fi

# 0. Build the gate from source and run its own tests, so the assertions below run
#    against the current source, not a stale binary.
CGO_ENABLED=0 "${GO[@]}" build -C core -o bin/leitwerk ./cmd/leitwerk
CGO_ENABLED=0 "${GO[@]}" test -C core ./...

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

# 2. The human-owned guard: a protected path is blocked, others are editable.
#    Test against the shipped default list so the assertion is deterministic.
export LEITWERK_TIERS="$PWD/core/leitwerk.tiers"
if "$CLI" guard leitwerk/constitution.md >/dev/null 2>&1; then
  echo "FAIL: guard allowed edit to a human-owned file" >&2; fail=1
fi
if "$CLI" guard /abs/path/to/leitwerk/tiers.conf >/dev/null 2>&1; then
  echo "FAIL: guard did not match a human-owned file by absolute-path suffix" >&2; fail=1
fi
if "$CLI" guard src/app.py >/dev/null 2>&1; then :; else
  echo "FAIL: guard blocked an ordinary editable file" >&2; fail=1
fi
unset LEITWERK_TIERS

# 3. The scaffolded review workflow must match this repo's own copy, so an
#    adopter running `leitwerk init` gets the same (advisory) orchestration this
#    repo dogfoods — the workflow is not plugin-packaged, so this is its only
#    guard against silent divergence.
if ! diff -q core/templates/workflows/leitwerk-review.mjs .claude/workflows/leitwerk-review.mjs >/dev/null 2>&1; then
  echo "FAIL: review workflow template and .claude/workflows/ copy diverged" >&2; fail=1
fi

# 4. The gate runs green on the bundled example (proves end-to-end execution).
if ( cd examples/reference-app && "$CLI" verify --tier T0 >/dev/null 2>&1 ); then
  :
else
  echo "FAIL: gate not green on reference-app" >&2; fail=1
fi

# 5. The documented scenarios hold (examples/scenarios/ are executable
#    documentation of the guarantees; a regression turns this check red).
if ! scen_out="$(examples/scenarios/run-all.sh "$CLI" 2>&1)"; then
  echo "$scen_out" >&2
  echo "FAIL: a documented scenario regressed (examples/scenarios/)" >&2; fail=1
fi

[ "$fail" -eq 0 ] && echo "CLI golden behaviour intact (built + tested + tiers + gate + scenarios)"
exit "$fail"
