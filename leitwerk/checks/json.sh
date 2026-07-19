#!/usr/bin/env bash
# Repo check: every JSON manifest must parse. These files configure the plugin
# and the marketplace; a broken one fails silently at install time otherwise.
# Uses find (not git ls-files) so it is correct even before the first commit and
# never passes vacuously. Exit 0 = all parse, 1 = a file is invalid, 2 = none.
set -euo pipefail

mapfile -t files < <(find . -type f -name '*.json' \
  -not -path './node_modules/*' -not -path './.git/*' | sort)

[ "${#files[@]}" -gt 0 ] || { echo "no JSON files found"; exit 2; }

# Pick a validator that actually executes here (probe with a trivial doc, so a
# broken interpreter shim is a skip, never a false "invalid").
if printf '{}' | node -e 'JSON.parse(require("fs").readFileSync(0,"utf8"))' 2>/dev/null; then
  parse() { node -e 'JSON.parse(require("fs").readFileSync(process.argv[1],"utf8"))' "$1"; }
elif command -v jq >/dev/null 2>&1; then
  parse() { jq empty "$1"; }
elif printf '{}' | python3 -c 'import json,sys; json.load(sys.stdin)' 2>/dev/null; then
  parse() { python3 -c 'import json,sys; json.load(open(sys.argv[1]))' "$1"; }
else
  echo "no working JSON validator (node/jq/python3)"; exit 2
fi

fail=0
for f in "${files[@]}"; do
  if ! parse "$f" 2>/dev/null; then echo "invalid JSON: $f" >&2; fail=1; fi
done

[ "$fail" -eq 0 ] && echo "${#files[@]} JSON manifest(s) parse"
exit "$fail"
