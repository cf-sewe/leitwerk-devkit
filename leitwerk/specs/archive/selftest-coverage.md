# Spec — selftest-coverage: close the real gaps in the CLI golden suite

Status: landed (2026-07-20) <!-- change record; landed at the T2 review 2026-07-20 -->

## Problem
M1.3 (`leitwerk/roadmap.md`) asks to extend the CLI golden suite to cover glob
edge cases, `leitwerk init`, `checks_for_tier`, and non-zero exit paths, with
acceptance "mutating the glob translation or the tier table fails `selftest`".

That acceptance **already holds** (CONFIRMED by execution 2026-07-20):
`leitwerk/checks/selftest.sh:24` runs `go test -C core ./...`, and the Go suite
already pins these. Verified by mutation:
- Breaking the `**/` translation in `core/internal/gate/glob.go:37`
  (`(.*/)?`→`(.*/)`) fails `TestMatchTier`/`TestGlobToRegex`
  (`glob_test.go:5,29`) → `go test` exit 1 → `selftest` red.
- Dropping `erosion` from the shipped `core/leitwerk.tiers` T2 line fails
  `TestShippedDefaults` (`tiers_test.go:120`) → `go test` exit 1 → `selftest` red.

The roadmap's problem statement ("the glob-engine edge cases … are untested")
predates that Go suite and is stale. `init` is covered
(`integration_test.go:154,266`), as are the exit paths (`:225,242,252`, guard
`:131`, drift skip `:208`).

What is genuinely **not** covered (the residual this change closes):
1. **`verify`'s `checks:` line.** `verify` prints the checks it will run
   (`checks: lint types …`); only the `ChecksForTier` *function* is asserted
   (`tiers_test.go:30`), never the built binary's observable output. A
   regression in that line would ship silently.
2. **Shipped `[paths]` first-match ordering.** `TestShippedDefaults`
   (`tiers_test.go:129-138`) checks four paths, none of which exercises a
   first-match conflict. A reorder of `core/leitwerk.tiers` `[paths]` that
   reclassified `docs/x.sql` (T0 today — `docs/**` precedes `**/*.sql`) or an
   `*.yaml`/`*.tf`-glob path would pass every existing test.
3. **The CLI's T1 fallback.** `core/cmd/leitwerk/main.go:64`
   (`if !ok { t = "T1" }`) is unreached with the shipped catch-all `*`; no
   integration test drives a no-catch-all tiers file to hit it.

## Behaviour (the observable contract)
After this change the Go suite (run by `selftest` §0) also pins:
- **Checks-line contract:** running the built binary `verify --tier <T>` against
  the shipped defaults prints a `checks:` line listing exactly that tier's
  checks, cumulative and in file order (T0 → `lint`; T2 → `lint types tests
  drift sast erosion`), and the gate is green when all checks skip.
- **Shipped ordering:** `tier docs/schema.sql` → `T0` (first-match: `docs/**`
  before `**/*.sql`); `tier k8s/deploy.yaml` → `T2` (`**/*.yaml`);
  `tier modules/net.tf` → `T2` (`**/*.tf`, not under `infra/**`). A reorder that
  changes these fails the suite. (`modules/net.tf`, not `infra/x.tf`: the latter
  matches `infra/**` first and would never exercise the `**/*.tf` glob.)
- **Shipped check lists:** `verify --tier T1` lists `lint types tests drift` —
  each tier's line is pinned independently (cumulativeness is a convention, not
  enforced), so weakening the default T1 tier reds the suite.
- **T1 fallback:** with a `[paths]` table that has no catch-all, `tier` on an
  unmatched path prints `T1` (the documented default), exercising
  `main.go:64`.

What must NOT happen: no assertion may depend on this repo's own
`leitwerk/tiers.conf` (non-deterministic across repos) — tests use the shipped
defaults or a fixture tiers file, matching the existing suite's discipline
(`selftest.sh:37`, `integration_test.go:116`).

## Design decisions
- **Land the coverage as Go tests (run under `selftest` §0), not new black-box
  shell assertions in `selftest` §1.** The gaps are about the tier engine and
  the binary's output; Go table tests are the more precise, maintainable layer,
  and `selftest` §0 already makes a Go-test failure a gate failure — so the
  acceptance ("fails selftest") is met without duplicating existing Go coverage
  into shell. *Rejected:* mirroring init/exit/glob into shell §1 (option C at
  scoping) — largely redundant with `integration_test.go`, more surface to
  maintain. The one genuinely observable-output gap (the `checks:` line) is a
  black-box integration test against the built binary, so it still tests the
  artifact, just via Go's harness — deliberately leaving §1 no longer the sole
  home of the observable contract (acceptable: the Go test execs the real
  binary, so there is no coverage gap).
- **Pin each tier's check line independently, incl. T1.** `ChecksForTier`
  returns a tier's configured line verbatim; T1/T2 are separate lines, so
  pinning only T2 leaves the default tier (T1) — the one `verify` uses when no
  tier is given — free to be silently weakened in the agent-editable
  `core/leitwerk.tiers`. The checks-line test asserts T0, T1, and T2.
- **Extend `TestShippedDefaults` in place** rather than a new test — it is the
  named home for "the shipped file carries the documented defaults", and adding
  ordering-sensitive paths there keeps one source of truth for the shipped
  contract.
- **Scope M1.3 to these residuals and close it.** Its headline acceptance is
  already met; gold-plating with redundant assertions is not warranted. The
  roadmap entry moves to "done" recording that the Go suite delivers it.

## Invariants touched
- *A check never fakes a pass.* The new checks-line test asserts a real green
  (all checks skip) with the correct listing — it cannot pass without observing
  the output.
- *Bindings never reimplement the gate.* Test-only change to `core/`; no gate
  logic added or moved; parity unaffected.
- *The gate config is human-owned.* Tests read `core/leitwerk.tiers` (the
  shipped default, not the human-owned `leitwerk/tiers.conf`) and never modify
  it; the roadmap close is a human-owned edit, proposed for approval.

## Blast radius
T2. Touched: `core/cmd/leitwerk/integration_test.go`,
`core/internal/gate/tiers_test.go` (`core/**` = T2), plus the human-owned
`leitwerk/roadmap.md` (close M1.3). Test-only — no production code path
changes. Worst case if wrong: a test that asserts the wrong expected output
(false red) or one too loose to catch a regression (false confidence);
mitigated by mutation-checking each new assertion catches the regression it
claims to.

## Acceptance checks
- `go test -C core ./...` green with the added assertions; and each new
  assertion, when its target is mutated, goes red (checks-line: change verify's
  output; ordering: reorder shipped `[paths]`; fallback: break `main.go:64`).
- `leitwerk verify --tier T2` on the devkit stays green.
- M1.3 recorded done in `leitwerk/roadmap.md`.

## Anchors
- `core/cmd/leitwerk/integration_test.go`
- `core/internal/gate/tiers_test.go`
- `core/leitwerk.tiers`

## Out of scope
- Duplicating existing Go coverage into the shell `selftest` §1 (rejected at
  scoping).
- Changing the shipped tier policy or the glob engine — this only adds tests
  around current behaviour.
- Property/mutation-testing tooling (M3.2) — assertions here are hand-written.
