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

## Decision routing
A decision is escalated to the human only if it (1) sets or changes intent —
scope, priorities, spec approval; (2) weakens or waives a guarantee —
thresholds, checks, tier downgrades; or (3) accepts irreversible or residual
risk — T2 sign-off, data/money/auth, shipping with an open finding, accepting
a known limitation. Everything else the agent decides, records in the spec's
Design decisions, and keeps reversible; domain judgment goes to a specialist
role before it goes to a human. Escalations are decision-ready (options,
evidence, recommendation, default) — durable ones as files in
`leitwerk/proposals/`, applied or rejected and then deleted by the human; the
`lifecycle` check keeps open proposals visible on every gate run. A question
class whose answers never diverge from the recommendation is retired to
agent-decided at a dream sweep (pruning; measured under the
`efficiency-evaluation` roadmap item).

Edits to human-owned files follow the same split. Judgment edits — invariants,
thresholds, tier downgrades, priorities, the Definition of Done — are made
only by the human, from a proposal. Mechanical consequence edits of an
already-approved action — reference/path updates after an approved move, and
additive check wiring the invariants already permit — may be applied by the
agent via the audited staged-copy route and must be listed in the change
summary. Weakening moves remain human-only in every case.

## Definition of Done
`leitwerk verify` is green at the change's tier. The authoritative tier→check
mapping is `leitwerk/tiers.conf` (human-owned); it is deliberately not restated
here, so this section cannot fall out of sync with the gate. (For the gate
itself, T2 is the widest set and adds the CLI golden-behaviour `selftest`.)

## Roles in play
`architect` for CLI/structure changes, `test-engineer` to extend `selftest` when
CLI behaviour changes, `scout` for read-only retrieval. `security-reviewer` now
applies to gate code that reads untrusted input: `drift` parses spec content
(2026-07-20), so a change to input-parsing gate code gets a security lens.

## Decisions of record
- 2026-07-19: Aligned with Anthropic's Claude Code steering guidance — absolute
  prohibitions became mechanical (`leitwerk guard` + a `PreToolUse` hook block
  edits to human-owned files), tier discipline moved into a path-scoped
  `.claude/rules/` rule, and a small `CLAUDE.md` bridges a fresh session to this
  constitution while procedures stay in skills. Rules and `CLAUDE.md` are
  repo-level, so `leitwerk init` scaffolds them from `core/templates/`. The guard
  is a guardrail; the gate remains the hard guarantee, so open-code parity is
  unaffected. See `leitwerk/specs/archive/steering-alignment.md`.
- 2026-07-19: Repo-local `leitwerk/checks/` per-check override added — surfaced by
  onboarding this repo onto itself (a consuming repo must not edit installed
  core). See `leitwerk/specs/self.md`.
- 2026-07-19: Layer-2 orchestration is realized by native dynamic workflows on
  Claude Code, not a hand-rolled orchestrator role (retired). Role subagents are
  reused as workflow `agentType`s; review scales by tier. The workflow's
  multi-agent verification is *soft* and never replaces the external gate —
  `leitwerk verify` stays authoritative. It ships as a core template `leitwerk
  init` scaffolds into `.claude/workflows/leitwerk-review.mjs`; the skills prefer
  it and fall back to spawning roles directly, so review never depends on it. See
  `leitwerk/specs/archive/workflow-orchestration.md`.
- 2026-07-19: Reimplemented the core CLI (`core/bin/leitwerk`) from Bash as a
  compiled Go binary — static-by-default binaries, simple cross-compilation, a
  zero-dependency stdlib build, `//go:embed` for checks/templates, and `go
  install` distribution. The external contract is unchanged; checks stay shell
  scripts. Toolchain pinned with mise (Go + Node LTS). See
  `leitwerk/specs/archive/go-cli.md`.
- 2026-07-19: Context budgets are policy — `CLAUDE.md` ≤ 200 lines, each
  `.claude/rules/*.md` ≤ 100 lines, each skill/agent frontmatter description
  ≤ 80 words, always-on total ≤ 2000 estimated tokens; enforced by the repo-local
  `context` check at every tier. Recorded here because a check script is
  agent-editable while the budgets are policy. See
  `docs/reviews/20260719_191453-leitwerk-concept-review.md`.
- 2026-07-20: Pre-build phase depth — the phase skills gained a research step
  (read-first, `scout` fan-out, facts carried as `file:line` tagged
  CONFIRMED/INFERRED) and a design step whose outcome lands in a spec **Design
  decisions** section; plans carry a per-step status convention (`[ ]`/`[x]`/
  `[~]`) and T2 manual-verification criteria so a cold session resumes from the
  plan alone. Durable output lands in the spec, the plan, or here — no parallel
  document tree. See `leitwerk/specs/archive/phase-depth.md`.
- 2026-07-20: The spec/plan lifecycle is defined and enforced. Specs are a living
  contract (`active`) or a change record, with states (`draft`/`active`/`landed`/
  `superseded`) defined normatively in `core/templates/spec.template.md`; landing
  runs a "dreaming" pass that merges the durable core out and archives the spec +
  plan. The repo-local `lifecycle` check enforces the states mechanically at every
  tier (a `landed`/`superseded` record must be in `archive/`, a plan may not
  outlive its spec, an unknown state is red). Decision (2026-07-24): the check
  stays repo-local; promoting it to a `core/` built-in (so adopters get lifecycle
  enforcement, not just the convention) is deferred low-priority work, tracked as
  the `lifecycle-core-promotion` roadmap item — artifact hygiene, not the code
  guarantee. See `leitwerk/specs/archive/lifecycle-check.md`.
- 2026-07-20: Escalate decisions, not questions — a decision reaches the human
  only to set or change intent, weaken or waive a guarantee, or accept
  irreversible/residual risk; everything else the agent decides and records in the
  spec's Design decisions, with specialist roles carrying domain judgment first.
  Durable escalations live as `leitwerk/proposals/` files, kept visible by the
  `lifecycle` check and (Claude binding) surfaced at `SessionStart` and presented
  by `leitwerk-review` as multiple-choice. The authority version is the "Decision
  routing" section above. See `leitwerk/specs/archive/decision-routing.md`.
- 2026-07-20: Drift detection is real. A spec declares the code it governs in an
  `## Anchors` section (`path` / `path#symbol`, globs allowed); the `drift` core
  check (every tier) goes red when an anchor no longer resolves and — given a diff
  base — when anchored code changed while its spec did not. It surfaces, never
  resolves; archived specs are ignored. Anchor paths are repo-confined because a
  spec is untrusted input — the "Roles in play" input-parsing line — and a
  security pass hardened it. Auto-provisioning the diff base landed (see the
  diff-derived tier entry below); drift's one-sided check and its living-contract
  exemption remain deferred and are not currently tracked in the roadmap. See
  `leitwerk/specs/archive/drift-detection.md`.
- 2026-07-20: Diff-derived tier selection. `leitwerk verify --auto` computes the
  highest blast-radius tier from the git diff inside the CLI, so the Stop hook and
  CI share one implementation (a binding reimplements less of the gate, not more).
  The diff base is taken ONLY from `--base`, never `LEITWERK_DIFF_BASE`: that env
  var is drift's Part-2 signal, and an ambient value must not silently flip
  `--auto` from working-tree to committed-range and hide uncommitted work.
  Precondition failures (no git, unresolvable base) error rather than fall back to
  a low tier — the gate never silently under-verifies. This repo's own Stop hook
  stays pinned T2 (highest-stakes); the plugin ships `--auto`. See
  `leitwerk/specs/archive/verify-auto.md`.
- 2026-07-24: Conventional Commits adopted repo-wide. `release-please` derives
  the version and changelog from commit history; because PRs are squash-merged,
  the PR **title** is the commit it reads, so a `semantic-pr` check enforces the
  Conventional-Commit format on PR titles. Allowed types and scopes live in
  `.gitmessage` (the commit template) and `CONTRIBUTING.md`; scopes are advisory.
  No leitwerk gate check parses commit messages — this is a PR-gating CI check,
  so the gate's open-code parity is unaffected. See
  `leitwerk/specs/cli-publish.md`.
