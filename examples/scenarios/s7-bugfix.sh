#!/usr/bin/env bash
# Scenario: the bugfix workflow (Workflow C — reproduce → pin → fix) on the
# reference-app. Unlike s6 (break a covered path -> red), this exercises a defect
# the existing suite MISSES: it stays green until a regression test pins it, then
# the fix greens the gate again. That is the whole point of Workflow C — a bug is
# invisible until an oracle exercises it, so you pin before you fix.
# Runs on a throwaway copy of examples/reference-app (with mise.toml alongside so
# the pinned Go toolchain resolves); the real example is never mutated.
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

# 1. Baseline. The tests check must actually execute go test, not skip. A missing
# marker with a zero exit is a genuine "no toolchain" skip; a non-zero exit is a
# broken baseline and must FAIL, not be masked.
grc=0
gout="$("$CLI" verify --tier T1 2>&1)" || grc=$?
case "$gout" in
  *"go test green"*) : ;;
  *)
    if [ "$grc" -eq 0 ]; then
      echo "SKIP: no working Go toolchain here — cannot exercise the bugfix workflow" >&2
      echo "SKIP: bugfix scenario (no Go toolchain)"; exit 0
    fi
    echo "FAIL: reference-app gate red at T1 and tests did not run" >&2; echo "$gout" >&2; exit 1 ;;
esac
if [ "$grc" -ne 0 ]; then
  echo "FAIL: baseline reference-app gate not green at T1" >&2; echo "$gout" >&2; exit 1
fi

# 2. Reproduce. Inject a latent defect: large quantities are silently dropped
# (Qty % 100). The existing suite uses Qty 2 and 3, so it does NOT catch this —
# the gate stays GREEN. The parens are load-bearing (* and % share precedence),
# so the code is gofmt/vet-clean: only a test can expose the bug.
sed 's/it\.UnitCents \* it\.Qty/it.UnitCents * (it.Qty % 100)/' orders.go > orders.go.tmp && mv orders.go.tmp orders.go
irc=0
iout="$("$CLI" verify --tier T1 2>&1)" || irc=$?
if [ "$irc" -ne 0 ]; then
  echo "FAIL: injected latent bug went red at T1 (want green — the suite should miss it)" >&2; echo "$iout" >&2; exit 1
fi
case "$iout" in
  *"go test green"*) : ;;
  *) echo "FAIL: injected-bug run did not report 'go test green' (tests must still run)" >&2; echo "$iout" >&2; exit 1 ;;
esac

# 3. Pin, then fail. Add a regression test that captures the dropped-quantity
# defect; it must fail on the current (buggy) code, turning the gate RED.
cat >> orders_test.go <<'EOF'

func TestOrderTotalCentsLargeQty(t *testing.T) {
	// Regression pin for the "large quantity dropped" defect (Workflow C, §8.3).
	if got := OrderTotalCents([]LineItem{{UnitCents: 500, Qty: 100}}); got != 50000 {
		t.Fatalf("large-qty order total = %d, want 50000", got)
	}
}
EOF
prc=0
pout="$("$CLI" verify --tier T1 2>&1)" || prc=$?
if [ "$prc" -ne 1 ]; then
  echo "FAIL: pinned bug verify exit = $prc, want 1" >&2; echo "$pout" >&2; exit 1
fi
case "$pout" in
  *"gate: FAIL"*) : ;;
  *) echo "FAIL: pinned bug did not report 'gate: FAIL'" >&2; echo "$pout" >&2; exit 1 ;;
esac
# The red must come from the `tests` check (the pin), not from lint/types/drift:
# if the tests check still reported green, some unrelated check reddened the gate
# and this stage would not prove the regression test failed.
case "$pout" in
  *"go test green"*) echo "FAIL: pin red but tests still green — wrong check reddened the gate" >&2; echo "$pout" >&2; exit 1 ;;
esac

# 4. Fix at the tier. Revert the injection (the minimal fix). The pin now passes
# and the prior tests stay green — the gate goes GREEN.
sed 's/it\.UnitCents \* (it\.Qty % 100)/it.UnitCents * it.Qty/' orders.go > orders.go.tmp && mv orders.go.tmp orders.go
frc=0
fout="$("$CLI" verify --tier T1 2>&1)" || frc=$?
if [ "$frc" -ne 0 ]; then
  echo "FAIL: fixed code gate not green at T1" >&2; echo "$fout" >&2; exit 1
fi
case "$fout" in
  *"go test green"*) : ;;
  *) echo "FAIL: fixed run did not report 'go test green'" >&2; echo "$fout" >&2; exit 1 ;;
esac

echo "PASS: bugfix workflow — a missed defect stays green, the pin reds the gate, the fix greens it"
