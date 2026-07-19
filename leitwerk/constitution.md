# Constitution — leitwerk-devkit

This repository is developed under its own framework. The gate that governs it
is the same `leitwerk verify` it ships. This file is human-owned; an agent may
propose changes but may not edit it unilaterally.

## Purpose & scope
Provide the Leitwerk framework: a tool-agnostic verification gate (`core/`) plus
thin per-tool bindings (`bindings/`). The repository is also the reference
implementation the whitepaper points to — it must actually run.

## Invariants (never violate)
- **The core never depends on an agent runtime.** `core/` must run with only a
  shell present. No binding may be required for `leitwerk verify` to work.
- **Bindings never reimplement the gate.** A binding invokes and enforces the
  core CLI; it does not fork its logic. (`bindings/claude/bin/leitwerk` is a
  launcher, not a second implementation.)
- **A check never fakes a pass.** A check with nothing to run exits 2 (skip). A
  green gate means checks ran or honestly abstained — never that they were faked.
- **The gate config is human-owned.** An agent may add a check but may not lower
  a threshold, remove a check, or downgrade a path's tier without a change here.
- **Consuming repos do not edit installed core.** Project-specific checks live in
  the repo's `leitwerk/checks/` (per-check override); the built-in `core/checks/`
  are generic templates.

## Blast-radius policy
See `leitwerk/tiers.conf`. `core/bin` and `core/checks` and all `*.sh` are **T2**:
a defect in the gate weakens every adopting repo, which is the worst case here.
Bindings are T1; docs are T0.

## Definition of Done
`leitwerk verify` is green at the change's tier. For T2 (the gate itself):
JSON manifests parse, shell is syntax-clean (+ shellcheck when available), and
the CLI golden behaviour (`selftest`) holds.

## Roles in play
`architect` for CLI/structure changes, `test-engineer` to extend `selftest` when
CLI behaviour changes, `scout` for read-only retrieval. No `security-reviewer`
until the CLI handles untrusted input.

## Decisions of record
- 2026-07-19: Repo-local `leitwerk/checks/` per-check override added — surfaced
  by onboarding this repo onto itself (a consuming repo must not edit installed
  core). See `leitwerk/specs/self.md`.
