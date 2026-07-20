# Spec — reference-app-real: real governance on a minimal real app

Status: landed (2026-07-20) <!-- durable content stays in this archived spec; no framework guarantee changed -->

## Problem
`examples/reference-app` has only a README and a filled-in constitution
(CONFIRMED — `find examples/reference-app -type f` = 2 files). The gate runs
there but governs nothing: at T0 the only check is `lint`, which skips. So the
example shows *execution*, not *governance* — there is no spec bound to code, no
test the gate runs, and no T2 path, so tier escalation is never demonstrated on
real code. The whitepaper points here as the worked example.

Constraints found by reading:
- The `tests`/`types` built-ins call **bare** `go` (`core/checks/tests.sh`,
  `core/checks/types.sh`, CONFIRMED). In this repo's environment bare `go` is a
  broken asdf shim (`go version` fails); the working toolchain is `mise exec --
  go` (CONFIRMED). CI has no mise but gets Go from `setup-go`, so bare `go`
  works there. A reference-app that runs Go tests must resolve Go robustly.
- The default tiers map `**/db/migrations/**` and `**/*.sql` to **T2**
  (`core/leitwerk.tiers`, CONFIRMED), and the reference-app has no `tiers.conf`,
  so it uses those defaults.
- `selftest` §4 already runs `leitwerk verify --tier T0` in the example
  (`leitwerk/checks/selftest.sh`, CONFIRMED); scenarios in `examples/scenarios/`
  are auto-discovered by `run-all.sh` and run in `selftest` §5.

## Behaviour (the observable contract)
- The reference-app is a minimal **Go** module (`examples/reference-app/go.mod`)
  with one package: an order-total function (`orders.go`) and a real test
  (`orders_test.go`) that passes on correct code and fails on a regression.
- A **spec** (`examples/reference-app/leitwerk/specs/orders.md`) states the
  app's contract and anchors it to the code (`orders.go#OrderTotalCents`), so
  `drift` governs it too.
- A **T2 path**: `examples/reference-app/db/migrations/001_create_orders.sql`.
  `leitwerk tier db/migrations/001_create_orders.sql` → **T2** there, so editing
  the migration escalates the reference-app's gate to the full check set.
- The reference-app ships **repo-local checks**
  (`examples/reference-app/leitwerk/checks/tests.sh`, `types.sh`) that resolve
  Go via `mise` then a working PATH `go`, else skip (2) — never a faked pass.
  This overrides the bare-`go` built-ins and doubles as a demonstration of the
  per-check repo-local override.
- **`leitwerk verify` runs actual tests, not skips**: at T1 in the reference-app
  the `tests` check runs `go test ./...` and prints `go test green`; the gate is
  green on correct code.
- **A deliberately broken change turns it red**: a regression to the app (or a
  build break) makes `go test` fail, so `leitwerk verify` exits 1 (`gate: FAIL`).
- Both are pinned as oracles in the devkit's own gate: `selftest` §4 runs the
  reference-app in place at T1 and asserts the tests actually ran; a scenario
  (`examples/scenarios/s6-reference-app.sh`) asserts green→broken→red on a
  throwaway copy.

What must NOT happen:
- No change to the shipped generic `core/checks/*` (the bare-`go` built-ins stay
  as generic templates; the reference-app overrides locally, per the
  "consuming repos do not edit installed core" invariant).
- No heavy dependencies: stdlib Go only (`testing`), no modules to download, so
  the test runs offline in CI.
- The reference-app's constitution/tiers stay honest: `sast`/`erosion` still
  skip cleanly at T2 (no analyzer present), never fake a pass.

## Design decisions
- **Go, not Node/other.** Go is the repo's pinned toolchain (mise), CI already
  sets it up, and stdlib `testing` needs no dependencies — a real test that runs
  offline. Rejected: Node (`node`/`npm` are also broken asdf shims here and a
  test runner adds deps) and a shell "app" (not a real app).
- **Repo-local mise-aware check overrides, not editing the built-ins.** Bare
  `go` is broken in this environment; the generic core checks must stay generic
  (invariant: consuming repos override per check, never edit installed core). A
  repo-local `tests.sh`/`types.sh` resolves Go via mise→PATH→skip and makes the
  example a faithful demonstration of the override mechanism. Rejected: making
  the shipped `core/checks/tests.sh` mise-specific (leaks one tool manager into
  a generic template).
- **Migration as the T2 path.** A SQL migration is the canonical irreversible
  change the tier policy already targets (`**/*.sql = T2`), so escalation is
  shown without inventing new policy. Rejected: infra/*.tf (heavier, no real use
  in a tiny app).
- **Oracle split: in-place green + throwaway red.** The green run is asserted in
  place (`selftest` §4) because Go resolves in the repo tree (mise finds the
  repo `mise.toml`); the broken→red case runs on a temp copy in a scenario, with
  `mise.toml` copied alongside so Go still resolves. Rejected: breaking the real
  app in place (dirties the tree and risks leaving it broken).

## Invariants touched
- *A check never fakes a pass* — the repo-local checks skip (2) only when no Go
  toolchain resolves; with Go present they run and can fail. `sast`/`erosion`
  keep skipping honestly.
- *Consuming repos do not edit installed core* — the reference-app overrides
  per check in its own `leitwerk/checks/`; the shipped `core/checks/` are
  untouched.
- *Drift is surfaced, not resolved* — the app's spec anchors its code so `drift`
  governs it, consistent with M1.1.
- *The gate is the guarantee* — the demonstration is wired into `selftest` and a
  scenario, so a regression in the example is caught by the devkit's own gate.

## Blast radius
T2 for the devkit change: it adds `**/*.sql` (T2) and edits `selftest.sh` /
adds a scenario (`*.sh`, T2). Worst case if wrong: a flaky reference-app gate
reds the devkit gate (fail-closed, caught here); the example itself has no real
users. The reference-app's own changes escalate T0→T2 by path, as intended.

## Acceptance checks
- In `examples/reference-app`: `leitwerk verify --tier T1` is green and its
  output shows `go test green` (tests ran, not skipped); `leitwerk verify
  --tier T2` is green (migration path's fuller set; sast/erosion skip cleanly).
- `leitwerk tier db/migrations/001_create_orders.sql` prints `T2` in the
  reference-app.
- A regression (break the test's expectation, or the build) makes
  `leitwerk verify` exit 1 with `gate: FAIL`.
- Devkit gate: `selftest` §4 asserts the reference-app ran real tests at T1;
  `examples/scenarios/s6-reference-app.sh` asserts green→broken→red; `leitwerk
  verify --tier T2` on the devkit stays green.

## Anchors
- `examples/reference-app/orders.go#OrderTotalCents`
- `examples/reference-app/leitwerk/checks/tests.sh`
- `leitwerk/checks/selftest.sh#reference-app`

## Out of scope
- The bugfix workflow (M1.5) that will reproduce→pin→fix a seeded bug on this
  substrate — this spec only provides the substrate.
- Multiple packages, a web layer, or a real database — kept minimal.
- Wiring language checks for non-Go toolchains in the example.
