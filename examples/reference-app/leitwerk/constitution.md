# Constitution — reference-app

## Purpose & scope
A demonstration service used to exercise the Leitwerk gate. It has no real
users; its only job is to show the framework running end to end.

## Invariants (never violate)
- The gate must pass before any change lands (this is the point of the example).
- No check may fake a pass: a check with nothing to run skips (`exit 2`).

## Blast-radius policy
- T0: `docs/**`, `**/*.md`.
- T2: none yet (no migrations/infra in this example).
- T1: everything else.

## Definition of Done
`leitwerk verify` is green at the change's tier.

## Roles in play
scout for retrieval, test-engineer once real code exists. No security-reviewer
until the app touches auth/data.

## Decisions of record
- Kept intentionally minimal so the gate's skip behaviour is visible.
