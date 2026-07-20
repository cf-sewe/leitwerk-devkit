#!/usr/bin/env bash
# Scenario: the reference-app demonstrates real governance — the gate runs its
# Go test suite (not a skip), and a deliberately broken change turns it red.
# Runs on a throwaway copy of examples/reference-app (with mise.toml alongside so
# the pinned Go toolchain resolves), so the real example is never mutated.
set -euo pipefail

CLI="${1:?usage: $0 /path/to/leitwerk}"
case "$CLI" in /*) : ;; *) CLI="$PWD/$CLI" ;; esac
[ -x "$CLI" ] || { echo "not an executable CLI: $CLI" >&2; exit 1; }

SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SELF_DIR/../.." && pwd)"
APP="$REPO_ROOT/examples/reference-app"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/app"
cp -R "$APP/." "$tmp/app/"
[ -f "$REPO_ROOT/mise.toml" ] && cp "$REPO_ROOT/mise.toml" "$tmp/app/mise.toml"
cd "$tmp/app"

# Green run: the tests check must actually execute go test, not skip. Decide SKIP
# only when the gate was green (grc==0) AND the marker is absent — a genuine
# "no toolchain" skip. A missing marker with a non-zero exit is a real failure
# (a broken baseline), not a skip, so it must FAIL rather than be masked.
grc=0
gout="$("$CLI" verify --tier T1 2>&1)" || grc=$?
case "$gout" in
  *"go test green"*) : ;;   # tests ran; proceed to the broken-change assertion
  *)
    if [ "$grc" -eq 0 ]; then
      echo "SKIP: no working Go toolchain here — cannot exercise the reference-app" >&2
      echo "SKIP: reference-app scenario (no Go toolchain)"; exit 0
    fi
    echo "FAIL: reference-app gate red at T1 and tests did not run" >&2; echo "$gout" >&2; exit 1 ;;
esac
if [ "$grc" -ne 0 ]; then
  echo "FAIL: reference-app gate not green at T1" >&2; echo "$gout" >&2; exit 1
fi

# Broken change: a real regression to the order total (multiply -> add) must be
# caught by the test and turn the gate red.
sed 's/it\.UnitCents \* it\.Qty/it.UnitCents + it.Qty/' orders.go > orders.go.tmp && mv orders.go.tmp orders.go

brc=0
bout="$("$CLI" verify --tier T1 2>&1)" || brc=$?
fail=0
if [ "$brc" -ne 1 ]; then echo "FAIL: broken reference-app verify exit = $brc, want 1" >&2; echo "$bout" >&2; fail=1; fi
case "$bout" in
  *"gate: FAIL"*) : ;;
  *) echo "FAIL: broken reference-app did not report 'gate: FAIL'" >&2; fail=1 ;;
esac

[ "$fail" -eq 0 ] && echo "PASS: reference-app runs real tests, and a broken change reds the gate"
exit "$fail"
