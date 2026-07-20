# Spec — leitwerk-devkit governs itself

Status: active <!-- the repo's living governance contract, not a change record -->

## Problem
The framework claims changes should be spec-anchored and verification-gated. If
its own repository is not, the claim is unproven and the whitepaper reads as
generic. This spec anchors the repository to its own gate.

## Behaviour (the observable contract)
- `leitwerk verify` runs from the repo root using `leitwerk/tiers.conf` and the
  repo-local checks in `leitwerk/checks/`, falling back to built-in `core/checks/`
  per check.
- Changing a `core/bin`, `core/checks`, `core/cmd`, `core/internal` (or other Go
  source), or any `*.sh` path selects tier **T2**.
- A T2 gate runs: `json` (all manifests parse), `shell` (`bash -n` + shellcheck
  when present), `drift` (spec `## Anchors` resolve against the code, and
  one-sided spec/code change is surfaced when a diff base is given), `selftest` (builds the Go CLI, runs its
  unit + integration tests, re-asserts the external contract, and executes the
  documented scenarios in `examples/scenarios/`), `parity` (the guarantee stays
  in `core/`, stdlib-only; bindings delegate), `context` (always-on steering
  files stay within the constitution's context budgets), `lifecycle` (spec/plan
  lifecycle states are consistent: valid `Status:` lines, terminal states only
  in `leitwerk/specs/archive/`, spec and plan agree).
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
of: a manifest is corrupted, a script has a syntax error, a tier boundary in
`selftest.sh` regresses, or a landed spec record sits outside
`leitwerk/specs/archive/`. Enforced locally by the Stop hook in `.claude/settings.json`
and authoritatively by `.github/workflows/leitwerk.yml`.

## Anchors
The gate code this contract governs — a rename here without updating this spec
surfaces as drift (the check dogfoods itself):
- `core/internal/gate/verify.go#RunVerify`
- `core/internal/gate/tiers.go#ChecksForTier`
- `core/cmd/leitwerk/main.go#cmdVerify`

## Out of scope
Wiring language-specific checks beyond the Go toolchain the gate itself needs,
and publishing the CLI (`go install` path / prebuilt binaries — roadmap M2.1).
