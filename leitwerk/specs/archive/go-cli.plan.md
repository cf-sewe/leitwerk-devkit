# Plan — reimplement the core CLI as a compiled Go binary

Status: landed (2026-07-19) <!-- all steps landed with the Go cutover -->

Spec: `leitwerk/specs/archive/go-cli.md`. The replacement is a strangler-fig: the Bash CLI
stays at `core/bin/leitwerk` and keeps the gate (and the PreToolUse guard hook)
working until the Go binary is built and proven, then the binary overwrites it.

## Steps
1. **Toolchain via mise** — `mise.toml` pins Go + Node LTS; drop `.tool-versions`
   (no asdf). Files: `mise.toml`. Tier T1. Proves: `mise exec -- go version`.
2. **Go module + gate library** — `core/go.mod` (module
   `github.com/cf-sewe/leitwerk-devkit/core`, stdlib only) and `core/internal/gate/`:
   `glob.go`, `tiers.go` (incl. guard matching), `resolve.go`, `verify.go`. Tier T2. Proves:
   `go build ./...`. Oracle written first (step 6).
3. **Embedded assets** — `core/assets.go` (`//go:embed checks templates
   leitwerk.tiers`) so the binary is self-contained; on-disk siblings win, embed is
   the fallback. Tier T2. Proves: an integration test running from a temp dir with
   no siblings.
4. **CLI entrypoint** — `core/cmd/leitwerk/main.go`: arg dispatch mirroring the
   Bash `case`, wiring the gate library + assets. Build output → `core/bin/leitwerk`
   (gitignored). Tier T2. Proves: black-box behaviour matches the old script.
5. **Build step** — `core/Makefile` (`build`, `test`, `install`, `clean`) with
   `CGO_ENABLED=0`; `.gitignore` the binary. Tier T2. Proves: `make -C core build`
   yields an executable at `core/bin/leitwerk`.
6. **Oracles (written before/with the code)** — `*_test.go` unit tests for
   glob→regex, tier-for-path, checks-for-tier, guard matching, tiers parsing, check
   resolution; integration tests for `verify` on `reference-app`, `init` output,
   and layout-independence via embed. Tier T2.
7. **Self-hosting checks** — rewrite `leitwerk/checks/selftest.sh` to build +
   `go test` + keep the black-box golden assertions; adjust
   `leitwerk/checks/shell.sh` to exclude the compiled binary; extend
   `leitwerk/checks/parity.sh` R1 to scan `core/cmd` + `core/internal`. Tier T2.
8. **CI + distribution docs** — add `actions/setup-go` + a build step to both jobs
   in `.github/workflows/leitwerk.yml`; update `core/README.md`, root `README.md`,
   `docs/adoption.md`, and `bindings/open/*` to document `go install` / prebuilt
   binary / vendoring instead of the non-existent npm package. Tier T2/T0.
9. **Cutover + gate** — build the binary over `core/bin/leitwerk`, run
   `leitwerk verify --tier T2` to green, then the review panel.

## Verification strategy
- New behaviour is proven by Go unit tests (table-driven, one table per engine
  function) plus integration tests that exec the built binary — added in step 6,
  before/with steps 2–4.
- The existing black-box `selftest` golden assertions are **kept** as a
  characterization test of the external contract, now run against the binary.
- The T2 gate's check list is unchanged by this change (`json shell drift
  selftest parity`; the `context` check was wired later, by a separate change);
  the Go build+test folds into `selftest`, so no human-owned `[tiers]` edit is
  needed.

## Risks & rollback
- **Guard-hook lockout:** while `core/bin/leitwerk` is mid-swap, the PreToolUse
  guard resolves the CLI to evaluate every Write/Edit. Mitigation: all `.go`/doc
  edits happen while the *Bash* CLI is still in place; the binary is written via a
  Bash `go build -o` (not a guarded Write/Edit tool) only after `go test` passes.
  Rollback: `git checkout core/bin/leitwerk` restores the Bash gate instantly.
- **Fresh-clone bootstrap:** the binary is gitignored, so a clone must `make -C
  core build` before the gate runs; documented in CLAUDE.md/README and built first
  in CI. Rollback is the same checkout.
- **Behaviour drift from the awk:** mitigated by porting the translation verbatim
  and asserting it in unit tests plus the retained black-box golden suite.

## Roles to wake
- `architect` — module boundary (core-only gate, launcher still delegates), embed
  fallback design, the proposed `.go`-path T2 rule.
- `test-engineer` — the Go test suite and the rewritten `selftest`; confirm each
  contract clause has an oracle.
- `security-reviewer` — T2 + the CLI now parses payload-derived paths (guard) and
  extracts embedded scripts to a temp dir; check path/temp handling.
