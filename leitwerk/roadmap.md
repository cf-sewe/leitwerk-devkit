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

**M1.3 · selftest-coverage** · tier **T2**
- *Problem:* `selftest` covers four tier assertions and one gate run; `init`,
  `drift`, `checks_for_tier`, the glob-engine edge cases, and error paths are
  untested.
- *Behaviour:* extend the CLI's golden suite to cover glob edge cases (`**/`
  optional segment, catch-all), `leitwerk init` output, and non-zero exit paths.
- *Acceptance:* mutating the glob translation or the tier table fails `selftest`.

### Milestone 2 — make it adoptable

**M2.1 · cli-publish** · tier **T1**
- *Problem:* the gate is distributed only as source; adopters and CI build it
  via `go install github.com/cf-sewe/leitwerk-devkit/core/cmd/leitwerk@latest`,
  which resolves only once the repository is public. Confirmed constraint: a
  Claude Code marketplace install sparse-copies **only** the plugin subdir, so
  the plugin's launcher cannot reach a sibling `core/` — an adopter must set
  `LEITWERK_HOME` or have a globally installed binary until publication. That
  makes publishing the real unblock for a marketplace-only adoption, not a
  nicety.
- *Behaviour:* make the module path publicly resolvable (`go install …@latest`
  works from a clean machine) and/or attach prebuilt static binaries to
  releases; document vendoring as the third path.
- *Acceptance:* a clean machine can obtain the binary (go install or release
  download) and run `leitwerk verify` in an unrelated repo.

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

**M2.4 · verify --auto** · tier **T2**
- *Problem:* the Stop hook verifies at a static tier (`$LEITWERK_TIER`, plugin
  default T1; this repo pins T2) — it cannot see the change's real blast
  radius. A T2 change can end a turn having passed only T1 locally; a docs-only
  turn overpays. Only CI derives the tier from the diff.
- *Behaviour:* `leitwerk verify --auto` computes the highest tier from the git
  diff (same first-match rule CI uses) and runs that tier; the hook templates
  and CI share it. Optional later extension: diff-signal triggers (auth paths →
  security review) once the tier derivation is proven.
- *Acceptance:* `selftest` covers tier derivation from a mixed diff (docs-only →
  T0, one migration → T2); the scaffolded hook uses `--auto`.
- *Roles/skills:* `architect` (diff-base semantics), `test-engineer`.

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

**M3.6 · required-checks** · tier **T2**
- *Problem:* a check that skips at T2 (e.g. `sast` when semgrep is absent)
  leaves the gate green; "a skipped security check on a T2 change is a blocker"
  is role prose, not mechanism.
- *Behaviour:* `[tiers]` syntax marks checks that may not skip at a tier (e.g.
  `T2 = … sast!`); a skip of a required check turns the gate red with a message
  naming the missing tool.
- *Acceptance:* scenario/`selftest` case: empty toolchain at T2 with `sast`
  required → exit 1; with the tool installed → normal pass/fail.

**M3.7 · repo-map** · tier **T2** (core) — design-proposal open decision O2
- *Problem:* the research step retrieves via scouts running grep/glob; P4
  (retrieve-don't-preload) calls for a structural symbol map, and context-rot
  evidence says preloading does not scale with repo size.
- *Behaviour:* the CLI builds/refreshes a tree-sitter symbol map (Aider-style
  ranking) that scouts query for "where is X / who calls X"; a typed code
  graph stays deferred until a repo exceeds ~1k files (per O2).
- *Acceptance:* on a fixture repo, a scout answers a where-is/who-calls
  question from the map without full-text search; the map refreshes
  incrementally and stale entries are detectable.
- *Roles/skills:* `architect` (map format/CLI surface), `test-engineer`.

### Milestone 4 — measure the framework itself

**M4.1 · efficiency-evaluation** · process (periodic, not per-change)
- *Problem:* deterministic proxies are gated (the `context` budget check), but
  outcome efficiency — tokens-to-green-gate, wall time to a landed change,
  human interventions per change — is unmeasured, and no off-the-shelf
  benchmark measures a governance framework's overhead.
- *Behaviour:* an A/B harness: a fixed task set (the runnable scenarios in
  `examples/scenarios/` plus reference-app changes) driven with and without
  Leitwerk, several runs per task, reporting cost and outcome distributions.
  Candidate substrates: long-horizon suites (SWE-bench-Pro-class) and
  terminal-agent suites (Terminal-Bench-class) with cost reporting.
- *Acceptance:* a first dated report in `docs/reviews/` comparing the two arms
  on cost and outcome.

## Recently decided (done)
- **Aligned with Claude Code steering guidance.** Turned prose prohibitions into
  mechanism: `leitwerk guard` (core) + a `PreToolUse` hook block edits to
  human-owned files; a path-scoped `.claude/rules/` rule carries T2 discipline; a
  small `CLAUDE.md` bridges a session to the constitution with procedures left in
  skills. Rules/CLAUDE.md are repo-level (not plugin-packageable) → `leitwerk
  init` scaffolds them, templates in `core/templates/`. Guard is a guardrail; the
  gate stays the hard guarantee (parity intact). Tested (guard + hook + selftest).
  See `leitwerk/specs/archive/steering-alignment.md`.
- **Open-code guarantee-parity guarded by construction.** Split compatibility
  into guarantee-parity (must hold) vs ergonomics-parity (don't chase). Added the
  `parity` structural check (fails if gate logic leaks into a binding), a
  constitution invariant, and a CI job that runs the core gate with no agent
  runtime. Negative-tested. Live open-code validation stays deferred to **M3.4**
  (do when stable). See `leitwerk/specs/archive/open-code-parity.md`.
- **Layer-2 orchestration → native dynamic workflows.** Retired the hand-rolled
  orchestrator role; roles reused as workflow `agentType`s; added
  `.claude/workflows/leitwerk-review.mjs`; review scales by tier; workflow
  verification is soft, the external gate stays authoritative. See
  `leitwerk/specs/archive/workflow-orchestration.md` and the constitution's decisions.
  (The JS syntax check for workflow scripts landed — `selftest` §9; see
  `leitwerk/specs/archive/workflow-syntax-check.md`.)
- **M1.1 · drift-detection — done (2026-07-20).** `drift` is real: specs declare
  `## Anchors` (`path`/`path#symbol`); unresolved anchors are red, one-sided
  spec/code change is red under a diff base, archived specs are ignored, and
  anchor paths are confined (untrusted input). See
  `leitwerk/specs/archive/drift-detection.md`.
- **M1.2 · reference-app-real — done (2026-07-20).** `examples/reference-app` is
  a minimal Go app with a spec, a real test the gate runs, and a T2 migration;
  `selftest` §4 runs its tests and scenario `s6` proves broken→red. Unblocks
  M1.5. See `leitwerk/specs/archive/reference-app-real.md`.
- **M1.4 · spec-lifecycle — done (2026-07-20).** Terminal states enforced by the
  `lifecycle` check, archived specs ignored by `drift` (via M1.1), the landing
  "dreaming" pass practiced at review; a mechanized periodic sweep stays a
  practice, not a tool. See `leitwerk/specs/archive/lifecycle-check.md` and the
  constitution's decisions.
- **M1.5 · bugfix-workflow — done (2026-07-20).** Workflow C (whitepaper §8.3) is
  a first-class entry path: the `leitwerk-fix` skill (reproduce → pin → fix at
  tier, composing `test-engineer`), with `bindings/open/AGENTS.md` mirroring the
  method, proven by `examples/scenarios/s7-bugfix.sh` — a defect that escapes the
  suite stays green until a regression test pins it (gate red), then the fix
  greens it. See `leitwerk/specs/archive/bugfix-workflow.md`.
- **Bash portability & workflow syntax (2026-07-20).** The gate's own checks run
  on macOS bash 3.2 (no `mapfile`), and `selftest` syntax-checks workflow
  `.mjs`. See `leitwerk/specs/archive/bash-portability.md` and
  `leitwerk/specs/archive/workflow-syntax-check.md`.

## Not planned (explicit non-goals for now)
- A bespoke spec DSL — specs stay Markdown.
- Auto-resolving drift — the framework surfaces, humans decide. This is a
  constitution invariant, not a backlog gap.
