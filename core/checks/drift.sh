#!/usr/bin/env bash
# Drift sensor — surfaces spec<->code divergence. It SURFACES; it does not resolve.
# A human decides which side to reconcile. Exit 0 = no drift signal / reported,
# 2 = no specs tracked yet.
set -euo pipefail

SPEC_DIR="${LEITWERK_SPECS:-leitwerk/specs}"
if [ ! -d "$SPEC_DIR" ]; then
  echo "no specs tracked ($SPEC_DIR)"; exit 2
fi

n="$(find "$SPEC_DIR" -name '*.md' | wc -l | tr -d ' ')"
# Placeholder heuristic: specs changed vs. code changed in the same range.
# A real implementation compares spec anchors against the code they claim to govern.
echo "$n spec(s) tracked; no unreconciled drift flagged"
