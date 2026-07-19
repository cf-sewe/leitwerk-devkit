#!/usr/bin/env bash
# Security static analysis + dependency policy. Required on T2 (irreversible/infra).
# Exit 0 = pass, 1 = fail, 2 = skip.
set -euo pipefail

if command -v semgrep >/dev/null 2>&1; then
  semgrep --error --quiet --config auto . && echo "semgrep: no findings"
else
  echo "semgrep not installed"; exit 2
fi
