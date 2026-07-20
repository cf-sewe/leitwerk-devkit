# Constitution — reference-app

## Purpose & scope
A demonstration service used to exercise the Leitwerk gate end to end. It has no
real users; its job is to show the framework governing real code — a spec bound
to a tested function, and a tier that escalates on a data migration.

## Invariants (never violate)
- The gate must pass before any change lands (this is the point of the example).
- No check may fake a pass: a check with nothing to run skips (`exit 2`).
- The order total is defined by `leitwerk/specs/orders.md`; the test is its
  oracle. A change that breaks it must turn the gate red.

## Blast-radius policy
- T0: `docs/**`, `**/*.md`.
- T2: `**/db/migrations/**`, `**/*.sql` (the `orders` migration) — irreversible
  data paths.
- T1: everything else (the Go application code).

## Definition of Done
`leitwerk verify` is green at the change's tier. At T1 that includes `go test`
(via the repo-local `tests` check); at T2 it also runs `sast`/`erosion`, which
skip cleanly when no analyzer is installed.

## Roles in play
scout for retrieval, test-engineer for the order-total oracle. No
security-reviewer until the app touches auth or real data.

## Decisions of record
- The `lint`/`types`/`tests` checks are repo-local overrides that resolve Go via
  `mise` (the pinned toolchain) or a PATH `go`, so the gate runs the real Go
  toolchain rather than the built-ins' bare `go`. They skip honestly if no Go is
  present.
- Kept minimal: one package, one migration — enough to demonstrate a spec bound
  to tested code and a T2 escalation, no more.
