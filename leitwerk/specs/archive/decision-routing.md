# Spec — decision routing: only relevant questions reach the human

Status: landed (2026-07-20) <!-- durable content: constitution Decision routing section + decision of record -->

## Problem
The framework routes decisions to the human by *file ownership* (the guard) and
by scattered "ask the human" prose in skills — not by decision *relevance*. The
measured symptom (2026-07-19 session, CONFIRMED): the human was asked three
times to approve, twice for mechanical follow-ups of already-approved actions,
and observed "ich sage eh nur 'du entscheidest'" — rubber-stamping. The
framework's own research base predicts this failure (automation bias; RADAR:
checkpoints humans approve ~99% unmodified should be pruned) and the agent has
specialist roles (`architect`, `security-reviewer`, `test-engineer`, `scout`)
that can carry domain judgment without the human. There is no stated rule for
*when a question may reach the human at all*.

## Behaviour (the observable contract)
- **The escalation test** is always-on context (repo `CLAUDE.md`; adopters get
  it via `core/templates/CLAUDE.template.md`). A decision is escalated to the
  human only if it:
  1. **sets or changes intent** — scope, priorities, spec approval;
  2. **weakens or waives a guarantee** — thresholds, checks, tier downgrades;
  3. **accepts irreversible or residual risk** — T2 sign-off, data/money/auth,
     shipping with an open finding, accepting a known limitation.
  Everything else the agent decides, records (spec's Design decisions, with
  rationale), and keeps reversible (git + gate).
- **Never ask what evidence can answer**: repo reads, tests, and scout runs
  come before any question; domain judgment goes to a specialist role first.
- **Escalations are decision-ready**: options, evidence, a recommendation, and
  a default — as `leitwerk/proposals/` files when the decision must outlive
  the session, in-session otherwise.
- **Pruning loop**: `leitwerk-review` reports which decisions were escalated
  and whether the human's answer diverged from the recommendation; a question
  class that never diverges is retired to agent-decided at the next dream
  sweep (measurement/mechanization deferred, fits M4.1).
- Skills align: `leitwerk-spec` (design step escalates only test-positive
  dimensions; research never asks what the repo answers), `leitwerk-review`
  (divergence reporting). `leitwerk-onboard` already carries the rule ("ask
  only where intent is genuinely unknowable") — checked, unchanged.
  `leitwerk-plan`/`leitwerk-build` contain no human-asks — checked, unchanged.
- `bindings/open/AGENTS.md` mirrors the test (working-method item).
- Whitepaper §10 gains the rule ("Escalate decisions, not questions") and the
  proposals-inbox mechanism; README states the three escalation classes where
  it describes the human's role.
- The constitution wording (authority version of the test + the pruning rule)
  is **proposed**, not edited: `leitwerk/proposals/`.
- **Follow-up is mechanical.** The `lifecycle` check reports open proposals on
  every gate run: the summary counts them, and a warning names files older
  than 30 days (age from the filename's timestamp prefix). An open decision
  cannot rot invisibly. It never turns the gate red — the waiting party is
  the human, and a red gate would punish the agent's unrelated work and
  reward deleting proposals.
- **Native wizard (Claude binding).** A `SessionStart` hook surfaces open
  proposals into session context; `leitwerk-review` presents each pending
  decision as a native multiple-choice question — the problem in a sentence,
  options with one-line descriptions, the recommendation marked. An accepted
  answer is the human's in-session authorization: the agent applies the
  change the file documents (staged copy for human-owned files) and deletes
  the file. Open-code has no wizard: the gate warning is the visibility, and
  AGENTS.md says to work the inbox at review.

What must NOT happen:
- No widening of agent authority over policy: the guard on human-owned files
  is untouched; weakening moves stay human-only. This change narrows
  *questions*, not *ownership*.
- The always-on addition must not breach the context budgets (the `context`
  check proves it).

## Design decisions
- **A three-condition test, not an enumerated decision list.** The test must
  live always-on, so it competes for context budget; enumerations belong in
  the constitution (proposed). Rejected: full taxonomy in CLAUDE.md.
- **Always-on placement (CLAUDE.md), authority in the constitution.** The
  failure mode is asking *before* any skill loads, so skills-only placement
  arrives too late. Rejected: skills-only.
- **Specialist roles substitute for the human on domain judgment.** The roles
  exist precisely so expensive judgment stays in-session. Rejected: routing
  security/architecture questions to the human that a role can answer.
- **Pruning stays procedural for now.** Divergence reporting in review +
  retirement at the dream sweep; objective measurement belongs to M4.1.
  Rejected: building tracking tooling now.
- **Open proposals warn, never red.** A red gate over an open decision would
  block work the agent can still do, punish the wrong party, and create an
  incentive to delete proposals. Rejected: hard red after N days. (Added from
  human feedback: follow-up must be automatic, not hoped-for.)

## Invariants touched
- *The gate config is human-owned* — untouched; escalation classes 2 and 3
  restate it.
- *Open-code guarantee-parity* — AGENTS.md mirrors the process rule; the
  guarantee (gate) is unaffected.
- *Drift is surfaced, not resolved* — irreconcilable spec↔code conflict
  remains an escalation (class 1: it is an intent question).

## Blast radius
T1 (binding skill markdown is the highest touched tier; templates, README,
whitepaper are T0). Worst case if wrong: the agent under-asks and a wrong
self-made decision ships behind a green gate — reversible and recorded; the
human over-asked is the status quo this change removes.

## Acceptance checks
- `leitwerk verify --tier T2` green; `context` proves the always-on additions
  stay within budget.
- Reviewable contract: each file listed under Behaviour contains the described
  rule; the escalation behaviour itself has no mechanical oracle (honest gap —
  it is judged by the human noticing fewer irrelevant questions).

## Out of scope
- Constitution/roadmap edits (proposed via `leitwerk/proposals/`).
- Objective escalation metrics (M4.1) and proposal-aging mechanization (the
  `lifecycle` check could warn on stale proposals — noted in the
  boundary-granularity proposal).
