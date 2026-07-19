# Spec — Layer-2 orchestration via native dynamic workflows

## Problem
Leitwerk's Layer 2 (roles / orchestration / review) was designed as a
hand-rolled `orchestrator` subagent plus a prose "trigger table" that wakes
roles. Claude Code now ships **dynamic workflows**: deterministic orchestration
scripts that fan out parallel subagents with built-in adversarial verification.
That is a better, native realization of exactly this layer. Keeping the
hand-rolled orchestrator would reinvent — worse and non-deterministically —
what the platform now does natively.

## Behaviour (the observable contract)
- On Claude Code, Layer-2 orchestration is realized by a **dynamic workflow**,
  not a bespoke orchestrator role. The workflow script is the orchestrator; its
  control flow is deterministic (code, not model whim).
- The role subagents (`architect`, `test-engineer`, `security-reviewer`,
  `scout`) are **kept** and reused as workflow `agentType`s.
- Review scales by blast-radius tier: T0/T1 stay lightweight (direct/inline);
  **T2 uses a workflow** (dimension fan-out + adversarial refutation).
- The workflow's multi-agent verification is **soft** (agents judging agents).
  It never replaces the hard gate: `leitwerk verify` (Layer 3) remains the
  external, authoritative oracle, enforced by the Stop hook / CI regardless of
  what the workflow concluded.
- Because **plugins cannot package workflows**, the workflow is shipped as a
  core template (`core/templates/workflows/leitwerk-review.mjs`) that
  `leitwerk init` scaffolds into a repo's `.claude/workflows/`. The
  `leitwerk-review` / `leitwerk-build` skills prefer the saved `/leitwerk-review`
  and **fall back to spawning the roles directly** when it is absent or workflows
  are disabled — so review degrades gracefully and never depends on the workflow
  being present. The template and this repo's own copy are kept identical by
  `selftest`, so an adopter gets the same orchestration this repo dogfoods.
- Open code (Codex/AGENTS.md) has **no** equivalent engine → it falls back to
  sequential role agents plus the CI gate. This asymmetry is stated, not hidden.

## Invariants touched
Constitution: "bindings never reimplement the gate" (unchanged — the workflow
sits above the gate) and the human-owned gate. Adds a new decision of record.

## Blast radius
T1 (bindings + docs). No change to `core/` or the gate logic.

## Acceptance checks
- The `orchestrator` subagent is removed and no artifact references it.
- `leitwerk-review` / `leitwerk-build` skills describe the workflow path and the
  soft-vs-hard-verification distinction.
- `.claude/workflows/leitwerk-review.mjs` exists and is syntactically valid as a
  workflow script (async-wrapped `node --check`; top-level `return`/`await` are
  legal in the workflow runtime), ending in a stage that runs `leitwerk verify`.
- Whitepaper, design proposal, and constitution record the realization.
- `leitwerk verify --tier T1` stays green.

## Full change list (what must be done)
1. **Constitution** — add decision of record (this change). [T-]
2. **Retire** `bindings/claude/agents/orchestrator.md`. [T1]
3. **`leitwerk-review` skill** — workflow-based review for T2; soft/hard split. [T0]
4. **`leitwerk-build` skill** — drop orchestrator delegation; workflow for heavy
   steps. [T0]
5. **Add** `.claude/workflows/leitwerk-review.mjs` (real example). [T1]
6. **`bindings/claude/README`** — agents list minus orchestrator; add a
   "Workflows" section (packaging constraint, opt-in, version min). [T0]
7. **`bindings/open/AGENTS.md` + README** — the no-workflow asymmetry. [T0]
8. **Whitepaper** — §7 realization note; §5 figcaption clause. [T0]
9. **Design proposal** — dated addendum. [T0]
10. **Roadmap** — mark integration; fold into M2.2 (live validation). [T0]
11. **Memory** — record decision + packaging constraint.

## Out of scope
A `leitwerk-build.mjs` workflow (review first); publishing; a JS syntax check in
the gate (roadmap item).
