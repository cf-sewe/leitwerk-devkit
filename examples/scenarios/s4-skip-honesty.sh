#!/usr/bin/env bash
# Scenario: a check with nothing to run skips honestly. In a fresh repo with no
# toolchain, the T1 checks (lint/types/tests/drift) all abstain — the gate is
# green but the output visibly says "(skipped)", never a fake pass.
set -euo pipefail

CLI="${1:?usage: $0 /path/to/leitwerk}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

"$CLI" init . >/dev/null

rc=0
out="$("$CLI" verify --tier T1 2>&1)" || rc=$?

fail=0
if [ "$rc" -ne 0 ]; then echo "FAIL: verify exit = $rc, want 0 (all skips)" >&2; fail=1; fi
case "$out" in
  *"(skipped)"*) : ;;
  *) echo "FAIL: no visible '(skipped)' marker in output" >&2; fail=1 ;;
esac
case "$out" in
  *"gate: PASS"*) : ;;
  *) echo "FAIL: gate did not report PASS" >&2; fail=1 ;;
esac

[ "$fail" -eq 0 ] && echo "PASS: skip honesty (green gate, skips reported, no fake pass)"
exit "$fail"
