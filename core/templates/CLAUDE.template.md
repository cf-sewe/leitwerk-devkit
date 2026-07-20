# <project> — working notes for Claude Code

Keep this file small (the guidance is: under ~200 lines); it loads into every
session. Always-on context goes here; procedures go in skills; absolute
prohibitions go in hooks.

## The one rule
Every change passes the gate at its blast-radius tier before it lands:

```
leitwerk verify --tier <T0|T1|T2>       # leitwerk tier <path> tells you the tier
```

The `Stop` hook runs this and blocks turn-end on a red gate.

## Authority (human-owned — propose, do not edit)
- `leitwerk/constitution.md` — invariants and Definition of Done.
- `leitwerk/tiers.conf` — tier and path policy, plus `[human-owned]`.

A `PreToolUse` hook blocks edits to human-owned files; `leitwerk guard <path>`
is the check.

## When to ask the human
Escalate a decision only if it (1) sets or changes intent — scope, priorities,
spec approval; (2) weakens or waives a guarantee — thresholds, checks, tiers;
or (3) accepts irreversible or residual risk — T2 sign-off, shipping with an
open finding. Everything else: decide, record it in the spec's Design
decisions, keep it reversible. Never ask what the repo can answer; wake a
specialist role for domain judgment. Escalations are decision-ready — options,
evidence, recommendation — as `leitwerk/proposals/` files when they must
outlive the session.

## Procedures (loaded on demand)
Use the Leitwerk skills (`leitwerk-onboard/spec/plan/build/review`). Do not
inline multi-step procedures into this file.

## Project specifics
<!-- Add build commands, layout, and conventions for THIS repo here. -->
