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

# 6. The repo-local lifecycle check enforces terminal spec/plan states
#    mechanically (see leitwerk/specs/archive/lifecycle-check.md). Fixture assertions:
#    consistent tree passes, misplaced/invalid states fail, no specs dir skips.
LC="$PWD/leitwerk/checks/lifecycle.sh"
if [ -x "$LC" ]; then
  lc_tmp="$(mktemp -d)"
  mkdir -p "$lc_tmp/specs/archive"
  printf 'Status: active (2026-07-19)\n' > "$lc_tmp/specs/ok.md"
  printf 'Status: landed (2026-07-19)\n' > "$lc_tmp/specs/archive/done.md"
  if ! LEITWERK_SPECS="$lc_tmp/specs" "$LC" >/dev/null 2>&1; then
    echo "FAIL: lifecycle check red on a consistent fixture" >&2; fail=1
  fi
  mkdir -p "$lc_tmp/props"
  printf '# p\n' > "$lc_tmp/props/20200101_000000-old-proposal.md"
  lc_out="$(LEITWERK_SPECS="$lc_tmp/specs" LEITWERK_PROPOSALS="$lc_tmp/props" "$LC" 2>&1 || true)"
  case "$lc_out" in
    *"proposal(s) open"*) : ;;
    *) echo "FAIL: lifecycle did not report open proposals" >&2; fail=1 ;;
  esac
  case "$lc_out" in
    *"open since 2020-01-01"*) : ;;
    *) echo "FAIL: lifecycle did not flag an overdue proposal" >&2; fail=1 ;;
  esac
  printf 'Status: landed (2026-07-19)\n' > "$lc_tmp/specs/stale.md"
  if LEITWERK_SPECS="$lc_tmp/specs" "$LC" >/dev/null 2>&1; then
    echo "FAIL: lifecycle check green on a landed record outside archive/" >&2; fail=1
  fi
  rm -f "$lc_tmp/specs/stale.md"
  printf 'no status line here\n' > "$lc_tmp/specs/broken.md"
  if LEITWERK_SPECS="$lc_tmp/specs" "$LC" >/dev/null 2>&1; then
    echo "FAIL: lifecycle check green on a spec without a Status line" >&2; fail=1
  fi
  rc=0
  LEITWERK_SPECS="$lc_tmp/nowhere" "$LC" >/dev/null 2>&1 || rc=$?
  if [ "$rc" -ne 2 ]; then
    echo "FAIL: lifecycle with no specs dir = exit $rc, want 2 (skip)" >&2; fail=1
  fi
  rm -rf "$lc_tmp"
else
  echo "FAIL: leitwerk/checks/lifecycle.sh missing or not executable" >&2; fail=1
fi

[ "$fail" -eq 0 ] && echo "CLI golden behaviour intact (built + tested + tiers + gate + scenarios + lifecycle)"
exit "$fail"
