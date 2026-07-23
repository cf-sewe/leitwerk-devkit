#!/usr/bin/env bash
# Lifecycle check — the spec/plan lifecycle, enforced mechanically (the states
# are defined in core/templates/spec.template.md; the incident that motivated
# this is recorded in leitwerk/specs/archive/lifecycle-check.md).
#   red  (1): missing/unknown Status line · landed/superseded outside archive/
#             · draft/active inside archive/ · plan without its spec · plan
#             still open while its spec is landed/superseded
#   warn (0): plan complete but still active ("ready to land") · change record
#             active for >30 days (dream-sweep candidate)
#   skip (2): no specs directory
set -euo pipefail

SPEC_DIR="${LEITWERK_SPECS:-leitwerk/specs}"
[ -d "$SPEC_DIR" ] || { echo "no specs tracked ($SPEC_DIR)"; exit 2; }
ROADMAP="${LEITWERK_ROADMAP:-leitwerk/roadmap.md}"

fail=0
active=0
archived=0

state_of() { # first word of the first Status: line; empty if none
  awk '/^Status:/ { sub(/^Status:[[:space:]]*/, ""); split($0, a, /[ (]/); print a[1]; exit }' "$1"
}

# aging threshold: 30 days ago, if the local date(1) can compute it (BSD/GNU)
cutoff="$(date -v-30d +%F 2>/dev/null || date -d '30 days ago' +%F 2>/dev/null || true)"

while IFS= read -r f; do
  rel="${f#"$SPEC_DIR"/}"
  st="$(state_of "$f")"
  case "$st" in
    draft|active|landed|superseded) ;;
    "") echo "lifecycle: $rel has no Status: line"; fail=1; continue ;;
    *)  echo "lifecycle: $rel has unknown state '$st'"; fail=1; continue ;;
  esac

  in_archive=0
  case "$rel" in archive/*) in_archive=1 ;; esac

  if [ "$in_archive" -eq 1 ]; then
    archived=$((archived + 1))
    case "$st" in
      draft|active) echo "lifecycle: $rel is $st but lies in archive/"; fail=1 ;;
    esac
  else
    case "$st" in
      landed|superseded)
        echo "lifecycle: $rel is $st but not in archive/ (dream pass missing)"; fail=1 ;;
      active) active=$((active + 1)) ;;
    esac
  fi

  # roadmap<->spec join (roadmap-spec-join): an active/draft change record may
  # name the roadmap item it realizes as `Roadmap: <slug>`; the slug must be an
  # open item (`**<slug>**`) in the roadmap. Archived specs are exempt (a landed
  # item has left the roadmap, mirroring drift). A `<slug>` placeholder or any
  # non-slug value is ignored, so a fresh draft never goes red.
  if [ "$in_archive" -eq 0 ]; then
    case "$f" in
      *.plan.md) ;;
      *)
        rslug="$(awk '/^Roadmap:/ { sub(/^Roadmap:[[:space:]]*/, ""); print $1; exit }' "$f")"
        case "$rslug" in
          "" | *[!a-z0-9-]*) ;;
          *)
            if [ -f "$ROADMAP" ] && ! grep -qE "^\*\*${rslug}\*\*( |\$)" "$ROADMAP"; then
              echo "lifecycle: $rel declares Roadmap: $rslug, not an open item in ${ROADMAP##*/}"; fail=1
            fi ;;
        esac ;;
    esac
  fi

  case "$f" in
    *.plan.md)
      spec="${f%.plan.md}.md"
      if [ ! -f "$spec" ]; then
        echo "lifecycle: $rel has no spec beside it (${rel%.plan.md}.md)"; fail=1
      else
        sst="$(state_of "$spec")"
        case "$sst" in
          landed|superseded)
            case "$st" in
              draft|active)
                echo "lifecycle: spec ${rel%.plan.md}.md is $sst but its plan is $st"; fail=1 ;;
            esac ;;
        esac
      fi
      # all step boxes done but not landed yet -> ready to land (advisory);
      # boxes may be backtick-quoted (`[x]`) or bare ([x])
      if [ "$st" = "active" ] && grep -qE '^[0-9]+\. `?\[' "$f" && ! grep -qE '^[0-9]+\. `?\[ \]' "$f"; then
        echo "lifecycle: warn — $rel: all steps done but still active (ready to land)"
      fi
      ;;
  esac

  # change records aging in the active set are dream-sweep candidates (advisory)
  if [ -n "$cutoff" ] && [ "$st" = "active" ] && [ "$in_archive" -eq 0 ]; then
    d="$(awk '/^Status:[[:space:]]*active[[:space:]]*\(/ { sub(/^[^(]*\(/, ""); sub(/\).*/, ""); print; exit }' "$f")"
    if [ -n "$d" ] && [[ "$d" < "$cutoff" ]]; then
      echo "lifecycle: warn — $rel: active since $d (dream-sweep candidate)"
    fi
  fi
done < <(find "$SPEC_DIR" -name '*.md' | sort)

# open proposals: decisions awaiting the human stay visible on every run —
# but never red: the waiting party is the human, and a red gate would punish
# unrelated work and reward deleting proposals
PROP_DIR="${LEITWERK_PROPOSALS:-leitwerk/proposals}"
proposals=0
if [ -d "$PROP_DIR" ]; then
  while IFS= read -r p; do
    base="$(basename "$p")"
    [ "$base" = "README.md" ] && continue
    proposals=$((proposals + 1))
    pd="$(printf '%s' "$base" | sed -n 's/^\([0-9]\{8\}\)_.*/\1/p')"
    if [ -n "$pd" ] && [ -n "$cutoff" ]; then
      pdate="${pd:0:4}-${pd:4:2}-${pd:6:2}"
      if [[ "$pdate" < "$cutoff" ]]; then
        echo "lifecycle: warn — proposal $base open since $pdate (overdue)"
      fi
    fi
  done < <(find "$PROP_DIR" -maxdepth 1 -name '*.md' | sort)
fi

[ "$fail" -eq 0 ] && echo "lifecycle intact: $active active, $archived archived, $proposals proposal(s) open"
exit "$fail"
