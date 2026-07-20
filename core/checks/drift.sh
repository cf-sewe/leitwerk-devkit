#!/usr/bin/env bash
# Drift sensor — surfaces spec<->code divergence. It SURFACES; it does not
# resolve. A human decides which side to reconcile (constitution invariant).
#
# A spec declares the code it governs in an "## Anchors" section, a list whose
# items each begin with a backtick-wrapped token `path` or `path#symbol`. This
# check flags:
#   (1) an anchor that no longer resolves — path gone / glob unmatched / symbol
#       absent from the file — always on;
#   (2) one-sided change — an anchored path changed in a range where its spec
#       did not — only when LEITWERK_DIFF_BASE names a git ref, so a local
#       working turn is never blocked mid-edit.
# Specs under an archive/ subdirectory are ignored: a landed record is not
# current contract. Anchor paths are confined to the repo (a spec is untrusted
# input): absolute or ..-escaping paths are rejected rather than probed.
#
# Exit 0 = no drift, 1 = drift surfaced (reconcile), 2 = no specs tracked.
# Tools: git, grep, awk, find, sort only (the core's stdlib-only shell floor).
set -euo pipefail

SPEC_DIR="${LEITWERK_SPECS:-leitwerk/specs}"
if [ ! -d "$SPEC_DIR" ]; then
  echo "no specs tracked ($SPEC_DIR)"; exit 2
fi

fail=0
n_specs=0
n_anchored=0
n_anchors=0

# The changed set for one-sided detection, computed once. Empty (Part 2 off)
# unless a base is given; the range mirrors what CI uses for tier selection.
# A base that is an option-looking token or not a resolvable commit is refused
# (loud warning, Part 2 skipped) — never silently treated as "no changes",
# which would fake a pass, and never interpolated as a git option.
base="${LEITWERK_DIFF_BASE:-}"
changed=""
if [ -n "$base" ]; then
  case "$base" in
    -*)
      echo "drift: warn — LEITWERK_DIFF_BASE '$base' looks like an option, not a ref; skipping the one-sided check" >&2
      base="" ;;
  esac
fi
if [ -n "$base" ]; then
  if git rev-parse --verify --quiet "$base^{commit}" >/dev/null 2>&1; then
    changed="$(git diff --name-only "$base...HEAD" 2>/dev/null || true)"
  else
    echo "drift: warn — LEITWERK_DIFF_BASE '$base' is not a resolvable commit; skipping the one-sided check" >&2
    base=""
  fi
fi

# Emit "lineno<TAB>anchor" for each item in a file's "## Anchors" section.
# Fenced code blocks are skipped so an *example* anchor block in a spec's prose
# is not mistaken for the spec's real anchors. The heading match tolerates
# trailing text ("## Anchors (optional)") so a small typo does not silently
# leave a spec unenforced.
anchors_of() {
  awk '
    /^```/ { fence = !fence; next }
    fence { next }
    /^##[[:space:]]/ { inblk = ($0 ~ /^##[[:space:]]+Anchors([[:space:]]|$)/) ? 1 : 0; next }
    inblk && /^[[:space:]]*[-*][[:space:]]+`/ {
      if (split($0, p, "`") >= 3) printf "%d\t%s\n", NR, p[2]
    }
  ' "$1"
}

# confined <path>: 0 if the anchor path stays inside the repo. A spec may carry
# attacker-influenced content; an unconfined path would turn this check into a
# file existence/content oracle for arbitrary host paths (e.g. /proc/self/environ).
confined() {
  case "$1" in
    /*) return 1 ;;            # absolute
  esac
  case "/$1/" in
    */../*) return 1 ;;        # a .. component anywhere in the path
  esac
  return 0
}

# path_exists <path-or-glob>: 0 if a literal path exists or a glob matches.
path_exists() {
  local p="$1" g
  case "$p" in
    *[*?[]*)
      for g in $p; do [ -e "$g" ] && return 0; done
      return 1 ;;
    *)
      [ -e "$p" ] ;;
  esac
}

# resolve_anchor <spec> <lineno> <token>: print a divergence line and return 1
# if the anchor does not resolve. <spec> is the repo-relative spec path.
resolve_anchor() {
  local spec="$1" ln="$2" tok="$3" path sym
  path="${tok%%#*}"
  if [ "$tok" = "$path" ]; then sym=""; else sym="${tok#*#}"; fi

  if ! confined "$path"; then
    echo "drift: $spec:$ln anchor $tok — path '$path' escapes the repo (absolute or ..); anchors must be repo-relative"
    return 1
  fi
  if ! path_exists "$path"; then
    echo "drift: $spec:$ln anchor $tok — path '$path' does not exist"
    return 1
  fi
  if [ -n "$sym" ]; then
    if [ ! -f "$path" ]; then
      echo "drift: $spec:$ln anchor $tok — symbol anchor needs a single file, '$path' is not one"
      return 1
    fi
    if ! grep -wF -- "$sym" "$path" >/dev/null 2>&1; then
      echo "drift: $spec:$ln anchor $tok — symbol '$sym' not found in $path"
      return 1
    fi
  fi
  return 0
}

# in_changed <path-or-glob>: 0 if any file in the changed set matches.
in_changed() {
  local pat="$1" ch
  while IFS= read -r ch; do
    [ -n "$ch" ] || continue
    case "$pat" in
      *[*?[]*)
        # $pat is a glob anchor (e.g. core/checks/*.sh); glob-matching it against
        # each changed path is the intent here, so the unquoted expansion is
        # deliberate.
        # shellcheck disable=SC2254
        case "$ch" in $pat) return 0 ;; esac ;;
      *) [ "$ch" = "$pat" ] && return 0 ;;
    esac
  done <<EOF
$changed
EOF
  return 1
}

while IFS= read -r f; do
  rel="${f#"$SPEC_DIR"/}"
  case "$rel" in archive/*) continue ;; esac   # archived record: not current contract
  n_specs=$((n_specs + 1))

  had_anchor=0
  spec_paths=""
  while IFS="$(printf '\t')" read -r ln tok; do
    [ -n "$tok" ] || continue
    had_anchor=1
    n_anchors=$((n_anchors + 1))
    resolve_anchor "$f" "$ln" "$tok" || fail=1
    spec_paths="$spec_paths ${tok%%#*}"
  done < <(anchors_of "$f")

  if [ "$had_anchor" -eq 1 ]; then
    n_anchored=$((n_anchored + 1))
    # Part 2: one-sided change. $f is the spec's repo-root-relative path, directly
    # comparable to `git diff --name-only` output (checks run from the root).
    if [ -n "$base" ] && ! in_changed "$f"; then
      for ap in $spec_paths; do
        if in_changed "$ap"; then
          echo "drift: $ap changed since $base but its spec $f did not — reconcile the spec or record why it is unchanged"
          fail=1
        fi
      done
    fi
  fi
done < <(find "$SPEC_DIR" -name '*.md' ! -name '*.plan.md' | sort)

if [ "$fail" -eq 0 ]; then
  echo "$n_specs spec(s) tracked, $n_anchored with anchors; $n_anchors anchor(s) resolve; no drift"
fi
exit "$fail"
