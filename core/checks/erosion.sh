#!/usr/bin/env bash
# Erosion budgets — complexity / duplication ceilings that keep long-lived code
# from degrading one accepted change at a time. Required on T2.
# Exit 0 = within budget, 1 = budget exceeded, 2 = no analyzer configured.
set -euo pipefail

if command -v jscpd >/dev/null 2>&1; then
  jscpd --silent --threshold 5 . && echo "duplication within 5% budget"
else
  echo "no complexity/duplication analyzer configured"; exit 2
fi
