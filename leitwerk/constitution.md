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
- **Open-code guarantee-parity.** The hard guarantee lives in `core/` and is
  reachable via `leitwerk verify` + CI with no agent runtime present; bindings add
  ergonomics only. Enforced structurally by the `parity` check so it cannot erode
  silently. (Ergonomics like Claude Code workflows are not expected to reach open
  code — only the guarantee is.)
- **A check never fakes a pass.** A check with nothing to run exits 2 (skip). A
  green gate means checks ran or honestly abstained — never that they were faked.
- **The gate config is human-owned.** An agent may add a check but may not lower
  a threshold, remove a check, or downgrade a path's tier without a change here.
  This is enforced, not merely stated: the human-owned files (`[human-owned]` in
  the tiers file) are protected by `leitwerk guard`, which the Claude binding
  wires to a `PreToolUse` hook that blocks the edit. Open-code relies on review
  of the same paths. An agent proposes changes to these files; a human makes them.
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
- 2026-07-19: Aligned with Anthropic's Claude Code steering guidance. Absolute
  prohibitions became mechanical, not prose: `leitwerk guard` + a `PreToolUse`
  hook block edits to human-owned files. Tier discipline moved into a path-scoped
  `.claude/rules/` rule (loads when a T2 path is touched, not only inside a
  skill). A small `CLAUDE.md` bridges a fresh Claude session to this constitution;
  procedures stay in skills. Rules and CLAUDE.md are repo-level (not
  plugin-packageable), so `leitwerk init` scaffolds them and templates ship in
  `core/templates/`. The guard is a guardrail; the gate remains the hard
  guarantee, so open-code parity is unaffected. See
  `leitwerk/specs/steering-alignment.md`.
- 2026-07-19: Repo-local `leitwerk/checks/` per-check override added — surfaced
  by onboarding this repo onto itself (a consuming repo must not edit installed
  core). See `leitwerk/specs/self.md`.
- 2026-07-19: Layer-2 orchestration is realized by **native dynamic workflows**
  on Claude Code, not a hand-rolled orchestrator role (now retired). Role
  subagents are reused as workflow `agentType`s; review scales by tier (workflow
  for T2). The workflow's multi-agent verification is *soft* and never replaces
  the external gate — `leitwerk verify` (Layer 3) stays authoritative. Plugins
  cannot package workflows, so the workflow ships as a core template that
  `leitwerk init` scaffolds into `.claude/workflows/leitwerk-review.mjs`; the
  skills prefer it and fall back to spawning roles directly, so review never
  depends on it being present. See `leitwerk/specs/workflow-orchestration.md`.
