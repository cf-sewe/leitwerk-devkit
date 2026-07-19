#!/usr/bin/env bash
# Structural fitness function for open-code guarantee-parity: the hard guarantee
# must live in core/ and stay reachable without any agent runtime; bindings add
# ergonomics only. This fails the gate if that boundary erodes — the mechanism
# that keeps open-code parity true by construction, not by hope.
# Exit 0 = boundary intact, 1 = violation.
set -euo pipefail

fail=0
note() { echo "PARITY VIOLATION: $*" >&2; fail=1; }

# R1: core/ must not depend on any binding (core is tool-agnostic). This covers
# the Go gate source (cmd/, internal/, assets.go, go.mod) as well as the checks,
# templates, and default tiers; grep -I skips the compiled binary under core/bin.
if grep -rnI 'bindings/' \
    core/bin core/checks core/templates core/leitwerk.tiers \
    core/cmd core/internal core/assets.go core/go.mod 2>/dev/null; then
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

# R3: a binding must not reimplement the gate in ANY language — no tier/check/glob
# logic and no parsing of the tiers policy file. Scan every binding bin/ file for
# gate-logic signals: the old Bash function names, the Go gate API, and references
# to the tiers file / its section headers. (The launcher and hook-guard delegate
# via the core CLI, so they carry none of these.)
reimpl='run_verify|checks_for_tier|tier_for_path|ChecksForTier|TierForPath|RunVerify|GlobToRegex|ParseTiers|leitwerk\.tiers|tiers\.conf|\[tiers\]|\[paths\]'
for f in bindings/*/bin/*; do
  [ -f "$f" ] || continue
  if grep -qE "$reimpl" "$f"; then
    note "$f contains gate logic — a binding must delegate to core, not reimplement it"
  fi
done
# The main launcher (named leitwerk) must additionally resolve+exec core.
for launcher in bindings/*/bin/leitwerk; do
  [ -f "$launcher" ] || continue
  if ! grep -qE 'exec |LEITWERK_HOME|core/bin/leitwerk' "$launcher"; then
    note "$launcher does not delegate to core (no exec of the core CLI found)"
  fi
done

# R4: the Go gate must stay standard-library-only, so no agent SDK or other runtime
# dependency can creep into core/ (the "core never depends on an agent runtime"
# invariant, enforced structurally rather than only by grepping for 'bindings/').
if [ -f core/go.mod ] && grep -qE '^[[:space:]]*require([[:space:]]|\()' core/go.mod; then
  note "core/go.mod has a require directive — the gate must stay standard-library-only"
fi
if [ -f core/go.sum ]; then
  note "core/go.sum exists — the gate must have zero third-party dependencies"
fi

[ "$fail" -eq 0 ] && echo "parity boundary intact (guarantee in core/, stdlib-only, bindings delegate)"
exit "$fail"
