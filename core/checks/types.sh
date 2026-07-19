#!/usr/bin/env bash
# Type / contract check — the cheapest external oracle. Auto-detects the toolchain.
# Exit 0 = pass, 1 = fail, 2 = skip.
set -euo pipefail

if [ -f tsconfig.json ]; then
  npx --no-install tsc --noEmit && echo "tsc: no type errors"
elif [ -x ./gradlew ]; then
  echo "javac type-check delegated to build"; exit 2
elif ls go.mod >/dev/null 2>&1; then
  go vet ./... && echo "go vet clean"
else
  echo "no type checker configured"; exit 2
fi
