# Spec — reimplement the core CLI as a compiled Go binary

Status: landed (2026-07-19) <!-- durable content: constitution decision of record (Go reimplementation) -->

## Problem
`core/bin/leitwerk` is a ~170-line Bash script. It is the one deterministic
artifact the whole framework rests on (CI, the Stop hook, and open-code all invoke
it), yet it is the least testable part: its glob→regex engine and tiers parser live
in inline `awk`, error handling relies on `set -euo pipefail`, and the only
coverage is the black-box `selftest`. A defect in this script silently weakens
every adopting repo (the constitution's worst case). We want the same external
contract delivered by a robust, unit-testable, single static binary that is easy
to cross-compile and distribute.

Request: reimplement the gate in a compiled language, preserving the exact
external contract and the framework invariants.

## Decision — Go over Rust (proposed decision-of-record)
Investigated both against the constraints (single static binary, easy
cross-compile, simple distribution to adopters + CI, mature CLI-arg and glob
libraries) for this specific workload (subcommand dispatch, INI-ish parsing,
glob→regex, shelling out to check scripts, asset embedding):

- **Static binary:** Go is static by default (`CGO_ENABLED=0`); Rust needs the
  `musl` target and sometimes a C toolchain.
- **Cross-compile:** Go uses `GOOS`/`GOARCH` with no cross-toolchain; Rust needs a
  per-triple target and often a cross linker.
- **Distribution/CI:** Go builds in seconds with a tiny toolchain and zero
  dependencies (stdlib only → empty lockfile, offline, reproducible); `go install`
  works out of the box. Rust recompiles crates (slower cold CI).
- **Libraries:** the glob engine must be hand-rolled in *either* language to match
  the awk's exact `**/`→`(.*/)?` + catch-all semantics, so Rust's `clap`/`globset`
  edge mostly evaporates; stdlib `flag` + `regexp` suffice.
- Rust's real advantages (memory safety, performance) do not pay off in an
  I/O-bound tool whose work is spawning shell scripts.

**Chosen: Go.** Recorded as a proposed decision-of-record for
`leitwerk/constitution.md` (human-owned — proposed here, applied by a human).

## Behaviour (the observable contract — unchanged)
The binary is a drop-in replacement for the Bash CLI at the same path
(`core/bin/leitwerk`). Everything below is byte-for-byte compatible unless noted.

- **Subcommands:** `verify [--tier T0|T1|T2]`, `tier <path>`, `guard <path>`,
  `drift`, `init [dir]`, `version` (`--version`, `-v`), and `help`/`--help`/`-h`/
  no-args.
- **Exit codes:** `0` = gate green / path editable; `1` = a check failed (gate
  red); `2` = usage error; `3` = path is human-owned (guard). A check's own exit
  `2` means *skip* (not a fail); any other non-zero from a check fails the gate.
- **`verify`:** default tier `T1`; reads the tier's cumulative check list from the
  tiers file; runs each resolved check in the caller's working directory (as the
  Bash CLI did — invoke from the repo root); prints one status line
  per check (`✓` pass / `–` skip / `✗` fail / `?` no-such-check) showing the last
  line of the check's combined output; prints `gate: PASS`/`gate: FAIL`. An empty
  check list for the tier, or an unknown `--tier`-adjacent option, is a usage error
  (exit 2). Colour is emitted when stdout is a TTY and `NO_COLOR` is unset.
- **`tier <path>`:** prints the tier for the path via the `[paths]` glob table
  (first match wins); prints `T1` when nothing matches. No arg → usage error.
- **`guard <path>`:** exit `0` if editable, `3` if the path suffix-matches a
  `[human-owned]` glob (matched with a `(^|/)` prefix so an absolute path from a
  hook payload resolves); reason to stderr. No arg → usage error. *Noted
  deviations (hardening, not in the Bash CLI):* the path is normalized before
  matching, so dot-segment spellings like `leitwerk/./constitution.md` or
  `x/../leitwerk/tiers.conf` are also blocked; matching is case-insensitive on
  case-insensitive filesystems (macOS/Windows).
- **`drift`:** runs the built-in `drift.sh` (embedded fallback if no on-disk copy)
  in the caller's working directory and propagates its exit code.
- **`init [dir]`:** scaffolds, into `dir` (default `.`):
  `leitwerk/constitution.md` and `leitwerk/tiers.conf` (always written), and
  `CLAUDE.md`, `.claude/rules/tier-discipline.md`,
  `.claude/workflows/leitwerk-review.mjs` (written only if absent). Same summary
  line as today.
- **Glob→regex engine:** exact awk parity — escape `.`; `**/`→`(.*/)?`; remaining
  `**`→`.*`; `*`→`[^/]*`; a bare `*` glob → catch-all `.*`; anchor `^…$` (tier) or
  `(^|/)…$` (guard).
- **tiers file resolution:** `$LEITWERK_TIERS` if set, else `leitwerk/tiers.conf`
  if present, else the built-in default (`core/leitwerk.tiers` on disk, or the
  embedded copy). `[tiers]` = tier→checks (cumulative), `[paths]` = glob→tier
  (first match wins), `[human-owned]` = glob list.
- **Check resolution:** for check `<name>`, prefer `$LEITWERK_CHECKS/<name>.sh`
  (default `leitwerk/checks`), else the built-in `core/checks/<name>.sh` on disk,
  else the embedded built-in extracted to a cache dir. Repo-local overrides
  built-in per check.
- **Robustness:** no panic on a malformed tiers file, a missing file, or bad args;
  errors go to stderr prefixed `leitwerk:`; output is deterministic.

## Invariants touched
- **Core never depends on an agent runtime** — the binary is plain, needs only a
  shell to run the checks it orchestrates; Go is a *build* toolchain, not a
  runtime dependency of running the gate once built.
- **Bindings never reimplement the gate** — the launcher
  `bindings/claude/bin/leitwerk` still only resolves and execs `core/bin/leitwerk`;
  the `parity` check keeps that true and is extended to scan the Go source dirs.
- **Open-code guarantee-parity** — the hard guarantee still lives in `core/`,
  reachable via `leitwerk verify` + CI with no agent runtime. Building from source
  (or a prebuilt binary) does not move the guarantee into a binding.
- **A check never fakes a pass** — skip (exit 2) vs fail semantics preserved
  exactly.
- **The gate config is human-owned** — `guard` behaviour and the human-owned list
  are unchanged; constitution/tiers/roadmap edits are proposed, not made.

## Blast radius
**T2.** The change replaces the gate itself and touches `core/bin`, `*.sh`
(checks), and `.github/`. Worst case if it ships wrong: the framework ships a gate
whose tier selection, guard, or pass/fail logic diverges from the contract, which
would silently mis-gate every adopting repo. Mitigated by keeping the black-box
`selftest` golden suite and adding unit + integration tests that must pass before
it lands.

## Acceptance checks
`leitwerk verify --tier T2` is green on a clean tree, where the T2 gate's
`selftest` check now:
1. builds the Go binary (`go build`),
2. runs `go test ./...` in `core/` — unit tests for the glob→regex engine,
   tier-for-path, guard matching, tiers parsing, and check resolution, plus an
   integration test that runs `verify` on `examples/reference-app` and one that
   runs the binary from a directory with no sibling `checks/`/`templates/` (proving
   the embedded assets make it independent of repo layout), and
3. runs the existing black-box golden assertions against the built binary (tier
   mapping on the shipped defaults, guard allow/deny incl. absolute-path suffix,
   the scaffolded review-workflow parity, and a green gate on `reference-app`).

The gate goes red if any of: the glob translation or tier table is mutated, the
guard list logic regresses, a manifest is corrupted, or a tracked shell script has
a syntax error.

## Out of scope
- Changing any tier→check policy or path→tier mapping (human-owned; a `.go`-path
  T2 rule is *proposed* for `leitwerk/tiers.conf`, not applied here).
- Publishing the package to a registry (roadmap M2.1 / M2.3) — this spec documents
  the `go install` / prebuilt-binary distribution path but does not publish.
- Reimplementing the check scripts in Go — they stay Bash and shell out to real
  toolchains; the binary only orchestrates them.
