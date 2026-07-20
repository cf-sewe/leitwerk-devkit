# Spec — bash-portability: the gate's own checks run on bash 3.2

Status: landed (2026-07-20) <!-- durable content (the no-mapfile decision + guard) stays in this archived spec -->

Anchors below are historical (this record is archived; drift ignores archive/).

## Problem
`leitwerk/checks/json.sh:8` and `leitwerk/checks/shell.sh:12` use `mapfile`
(CONFIRMED), a bash 4+ builtin. macOS ships bash 3.2 as `/bin/bash`; under it
`json` silently skips (mapfile missing → empty array → "no JSON files", exit 2)
and `shell` fails to run at all — reproduced on `GNU bash 3.2.57`. The checks
pass in this repo only because a bash ≥ 4 happens to be first on PATH for
`#!/usr/bin/env bash`. That contradicts the invariant that the core "runs with
only a shell present": the gate silently degrades (json fakes a skip) or breaks
on a stock macOS shell. A sweep found no other bash-4 feature in any check
(mapfile/readarray/associative arrays/case-conversion).

## Behaviour (the observable contract)
- `json.sh` and `shell.sh` collect their file lists with a portable
  `arr=(); while IFS= read -r x; do arr+=("$x"); done < <(…)` loop instead of
  `mapfile`. Arrays and `+=` append are bash-3.2 features; the rest of each
  script is unchanged.
- On bash ≥ 4 the observable output is identical. On bash 3.2 both checks now
  run correctly: `json` validates the JSON files (not a skip), `shell` runs
  `bash -n`/shellcheck over every collected script.
- A `selftest` guard (environment-independent) fails if any check reintroduces a
  bash-4 array builtin (`mapfile`/`readarray`), so the portability cannot
  silently regress.
- `shell.sh` additionally avoids a `case` nested inside a process substitution
  (bash 3.2 mis-parses that): its collection is two top-level `while` loops over
  the two disjoint file sources.

## Design decisions
- **Replace `mapfile`, don't document a bash-4 requirement.** The framework's
  headline is "runs with only a shell present"; requiring bash ≥ 4 to run the
  gate honestly (rather than degrade) would narrow that claim. The replacement
  is a small, behaviour-preserving change. Rejected: documenting the
  prerequisite in README + `leitwerk-onboard` (weaker — it leaves a silent skip
  on the most common developer shell).

## Invariants touched
- *The core never depends on an agent runtime* / runs with only a shell —
  restored for the stock-macOS shell.
- *A check never fakes a pass* — the bug made `json` skip on bash 3.2 when there
  were files to check; the fix removes that false skip.

## Blast radius
T2 (`*.sh`, gate behaviour). Worst case if wrong: a mis-collected file list;
mitigated by testing both files under bash 3.2 and the PATH bash, and the gate.

## Acceptance checks
- Under `/bin/bash` (3.2): `json.sh` and `shell.sh` run and report correctly
  (json validates files, shell checks scripts), with no mapfile error/skip.
- `leitwerk verify --tier T2` stays green on the PATH bash (≥ 4).

## Anchors
- `leitwerk/checks/json.sh`
- `leitwerk/checks/shell.sh`

## Out of scope
- Promoting `json`/`shell` into `core/checks/` (they are repo-local); any future
  core check should follow the same no-`mapfile` rule.
