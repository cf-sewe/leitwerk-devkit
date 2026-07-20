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
agent-decided at a dream sweep (pruning; measurement is M4.1).

Edits to human-owned files follow the same split. Judgment edits — invariants,
thresholds, tier downgrades, priorities, the Definition of Done — are made
only by the human, from a proposal. Mechanical consequence edits of an
already-approved action — reference/path updates after an approved move, and
additive check wiring the invariants already permit — may be applied by the
agent via the audited staged-copy route and must be listed in the change
summary. Weakening moves remain human-only in every case.

## Definition of Done
`leitwerk verify` is green at the change's tier. For T2 (the gate itself):
JSON manifests parse, shell is syntax-clean (+ shellcheck when available), and
the CLI golden behaviour (`selftest`) holds.

## Roles in play
`architect` for CLI/structure changes, `test-engineer` to extend `selftest` when
CLI behaviour changes, `scout` for read-only retrieval. `security-reviewer` now
applies to gate code that reads untrusted input: `drift` parses spec content
(2026-07-20, M1.1), so a change to input-parsing gate code gets a security lens.

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
  `leitwerk/specs/archive/steering-alignment.md`.
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
  depends on it being present. See `leitwerk/specs/archive/workflow-orchestration.md`.
- 2026-07-19: Reimplemented the core CLI (`core/bin/leitwerk`) from Bash as a
  compiled Go binary. Go over Rust for an I/O-bound orchestration gate:
  static-by-default binaries, simple cross-compilation, a zero-dependency
  stdlib build, `//go:embed` for checks/templates (layout-independent), and
  `go install` distribution. The external contract is unchanged; checks stay
  shell scripts. Toolchain pinned with mise (Go + Node LTS). See
  `leitwerk/specs/archive/go-cli.md`.
- 2026-07-19: Context budgets are policy: `CLAUDE.md` ≤ 200 lines, each
  `.claude/rules/*.md` ≤ 100 lines, each skill/agent frontmatter description
  ≤ 80 words, always-on total ≤ 2000 estimated tokens. Enforced by the
  repo-local `context` check at every tier. The numbers are recorded here
  because a check script is agent-editable while budgets are policy. See
  `docs/reviews/20260719_191453-leitwerk-concept-review.md`.
- 2026-07-20: Pre-build phase depth. The phase skills gained an explicit
  research step (read-first, `scout` fan-out, facts carried as `file:line`
  tagged CONFIRMED/INFERRED) and a design step whose outcome lands in a new
  **Design decisions** section of the spec; plans carry a per-step status
  convention (`[ ]`/`[x]`/`[~]`) and T2 manual-verification criteria so a cold
  session resumes from the plan alone. No parallel document tree — durable
  output lands in the spec, the plan, or (after landing) here; the gate remains
  the only guarantee. See `leitwerk/specs/archive/phase-depth.md`.
- 2026-07-20: The spec/plan lifecycle is defined and enforced. Specs are one of
  two kinds — a living contract (`active`) or a change record — with states
  (`draft`/`active`/`landed`/`superseded`), owners, and transitions defined
  normatively in `core/templates/spec.template.md`; landing runs a "dreaming"
  pass that merges the durable core out and archives the spec + plan. The
  repo-local `lifecycle` check enforces the states mechanically at every tier (a
  `landed`/`superseded` record must be in `archive/`, a plan may not outlive its
  spec, an unknown or missing state is red), replacing prose the framework
  itself says cannot be a guarantee. Promotion of the check into `core/` ships
  with M1.4. See `leitwerk/specs/archive/lifecycle-check.md`.
- 2026-07-20: Decision routing — escalate decisions, not questions. A decision
  reaches the human only if it sets or changes intent, weakens or waives a
  guarantee, or accepts irreversible/residual risk; everything else the agent
  decides and records in the spec's Design decisions, with specialist roles
  carrying domain judgment before a human is asked. Durable escalations live as
  `leitwerk/proposals/` files, kept visible by the `lifecycle` check and (Claude
  binding) surfaced by a `SessionStart` hook and presented by `leitwerk-review`
  as multiple-choice. The authority version is the "Decision routing" section
  above. See `leitwerk/specs/archive/decision-routing.md`.
- 2026-07-20: Drift detection is real (M1.1). A spec declares the code it
  governs in an `## Anchors` section (`path` / `path#symbol`, globs allowed);
  the `drift` core check (wired at every tier) goes red when an anchor no longer
  resolves and — when a diff base is provided — when anchored code changed while
  its spec did not. It surfaces, never resolves; archived specs are ignored.
  Anchor paths are confined to the repo because a spec is untrusted input — this
  change crosses the "CLI handles untrusted input" line the "Roles in play" note
  named, and a security pass hardened it. The placeholder that always passed is
  gone. Auto-provisioning the diff base (CI / `verify --auto`) and a
  living-contract exemption for the one-sided check are deferred to M2.4 / M2.3.
  See `leitwerk/specs/archive/drift-detection.md`.
- 2026-07-20: The bugfix workflow (whitepaper §8.3, Workflow C) is a first-class
  entry path, not role prose. The `leitwerk-fix` skill sequences reproduce → pin
  (a failing regression test, sourced from the `test-engineer` charter) → fix at
  the change's tier → gate → review proportional to risk; it composes the role
  rather than restating it, and `bindings/open/AGENTS.md` mirrors the method for
  open code. The claim is proven executably by
  `examples/scenarios/s7-bugfix.sh`, which shows a defect the existing suite
  misses staying green until the pin reds the gate and the fix greens it — so
  "pin before you fix" is a runnable check, not advice. See
  `leitwerk/specs/archive/bugfix-workflow.md`.
- 2026-07-20: Diff-derived tier selection (M2.4). `leitwerk verify --auto`
  computes the highest blast-radius tier from the git diff inside the CLI and
  runs it, replacing the shell rank-loop CI duplicated — so the Stop hook and CI
  share one implementation (a binding reimplements less of the gate, not more).
  The diff base is taken ONLY from the `--base` flag, never `LEITWERK_DIFF_BASE`:
  that env var is drift's Part-2 signal, and an ambient value must not silently
  flip `--auto` from working-tree to committed-range and hide uncommitted work.
  Precondition failures (no git, unresolvable base) error rather than fall back
  to a low tier — the gate never silently under-verifies. This repo's own Stop
  hook stays pinned T2 (highest-stakes); the plugin ships `--auto`. Drift's
  one-sided check and its living-contract exemption remain deferred to M2.3,
  which will set `LEITWERK_DIFF_BASE` alongside the flag. See
  `leitwerk/specs/archive/verify-auto.md`.
