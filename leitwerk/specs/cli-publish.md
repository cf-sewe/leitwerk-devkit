# Spec — cli-publish: obtain the gate on a clean machine (release the CLI)

Status: active (2026-07-23) <!-- draft → active → landed YYYY-MM-DD → superseded by <slug> -->

Roadmap: cli-publish

Promotes the `cli-publish` roadmap item from proto-spec to spec. This is the
release mechanism for the `leitwerk` gate binary.

## Problem
The gate is distributed only as source. Every adopter and every CI run today
obtains it one of two ways: a full checkout with `LEITWERK_HOME=…/core` on PATH,
or `go install github.com/cf-sewe/leitwerk-devkit/core/cmd/leitwerk@latest`
(`README.md:197-200`, `docs/adoption.md:9-10`, `core/README.md:68-75`,
CONFIRMED). Both are blocked from a clean machine right now:

- The module path `github.com/cf-sewe/leitwerk-devkit/core` (`core/go.mod:1`,
  CONFIRMED) resolves only once the repository is public. It is not — there is
  no git remote and no tags (`git remote -v` / `git tag -l` both empty,
  CONFIRMED). Every install doc that names `go install …@latest` carries a
  "path provisional until published" caveat (`README.md:200`, CONFIRMED).
- A Claude Code marketplace install sparse-copies **only** the plugin subdir, so
  the plugin launcher cannot reach a sibling `core/`; an adopter must set
  `LEITWERK_HOME` or already have the binary on PATH (`docs/adoption.md:37-38`,
  `bindings/claude/README.md`, CONFIRMED). There is no path for an adopter who
  has no Go toolchain and no checkout — no prebuilt binary is published anywhere
  (no release workflow, no `.goreleaser*`, CONFIRMED by search).

M2.1 makes the module path publicly resolvable and adds a prebuilt-binary path,
so a clean machine can obtain a working gate. This unblocks a marketplace-only
Claude Code adoption (the plugin's launcher can find a globally installed
binary) and the M2.2/M2.3 live validations, which both assume a public repo.

## Behaviour (the observable contract)
After this change, someone on a machine with **no checkout of this repo** can
obtain a working `leitwerk` binary by either path, and run the gate in an
unrelated repo:

- **Path A — `go install` (requires a Go toolchain):**
  `go install github.com/cf-sewe/leitwerk-devkit/core/cmd/leitwerk@latest`
  produces a binary that carries its checks/templates via `//go:embed`
  (`core/assets.go`, CONFIRMED) and runs `leitwerk verify` / `leitwerk init`
  with no sibling files on disk. Works because the repo is public and tagged.
- **Path B — release download (requires no Go toolchain):** each tagged release
  `vX.Y.Z` on GitHub carries prebuilt **static** binaries (CGO disabled,
  matching `core/Makefile`) for `{linux, darwin, windows} × {amd64, arm64}`,
  plus a checksums file. Downloading the binary for the host platform, marking
  it executable, and putting it on PATH yields the same working gate. The asset
  names and the `checksums.txt` file are a **stable, documented contract**: the
  archive name template and the checksums filename do not change across releases
  without a deliberate, recorded decision, so downstream tooling that fetches and
  verifies the binary (a future plugin bootstrap, proposed) can rely on them.
- **Path C — full checkout / vendoring:** unchanged; `LEITWERK_HOME=…/core` with
  `$LEITWERK_HOME/bin` on PATH stays documented as the third path.

Version reporting:

- `leitwerk version` prints one line `leitwerk <version>\n` and exits 0 (the
  existing contract, `core/cmd/leitwerk/main.go:89-91`, CONFIRMED).
- For a **release** build the `<version>` is the release tag (`vX.Y.Z`), injected
  at build time — not a value hand-edited in source.
- For a plain `make build` / `go build` with no injection the `<version>` is a
  stable non-release sentinel (`dev`). The value is never empty and the line
  format never changes.

Distribution docs: the "provisional until published" caveats
(`README.md:200`, and the equivalents in `core/README.md`, `docs/adoption.md`)
are removed; the three obtain paths above are documented and each works.

Release process (contributor-facing contract — mechanism in D2):

- Commits on `main` follow **Conventional Commits** (`feat:`/`fix:`/… , `!` or
  `BREAKING CHANGE:` for a major). A non-conforming commit is not a release
  event; it just does not contribute a changelog entry or a version bump.
- release-please keeps a **Release-PR** open reflecting the pending version and
  changelog. A release happens when — and only when — a human merges it; that
  merge creates the tag and the GitHub Release, and the tag triggers the binary
  build. Nothing releases automatically without that merge.

What must NOT happen:

- The module path does **not** change (`github.com/cf-sewe/leitwerk-devkit/core`
  stays; the repo is published under `cf-sewe`, per the human decision
  2026-07-21). No import rewrite, no `go.mod` module rename.
- A release build never ships with the `dev` sentinel as its version; a
  non-release build never claims a release tag.
- The release mechanism carries no gate logic — it builds and ships the binary
  the gate already is; it does not reimplement any check (parity invariant).
- `leitwerk verify` / `tier` / `guard` / `drift` / `init` behaviour is unchanged
  by this work; only how the binary is *obtained* and what `version` prints
  changes.

Preconditions (ops actions, outside the gate — see Design decisions D6):

- The repository is published as a **public** GitHub repo at
  `github.com/cf-sewe/leitwerk-devkit` (matching the module path). This is the
  human-owned, effectively irreversible action that flips Path A/B from
  "provisional" to working; the code in this spec assumes it.

## Design decisions

- **D1 · Ship all three obtain paths; the two "clean-machine" paths are `go
  install` (A) and release binaries (B).** Acceptance names "go install or
  release download"; A is free once the repo is public (the embed already makes
  a lone binary self-sufficient), B serves adopters and CI runners with no Go
  toolchain (the common case for a marketplace-only Claude Code user). C
  (vendoring/`LEITWERK_HOME`) already exists and stays. *Rejected:* `go install`
  only — leaves the no-Go adopter (the marketplace-only case that motivated
  M2.1) with no path.

- **D2 · release-please owns version/tag/changelog; GoReleaser builds and
  attaches the binaries.** They solve different problems and compose; release-
  please compiles nothing, so it does not replace GoReleaser for Path B. Two
  GitHub Actions workflows, two responsibilities:
  - `.github/workflows/release-please.yml` (on push to `main`): release-please
    maintains a **Release-PR** that computes the next SemVer version from
    Conventional Commits (D7) and updates `CHANGELOG.md`. Merging that PR is the
    per-release human gate; on merge release-please creates the tag `vX.Y.Z` and
    the GitHub Release with generated notes.
  - `.github/workflows/release.yml` (on push of a `v*` tag — the tag release-
    please just created): sets up Go from `core/go.mod` and runs `goreleaser
    release`. GoReleaser builds the static cross-platform matrix (D1/D3) plus
    checksums and uploads them to the **existing** release with
    `release.mode: append`; `changelog.disable: true`, because release-please
    owns the notes.
  Division of labour: release-please answers *which version, when, what
  changelog*; GoReleaser answers *given a tag, build and attach the artifacts*.
  Both run only in CI, carry no gate logic, and touch no check — parity intact.
  Config: `release-please-config.json` + `.release-please-manifest.json` and
  `.goreleaser.yaml`. *Rejected:* GoReleaser with hand-cut tags (loses the
  Release-PR human gate and the automated changelog — viable but less aligned
  with human-owns-intent); release-please alone (compiles nothing → Path B has
  no binaries); a hand-rolled build matrix (reimplements the archive/checksum/
  upload/version-injection GoReleaser already does correctly).

- **D3 · Version stamped by ldflags from the tag; source default is `dev`.**
  Change `core/cmd/leitwerk/main.go` from `const version = "0.1.0"` to
  `var version = "dev"`, injected at build with
  `-ldflags "-X main.version=<tag>"` (GoReleaser does this; `core/Makefile`
  gains the same injection for a local release-style build, defaulting to `dev`
  otherwise). A release binary then reports its tag; a `make build` reports
  `dev`. The tag is the single source of truth: release-please creates it (D2),
  GoReleaser injects it; no version literal is hand-maintained in source, and
  release-please does not write one into `main.go` (it tracks the version in
  `.release-please-manifest.json`). *Rejected:* keeping a hand-edited `const
  version` bumped per release —
  it silently drifts from the tag and defeats the point of a trustworthy
  version string; the whole reason to publish releases is that the reported
  version is authoritative. **Consequence (CONFIRMED):** `TestIntegrationVersion`
  (`core/cmd/leitwerk/integration_test.go:252-257`) pins the literal
  `leitwerk 0.1.0`; it must move to asserting the line *format* and the injected
  value (build the test binary with a known `-X main.version` and assert it
  round-trips), not a hard-coded release number. This is part of acceptance.

- **D4 · Blast radius is T2, higher than the roadmap's T1 estimate — recorded,
  not silently accepted.** The roadmap lists M2.1 as T1, but the concrete
  mechanism edits `core/cmd/leitwerk/main.go` (T2), `core/Makefile` (T2), and
  adds `.github/workflows/release.yml` (T2) — verified with `leitwerk tier`. The
  honest tier is therefore **T2** (see Blast radius). This is a *strengthening*
  (more verification), safe to raise; the roadmap's tier was a proto-spec
  estimate and the spec's tier is set from the files actually touched. The T2
  landing sign-off is human-owned (constitution) — flagged for the landing
  review, not decided here.

- **D5 · Versions from Conventional Commits; seed `0.1.0`, pre-1.0 = unstable
  surface.** release-please derives each bump from commit types (`fix:`→patch,
  `feat:`→minor, `!`/`BREAKING CHANGE`→major). `.release-please-manifest.json`
  is seeded at `0.1.0` (the value the source already carried) so the first
  Release-PR proposes `v0.1.0`; the `v`-prefixed SemVer tag is the Go-module
  convention required for `go install …@vX.Y.Z`/`@latest`. Pre-1.0 signals the
  CLI surface may still change. The per-release human decision is merging the
  Release-PR, not a hand-cut tag. *Rejected:* starting at `v1.0.0` (over-claims
  stability before the M2.2/M2.3 live validations).

- **D6 · Publishing the repo is a one-time ops action, not part of the gated
  diff.** `gh repo create --public` / `git remote add` / `git push` are human-
  owned, effectively irreversible, outward-facing actions (the human approved
  `cf-sewe` public on 2026-07-21). They are a *precondition* this spec's code
  assumes, not code the gate verifies. The first *tag* is **not** a manual ops
  step under D2 — it is created by merging the first Release-PR (release-please),
  which is the recurring per-release human gate. So the one-time human action is
  "make the repo public"; the per-release human action is "merge the Release-PR".
  *Rejected:* scripting the publish inside the repo (a one-time irreversible
  action does not belong behind an automated check).

- **D7 · Adopt Conventional Commits — the cost of D2's automation.** release-
  please computes the version and changelog from commit-message prefixes
  (`feat:`, `fix:`, `feat!:`/`BREAKING CHANGE:`, plus `docs:`/`chore:`/… for
  grouping). This repo's history uses milestone prefixes (`M2.4: …`, CONFIRMED
  via `git log`), so going forward commits must follow Conventional Commits for
  release-please to work. This is a workflow/intent change accepted by the human
  (2026-07-21). The convention is documented for humans and agents as part of
  this change (a commit-convention note and a line in the open-code
  `bindings/open/AGENTS.md`); recording it in the constitution's decisions of
  record is *proposed* to the human (that file is human-owned). No gate check
  parses commit messages — enforcement is release-please's behaviour plus
  review, not a new check. *Rejected:* keeping ad-hoc commit messages (release-
  please could then derive neither a version nor a changelog).

## Invariants touched
- **Bindings never reimplement the gate / parity holds.** The release mechanism
  builds and ships the existing binary; it adds no check and no gate logic. The
  `parity` structural check stays green (nothing in a binding gains gate logic).
- **A check never fakes a pass / the gate never under-verifies.** Unaffected —
  `verify`'s behaviour is untouched. The version string change is observable
  metadata only.
- **The gate config is human-owned.** Untouched — no change to `tiers.conf`,
  the constitution, or the tier→path policy. The roadmap tier estimate is
  re-recorded (D4) as a proposal to the human, not an edit to a human-owned file.

## Blast radius
**T2.** Highest tier among the files touched (each via `leitwerk tier`,
CONFIRMED):
- `core/cmd/leitwerk/main.go` — T2 (the `version` var + ldflags seam).
- `core/Makefile` — T2 (ldflags injection for local builds).
- `.github/workflows/release.yml` — T2 (new tag-triggered GoReleaser workflow).
- `.github/workflows/release-please.yml` — T2 (new release-please workflow).
- `.goreleaser.yaml`, `release-please-config.json`,
  `.release-please-manifest.json` — T1 (repo-root config; the `.json` pair is
  parsed by the existing `json` check).
- `README.md`, `core/README.md`, `docs/adoption.md`,
  `bindings/open/AGENTS.md` — T0/T1 (doc caveats removed; commit convention
  documented, D7).

Worst case if it ships wrong: a published binary reports a wrong or empty
version, or the release matrix produces a broken/non-static binary for some
platform, so an adopter obtains a gate that misidentifies itself or does not run
— an adoption/trust failure, not a weakening of the gate's verification (the
checks are unchanged). Mitigated by: the version round-trip test (D3), a
static-build assertion, `goreleaser check` on the config, and the T2 review
panel. Publishing the repo public is irreversible — mitigated by D6 (explicit
human sign-off before making it public) and by the Release-PR gate (no release
cuts without a human merge).

## Acceptance checks
Gate-verifiable (run in `leitwerk verify` for this area):
- **Version format + injection:** `TestIntegrationVersion` asserts `leitwerk
  version` prints one line matching `^leitwerk \S+$`, exit 0; and a binary built
  with `-ldflags "-X main.version=vTEST"` reports `leitwerk vTEST`. No pinned
  release literal remains.
- **Local build unchanged + static:** `make -C core build` still succeeds and
  the default build reports `leitwerk dev`; the release build flags keep
  `CGO_ENABLED=0` (static).
- **Release config valid:** `.goreleaser.yaml` passes `goreleaser check`, and a
  `goreleaser build --snapshot --clean` (no publish) produces the full
  `{linux,darwin,windows}×{amd64,arm64}` matrix locally. The archive
  `name_template` and the `checksums.txt` filename are pinned in
  `.goreleaser.yaml`, and the snapshot build produces assets with those
  documented names (the stable contract in §Behaviour Path B). The release-please
  config + manifest parse (the existing `json` check covers the `.json` pair).
- **The devkit's own gate stays green:** `leitwerk verify --tier T2`.

Live proof (one-time, recorded like an M2.2/M2.3 validation — not a per-change
gate check, because it needs the network and a real GitHub release):
- The first Release-PR opens proposing `v0.1.0` with a changelog; merging it
  creates the `v0.1.0` tag and GitHub Release; the tag triggers GoReleaser,
  which appends the binaries + checksums to that release.
- On a clean machine with a Go toolchain: `go install
  github.com/cf-sewe/leitwerk-devkit/core/cmd/leitwerk@latest`, then `leitwerk
  verify --tier T1` runs in an **unrelated** repo (embed self-sufficiency), and
  `leitwerk version` prints the released tag.
- On a clean machine with **no** Go toolchain: download the release binary for
  the platform, verify its checksum, and run `leitwerk verify` in an unrelated
  repo.
- The plugin launcher, given only a globally installed `leitwerk`, resolves the
  core CLI without a sibling `core/` (the marketplace-only case).

## Anchors
Only files that already exist are anchored (drift resolves anchors eagerly and a
not-yet-created path would go red on a draft). The release configs/workflows
(`.github/workflows/release.yml`, `.github/workflows/release-please.yml`,
`.goreleaser.yaml`, `release-please-config.json`,
`.release-please-manifest.json`) become anchors when the build creates them —
bidirectional refinement, added in the same change that lands the files.
- `core/cmd/leitwerk/main.go`
- `core/Makefile`

## Out of scope
- **M2.2 · plugin-live-validation** and **M2.3 · ci-live** — separate roadmap
  items; this spec unblocks them (public repo) but does not perform them.
- Additional distribution channels (Homebrew tap, apt/deb, container image) —
  add later only if adoption demands them.
- Binary signing / provenance (cosign, macOS notarization, SLSA attestation) —
  a possible later hardening; not required to obtain a working gate.
- Any change to the module path, the checks, the tier ladder, or tier→path
  policy.
