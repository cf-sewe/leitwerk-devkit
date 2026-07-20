# Spec — lifecycle-check: enforce spec/plan lifecycle states in the gate

Status: landed (2026-07-20) <!-- durable content: the lifecycle check (mechanism) + constitution decision of record -->

## Problem
The spec lifecycle was convention only: `leitwerk-review` step 6 is prose, and
no gate check reads `Status:` lines — `drift` is a placeholder that counts
files (`core/checks/drift.sh:12-15`, CONFIRMED). The measured incident: after
the Go-CLI change merged, `go-cli.md` sat `active` ("mark landed when merged")
and three `landed` specs lay unarchived in the active set until a review found
them by chance (fixed by hand in the first dream pass, 2026-07-19). Prose rules
are exactly what this framework says cannot be a guarantee.

## Behaviour (the observable contract)
- A new repo-local check `leitwerk/checks/lifecycle.sh` runs at every tier
  (like `context`).
- Exit 2 (skip) when the specs directory is absent (`LEITWERK_SPECS` overrides
  the default `leitwerk/specs`, as in `drift`).
- Red (exit 1) when any `*.md` under the specs tree:
  - lacks a `Status:` line, or names an unknown state (valid: `draft`,
    `active`, `landed`, `superseded`);
  - is `landed`/`superseded` but lies outside `archive/`;
  - is `draft`/`active` but lies inside `archive/`;
  - is a `<slug>.plan.md` with no `<slug>.md` in the same directory;
  - is a plan still `draft`/`active` while its spec is `landed`/`superseded`.
- Warning (printed, still exit 0):
  - a plan whose step boxes are all `[x]`/`[~]` but whose Status is `active`
    ("ready to land");
  - a change record `active` since more than 30 days (dream-sweep candidate);
    silently skipped when the local `date` cannot do the arithmetic.
- The green summary counts active vs archived records.

## Design decisions
- **Complete-plan-but-active is a warning, not red.** Between the last build
  step and landing, that state is legitimate within a session; a hard red
  would block every turn end in that window. Hard red is reserved for
  misplaced terminal states and inconsistent spec/plan pairs. Rejected: red
  (false positives mid-change).
- **Repo-local first, not a core built-in.** Proves the check on this repo
  before adopters get it; promotion into `core/checks/` + shipped defaults
  belongs to M1.4. Rejected: shipping it untested.
- **Aging is advisory (30 days).** The threshold is a heuristic; making it
  policy would belong in the constitution. Rejected: hard red on age.

## Invariants touched
- *A check never fakes a pass* — nothing to scan exits 2.
- *The gate config is human-owned* — the `[tiers]` wiring adds a check, which
  the constitution explicitly permits an agent to do ("An agent may add a
  check"); the mechanical guard still requires the staged-copy route, executed
  under the human's in-session directive.
- *Drift is surfaced, not resolved* — the check reports; it never edits a
  `Status:` line.

## Blast radius
T2 (`*.sh` → gate behaviour). Worst case: a false red blocks turn-end and
merges (fail-closed, annoying); a false green is today's status quo.

## Acceptance checks
`selftest` asserts against fixtures: a consistent fixture exits 0; a landed
record outside `archive/` exits 1; a file without a `Status:` line exits 1; a
missing specs directory exits 2 (skip). `leitwerk verify --tier T2` stays
green on the clean repo.

## Out of scope
- Detecting "merged but never landed" — needs spec↔code anchors (M1.1).
- Archive-aware `drift` and promotion to a core built-in (M1.4).
- Prose truthfulness of docs — review at the tier's weight; the whitepaper's
  "keeping artifacts current" passage documents the split between what the
  gate enforces and what stays with review.
