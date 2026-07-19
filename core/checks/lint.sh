#!/usr/bin/env bash
# Style / static conventions. Auto-detects the project's linter.
# Exit 0 = pass, 1 = fail, 2 = skip (no linter configured here).
set -euo pipefail

if [ -f package.json ] && grep -q '"lint"' package.json; then
  npm run --silent lint && echo "eslint clean"
elif [ -f go.mod ] && command -v golangci-lint >/dev/null 2>&1; then
  golangci-lint run && echo "golangci-lint clean"
elif [ -x ./gradlew ]; then
  echo "gradle checkstyle delegated to build"; exit 2
else
  echo "no linter configured"; exit 2
fi
