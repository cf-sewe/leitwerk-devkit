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

- The repository is now public at `github.com/cf-sewe/leitwerk-devkit` (`origin`
  set, `isPrivate: false`, CONFIRMED 2026-07-23), so the module path is
  reachable — but there are no version tags yet (`git tag -l` empty, CONFIRMED),
  so `go install …@latest` has no version to resolve. Compounding this: the
  module lives in the `core/` **subdirectory** (`core/go.mod:1`, no repo-root
  `go.mod`, CONFIRMED), and Go's rule for a subdirectory module is that version
  tags must be prefixed with the subdirectory — `core/vX.Y.Z`, not a bare
  `vX.Y.Z`. No such tag exists. Every install doc that names `go install …@latest`
  carried a "path provisional until published" caveat (CONFIRMED at spec time;
  removed by this change).
- A Claude Code marketplace install sparse-copies **only** the plugin subdir, so
  the plugin launcher cannot reach a sibling `core/`; an adopter must set
  `LEITWERK_HOME` or already have the binary on PATH (`docs/adoption.md:37-38`,
  `bindings/claude/README.md`, CONFIRMED). There is no path for an adopter who
  has no Go toolchain and no checkout — no prebuilt binary is published anywhere
  (no release workflow and no release/build config of any kind, CONFIRMED by
  search).

This change cuts the first `core/`-prefixed release tags and adds a prebuilt-
binary path, so a clean machine can obtain a working gate. This unblocks a
marketplace-only Claude Code adoption (the plugin's launcher can find a globally
installed binary) and the plugin-live-validation / ci-live validations.

## Behaviour (the observable contract)
After this change, someone on a machine with **no checkout of this repo** can
obtain a working `leitwerk` binary by either path, and run the gate in an
unrelated repo:

- **Path A — `go install` (requires a Go toolchain):**
  `go install github.com/cf-sewe/leitwerk-devkit/core/cmd/leitwerk@latest`
  produces a binary that carries its checks/templates via `//go:embed`
  (`core/assets.go`, CONFIRMED) and runs `leitwerk verify` / `leitwerk init`
  with no sibling files on disk. Works because the repo is public and carries
  `core/vX.Y.Z` tags (the subdirectory-module tag form `@latest` resolves).
- **Path B — release download (requires no Go toolchain):** each tagged release
  `core/vX.Y.Z` on GitHub carries prebuilt **static** binaries
  (`CGO_ENABLED=0`) for `{linux, darwin} × {amd64, arm64}` — four
  targets; Windows is out of scope (see Out of scope) — plus a `checksums.txt`.
  Downloading the binary for the host platform, marking it executable, and
  putting it on PATH yields the same working gate. The asset names and the
  `checksums.txt` filename are a **stable, documented contract**: they do not
  change across releases without a deliberate, recorded decision, so downstream
  tooling that fetches and verifies the binary (the proposed plugin-bootstrap)
  can rely on them.
- **Path C — full checkout / vendoring:** unchanged; `LEITWERK_HOME=…/core` with
  `$LEITWERK_HOME/bin` on PATH stays documented as the third path.

Version reporting:

- `leitwerk version` prints one line `leitwerk <version>\n` and exits 0 (the
  existing contract, `core/cmd/leitwerk/main.go:89-91`, CONFIRMED).
- For a **release** build the `<version>` is the SemVer of the release tag
  (`vX.Y.Z` — the `core/` tag prefix is stripped so `leitwerk version` prints
  `leitwerk v0.1.0`, not the tag path), injected at build time — not a value
  hand-edited in source.
- For a plain `mise run build` / `go build` with no injection the `<version>` is a
  stable non-release sentinel (`dev`). The value is never empty and the line
  format never changes.

Distribution docs: the "provisional until published" caveats in `README.md` and
`core/README.md` are removed; the three obtain paths above are documented and
each works.

Release process (contributor-facing contract — mechanism in D2):

- Commits on `main` follow **Conventional Commits** (`feat:`/`fix:`/… , `!` or
  `BREAKING CHANGE:` for a major). A non-conforming commit is not a release
  event; it just does not contribute a changelog entry or a version bump.
- release-please keeps a **Release-PR** open reflecting the pending version and
  changelog. A release happens when — and only when — a human merges it; that
  merge creates the `core/vX.Y.Z` tag and the GitHub Release, and the tag
  triggers the binary build. Nothing releases automatically without that merge.

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
  `github.com/cf-sewe/leitwerk-devkit` (matching the module path) — **satisfied
  2026-07-23** (`origin` set, `isPrivate: false`, CONFIRMED). This is the
  human-owned, effectively irreversible action that flipped Path A/B from
  "provisional" to reachable; what remains is cutting the first tag and binaries.
- **Squash-merge is enabled** on the repo (GitHub → Settings → "Allow squash
  merging", PR title as the default squash commit message; other merge methods
  off if only squash is intended). This makes the PR title — validated by
  `semantic-pr` — the commit release-please reads (D7). Human-owned repo setting,
  not gated code.

## Design decisions

- **D1 · Ship all three obtain paths; the two "clean-machine" paths are `go
  install` (A) and release binaries (B).** Acceptance names "go install or
  release download"; A is free once the repo is public (the embed already makes
  a lone binary self-sufficient), B serves adopters and CI runners with no Go
  toolchain (the common case for a marketplace-only Claude Code user). C
  (vendoring/`LEITWERK_HOME`) already exists and stays. *Rejected:* `go install`
  only — leaves the no-Go adopter (the marketplace-only case that motivated
  cli-publish) with no path.

- **D2 · release-please owns version/tag/changelog; a second job in the same
  workflow builds and attaches the binaries — no GoReleaser.** For a single
  static binary from one `main` package, the build+attach step is a short matrix
  loop; release-please compiles nothing, so it needs a partner for Path B, but
  that partner does not need to be GoReleaser. **One** workflow
  (`.github/workflows/release-please.yml`, on push to `main`), two jobs:
  - `release-please`: maintains a **Release-PR** that computes the next SemVer
    version from Conventional Commits (D7) and updates the changelog. Merging
    that PR is the per-release human gate; on merge release-please creates the
    **`core/`-prefixed** tag `core/vX.Y.Z` (D5) and the GitHub Release with
    generated notes, and emits `release_created` / `tag_name` outputs.
  - `build-and-upload` (`needs:` the first job, `if: release_created`): checks
    out the tag, builds the `{linux, darwin} × {amd64, arm64}` matrix with
    `CGO_ENABLED=0 GOOS=… GOARCH=… go build -C core -trimpath -ldflags "-s -w -X
    main.version=${TAG#core/}"` (D3), computes `sha256` into `checksums.txt`, and
    uploads the four binaries + checksums to the release release-please just
    created (`gh release upload "$TAG"`). Go is set up from `core/go.mod`.
  - *Why one workflow, not a tag-triggered `release.yml`:* a tag pushed by the
    default `GITHUB_TOKEN` does not trigger another workflow (GitHub's recursion
    guard), so a separate `on: push: tags` build would silently never run without
    a long-lived PAT secret. Chaining the build as a dependent job in the same
    run avoids both the missed trigger and the secret — a smaller supply-chain
    surface.
  Division of labour: release-please answers *which version, when, what
  changelog*; the build job answers *given a release, build and attach the
  artifacts*. Both run only in CI, carry no gate logic, and touch no check —
  parity intact. Config: `release-please-config.json` +
  `.release-please-manifest.json` (no `.goreleaser.yaml`). *Rejected:*
  **GoReleaser** — it adds a toolchain not otherwise pinned (`mise.toml`) and a
  config DSL, and its value (archives, nfpm/deb, Homebrew, Docker, signing) is
  all out of scope here; it also assumes a plain semver tag and is awkward with
  the `core/`-prefixed tag D5 requires, whereas the matrix strips the prefix with
  `${TAG#core/}`. If a later item adds those channels, adopting GoReleaser then
  is a contained, reversible change. *Also rejected:* a separate tag-triggered
  `release.yml` with a PAT (extra secret, larger attack surface); release-please
  alone (compiles nothing → Path B has no binaries); hand-cut tags without
  release-please (loses the Release-PR human gate and the automated changelog).

- **D3 · Version stamped by ldflags from the tag; source default is `dev`.**
  Change `core/cmd/leitwerk/main.go` from `const version = "0.1.0"` to
  `var version = "dev"`, injected at build with
  `-ldflags "-X main.version=<version>"` where `<version>` is the tag with the
  `core/` prefix stripped (the release workflow does this via `${TAG#core/}`;
  the `mise run build` task carries the same injection for a local release-style
  build, defaulting to `dev` otherwise). A release binary then reports its
  version; a plain `mise run build` reports `dev`; and a `go install`-obtained
  binary (Path A) — which gets no ldflags — falls back via `resolveVersion()` to
  the module version in `runtime/debug.BuildInfo`, so it reports the version the
  user installed. The tag is the single source of
  truth: release-
  please creates it (D2), the release workflow injects it; no version literal is
  hand-maintained in source, and release-please does not write one into
  `main.go` (it tracks the version in `.release-please-manifest.json`). *Rejected:* keeping a hand-edited `const
  version` bumped per release —
  it silently drifts from the tag and defeats the point of a trustworthy
  version string; the whole reason to publish releases is that the reported
  version is authoritative. **Consequence (done):** `TestIntegrationVersion` no
  longer pins the literal `leitwerk 0.1.0`; it asserts the line *format*
  (`^leitwerk \S+$`), the default-build `dev` sentinel, and the injection
  round-trip (a binary built with `-X main.version=vTEST` reports `leitwerk
  vTEST`). This is part of acceptance.

- **D4 · Blast radius is T2, higher than the roadmap's T1 estimate — recorded,
  not silently accepted.** The roadmap lists cli-publish as T1, but the concrete
  mechanism edits `core/cmd/leitwerk/main.go` (T2) and adds
  `.github/workflows/release-please.yml` + `.github/workflows/leitwerk.yml` (T2) —
  verified with `leitwerk tier`. The
  honest tier is therefore **T2** (see Blast radius). This is a *strengthening*
  (more verification), safe to raise; the roadmap's tier was a proto-spec
  estimate and the spec's tier is set from the files actually touched. The T2
  landing sign-off is human-owned (constitution) — flagged for the landing
  review, not decided here.

- **D5 · Versions from Conventional Commits; seed `0.1.0`; tags are
  `core/`-prefixed.** release-please derives each bump from commit types
  (`fix:`→patch, `feat:`→minor, `!`/`BREAKING CHANGE`→major).
  `.release-please-manifest.json` is seeded at `0.1.0` (the value the source
  already carried) so the first Release-PR proposes the release, and release-
  please is configured — the package keyed at path `core`,
  `include-component-in-tag: true`, `tag-separator: "/"` (release-please's
  documented Go-subdir convention), `include-v-in-tag: true` — to cut the tag as
  **`core/v0.1.0`** (CONFIRMED against the release-please config schema). The
  `core/` prefix is not cosmetic: the module lives in the `core/` subdirectory
  (`core/go.mod`, no repo-root module), and Go resolves a subdirectory module's
  versions only from tags of the form `core/vX.Y.Z`. So `go install
  …/core/cmd/leitwerk@latest`/`@vX.Y.Z` requires exactly this tag form (Path A).
  Pre-1.0 signals the CLI surface may still change. The per-release human
  decision is merging the Release-PR, not a hand-cut tag. *Rejected:* a bare
  `vX.Y.Z` tag — invisible to `go install …/core` (Go would read it as a version
  of the non-existent repo-root module), silently breaking Path A. *Rejected:*
  starting at `v1.0.0` (over-claims stability before the live validations).

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
  via `git log`), so going forward commits follow Conventional Commits. Workflow
  change accepted by the human (2026-07-21). Documented in `.gitmessage` (the
  canonical types + scopes, usable as `git config commit.template .gitmessage`),
  `CONTRIBUTING.md`, and a pointer in `bindings/open/AGENTS.md`; recording it in
  the constitution's decisions of record is *proposed* to the human (that file is
  human-owned).
  - **Squash-merge → PR titles are load-bearing.** PRs are squash-merged, so the
    PR **title** becomes the commit on `main` that release-please reads. A
    `semantic-pr` workflow (`amannn/action-semantic-pull-request`) validates each
    PR title's type + Conventional-Commit format; scopes stay advisory in
    `.gitmessage`. This is a **PR-gating CI check, not part of the leitwerk gate**
    — the gate parses no commit messages (parity intact). Squash-merge itself is
    a **repo setting** the human enables (see Preconditions).
  *Rejected:* keeping ad-hoc commit messages (release-please could derive neither
  a version nor a changelog); relying on `.gitmessage` alone (opt-in, unenforced,
  and it does not touch the PR title that squash-merge actually commits).

- **D8 · mise is the single toolchain + build entry point; the Makefile is
  removed.** `mise.toml` already pins `go` and `node` for local dev; the CI
  workflows install the toolchain the same way with `jdx/mise-action@v4`, and
  build/test live in `mise.toml` `[tasks]` (`mise run build/test/vet/install/
  clean`) — the `build` task carries the ldflags version injection (D3). The
  `core/Makefile` is **removed**: one entry point for both toolchain and build,
  local and CI. Scope: the gate job installs Go + Node; the **parity** job and
  the release build install **Go only** (`install: false` + `mise install go`) —
  parity must not gain Node or it would stop proving the gate runs without it.
  Removing the Makefile does not weaken the adopter story — adopters obtain the
  **prebuilt binary** (cli-publish's whole point), not a source build, so the
  build tooling only affects leitwerk contributors and CI, who already run mise.
  Matches the cplace-ops-cloud reference (build/test/lint are `mise run` tasks,
  no Makefile). Note: `mise.toml` is currently **T1** whereas the Makefile was
  T2; the version-injection *contract* stays gated at T2 by
  `TestIntegrationVersion`, and CI runs `mise run build` explicitly, so the drop
  is covered — but see the landing escalation on whether to raise `mise.toml` to
  T2 in `tiers.conf`. Human decision, 2026-07-23. *Rejected:* keeping the
  Makefile (two build entry points that can drift; the earlier worry that mise
  tasks "force mise on adopters" is moot — adopters download the binary); keeping
  `actions/setup-go` (two toolchain sources — mise locally, setup-go in CI — that
  can drift).

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
- `mise.toml` — T1 (`[tasks]` build wrapper incl. the ldflags version injection;
  see D8 on the T1-vs-T2 tier).
- `.github/workflows/release-please.yml` — T2 (new release automation: the
  `release-please` job + the dependent `build-and-upload` job).
- `.github/workflows/leitwerk.yml` — T2 (existing gate CI switched to mise for the
  toolchain, D8).
- `.github/workflows/semantic-pr.yml` — T2 (new PR-title Conventional-Commit
  check; D7).
- `release-please-config.json`, `.release-please-manifest.json` — T1 (repo-root
  config; the `.json` pair is parsed by the existing `json` check).
- `README.md`, `core/README.md`, `docs/adoption.md`, `bindings/open/AGENTS.md`,
  `CONTRIBUTING.md` (new), `.gitmessage` (new) — T0/T1 (doc caveats removed;
  commit convention documented in `.gitmessage` + `CONTRIBUTING.md`, D7).

Worst case if it ships wrong: a published binary reports a wrong or empty
version, or the release matrix produces a broken/non-static binary for some
platform, so an adopter obtains a gate that misidentifies itself or does not run
— an adoption/trust failure, not a weakening of the gate's verification (the
checks are unchanged). Mitigated by: the version round-trip test (D3), a
static-build assertion, the four-target matrix building green in CI, and the T2
review panel. Publishing the repo public is irreversible — already done (D6);
the recurring guard is the Release-PR gate (no release cuts without a human
merge).

## Acceptance checks
Gate-verifiable (run in `leitwerk verify` for this area):
- **Version format + injection:** `TestIntegrationVersion` asserts `leitwerk
  version` prints one line matching `^leitwerk \S+$`, exit 0; and a binary built
  with `-ldflags "-X main.version=vTEST"` reports `leitwerk vTEST`. No pinned
  release literal remains.
- **Local build + static:** `mise run build` succeeds and the default build
  reports `leitwerk dev`; the release build flags keep `CGO_ENABLED=0` (static).
- **Release build produces the matrix:** locally, a
  `CGO_ENABLED=0 GOOS=… GOARCH=…` build for each of `{linux,darwin}×{amd64,arm64}`
  succeeds and yields a static binary with the documented asset name (the stable
  contract in §Behaviour Path B); `sha256` over the four produces `checksums.txt`.
  The same matrix is what the `build-and-upload` job runs in CI. The release-please
  config + manifest parse (the existing `json` check covers the `.json` pair); the
  tag it is configured to cut is `core/v0.1.0`.
- **The devkit's own gate stays green:** `leitwerk verify --tier T2`.

Live proof (one-time, recorded like a plugin-live-validation / ci-live validation
— not a per-change gate check, because it needs the network and a real GitHub
release):
- The first Release-PR opens proposing `v0.1.0` with a changelog; merging it
  creates the `core/v0.1.0` tag and GitHub Release; the same workflow run
  continues to the `build-and-upload` job, which appends the four binaries +
  `checksums.txt` to that release.
- On a clean machine with a Go toolchain: `go install
  github.com/cf-sewe/leitwerk-devkit/core/cmd/leitwerk@latest` resolves the
  `core/v0.1.0` tag, then `leitwerk verify --tier T1` runs in an **unrelated**
  repo (embed self-sufficiency), and `leitwerk version` prints `leitwerk v0.1.0`.
- On a clean machine with **no** Go toolchain: download the release binary for
  the platform, verify its checksum, and run `leitwerk verify` in an unrelated
  repo.
- The plugin launcher, given only a globally installed `leitwerk`, resolves the
  core CLI without a sibling `core/` (the marketplace-only case).

## Anchors
Only files that already exist are anchored (drift resolves anchors eagerly and a
not-yet-created path would go red on a draft). The release config/workflow
(`.github/workflows/release-please.yml`, `release-please-config.json`,
`.release-please-manifest.json`) become anchors when the build creates them —
bidirectional refinement, added in the same change that lands the files.
- `core/cmd/leitwerk/main.go`
- `mise.toml`

## Out of scope
- **plugin-live-validation** and **ci-live** — separate roadmap
  items; this spec unblocks them (public repo) but does not perform them.
- **Windows builds.** Only `{linux, darwin} × {amd64, arm64}` ship (the arches
  adopters and CI actually run); add Windows only if adoption demands it.
- Additional distribution channels (Homebrew tap, apt/deb, container image) and
  GoReleaser to produce them — add later only if adoption demands them.
- Binary signing / provenance (cosign, macOS notarization, SLSA attestation) —
  a possible later hardening; not required to obtain a working gate.
- Any change to the module path, the checks, the tier ladder, or tier→path
  policy.
