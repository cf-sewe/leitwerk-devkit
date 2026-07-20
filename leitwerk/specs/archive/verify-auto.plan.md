# Plan — verify --auto

Status: landed (2026-07-20) <!-- landed with its spec at the T2 review -->

Adds `leitwerk verify --auto`. Splits the pure tier-derivation (gate package,
unit-tested) from the git/flag I/O (main.go, integration-tested), then rewires
the two consumers (Stop hook, CI). One strand, one gate-green commit. Spec
anchors are existing paths, so `drift` stays green throughout.

## Steps
Step status: `[ ]` open · `[x]` done · `[~]` deviated — one line why.

1. `[x]` **Pure tier derivation** — new `core/internal/gate/autotier.go` (T2):
   `RankTier(tier string) int` (T0=0,T1=1,T2=2, unknown=high/conservative) and
   `HighestTier(t *Tiers, paths []string) string` (start T0; per path
   `TierForPath` else `T1`; keep max; empty → T0). Unit tests
   `autotier_test.go`: mixed-with-migration → T2, docs-only → T0, empty → T0,
   unknown-tier → highest. Proves it: `go test` (→ `selftest` §0). Gate green
   (nothing else changes).

2. `[x]` **`--auto`/`--base` wiring** — `core/cmd/leitwerk/main.go` (T2):
   extend `cmdVerify` to parse `--auto` and `--base <ref>` (mutually exclusive
   with `--tier` → exit 2). Add `changedFiles(base string) ([]string, error)`:
   base given → validate (reject `-*`; `git rev-parse --verify --quiet
   <base>^{commit}`) then `git diff --name-only <base>...HEAD`; no base →
   `git diff --name-only HEAD` ∪ `git ls-files --others --exclude-standard`.
   Not a git work tree / bad base → usage error (exit 2). Derive via
   `gate.HighestTier`, print `auto: tier <T> (<deciding-file>); N changed
   file(s)`, then `RunVerify`. Integration tests (`integration_test.go`, run the
   built binary in a git fixture): docs-only commit → T0, migration commit → T2,
   `--auto --tier T2` → 2, bad `--base` → 2, non-git dir → 2. Proves it:
   `go test` (→ `selftest` §0) — this IS the roadmap's "selftest covers tier
   derivation from a mixed diff" (Go layer, per M1.3's precedent; no redundant
   shell section). Gate green. **Manual (T2):** run `core/bin/leitwerk verify
   --auto` at the repo root and confirm it derives T2 (core files dirty) and runs.

3. `[x]` **Rewire consumers** — T1:
   - `bindings/claude/hooks/hooks.json` Stop → `leitwerk verify --auto`.
   - `bindings/open/ci/leitwerk-verify.yml` → drop the shell tier-loop; one step
     `leitwerk verify --auto --base "$base"` (flag, so drift Part-2 stays off).
   - `.claude/settings.json` Stop → `--auto` (dogfood), keeping the
     binary-built guard. **Change this one LAST and test the exact hook command**
     before relying on it (it is this session's own turn-end gate). Proves it:
     `verify --tier T2` green; the hook command runs green by hand.

4. `[x]` **Review & land** — T2. `verify --tier T2` green; `drift`; spec
   fidelity; adversarial panel (documented read-only fallback: architect on the
   diff-base semantics, test-engineer on the derivation oracle, security on the
   git-arg handling). Landing ritual: T2 sign-off (human), spec→landed, archive
   spec+plan, close M2.4 in `roadmap.md` (staged-copy), commit gate-green.

## Verification strategy
- New oracles: `autotier_test.go` (pure derivation, incl. mixed/docs/empty/unknown)
  and `integration_test.go` cases (built-binary `--auto` over a git fixture,
  incl. the error paths). Both run under `go test` → `selftest` §0 → the gate.
- Each derivation assertion is checked to red on a mutation (flip the rank order;
  break the base validation) during build.

## Risks & rollback
- **Under-selection (the core risk):** a bug makes `--auto` pick too low a tier →
  silent gate weakening. Mitigation: precondition failures error (exit 2) rather
  than fall back low; unknown tier ranks high; tests cover the mixed diff and the
  error paths. Rollback: revert the commit — consumers return to static `--tier`.
- **Live Stop-hook breakage** (step 3, `.claude/settings.json`): a `--auto` edge
  bug would block this session's turn-end. Mitigation: change it last, after the
  integration tests pass and the hand-run succeeds; `LEITWERK_TIER`/`--tier`
  stays as the override. Rollback: restore the `--tier ${LEITWERK_TIER:-T2}` line.
- **Git-arg injection:** a crafted base could inject a git option. Mitigation:
  reject `-*` and `rev-parse --verify` before use (mirrors `drift.sh`); pass base
  only in the fixed `<base>...HEAD` position, never as a bare arg.

## Roles to wake
- `architect` — the diff-base semantics (flag vs env; Part-2 coupling; the M2.3
  boundary) and the CLI surface.
- `test-engineer` — the derivation oracle and the error-path coverage.
- `security-reviewer` — `--auto` now shells out to `git` with a
  caller-influenced base; confirm the option-injection defence and that no base
  is interpolated unquoted.

## Review outcome (2026-07-20)
T2 gate green; adversarial panel via the documented read-only fallback. Two HIGHs
plus several MED/LOW, all fixed and re-verified:
- *security:* PASS — no injection vector (empirically confirmed); added the
  optional `--` end-of-options hardening on `git diff`.
- *architect:* **HIGH** — `--auto` read `LEITWERK_DIFF_BASE`, so an ambient value
  would flip the Stop hook to committed-range and hide uncommitted T2 edits →
  **decoupled: base is `--base`-flag only.** MEDIUM — dropped the dogfood of this
  repo's Stop hook; it stays pinned T2 (plugin ships `--auto`). MEDIUM — fresh
  repo (no HEAD) blocked every turn → now derives from all tracked+untracked.
  LOW — unknown-tier caveat added.
- *test-engineer:* **HIGH** — the `HighestTier` no-match→T1 fallback was uncovered
  and its regression self-masking (under-selects to T1) → added a no-catch-all
  unit test (mutation-verified). MED — working-tree mode was untested → added an
  untracked-migration case; option-guard assertion was non-discriminating → now
  asserts per-case stderr + no leaked `--output` file. LOW — docs case asserts
  the file count; migration case asserts the T2 checks line; unknown-tier
  exercised through `HighestTier`.
Human: T2 sign-off granted; roadmap close + constitution decision-of-record
approved (staged-copy). No escalated decision diverged from the recommendation.
