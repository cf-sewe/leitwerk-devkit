#!/usr/bin/env bash
# Repo check: the shell scripts ARE the gate, so they get the strictest check.
# `bash -n` (syntax) always runs; shellcheck runs when available. Uses find so it
# is correct before the first commit. Exit 0 = clean, 1 = finding, 2 = no scripts.
set -euo pipefail

# Collect every shell script: all *.sh, plus extensionless scripts (CLI
# launchers, hook wrappers) detected by shebang — so a new one is never silently
# skipped, which would let the gate report clean without checking it.
mapfile -t scripts < <(
  {
    find . -type f -name '*.sh' -not -path './node_modules/*' -not -path './.git/*'
    while IFS= read -r f; do
      IFS= read -r first < "$f" 2>/dev/null || continue
      case "$first" in
        '#!'*bash*|'#!'*/sh|'#!'*' sh') echo "$f" ;;
      esac
    done < <(find . -type f ! -name '*.sh' -not -path './node_modules/*' -not -path './.git/*')
  } | sort -u
)

[ "${#scripts[@]}" -gt 0 ] || { echo "no shell scripts found"; exit 2; }

fail=0
for s in "${scripts[@]}"; do
  if ! err="$(bash -n "$s" 2>&1)"; then
    echo "syntax error in $s:" >&2; echo "$err" >&2; fail=1
  fi
done

if command -v shellcheck >/dev/null 2>&1; then
  shellcheck -S warning "${scripts[@]}" || fail=1
  msg="${#scripts[@]} script(s): bash -n + shellcheck clean"
else
  msg="${#scripts[@]} script(s): bash -n clean (shellcheck not installed)"
fi

[ "$fail" -eq 0 ] && echo "$msg"
exit "$fail"
