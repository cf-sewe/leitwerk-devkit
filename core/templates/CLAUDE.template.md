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

## Procedures (loaded on demand)
Use the Leitwerk skills (`leitwerk-onboard/spec/plan/build/review`). Do not
inline multi-step procedures into this file.

## Project specifics
<!-- Add build commands, layout, and conventions for THIS repo here. -->
