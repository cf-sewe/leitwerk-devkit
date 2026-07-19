#!/usr/bin/env bash
# Structural fitness function for open-code guarantee-parity: the hard guarantee
# must live in core/ and stay reachable without any agent runtime; bindings add
# ergonomics only. This fails the gate if that boundary erodes — the mechanism
# that keeps open-code parity true by construction, not by hope.
# Exit 0 = boundary intact, 1 = violation.
set -euo pipefail

fail=0
note() { echo "PARITY VIOLATION: $*" >&2; fail=1; }

# R1: core/ must not depend on any binding (core is tool-agnostic).
if grep -rnI 'bindings/' core/bin core/checks core/templates core/leitwerk.tiers 2>/dev/null; then
  note "core/ references bindings/ — the gate must not depend on a binding"
fi

# R2: gate/check logic lives only in core/ (or a consuming repo's leitwerk/).
# Bindings must ship no checks and no tier policy.
if find bindings -type d -name checks 2>/dev/null | grep -q .; then
  note "a checks/ dir exists under bindings/ — checks belong in core/"
fi
if find bindings -type f \( -name '*.tiers' -o -name 'tiers.conf' \) 2>/dev/null | grep -q .; then
  note "a tiers file exists under bindings/ — tier policy is not a binding's job"
fi

# R3: a binding launcher may only resolve and exec core, never reimplement the gate.
for launcher in bindings/*/bin/leitwerk; do
  [ -f "$launcher" ] || continue
  if grep -qE 'run_verify|checks_for_tier|tier_for_path|leitwerk\.tiers' "$launcher"; then
    note "$launcher reimplements gate logic — it must only delegate to core"
  fi
  if ! grep -qE 'exec |LEITWERK_HOME|core/bin/leitwerk' "$launcher"; then
    note "$launcher does not delegate to core (no exec of the core CLI found)"
  fi
done

[ "$fail" -eq 0 ] && echo "parity boundary intact (guarantee in core/, bindings delegate)"
exit "$fail"
