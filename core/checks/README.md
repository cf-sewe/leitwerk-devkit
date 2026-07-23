# core/checks/ — the shipped (default) gate checks

The generic check library every adopter gets: the binary embeds these scripts
(`core/assets.go`) and also runs them from disk in a checkout. They are what
`leitwerk verify` orchestrates when a repo has not overridden a check.

Contrast with `leitwerk/checks/` — that directory is *this* repo's own,
atypical, override set (framework-integrity checks). The scripts here are the
generic **application** set an adopter starts from. The authoritative tier→check
map is the human-owned policy file `core/leitwerk.tiers` (an adopter's working
copy is `leitwerk/tiers.conf`); this README only narrates the scripts.

## The contract

Each check is a shell script `<name>.sh` with one contract:

- `exit 0` — pass
- `exit 1` — fail (gate red)
- `exit 2` — **skip**: nothing to run here. A skip is never a fake pass.

Honesty caveat (whitepaper §9): **the gate cannot verify a check's own honesty.**
For the shipped checks below the exit-2-means-skip property holds *by
construction*. For a repo-local override it is a constitutional convention that
review must uphold — a script that `exit 0`s without really testing is not
something the gate can detect. See "Are these strong tests?" below.

## Activation is two-part: project marker × tool

Whether a shipped check *runs* depends on two conditions — a project marker
**and**, for some checks, an installed tool. If either is missing the check
skips (exit 2); the gate stays green but that check did not run.

| Check | File | Runs when… | Detected toolchains |
|---|---|---|---|
| `lint` | `lint.sh` | marker **and** tool | `package.json` w/ `"lint"` → `npm run lint`; `go.mod` **+ `golangci-lint`** → `golangci-lint run`; `./gradlew` → delegated (skip) |
| `types` | `types.sh` | marker present | `tsconfig.json` → `tsc --noEmit`; `go.mod` → `go vet ./...`; `./gradlew` → delegated (skip) |
| `tests` | `tests.sh` | marker present | `package.json` w/ `"test"` → `npm test`; `./gradlew` → `gradlew test`; `go.mod` → `go test ./...` |
| `sast` | `sast.sh` | tool present | `semgrep` → `semgrep --config auto`; else skip |
| `erosion` | `erosion.sh` | tool present | `jscpd` → duplication ≤ 5%; else skip |
| `drift` | `drift.sh` | specs present | spec `## Anchors` resolve; one-sided spec/code change (under a diff base) is red; no specs dir → skip |

Note the split: `go vet` / `go test` ship **with Go**, so `types`/`tests`
activate the moment a `go.mod` exists — zero extra setup. `lint`/`sast`/`erosion`
need a separate tool on PATH.

## Activate a check = install (and pin) its tool

To turn on the Go linter, put `golangci-lint` on PATH (pin it in `mise.toml` so a
fresh checkout and CI resolve the same version); `lint.sh` picks it up on its own
— no script edit. Likewise `semgrep` for `sast`, `jscpd` for `erosion`. Until the
tool is present the check skips honestly rather than blocking.

## Custom languages & non-standard suites — override or add

A check is *any deterministic exit-code command*; the names `lint`/`types`/`tests`
are not privileged. Resolution is first-match-wins:

**repo-local `leitwerk/checks/` → built-in `core/checks/` → embedded in the binary**
(`LEITWERK_CHECKS` repoints the local dir).

- **A language the defaults don't detect** (Rust, Python, Elixir, …): drop a
  repo-local `leitwerk/checks/lint.sh` (or `types`/`tests`) that calls the native
  tool — `cargo clippy`, `ruff`, `mix test`. It overrides the generic script of
  the same name.
- **Non-standard suites** (integration, e2e, contract, fuzz): add them as their
  **own** checks — `leitwerk/checks/integration.sh` (e.g. `docker compose up -d
  && run-suite`), `contract.sh` — and wire each into `tiers.conf` at the tier
  that should trigger it (integration/migration suites usually gate T1/T2, a fast
  unit run T0). This is how "integration tests run on risky changes" becomes
  mechanism rather than discipline.

Never edit the installed `core/checks/`; a repo customizes only through its own
`leitwerk/checks/` (per the onboarding skill).

## Are these strong tests? (check quality)

Running is not the same as asserting. AI-written tests often reach high line
coverage while asserting nothing, and `go test ./...` with no tests passes
vacuously. The framework's stated answer is a **mutation-score floor** — inject
faults and require the suite to catch them — which the whitepaper names as the
T1+ design target ("Mutation score, not line coverage", whitepaper §9, Fig. 6).
That is a *design target a mature repo grows into*, not a shipped default: today
check honesty is a review-upheld convention (the gate cannot verify it), and the
mechanization is roadmap work — **M3.2 · verification-helpers** (property /
mutation / characterization oracles) and **M3.6 · required-checks** (a check may
not skip at a tier). This README documents what runs and how to wire it; it does
not claim the shipped checks enforce test *quality*.

## Adding or changing a check

An agent may **add** a check (drop `<name>.sh` following the exit contract and
**propose** its wiring in `tiers.conf`). Changing a shipped check is a gated
change like any other (`core/checks/*` is T2 — a defect here weakens every
adopter). Lowering a threshold, removing a check, or downgrading a path's tier
lives in the human-owned `leitwerk/tiers.conf` / `leitwerk/constitution.md` —
propose, do not edit (enforced by `leitwerk guard` + the `PreToolUse` hook).
