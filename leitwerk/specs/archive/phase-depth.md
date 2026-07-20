# Spec — phase depth: research, design, and progress in the phase skills

Status: landed (2026-07-20) <!-- durable content: constitution decision of record (pre-build phase depth) -->

## Problem
The framework's own papers prescribe pre-build structure that the shipped phase
skills do not implement:

- The research briefing's reference architecture includes an explicit explore
  step ("explore → plan → decompose → implement") and retrieve-don't-preload
  (P4); `leitwerk-spec` starts at "write the problem" and no phase skill ever
  invokes the `scout` role.
- MAST attributes ~42% of multi-agent failures to specification/decision issues;
  the design proposal lists a decision log in Layer 1 and marks the architect
  "human-collaborative" — but pre-change design decisions (chosen approach,
  rejected alternatives) have no place in any artifact. Decisions of record
  cover only landed changes.
- P5 makes cold-start recovery the design default (durable plan + progress),
  yet plan execution state lives only in chat scrollback: nothing records which
  plan steps are done or where the build deviated.
- Whitepaper §8.2 requires provenance labels (CONFIRMED/INFERRED/GAP) when
  reverse-deriving specs from brownfield code; `leitwerk-onboard` does not
  mention them (only `leitwerk-review` tags this way).
- `leitwerk-plan` never requires verifying assumed file/symbol references
  against the actual code, and plans carry no manual-verification criteria for
  T2 steps.

## Behaviour (the observable contract)
- `leitwerk-spec` instructs: read files the request mentions fully before
  anything else; fan out `scout` subagents for locating and analyzing relevant
  code; carry facts into the spec as `file:line` references tagged
  CONFIRMED/INFERRED. It then decides implementation-level design choices
  itself and records them; dimensions that set intent, weaken a guarantee, or
  accept irreversible risk it walks with the human one at a time (options with
  trade-offs; the human decides). Outcomes land in the spec's Design decisions
  section; no separate research or design document is produced. (Wording
  aligned with the later decision-routing change.)
- `core/templates/spec.template.md` gains a **Design decisions** section:
  chosen approach, alternatives considered, and why they were rejected.
- `leitwerk-plan` instructs: no file/symbol reference enters the plan without
  being verified against the code; T2 steps carry manual verification criteria
  (what a human must eyeball) alongside the automated checks.
- `core/templates/plan.template.md` documents a per-step status convention
  (`[ ]` open · `[x] done` · `[~] deviated — <one line>`) and a manual-check
  line for T2 steps.
- `leitwerk-build` updates the step's status line in the plan when a step
  lands, recording deviations in one line, so a cold session can resume from
  the plan file alone.
- `leitwerk-review` reports deviations between plan and implementation in its
  summary for the human reviewer.
- `leitwerk-onboard` labels reverse-derived facts CONFIRMED/INFERRED/GAP;
  INFERRED and GAP require human validation before they gain authority
  (aligns the skill with whitepaper §8.2).
- `bindings/open/AGENTS.md` mirrors the working-method additions so the
  process description stays equivalent across bindings.
- `core/templates/spec.template.md` states the lifecycle normatively: two spec
  kinds (living contract vs change record) and the states with their triggers
  and owners — human approval sets `active`; the landing review sets `landed`,
  merges the durable core out, and archives spec + plan ("dreaming"). This
  replaces the ambiguous "archive when superseded" wording; `leitwerk-spec`
  and `leitwerk-review` carry the transitions they own.

What must NOT happen:
- No parallel document tree (no research/, designs/, PROGRESS_* files). All
  durable output lands in the existing artifacts: the spec, the plan, and — for
  durable cross-change decisions, after landing — the constitution's decisions
  of record (human-owned, proposed not edited).
- No new enforcement claims from prose: these are procedure improvements; the
  gate remains the only guarantee.
- Design exploration is not fanned out to parallel agents (the briefing
  documents multi-agent failure on tightly-coupled architectural work); it is
  one context plus the human.

## Design decisions
- **Design phase as a step inside `leitwerk-spec`, not a sixth skill.** The
  framework's positioning is "lightweight phases around a hard gate"; the
  design outcome lands in the spec anyway, and a separate skill invites a
  separate artifact. Rejected: a standalone `leitwerk-design` skill.
- **Findings land in existing artifacts, not a document tree.** A
  docs/thoughts-style tree grows monotonically, which conflicts with the spec
  lifecycle (roadmap M1.4) and the context-rot evidence. Rejected: persisted
  research/design documents with their own lifecycle.
- **Progress lives in the plan file itself.** The plan is already per-change,
  perishable, and archived at landing; a status line per step adds recovery
  without new files. Rejected: separate PROGRESS_PHASE files and a
  `progress.json` (the latter stays a design-proposal option, out of scope).

## Invariants touched
- *Procedures live in skills, loaded on demand* — additions go into skill
  bodies and templates only; always-on context (CLAUDE.md, rules) is untouched,
  so the `context` budgets are unaffected.
- *The gate is the guarantee; prose is advisory* — unchanged; this spec only
  improves the advisory layer.
- *Humans own judgment* — the design step operationalizes the constitution's
  human-owned intent: on escalated dimensions the agent presents options and
  the human decides (which dimensions escalate is governed by the
  decision-routing spec).
- *Drift is surfaced, not resolved* — unchanged.
- *Open-code guarantee-parity* — untouched (no gate logic moves); AGENTS.md is
  updated for process-description equivalence only.

## Blast radius
T1 — Markdown-only changes; no gate logic, no Go code. Core templates and
repo docs are T0, but the binding markdown (skills, AGENTS.md) is T1 since the
human raised `bindings/**/*.md` to T1 in `tiers.conf` (2026-07-19, resolving a
constitution↔tiers.conf tension this change surfaced), and the highest touched
tier governs. Worst case if wrong: agents follow a worse procedure; no
mechanical guarantee weakens. (Templates are embedded in the binary via
go:embed, so `selftest` rebuilds and re-embeds them.)

## Acceptance checks
- `leitwerk verify --tier T2` stays green (json/shell/drift/selftest/parity/
  context) after every step.
- Reviewable contract: each file listed under Behaviour contains the described
  addition; the skill additions stay tight (guideline: ≤ 15 lines per skill) so
  on-demand load cost stays low.
- The prose itself is reviewed by the panel/human — a text's quality has no
  mechanical oracle; this spec does not pretend otherwise.

## Out of scope
- A bugfix workflow skill (whitepaper §8.3: reproduce → pin → fix) — proposed
  as a roadmap entry (human-owned; relates to M1.2).
- A repo map / code graph (design-proposal open decision O2) — proposed as a
  roadmap entry.
- `progress.json` / feature lists, diff-signal role triggers, provenance
  *tooling* (M3.3) — unchanged design targets.
