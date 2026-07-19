#!/usr/bin/env bash
# Executable behaviour oracle — the golden suite plus any characterization tests.
# Exit 0 = pass, 1 = fail, 2 = skip.
set -euo pipefail

if [ -f package.json ] && grep -q '"test"' package.json; then
  npm test --silent && echo "test suite green"
elif [ -x ./gradlew ]; then
  ./gradlew test && echo "gradle test green"
elif ls go.mod >/dev/null 2>&1; then
  go test ./... && echo "go test green"
else
  echo "no test runner configured"; exit 2
fi
