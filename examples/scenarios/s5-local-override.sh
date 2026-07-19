#!/usr/bin/env bash
# Scenario: a repo-local check overrides the built-in of the same name — how a
# consuming repo wires its real toolchain without editing installed core files.
set -euo pipefail

CLI="${1:?usage: $0 /path/to/leitwerk}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

"$CLI" init . >/dev/null

mkdir -p leitwerk/checks
cat > leitwerk/checks/lint.sh <<'EOF'
#!/usr/bin/env bash
echo "repo-local lint ran"
exit 0
EOF
chmod +x leitwerk/checks/lint.sh

rc=0
out="$("$CLI" verify --tier T0 2>&1)" || rc=$?   # default T0 = lint only

fail=0
if [ "$rc" -ne 0 ]; then echo "FAIL: verify exit = $rc, want 0" >&2; fail=1; fi
case "$out" in
  *"repo-local lint ran"*) : ;;
  *) echo "FAIL: built-in lint ran instead of the repo-local override" >&2; fail=1 ;;
esac

[ "$fail" -eq 0 ] && echo "PASS: local override (leitwerk/checks/lint.sh shadowed the built-in)"
exit "$fail"
