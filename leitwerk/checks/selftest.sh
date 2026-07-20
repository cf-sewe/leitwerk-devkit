#!/usr/bin/env bash
# Repo check: the golden behaviour of the CLI itself — the executable oracle the
# whole framework rests on. The CLI is now a compiled Go binary, so this check
# (1) builds it from source, (2) runs its Go unit + integration tests, and
# (3) re-asserts the external contract against the built binary. If tier selection,
# the guard, or the gate regresses, this fails. Exit 0 = intact, 1 = a regression.
set -euo pipefail

# Resolve a `go` command. Prefer mise (the repo's declared toolchain manager, per
# mise.toml) so we get the pinned Go and never a stale shim from another version
# manager on PATH. CI has no mise (it uses setup-go), so fall back to a bare `go`.
if command -v mise >/dev/null 2>&1; then
  GO=(mise exec -- go)
elif command -v go >/dev/null 2>&1; then
  GO=(go)
else
  echo "go toolchain not found (install it; the repo pins it in mise.toml)" >&2
  exit 1
fi

# 0. Build the gate from source and run its own tests, so the assertions below run
#    against the current source, not a stale binary.
CGO_ENABLED=0 "${GO[@]}" build -C core -o bin/leitwerk ./cmd/leitwerk
CGO_ENABLED=0 "${GO[@]}" test -C core ./...

CLI="$PWD/core/bin/leitwerk"
[ -x "$CLI" ] || { echo "CLI not found at $CLI" >&2; exit 1; }

fail=0
assert() { # <description> <expected> <actual>
  if [ "$2" != "$3" ]; then echo "FAIL: $1 (expected '$2', got '$3')" >&2; fail=1; fi
}

# 1. Tier mapping is the safety-critical logic. Test the glob ENGINE against the
#    shipped default tiers (not this repo's policy), so the assertions are
#    deterministic regardless of leitwerk/tiers.conf.
export LEITWERK_TIERS="$PWD/core/leitwerk.tiers"
assert "migration is T2"  T2 "$("$CLI" tier db/migrations/001.sql)"
assert "infra is T2"      T2 "$("$CLI" tier infra/main.tf)"
assert "markdown is T0"   T0 "$("$CLI" tier docs/guide.md)"
assert "app code is T1"   T1 "$("$CLI" tier src/app.py)"
unset LEITWERK_TIERS

# 2. The human-owned guard: a protected path is blocked, others are editable.
#    Test against the shipped default list so the assertion is deterministic.
export LEITWERK_TIERS="$PWD/core/leitwerk.tiers"
if "$CLI" guard leitwerk/constitution.md >/dev/null 2>&1; then
  echo "FAIL: guard allowed edit to a human-owned file" >&2; fail=1
fi
if "$CLI" guard /abs/path/to/leitwerk/tiers.conf >/dev/null 2>&1; then
  echo "FAIL: guard did not match a human-owned file by absolute-path suffix" >&2; fail=1
fi
if "$CLI" guard src/app.py >/dev/null 2>&1; then :; else
  echo "FAIL: guard blocked an ordinary editable file" >&2; fail=1
fi
unset LEITWERK_TIERS

# 3. The scaffolded review workflow must match this repo's own copy, so an
#    adopter running `leitwerk init` gets the same (advisory) orchestration this
#    repo dogfoods — the workflow is not plugin-packaged, so this is its only
#    guard against silent divergence.
if ! diff -q core/templates/workflows/leitwerk-review.mjs .claude/workflows/leitwerk-review.mjs >/dev/null 2>&1; then
  echo "FAIL: review workflow template and .claude/workflows/ copy diverged" >&2; fail=1
fi

# 4. The gate runs green on the bundled example AND runs its real Go tests there
#    (not a skip): the reference-app is a spec-anchored Go app with a T2
#    migration (M1.2). Run in place at T1 so mise resolves the repo's pinned Go
#    toolchain; assert the tests check actually executed.
ra_rc=0
ra_out="$( cd examples/reference-app && "$CLI" verify --tier T1 2>&1 )" || ra_rc=$?
if [ "$ra_rc" -ne 0 ]; then
  echo "FAIL: gate not green on reference-app at T1" >&2; echo "$ra_out" >&2; fail=1
fi
case "$ra_out" in
  *"go test green"*) : ;;
  *) echo "FAIL: reference-app did not run real tests (expected 'go test green')" >&2; fail=1 ;;
esac

# 5. The documented scenarios hold (examples/scenarios/ are executable
#    documentation of the guarantees; a regression turns this check red).
if ! scen_out="$(examples/scenarios/run-all.sh "$CLI" 2>&1)"; then
  echo "$scen_out" >&2
  echo "FAIL: a documented scenario regressed (examples/scenarios/)" >&2; fail=1
fi

# 6. The repo-local lifecycle check enforces terminal spec/plan states
#    mechanically (see leitwerk/specs/archive/lifecycle-check.md). Fixture assertions:
#    consistent tree passes, misplaced/invalid states fail, no specs dir skips.
LC="$PWD/leitwerk/checks/lifecycle.sh"
if [ -x "$LC" ]; then
  lc_tmp="$(mktemp -d)"
  mkdir -p "$lc_tmp/specs/archive"
  printf 'Status: active (2026-07-19)\n' > "$lc_tmp/specs/ok.md"
  printf 'Status: landed (2026-07-19)\n' > "$lc_tmp/specs/archive/done.md"
  if ! LEITWERK_SPECS="$lc_tmp/specs" "$LC" >/dev/null 2>&1; then
    echo "FAIL: lifecycle check red on a consistent fixture" >&2; fail=1
  fi
  mkdir -p "$lc_tmp/props"
  printf '# p\n' > "$lc_tmp/props/20200101_000000-old-proposal.md"
  lc_out="$(LEITWERK_SPECS="$lc_tmp/specs" LEITWERK_PROPOSALS="$lc_tmp/props" "$LC" 2>&1 || true)"
  case "$lc_out" in
    *"proposal(s) open"*) : ;;
    *) echo "FAIL: lifecycle did not report open proposals" >&2; fail=1 ;;
  esac
  case "$lc_out" in
    *"open since 2020-01-01"*) : ;;
    *) echo "FAIL: lifecycle did not flag an overdue proposal" >&2; fail=1 ;;
  esac
  printf 'Status: landed (2026-07-19)\n' > "$lc_tmp/specs/stale.md"
  if LEITWERK_SPECS="$lc_tmp/specs" "$LC" >/dev/null 2>&1; then
    echo "FAIL: lifecycle check green on a landed record outside archive/" >&2; fail=1
  fi
  rm -f "$lc_tmp/specs/stale.md"
  printf 'no status line here\n' > "$lc_tmp/specs/broken.md"
  if LEITWERK_SPECS="$lc_tmp/specs" "$LC" >/dev/null 2>&1; then
    echo "FAIL: lifecycle check green on a spec without a Status line" >&2; fail=1
  fi
  rc=0
  LEITWERK_SPECS="$lc_tmp/nowhere" "$LC" >/dev/null 2>&1 || rc=$?
  if [ "$rc" -ne 2 ]; then
    echo "FAIL: lifecycle with no specs dir = exit $rc, want 2 (skip)" >&2; fail=1
  fi
  rm -rf "$lc_tmp"
else
  echo "FAIL: leitwerk/checks/lifecycle.sh missing or not executable" >&2; fail=1
fi

# 7. The drift check anchors specs to code (see
#    leitwerk/specs/archive/drift-detection.md). One fixture per acceptance
#    bullet: resolving anchors pass with the enriched summary; a renamed symbol
#    and a missing path fail with a readable, located line; an out-of-repo
#    anchor is refused (not probed); a broken anchor inside archive/ is ignored;
#    a fenced example block is not parsed; a one-sided change fails when a diff
#    base is given; an option-like base is refused; no specs dir skips.
DR="$PWD/core/checks/drift.sh"
if [ -x "$DR" ]; then
  dr_tmp="$(mktemp -d)"
  mkdir -p "$dr_tmp/specs/archive" "$dr_tmp/code"
  printf 'package x\nfunc Alpha() {}\n' > "$dr_tmp/code/a.go"

  # consistent: a path anchor and a symbol anchor both resolve, and the summary
  # reports the anchor counts (not just the "spec(s) tracked" golden substring)
  printf 'Status: active (2026-07-20)\n## Anchors\n- `code/a.go`\n- `code/a.go#Alpha`\n' > "$dr_tmp/specs/ok.md"
  dr_rc=0
  dr_out="$(cd "$dr_tmp" && LEITWERK_SPECS=specs "$DR" 2>&1)" || dr_rc=$?
  assert "drift consistent fixture exit" 0 "$dr_rc"
  case "$dr_out" in *"spec(s) tracked"*) : ;; *) echo "FAIL: drift green summary missing 'spec(s) tracked'" >&2; fail=1 ;; esac
  case "$dr_out" in *"with anchors"*"anchor(s) resolve"*) : ;; *) echo "FAIL: drift summary did not report anchor counts" >&2; fail=1 ;; esac

  # renamed/removed symbol -> red, with a line naming the spec:line, the word
  # "anchor", and the missing symbol
  printf 'Status: active (2026-07-20)\n## Anchors\n- `code/a.go#Renamed`\n' > "$dr_tmp/specs/ok.md"
  dr_rc=0
  dr_out="$(cd "$dr_tmp" && LEITWERK_SPECS=specs "$DR" 2>&1)" || dr_rc=$?
  assert "drift missing-symbol exit" 1 "$dr_rc"
  case "$dr_out" in *ok.md:*anchor*Renamed*) : ;; *) echo "FAIL: drift missing-symbol line lacks spec:line / anchor / symbol" >&2; fail=1 ;; esac

  # missing path -> red
  printf 'Status: active (2026-07-20)\n## Anchors\n- `code/gone.go`\n' > "$dr_tmp/specs/ok.md"
  if (cd "$dr_tmp" && LEITWERK_SPECS=specs "$DR" >/dev/null 2>&1); then
    echo "FAIL: drift green on a missing-path anchor" >&2; fail=1
  fi

  # an out-of-repo anchor is refused, not turned into a file-existence oracle
  printf 'Status: active (2026-07-20)\n## Anchors\n- `/etc/hosts`\n' > "$dr_tmp/specs/ok.md"
  dr_rc=0
  dr_out="$(cd "$dr_tmp" && LEITWERK_SPECS=specs "$DR" 2>&1)" || dr_rc=$?
  assert "drift absolute-anchor refused exit" 1 "$dr_rc"
  case "$dr_out" in *"escapes the repo"*) : ;; *) echo "FAIL: drift did not refuse an out-of-repo (absolute) anchor" >&2; fail=1 ;; esac
  printf 'Status: active (2026-07-20)\n## Anchors\n- `../../../../etc/hosts`\n' > "$dr_tmp/specs/ok.md"
  if (cd "$dr_tmp" && LEITWERK_SPECS=specs "$DR" >/dev/null 2>&1); then
    echo "FAIL: drift green on a ..-escaping anchor" >&2; fail=1
  fi

  # a broken anchor that lives in archive/ is ignored
  printf 'Status: active (2026-07-20)\n## Anchors\n- `code/a.go`\n' > "$dr_tmp/specs/ok.md"
  printf 'Status: landed (2026-07-19)\n## Anchors\n- `code/gone.go`\n' > "$dr_tmp/specs/archive/old.md"
  if (cd "$dr_tmp" && LEITWERK_SPECS=specs "$DR" >/dev/null 2>&1); then :; else
    echo "FAIL: drift red on a broken anchor inside archive/ (should be ignored)" >&2; fail=1
  fi
  command rm -f "$dr_tmp/specs/archive/old.md"

  # a fenced ``` example anchor block (even with a broken path) is not parsed;
  # only the spec's real ## Anchors section counts
  {
    printf 'Status: active (2026-07-20)\n\n```\n## Anchors\n- `code/example-only.go`\n```\n\n## Anchors\n- `code/a.go`\n'
  } > "$dr_tmp/specs/ok.md"
  if (cd "$dr_tmp" && LEITWERK_SPECS=specs "$DR" >/dev/null 2>&1); then :; else
    echo "FAIL: drift parsed an anchor from inside a fenced code block" >&2; fail=1
  fi

  # no specs dir -> skip (exit 2), never a faked pass
  dr_rc=0
  (cd "$dr_tmp" && LEITWERK_SPECS=nowhere "$DR" >/dev/null 2>&1) || dr_rc=$?
  assert "drift no-specs skip" 2 "$dr_rc"
  command rm -rf "$dr_tmp"

  # one-sided change (Part 2). The git fixture is built OUTSIDE the assertion so
  # a setup failure cannot masquerade as the check passing: we capture the exit
  # code and require exactly 1, and assert the line names the code and the spec.
  if command -v git >/dev/null 2>&1; then
    dg="$(mktemp -d)"
    (
      cd "$dg"
      git init -q && git config user.email t@t && git config user.name t
      mkdir -p specs code
      printf 'Status: active (2026-07-20)\n## Anchors\n- `code/a.go`\n' > specs/x.md
      printf 'package x\nfunc Alpha() {}\n' > code/a.go
      git add -A && git commit -qm base
      git rev-parse HEAD > base.txt
      printf 'package x\nfunc Alpha() { return }\n' > code/a.go
      git add -A && git commit -qm code-only
    ) >/dev/null 2>&1
    dg_base="$(cat "$dg/base.txt" 2>/dev/null || true)"
    if [ -n "$dg_base" ]; then
      dr_rc=0
      dr_out="$(cd "$dg" && LEITWERK_SPECS=specs LEITWERK_DIFF_BASE="$dg_base" "$DR" 2>&1)" || dr_rc=$?
      assert "drift one-sided change exit" 1 "$dr_rc"
      case "$dr_out" in *"code/a.go changed"*"specs/x.md did not"*) : ;; *) echo "FAIL: drift one-sided line did not name the code and the spec" >&2; fail=1 ;; esac

      # an option-like base is refused before it can reach git (no injected file,
      # Part 2 skipped so the consistent fixture stays green)
      dr_rc=0
      (cd "$dg" && LEITWERK_SPECS=specs LEITWERK_DIFF_BASE='--output=INJECTED' "$DR" >/dev/null 2>&1) || dr_rc=$?
      assert "drift option-like base refused exit" 0 "$dr_rc"
      if [ -e "$dg/INJECTED...HEAD" ]; then echo "FAIL: drift let an option-like base reach git" >&2; fail=1; fi
    else
      echo "FAIL: drift one-sided git fixture failed to build" >&2; fail=1
    fi
    command rm -rf "$dg"
  fi
else
  echo "FAIL: core/checks/drift.sh missing or not executable" >&2; fail=1
fi

# 8. The gate's own checks must run on the stock macOS shell (bash 3.2): no
#    bash-4 array builtins (mapfile/readarray) in any check — a portable
#    while-read loop is required instead (see
#    leitwerk/specs/archive/bash-portability.md). Environment-independent guard;
#    selftest.sh itself is excluded (it names the builtins here).
mf="$(grep -rnE '(^|[^[:alnum:]_])(mapfile|readarray)([[:space:]]|$)' leitwerk/checks core/checks 2>/dev/null | grep -v '/selftest\.sh:' || true)"
if [ -n "$mf" ]; then
  echo "FAIL: a check uses a bash-4 array builtin — use a portable while-read loop:" >&2
  echo "$mf" >&2; fail=1
fi

# 9. Workflow scripts are valid JavaScript. .claude/workflows/*.mjs and the
#    shipped template core/templates/workflows/*.mjs run inside Claude Code's
#    async wrapper (top-level await/return, one `export const meta`), so a plain
#    `node --check` misreads them. Mirror the runtime: strip the leading `export`
#    and wrap the body in an async function, then syntax-check that. A malformed
#    workflow would otherwise fail only when a human invokes it. Node is resolved
#    via mise then PATH; absent node is a skip-with-note, never a faked result.
#    See leitwerk/specs/archive/workflow-syntax-check.md.
if command -v mise >/dev/null 2>&1 && mise exec -- node --version >/dev/null 2>&1; then
  NODE=(mise exec -- node)
elif command -v node >/dev/null 2>&1 && node --version >/dev/null 2>&1; then
  NODE=(node)
else
  NODE=()
fi
if [ "${#NODE[@]}" -gt 0 ]; then
  wf_found=0
  while IFS= read -r mjs; do
    wf_found=1
    wf_dir="$(mktemp -d)"
    { echo 'async function __leitwerk_wf__() {'; sed -E 's/^export ([a-z])/\1/' "$mjs"; echo; echo '}'; } > "$wf_dir/wf.js"
    if ! wf_err="$("${NODE[@]}" --check "$wf_dir/wf.js" 2>&1)"; then
      echo "FAIL: workflow syntax error in $mjs:" >&2; echo "$wf_err" >&2; fail=1
    fi
    rm -rf "$wf_dir"
  done < <(find .claude/workflows core/templates/workflows -type f -name '*.mjs' 2>/dev/null | sort)
  [ "$wf_found" -eq 1 ] || echo "note: no workflow .mjs found to syntax-check" >&2
else
  echo "note: node not available; skipped workflow .mjs syntax check" >&2
fi

[ "$fail" -eq 0 ] && echo "CLI golden behaviour intact (built + tested + tiers + gate + scenarios + lifecycle + drift + portability + workflows)"
exit "$fail"
