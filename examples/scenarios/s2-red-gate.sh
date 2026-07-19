#!/usr/bin/env bash
# Scenario: a failing check turns the gate red. The fixture wires a check that
# always fails and asserts `verify` exits 1 and says so — the mechanism that
# blocks a broken change from landing.
set -euo pipefail

CLI="${1:?usage: $0 /path/to/leitwerk}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

"$CLI" init . >/dev/null

# Wire a deliberately failing repo-local check as the fixture's whole T1 gate.
# (In the fixture we edit tiers.conf directly: the human-owned guard constrains
# agent tool calls, not this test harness standing in for the human.)
cat > leitwerk/tiers.conf <<'EOF'
[tiers]
T1 = boom
[human-owned]
leitwerk/constitution.md
[paths]
* = T1
EOF
mkdir -p leitwerk/checks
cat > leitwerk/checks/boom.sh <<'EOF'
#!/usr/bin/env bash
echo "deliberate failure"
exit 1
EOF
chmod +x leitwerk/checks/boom.sh

rc=0
out="$("$CLI" verify --tier T1 2>&1)" || rc=$?

fail=0
if [ "$rc" -ne 1 ]; then echo "FAIL: verify exit = $rc, want 1" >&2; fail=1; fi
case "$out" in
  *"gate: FAIL"*) : ;;
  *) echo "FAIL: output does not report 'gate: FAIL'" >&2; fail=1 ;;
esac

[ "$fail" -eq 0 ] && echo "PASS: red gate (failing check -> exit 1, gate: FAIL)"
exit "$fail"
