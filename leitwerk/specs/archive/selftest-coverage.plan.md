# Plan — selftest-coverage

Status: landed (2026-07-20) <!-- landed with its spec at the T2 review -->

Small, test-only hardening (three residual gaps) + the M1.3 roadmap close. All
Go-test changes land under `selftest` §0. One strand, one gate-green commit.

## Steps
Step status: `[ ]` open · `[x]` done · `[~]` deviated — one line why.

1. `[x]` **Checks-line contract** — add a test to
   `core/cmd/leitwerk/integration_test.go` (T2) that runs the built binary
   `verify --tier T0` and `--tier T2` in an empty temp dir (shipped defaults via
   the binary's sibling `leitwerk.tiers`), asserting the `checks:` line lists
   exactly `lint` (T0) and `lint types tests drift sast erosion` (T2) and the
   gate is PASS (all checks skip). Mutation-check: temporarily reorder/trim the
   shipped T2 line → test red.
2. `[~]` **Shipped `[paths]` ordering** — extend `TestShippedDefaults`
   (`core/internal/gate/tiers_test.go`, T2) with first-match-sensitive paths:
   `docs/schema.sql`→T0, `k8s/deploy.yaml`→T2, **`modules/net.tf`→T2** (deviated:
   the planned `infra/x.tf` matches `infra/**` first and never reaches `**/*.tf`,
   so it would not have closed the `*.tf` gap). Mutation-check: remove
   `docs/** = T0` from the shipped file → test red on `docs/schema.sql`.
3. `[x]` **T1 fallback** — add a test to `integration_test.go` (T2): a temp dir
   with a `leitwerk/tiers.conf` fixture whose `[paths]` has no catch-all; `tier`
   on an unmatched path prints `T1`, exercising `main.go:64`. Mutation-check:
   change the fallback literal in `main.go` → test red.
4. `[x]` **Review & land** — T2. `leitwerk verify --tier T2` green; `leitwerk
   drift`; spec fidelity; T2 review (documented read-only role fallback —
   test-engineer primary). Landing ritual: T2 sign-off (human), spec→landed,
   archive spec+plan, close M1.3 in `roadmap.md` (staged-copy), commit
   gate-green.

## Verification strategy
Each new assertion is mutation-checked: it must go red when the behaviour it
pins is broken, and green on correct code (so it is a real oracle, not decoration).
Runs under `go test` → `selftest` §0 → the devkit gate at T2.

## Risks & rollback
- **False red / brittle output match:** the `checks:` assertion pins output
  format. Mitigation: match the substring `checks: <list>` (the documented
  contract line), not the whole verify banner. Rollback: revert the test file.
- **T2 rollback (whole strand):** `git revert` the single commit — tests are
  additive, no production code touched; roadmap close reverts with it.

## Roles to wake
- `test-engineer` — primary: these are oracles; the mutation-check is the pin.
- `architect` — light: confirm the Go-test layer choice (vs shell §1) holds and
  no coverage is duplicated.

## Review outcome (2026-07-20)
T2 gate green; panel via the documented read-only fallback.
- *architect:* sound — close M1.3 (acceptance already met; Go-test layer honors
  the black-box contract; no duplication). 3 LOW fixed: `infra/x.tf`→`modules/net.tf`
  (spec + step 2 `[~]`), init-output banner assertion added, layer split noted.
- *test-engineer:* 1 HIGH + 2 MED + 1 LOW, all fixed and mutation-re-verified —
  HIGH: shipped T1 checks line now pinned (T0/T1/T2); MED-1: fallback test made
  discriminating (`x.tf` probe); MED-2: `runBin` env scrubbed (hermetic); LOW:
  dropped the tool-conditional exit-0 assertion.
Human: T2 sign-off granted; roadmap close approved (staged-copy). No escalated
decision diverged from the recommendation.
