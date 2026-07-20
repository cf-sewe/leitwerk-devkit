# Plan — reference-app-real

Status: landed (2026-07-20) <!-- all steps landed -->

Spec: `leitwerk/specs/archive/reference-app-real.md`. Verified before planning:
`examples/reference-app` has 2 files; `core/checks/{tests,types}.sh` call bare
`go`; bare `go` is a broken asdf shim here while `mise exec -- go` works (go
1.26.5) — confirmed from the reference-app subdir; default tiers map
`**/*.sql`/`**/db/migrations/**` to T2; `selftest` §4 runs the example at T0;
scenarios auto-discover via `run-all.sh`.

Every step leaves the devkit `leitwerk verify --tier T2` green. The devkit gate
does not run the reference-app's Go tests until step 3 wires §4 to T1, so
steps 1–2 keep it green while the app takes shape.

## Steps
Step status: `[ ]` open · `[x]` done · `[~]` deviated — one line why.

1. `[x]` **App + spec + T2 path + local checks.** Add `examples/reference-app/`:
   `go.mod`, `orders.go` (`OrderTotalCents`), `orders_test.go` (passing tests,
   written first), `leitwerk/specs/orders.md` (anchored to `orders.go`),
   `db/migrations/001_create_orders.sql` (T2), and repo-local
   `leitwerk/checks/{tests.sh,types.sh}` (mise→PATH→skip). Verify **in place**:
   `leitwerk verify --tier T1` green with `go test green`; `--tier T2` green;
   `leitwerk tier db/migrations/001_create_orders.sql` = T2. Tier T2 (adds a
   `*.sql`). Manual check (T2): the tests check genuinely runs (`go test`), not
   a skip.

2. `[x]` **Docs.** Update `examples/reference-app/README.md` and
   `leitwerk/constitution.md` (the example now has a T2 migration and a real
   test; the gate no longer only skips). Tier T0.

3. `[x]` **Wire the devkit oracle.** `selftest` §4: run the reference-app in
   place at **T1** and assert the output shows real tests ran (`go test green`),
   not a skip. Add `examples/scenarios/s6-reference-app.sh`: copy the app to a
   temp dir (with `mise.toml` alongside so Go resolves), assert `verify --tier
   T1` green + tests ran, then break it and assert exit 1 / `gate: FAIL`; update
   the scenarios `README.md` list. Tier T2. Manual check (T2): `s6` passes
   standalone and via `run-all.sh`; devkit gate T2 green.

4. `[x]` **Review + land.** Devkit `leitwerk verify --tier T2` green; review at
   T2 weight; commit gate-green.

## Review findings addressed (T2 review, 2026-07-20)
A focused adversarial reviewer exercised the gate end-to-end (verify T1/T2,
tier, s6, run-all, selftest, drift) and returned SHIP with two minor findings,
both bending "a check never fakes a pass"; both fixed and re-verified:
- **s6 SKIP hole:** SKIP was decided on the missing `go test green` marker
  before consulting the exit code, so a broken baseline (Go present, gate red)
  read as SKIP. Now SKIP only when the green run exited 0; a non-zero exit with
  no marker FAILs.
- **reference-app `lint.sh`:** `gofmt -l` exits 2 on an unparseable file; under
  `set -e` that propagated as a skip → green at T0. Now any gofmt error maps to
  exit 1; exit 2 is reserved for the no-toolchain branch. Verified: a
  parse-error `.go` reds `lint` at T0.
Confirmed sound by the reviewer: selftest §4 is fail-closed and genuinely
proves real tests ran; the sed break self-detects source drift and never
mutates the real app; the nested module is isolated; anchors resolve; the T2
escalation on the migration holds.

No constitution decision of record: M1.2 changes no framework guarantee or
invariant (it builds a demonstration on existing features), so the durable
content stays in this archived spec.

## Verification strategy
Oracle-first per step: `orders_test.go` is written before `orders.go`; the
devkit-level oracles are `selftest` §4 (real tests ran, in place) and
`examples/scenarios/s6-reference-app.sh` (green→broken→red). No behaviour lands
without a failing→passing test.

## Risks & rollback
- **Toolchain resolution** — bare `go` broken locally; mitigated by the
  mise→PATH→skip resolver in the repo-local checks (works in-repo via mise, in
  CI via setup-go). If neither resolves, the checks skip honestly (the scenario
  detects a skip and does not assert red on a skip).
- **Temp-copy Go resolution** — a copy outside the repo loses `mise.toml`;
  the scenario copies `mise.toml` alongside so mise resolves Go there.
- **Rollback (T2):** `git checkout -- leitwerk/checks/selftest.sh` and remove
  `examples/reference-app/{go.mod,orders*.go,db,leitwerk/checks,leitwerk/specs}`
  and `examples/scenarios/s6-reference-app.sh`; the example returns to its
  skip-only state, which is green. No external state touched.

## Roles to wake
- `test-engineer` — the reference-app test and the two devkit oracles (real
  tests ran; broken→red).
- `architect` — the toolchain-resolution boundary (repo-local override vs
  generic built-in) and the T2 escalation demonstration.
- Review at T2 weight before landing (no separate human sign-off reserved for
  M1.2; escalate only if a genuine guarantee/risk decision surfaces).
