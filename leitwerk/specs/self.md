# Spec — leitwerk-devkit governs itself

## Problem
The framework claims changes should be spec-anchored and verification-gated. If
its own repository is not, the claim is unproven and the whitepaper reads as
generic. This spec anchors the repository to its own gate.

## Behaviour (the observable contract)
- `leitwerk verify` runs from the repo root using `leitwerk/tiers.conf` and the
  repo-local checks in `leitwerk/checks/`, falling back to built-in `core/checks/`
  per check.
- Changing a `core/bin`, `core/checks`, or any `*.sh` path selects tier **T2**.
- A T2 gate runs: `json` (all manifests parse), `shell` (`bash -n` + shellcheck
  when present), `drift` (specs tracked), `selftest` (CLI tier mapping + a green
  gate on the reference app).
- The gate is green on a clean tree; a real defect (invalid JSON, shell syntax
  error, or a broken tier boundary) turns it red.

## Invariants touched
All constitution invariants — this is the change that makes them enforced rather
than aspirational.

## Blast radius
T2. The change installs the gate over the gate; a mistake here would let the
framework ship an untested gate.

## Acceptance checks
`leitwerk verify --tier T2` exits 0 on the clean repo, and exits non-zero if any
of: a manifest is corrupted, a script has a syntax error, or a tier boundary in
`selftest.sh` regresses. Enforced locally by the Stop hook in `.claude/settings.json`
and authoritatively by `.github/workflows/leitwerk.yml`.

## Out of scope
Wiring language-specific checks (there is no application code here yet) and
publishing the npm package.
