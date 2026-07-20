# leitwerk-devkit — working notes for Claude Code

This repo builds Leitwerk and is governed by it. Keep this file small; it loads
into every session. Authority and procedures live elsewhere, linked below.

## The one rule
Every change passes the gate at its blast-radius tier before it lands:

```
leitwerk verify --tier <T0|T1|T2>       # leitwerk tier <path> tells you the tier
```

A `Stop` hook runs this automatically and blocks turn-end on a red gate. Do not
argue with a red result — fix the cause.

## Authority (human-owned — propose, do not edit)
- `leitwerk/constitution.md` — invariants, blast-radius policy, Definition of Done.
- `leitwerk/tiers.conf` — tier→checks and path→tier policy, plus `[human-owned]`.
- `leitwerk/roadmap.md` — ordered backlog of future specs.

A `PreToolUse` hook blocks edits to these; `leitwerk guard <path>` is the check.

## When to ask the human
Escalate a decision only if it (1) sets or changes intent — scope, priorities,
spec approval; (2) weakens or waives a guarantee — thresholds, checks, tiers;
or (3) accepts irreversible or residual risk — T2 sign-off, shipping with an
open finding. Everything else: decide, record it in the spec's Design
decisions, keep it reversible. Never ask what the repo can answer; wake a
specialist role for domain judgment. Escalations are decision-ready — options,
evidence, recommendation — as `leitwerk/proposals/` files when they must
outlive the session.

## Procedures (loaded on demand — do not inline them here)
Use the skills: `leitwerk-onboard`, `leitwerk-spec`, `leitwerk-plan`,
`leitwerk-build`, `leitwerk-review`. Specialist roles live in
`bindings/claude/agents/`; the review workflow is `.claude/workflows/leitwerk-review.mjs`.

## Layout
- `core/` — the tool-agnostic gate (`core/bin/leitwerk`, `core/checks/*.sh`). T2.
  The binary is gitignored — on a fresh clone run `make -C core build` once.
- `bindings/claude/` — the Claude Code plugin. `bindings/open/` — AGENTS.md + CI.
- `leitwerk/` — this repo's own governance, specs, and repo-local checks.
