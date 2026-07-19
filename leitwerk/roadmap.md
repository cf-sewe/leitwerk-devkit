# Roadmap — leitwerk-devkit

The ordered backlog of work needed to take Leitwerk from a working scaffold to a
framework whose claims are all backed by running code.

## How this fits the Leitwerk process

Leitwerk defines a per-change lifecycle — spec → plan → build → review, gated at
the change's blast-radius tier. It does **not** define a separate "roadmap"
artifact, and it does not need one: a roadmap is the ordered set of *future
specs*. Each item below is a **proto-spec** — problem, target behaviour, tier,
and the acceptance condition that makes the gate green. Nothing here bypasses the
gate; the roadmap only decides order.

Lifecycle of a roadmap item:

1. **Proposed** — an entry in this file (spec vocabulary, not yet started).
2. **Active** — promoted to `leitwerk/specs/<slug>.md` via the `leitwerk-spec`
   skill; if non-trivial, `leitwerk-plan` breaks it into gated steps.
3. **Built** — `leitwerk-build` implements each step; `leitwerk verify` must be
   green at the step's tier before it lands.
4. **Reviewed** — `leitwerk-review` runs the tier's roles and confirms the spec
   still matches the code, then the change merges on a green gate.

This file is human-owned, like the constitution: agents may propose entries or
reorder by evidence, but the human owns priority. Completed items move to the
constitution's decisions of record.

> **Proposed decision of record (needs human acceptance):** add "roadmap =
> ordered backlog of future specs" to the framework's artifact model, and record
> it in `leitwerk/constitution.md`. Per the constitution's own rule, an agent may
> not edit it unilaterally — this is the proposal.

## Backlog

Tiers refer to `leitwerk/tiers.conf`. Order is by leverage: make the claims true
before making it broadly adoptable.

### Milestone 1 — make the framework's claims true

**M1.1 · drift-detection** · tier **T2** (`core/checks/drift.sh`)
- *Problem:* "surface spec↔code drift, don't resolve it" is the headline
  principle, but `drift.sh` only counts spec files and always reports no drift.
- *Behaviour:* a spec declares the code it governs (path/symbol anchors); the
  check flags (a) an anchor that no longer resolves, and (b) governed code that
  changed in a range where its spec did not, or vice-versa. It surfaces and
  exits non-zero for a human to reconcile; it never edits either side.
- *Acceptance:* on a repo where a spec's anchored symbol is renamed without
  updating the spec, `leitwerk verify` goes red with a specific, human-readable
  divergence; on a consistent repo it stays green.
- *Roles/skills:* `architect` (anchor format), `test-engineer` (add drift cases
  to `selftest`).

**M1.2 · reference-app-real** · tier **T1/T2**
- *Problem:* `examples/reference-app` has only a constitution; the gate runs but
  governs nothing, so it shows execution, not real governance.
- *Behaviour:* a minimal real app with a spec, a failing→passing test wired into
  the `tests` check, and at least one T2 path (e.g. a migration) so tier
  escalation is demonstrated on real code.
- *Acceptance:* `leitwerk verify` on the example runs actual tests (not skips)
  and a deliberately broken change turns it red.
- *Roles/skills:* `leitwerk-spec`, `leitwerk-build`, `test-engineer`.

**M1.3 · selftest-coverage** · tier **T2**
- *Problem:* `selftest` covers four tier assertions and one gate run; `init`,
  `drift`, `checks_for_tier`, the glob-engine edge cases, and error paths are
  untested.
- *Behaviour:* extend the CLI's golden suite to cover glob edge cases (`**/`
  optional segment, catch-all), `leitwerk init` output, and non-zero exit paths.
- *Acceptance:* mutating the glob translation or the tier table fails `selftest`.

### Milestone 2 — make it adoptable

**M2.1 · cli-publish** · tier **T1**
- *Problem:* CI and `bindings/open` reference `npm i -g @cf-sewe/leitwerk`, which
  does not exist; only the in-repo path works. Confirmed constraint: a Claude Code
  marketplace install sparse-copies **only** the plugin subdir, so the plugin's
  launcher cannot reach a sibling `core/` — an adopter must set `LEITWERK_HOME`
  until the package exists. This makes publishing the real unblock for a
  marketplace-only adoption, not a nicety.
- *Behaviour:* publish `core/` as the `@cf-sewe/leitwerk` package (or document a
  vendoring path) so adopters and CI do not depend on the repo layout.
- *Acceptance:* a clean machine can `npm i -g @cf-sewe/leitwerk` and run
  `leitwerk verify` in an unrelated repo.

**M2.2 · plugin-live-validation** · process (no tier — validation, not a code change)
- *Problem:* the plugin, skills, agents, and the review **workflow** are written
  but never run in a live Claude Code session; their composition is inferred,
  not proven.
- *Behaviour:* install the plugin from the marketplace and drive one real feature
  end to end (spec → plan → build → review) with the Stop-hook gate active; for a
  T2 change, launch `.claude/workflows/leitwerk-review.mjs` and confirm it spawns
  the role subagents as `agentType`s and its findings are advisory to the gate.
- *Acceptance:* a documented session transcript where the roles are spawned (incl.
  via the workflow), the gate blocks a red turn-end, and a green change lands.
  Findings feed back into the skills/agents/workflow (bidirectional refinement).

**M2.3 · ci-live** · tier **T2** (`.github/`)
- *Problem:* the CI workflow has never executed (no remote repo yet).
- *Behaviour:* push to `cf-sewe`, run the gate as a required check on a PR, and
  confirm tier selection from the diff works on GitHub's runner.
- *Acceptance:* a PR with a T2 change is blocked until the gate is green.

### Milestone 3 — verification depth

**M3.1 · erosion-budgets** · tier **T2** — make `erosion.sh` enforce real
complexity/duplication ceilings instead of skipping when no analyzer is present;
define default budgets in the constitution.

**M3.2 · verification-helpers** — provide the substance the whitepaper describes:
scaffolds/guidance for property, mutation, and characterization tests, so
`leitwerk-build` has concrete oracles to reach for rather than prose.

**M3.3 · provenance-tooling** — turn CONFIRMED/INFERRED/GAP from a review
convention into something the review output actually carries and the gate can
check for (e.g. no GAP left on a T2 change without a human sign-off).

**M3.4 · open-code-live-validation** — the M2.2 equivalent for Codex/AGENTS.md:
drive one change under an open-code agent and confirm the CI gate is the binding
constraint.

**M3.5 · monorepo-affected-scope** · tier **T2** (`core/bin/leitwerk` + checks) ·
*proposed (agent-suggested, needs human priority)*
- *Problem:* the gate is diff-aware only for **tier selection**; the checks
  themselves run project-wide (`go test ./...`, `tsc`, …). On a monorepo that
  re-validates untouched code on every change — slow, and it blurs which package
  a red result belongs to. But naive "validate only touched files" is unsafe: a
  change to a shared package can break an untouched **dependent**.
- *Behaviour:* the gate validates the **affected set** — changed files plus their
  reverse-dependencies — not the whole repo and not only the literally-touched
  files. The affected set is computed deterministically from the diff plus a
  dependency graph and passed to the checks (e.g. a `$LEITWERK_CHANGED` set the
  check scripts scope themselves to). Package boundaries and per-package checks
  are human-owned policy, expressed with the existing repo-local `leitwerk/`
  override (a per-package `tiers.conf` / `checks/`). The agent may not narrow its
  own scope — scoping is part of the gate, not agent judgment.
- *Safety:* scoping trades completeness for speed, so pair a **scoped per-change
  gate** (fast, pre-merge) with a **periodic unscoped full gate** (nightly / on
  the protected branch) that catches what a stale graph or wrong exclusion missed.
  A T2 change may widen the scope.
- *Acceptance:* on a two-package monorepo fixture, editing package A runs A's
  checks (and B's iff B depends on A) and skips unrelated package C; a
  deliberately broken dependent still turns the gate red; the full nightly gate
  runs everything.
- *Roles/skills:* `architect` (affected-set model + boundary policy),
  `test-engineer` (fixture + `selftest` cases for scope correctness).

## Recently decided (done)
- **Aligned with Claude Code steering guidance.** Turned prose prohibitions into
  mechanism: `leitwerk guard` (core) + a `PreToolUse` hook block edits to
  human-owned files; a path-scoped `.claude/rules/` rule carries T2 discipline; a
  small `CLAUDE.md` bridges a session to the constitution with procedures left in
  skills. Rules/CLAUDE.md are repo-level (not plugin-packageable) → `leitwerk
  init` scaffolds them, templates in `core/templates/`. Guard is a guardrail; the
  gate stays the hard guarantee (parity intact). Tested (guard + hook + selftest).
  See `leitwerk/specs/steering-alignment.md`.
- **Open-code guarantee-parity guarded by construction.** Split compatibility
  into guarantee-parity (must hold) vs ergonomics-parity (don't chase). Added the
  `parity` structural check (fails if gate logic leaks into a binding), a
  constitution invariant, and a CI job that runs the core gate with no agent
  runtime. Negative-tested. Live open-code validation stays deferred to **M3.4**
  (do when stable). See `leitwerk/specs/open-code-parity.md`.
- **Layer-2 orchestration → native dynamic workflows.** Retired the hand-rolled
  orchestrator role; roles reused as workflow `agentType`s; added
  `.claude/workflows/leitwerk-review.mjs`; review scales by tier; workflow
  verification is soft, the external gate stays authoritative. See
  `leitwerk/specs/workflow-orchestration.md` and the constitution's decisions.
  A JS syntax check for workflow scripts in the gate is a small open follow-up.

## Not planned (explicit non-goals for now)
- A bespoke spec DSL — specs stay Markdown.
- Auto-resolving drift — the framework surfaces, humans decide. This is a
  constitution invariant, not a backlog gap.
