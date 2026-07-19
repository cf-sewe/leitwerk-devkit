# Design Proposal: A Framework for AI-Native Application Development

**Date:** 2026-07-19
**Status:** Draft for review
**Companion:** `20260719_101757-Research-Briefing.md` (evidence base, ~90 sources)

## 1. Purpose and scope

A generic, shareable framework for building and maintaining applications where
AI agents write most of the code and humans own requirements and review. It must
work across the full project lifecycle — **greenfield** (new), **brownfield**
(existing/legacy), and **long-lived** (evolving over months/years and many
sessions) — and degrade gracefully when a human expert writes code by hand.

Packaging target: **one or more Claude Code skills** plus a small set of
repo-local artifacts and one portable verification command. Existing skills
(research / plan / implement / validate) are retired once this is proven.

Working name for the framework in this document: **Keel** (a placeholder — the
structural spine a hull is built around; rename freely). Skills are prefixed
`keel-*`.

Non-goals: not a new IDE or model; not tied to any one product or stack; not full
spec-as-source regeneration (see §3, decision D1).

## 2. Design principles (the non-negotiables)

These are derived directly from the evidence and are the parts we should not
compromise, because the research shows each failure mode is real and measured.

- **P1 — Generation is cheap; acceptance is strict.** The LLM is treated as a
  fallible generator. Correctness is decided by an external gate the generating
  agent cannot author or weaken for itself. (Clover: 0 false positives is the
  target property; Veracode: ~45% of AI code fails security tests; iterative
  self-refinement *raised* critical vulns +37.6% — self-review is not a gate.)
- **P2 — Enforce with executable artifacts, never prose.** Requirements are kept
  honest by tests, contracts, and CI checks, not by documents. Prose specs are
  context, not enforcement. (Amazon found tests-alone insufficient for *intent*,
  but still enforce via executables and use prose only as context.)
- **P3 — Pin behavior before changing it.** On any code the framework did not
  itself generate under gate, characterization/golden-master tests capture
  current behavior first. (Feathers; Google's migration work depends on a strong
  build+test signal existing first.)
- **P4 — Retrieve, don't preload.** A structural repo map / code graph is a
  first-class always-available object; the context window stays small and
  high-signal. (Context rot: every model degrades as input grows — "a 1M window
  still rots at 50K"; graph navigation beat semantic-only by ~20pp on
  hidden-dependency tasks; weaker-model-with-context beat stronger-without.)
- **P5 — Recovery from a cold context window is the default.** Durable artifacts
  (plan, progress, decisions) live on disk and are re-read every session; no
  critical state lives only in a conversation. (Anthropic long-running-agent
  harness; self-conditioning: past errors compound within a session.)
- **P6 — Tier rigor by blast radius.** Read-only/display code gets light gates;
  state-mutating gets behavioral + invariant gates; irreversible/infra gets
  safety checks + mandatory human sign-off. Rigor goes where the damage is.
- **P7 — Review behavior, not diffs; surface drift, don't auto-resolve.** Humans
  review specs, tests, and rendered results — the things they judge well — and
  are pointed at spec↔code divergence to decide, never handed raw diffs to
  rubber-stamp. (AI code "looks clean even when wrong"; automation bias raises
  following-bad-advice ~26%; METR perception gap: devs felt +20% but were −19%.)
- **P8 — Budget against erosion.** Duplication, complexity, and churn have hard
  ceilings that fail the build, plus scheduled consolidation passes. (GitClear:
  duplication up 8–12%, refactoring down 25%→<10% under AI authorship; LinearB:
  ~1.7x issues/PR, 30–41% debt increase.)
- **P9 — Human owns decomposition, seams, and architecture.** Agents excel at
  local/mechanical/high-volume change and are weak at system-wide reasoning
  (43% vs 55% high-level refactors; raw spec-extraction F1 ~0.64). The framework
  keeps humans on the calls agents demonstrably get wrong.

## 3. Key decisions (made, with rationale)

- **D1 — Baseline is spec-anchored, not spec-as-source.** Spec and code
  co-evolve, bound by executable checks. Full regenerate-only (Tessl) is
  pre-production and fights non-determinism; it also can't accommodate a human
  writing code by hand, which is a stated requirement. *Spec-as-source is
  allowed opportunistically per well-bounded module (e.g. a generated client, a
  pure UI page) but never required.*
- **D2 — The gate is one deterministic, portable command, hardened by an
  optional hook.** `keel verify` is a plain executable (script/CLI) that any CI
  can run — this keeps the framework shareable and stack-agnostic. Teams on
  Claude Code additionally wire it as a Stop / pre-commit **hook** so the agent
  cannot end a task or commit without a green gate. Portability first, harness
  hardening opt-in. Rationale: the strongest guarantee (agent literally cannot
  skip the gate) shouldn't be the thing that blocks other teams from adopting.
- **D3 — Five skills covering the whole lifecycle**, not one monolith and not the
  current four. The onboarding step (greenfield bootstrap vs. brownfield
  adoption) is distinct enough to warrant its own skill, and behavior review is
  distinct from building. Set: `keel-onboard`, `keel-spec`, `keel-plan`,
  `keel-build`, `keel-review` (§10). Verification is not a skill — it is D2's
  executable, so it can't be talked around.
- **D4 — Drift is surfaced, not auto-resolved.** A drift sensor reports spec↔code
  divergence on every change and escalates to a human decision point; no
  automatic winner. This makes bidirectional refinement safe without relying on
  unproven auto-regeneration.
- **D5 — Every derived artifact carries provenance and freshness.** Extracted
  specs and generated docs are labeled CONFIRMED / INFERRED / GAP (after Reversa)
  and anchored to source so staleness is detectable — because doc/spec rot is the
  #1 failure mode of every "codified context" approach.

## 4. Architecture: three layers

The central idea is to separate what the agent *drives* from what it *cannot
skip or edit*.

```
┌─────────────────────────────────────────────────────────────┐
│ LAYER 1 — DURABLE ARTIFACTS (git-versioned, re-read each session) │
│  constitution · living specs · decision log · progress/feature  │
│  list · blast-radius map · repo map/code graph                   │
└─────────────────────────────────────────────────────────────┘
              ▲ read every session          ▲ updated under gate
┌─────────────────────────────────────────────────────────────┐
│ LAYER 2 — SKILLS (agent-driven phases)                          │
│  keel-onboard · keel-spec · keel-plan · keel-build · keel-review │
└─────────────────────────────────────────────────────────────┘
              │ every change must pass ▼
┌─────────────────────────────────────────────────────────────┐
│ LAYER 3 — GATES (deterministic, agent-cannot-edit)              │
│  keel verify (tiered Definition of Done) · drift sensor ·       │
│  erosion budgets   → run in CI, optionally as Claude Code hooks  │
└─────────────────────────────────────────────────────────────┘
```

## 5. Layer 1 — Durable artifacts

All live in the repo under `keel/` (or the team's chosen path) and in
`AGENTS.md` / `CLAUDE.md`, versioned with the code.

- **Constitution** (`keel/constitution.md`) — small (~1 page), stable. Human-owned.
  Architectural principles, the blast-radius tier definitions, which gates are
  mandatory per tier, and hard boundaries (✅ autonomous / ⚠️ ask-first /
  🚫 never — after Osmani). Re-read every session. This is the one document that
  is *authority*, not advice.
- **Living specs** (`keel/specs/<feature>.md`) — per feature/module. Behavior and
  acceptance criteria as Given/When/Then. Human-owned, AI-refinable **with
  confirmation**. Each spec fragment is anchored to the code/tests that satisfy
  it (for drift detection, §8). Not a rigid upfront contract — a description kept
  consistent with reality.
- **Decision log** (`keel/decisions/NNNN-*.md`) — ADRs: context, decision,
  alternatives, consequences. Captures the *why* so later sessions don't
  re-litigate (MAST: ~42% of multi-agent failures are specification/decision
  issues). Loaded on demand, keyword-searchable.
- **Progress / feature list** (`keel/progress.json`) — granular features with
  pass/fail status; every planned feature starts "failing." Prevents premature
  "done" and drives cold-start recovery (Anthropic harness pattern).
- **Blast-radius map** (`keel/blast-radius.md` + per-task tags) — classification
  rules and current assignments (§6).
- **Repo map / code graph** (`keel/repomap` cache) — tree-sitter AST +
  PageRank-ranked symbol map (Aider-style), optionally a typed code graph
  (IMPORTS/CALLS/INHERITS) + hybrid semantic index for large repos. Regenerated
  incrementally; the always-on answer to P4.

Freshness discipline (D5): a nightly/`keel verify` check confirms anchors still
resolve; unresolved anchors = drift, surfaced to a human.

## 6. Blast-radius tiering

The organizing axis for how much rigor applies. Defined in the constitution so
it is human-anchored, not agent-chosen.

| Tier | Examples | Required gate (adds to lighter tiers) | Human checkpoint |
|---|---|---|---|
| **T0 read-only / display** | UI rendering, formatting, read queries, logging | compile/types · lint · SAST on diff · dep allowlist · erosion budgets | visual/functional review only |
| **T1 state-mutating** | writes, business logic, workflows, non-destructive APIs | + property/contract tests · state invariants · mutation-score floor · behavior-level acceptance tests | spec + behavior review |
| **T2 irreversible / infra** | deletes, migrations, money/permissions, infra/config, auth | + explicit safety checks on the destructive op · characterization tests on affected legacy · dry-run/rollback proof | **mandatory human sign-off** |

Rationale: correctness is largely invisible in the UI, so "the page looks good"
covers T0 but is necessary-not-sufficient for T1/T2. This also matches the
standing rule that irreversible/destructive operations default to admin-level
rigor rather than reusing ordinary edit paths.

## 7. Lifecycle: how the framework enters a project

`keel-onboard` detects the situation and does the right thing.

### 7a. Greenfield
1. Interview to draft the **constitution** (stack, principles, tier rules,
   boundaries) — the one heavy human-input moment.
2. Scaffold the artifact tree, CI wiring for `keel verify`, and the optional hook.
3. Seed one exemplar per code type (a model test, a component, an API handler) at
   high quality — the agent thereafter matches this style (Willison: quality is
   self-reinforcing or self-degrading from the seed).
4. First feature list → enter the core loop (§8).

### 7b. Brownfield (existing / legacy)
1. **Index, don't ingest** — build the repo map / code graph (P4).
2. **Reverse-derive specs** with provenance: label every claim CONFIRMED /
   INFERRED / GAP (D5); INFERRED/GAP require human validation before they gain
   authority. (Raw LLM spec extraction is ~0.64 F1 — never trusted unlabeled.)
3. **Pin behavior** — generate characterization/golden-master tests over the
   areas about to change *before* any edit (P3). These become the regression net.
4. Generate the constitution *from* observed conventions, human-confirmed.
5. Enter the loop using the **strangler-fig / seam** discipline (§8, §9): small,
   reversible, seam-bounded units; legacy path kept as fallback; dual-write +
   compare where data moves (Shopify's 7-step recipe is directly automatable).

### 7c. Long-lived (continuous)
- **Drift sensor** runs every change (§8); unresolved anchors escalate.
- **Scheduled consolidation passes** — agent-run dead-code removal, de-duplication,
  mechanical refactors (agents are strong here); humans own architectural moves
  (P9). Budgeted, because debt compounds faster under AI (P8).
- **Golden suite grows** — every past failure becomes a permanent regression case;
  per-cohort monitoring (aggregate pass rate can rise while a cohort silently
  regresses — Google Cloud eval guidance).
- **Artifacts are versioned and pruned** — constitution/specs/ADRs reviewed on a
  cadence; MEMORY-style index kept small (instruction-slot ceiling is real).

## 8. The core loop and bidirectional refinement

```
requirements ─► keel-spec ─► keel-plan (tier each task) ─► keel-build (one task)
                   ▲                                            │
      human confirms│                                  ┌─────────▼─────────┐
      spec update   │                                  │  keel verify      │  D2: deterministic,
                   │                                  │  (tiered DoD) +    │  agent-cannot-skip
                   │                                  │  drift sensor      │  (CI + optional hook)
                   │                                  └─────────┬─────────┘
                   │              green + drift clean            │
   feedback / expert hand-edit ◄──── keel-review ◄───────────────┘
   (test result, code change)      (behavior/tests/screenshots + drift report)
```

**Bidirectional refinement (the hard, original part).** Two entry points feed
back into the spec:
- *Human feedback after results* ("the test shows X, we actually want Y") →
  `keel-spec` amends the living spec; the change re-enters plan/build.
- *Expert hand-edits code* → the **drift sensor** detects the anchored spec no
  longer matches; `keel-spec` proposes a spec update reverse-derived from the
  change, labeled INFERRED, for human confirmation (D4/D5). Nothing auto-commits
  a spec change; the human decides which side is authoritative.

This closes the bidirectional requirement without unproven auto-regeneration:
the spec stays the source of truth, but the *code* is allowed to lead when an
expert wants it to, and the divergence is made "visible and painful rather than
silent."

## 9. Human checkpoints and review ergonomics

Minimize human effort without losing safety (P7), using the RADAR-style
risk-tiered funnel that is the strongest empirical result here.

- **Checkpoint placement by tier** (§6): T0 → sampled/auto-approve when the gate
  is green; T1 → behavior/spec review; T2 → mandatory sign-off. Start stricter
  and **prune checkpoints on data** — where humans approve ~99% and never modify,
  remove the gate (Meta RADAR relaxed its threshold to auto-land ~60% of diffs
  while keeping revert rate ~1/3 and incident rate ~1/50 of manual).
- **Review modality = what humans judge well:** rendered screenshots / visual
  regression (approve / reject / ignore, AI-grouped by cosmetic/functional/
  critical), acceptance-test scenarios in business language, and the drift
  report — not raw diffs.
- **Design against over-trust:** self-reported confidence is unreliable (METR
  gap). Instrument objective signals — revert rate, incident rate, regression
  escapes, per-cohort pass rates — and drive checkpoint pruning from those, not
  from feeling. Detailed AI explanations *increase* trust even when baseless, so
  the human's job is framed as *verify the behavior*, not *judge the reasoning*.

## 10. Layer 2 — the five skills

Each is a Claude Code skill; each begins by reading the constitution + relevant
specs/decisions and the repo map (P4/P5).

1. **`keel-onboard`** — detects greenfield vs brownfield (§7); produces/repairs
   the artifact tree, repo map, extracted specs (with provenance),
   characterization tests, and CI/hook wiring.
2. **`keel-spec`** — create and *refine* living specs. Bidirectional: ingests
   human feedback and hand-edits, proposes spec updates for confirmation, keeps
   spec↔code anchors current.
3. **`keel-plan`** — decompose into small, seam-bounded, independently reversible
   tasks; assign a blast-radius tier to each (which selects its gate set); update
   the feature list.
4. **`keel-build`** — implement one task (or wrap/validate human-written code);
   run the tiered `keel verify` loop locally until green; bounded self-repair
   (~a few iterations, then escalate — self-repair saturates fast).
5. **`keel-review`** — assemble the human checkpoint package: screenshots/visual
   diff, acceptance-test results, drift report, and the tier's required sign-offs.

For large parallelizable work (migrations, sweeps), `keel-plan` may group tasks
into non-conflicting packages run by parallel sub-agents (Devin-playbook /
Google-migration shape), each verifying independently before a coordinator
consolidates one reviewable change. Not used for tightly-coupled architectural
work (multi-agent parallelism is documented to fail there).

## 11. Layer 3 — the gate (`keel verify`)

A single deterministic command, composed of tier-selected checks (§6). The agent
runs it; CI runs it; optionally a hook enforces it (D2). It is the operational
form of P1–P2 and aggregates:

- **All tiers:** compile/type-check, lint, dependency allowlist + SBOM/checksum
  (anti-slopsquatting), SAST on the diff, erosion budgets (duplication /
  complexity / churn ceilings).
- **T1+:** property-based + contract tests, state invariants, mutation-score
  floor (kills "high-coverage asserts-nothing" tests), behavior-level acceptance
  tests derived from the spec.
- **T2+:** explicit safety assertions on the destructive operation, dry-run /
  rollback proof, characterization tests green on affected legacy.
- **Cross-cutting:** the **drift sensor** (anchors resolve; spec↔code consistent)
  and, as an *advisory* layer only, LLM-as-judge / multi-agent review — never the
  sole gate, because same-model reviewers share blind spots.

Gate definitions (which checks, which thresholds) live in the constitution and
are human-owned. An agent may *propose* new checks but cannot lower a threshold
or disable a check without a human-approved constitution change — the guardrail
the generating agent cannot edit for itself.

## 12. How this addresses the known failure modes

| Failure mode (evidence) | Mitigation |
|---|---|
| AI code fails security ~45%; iterating worsens it | P1 external gate; SAST every diff; security in `keel verify`, not self-review |
| Duplication ↑, refactoring ↓ under AI | P8 erosion budgets fail the build; scheduled consolidation passes |
| Slopsquatting (~20% hallucinated deps) | dependency allowlist + checksum/SBOM gate |
| Context rot / lost-in-the-middle | P4 retrieve-don't-preload; repo map; small windows |
| Self-conditioning / error compounding | P5 cold-start recovery; bounded self-repair then escalate; checkpoints |
| Spec/doc rot | D5 provenance + anchored fragments; drift sensor makes rot fail CI |
| Premature "done" | feature list starts "failing"; end-to-end evidence required in review |
| Over-trust / review fatigue | P7 review behavior not diffs; RADAR risk-tiering; objective metrics drive pruning |
| Agents weak at architecture | P9 humans own decomposition/seams; agents scoped to local/mechanical |
| Multi-agent coordination failure | parallelism only for loosely-coupled work; single context for cross-cutting |

## 13. Proposed rollout (prove before retiring the old skills)

1. **Phase 0 — gate first.** Build `keel verify` (tiered DoD + drift sensor +
   erosion budgets) and wire it into CI on one existing app. This alone delivers
   value and is the highest-leverage, best-evidenced piece.
2. **Phase 1 — loop on greenfield.** Ship `keel-onboard` (greenfield) +
   `keel-spec/plan/build/review`; build one new small app end-to-end.
3. **Phase 2 — brownfield adoption.** Add repo-map, spec extraction with
   provenance, and characterization-test generation; adopt one existing app.
4. **Phase 3 — long-lived hardening.** Consolidation passes, checkpoint pruning
   from metrics, golden-suite growth.
5. **Retire** research/plan/implement/validate once Phases 1–2 are proven on real
   apps; migrate their useful prompts into the `keel-*` skills.

## 14. Open decisions to settle during build

These are genuine forks I'll resolve with evidence as we build Phase 0, noting
them so they aren't lost:

- **O1** — spec fragment ↔ code anchoring mechanism (tree-sitter + git AST
  fingerprint vs. explicit spec-id comments vs. test-name convention). Affects
  drift-sensor precision. Lean: AST fingerprint + test mapping, no source
  pollution.
- **O2** — repo-map only vs. full typed code graph + semantic index. Scale-
  dependent; start with repo-map, add graph when a repo exceeds ~1k files.
- **O3** — how tier assignment is decided (static rules in constitution vs.
  agent-proposed + human-confirmed vs. learned risk score à la RADAR). Start
  rule-based; add a risk score once we have revert/incident telemetry.
- **O4** — where erosion-budget thresholds start and how they ratchet.

## 15. Summary

Keel is a spec-anchored, gate-centric framework. Humans own two things — the
**constitution/specs** and the **behavior/visual review** — and everything
between is enforced by a deterministic gate the agent cannot weaken, tiered by
blast radius. It supports human-written code and continuous bidirectional
refinement by *surfacing* spec↔code drift for a human to resolve rather than
auto-regenerating. It covers greenfield, brownfield, and long-lived projects
through one onboarding skill and a small durable-artifact layer designed for
cold-start recovery. Every principle traces to a measured failure mode in the
companion research briefing.
