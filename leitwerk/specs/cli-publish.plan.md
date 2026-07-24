# Plan — cli-publish (build & release the gate binary)

Status: active (2026-07-23) <!-- per-change and perishable: when every step has landed, mark it landed and move it to leitwerk/specs/archive/ -->

Turns `leitwerk/specs/cli-publish.md` into gated steps. Strategy: land the
mechanism inward-out so the gate stays green and `main` stays shippable at every
step — first the **version seam** in the binary (it reports `dev` honestly until
a tag exists), then the **release config** (release-please), then the **CI
release workflow** (`release-please.yml`, two jobs), then the
**docs/convention**, then **review & land**. No release is cut by landing this;
the first release is the post-merge live proof (spec §Acceptance).

No GoReleaser: for one static binary from one `main` package the build+attach job
is a short workflow (spec D2). The release tag is **`core/v0.1.0`** — the module
lives in the `core/` subdirectory, and Go resolves a subdirectory module's
versions only from `core/`-prefixed tags (spec D5). Targets are
`{linux,darwin}×{amd64,arm64}`; Windows is out of scope (spec §Out of scope).

The reported version is a real behaviour change on `main`: `leitwerk version`
goes from `leitwerk 0.1.0` to `leitwerk dev` the moment step 1 lands, and stays
`dev` until the first tag is cut by merging the first Release-PR — then a release
binary reports `leitwerk v0.1.0` (the `core/` prefix is stripped at build,
`${TAG#core/}`). That is the intended contract (spec D3): a non-release build
must not claim a release tag.

## Steps
Step status: `[ ]` open · `[x]` done · `[~]` deviated — one line why. Kept current
by whoever executes the steps; a cold session resumes from these lines.

0. `[x]` **De-risk the load-bearing assumption: release-please can cut a
   `core/v*` tag.** — **CONFIRMED 2026-07-23.** The release-please config schema
   defines `tag-separator` ("Customize the separator between the component and
   version in the GitHub tag", type string, **no character restriction — `/` is
   valid**), `include-component-in-tag` (default true), and `include-v-in-tag`
   (default true). release-please explicitly supports the Go-subdir case: "Go
   uses `/` for a tag separator, rather than `-`, … important when releasing Go
   modules in subdirectories" (googleapis docs). So the package keyed at path
   `core` + `include-component-in-tag: true` + `tag-separator: "/"` yields the
   tag `core/v0.1.0`. This is the linchpin: `go install …/core@latest` resolves
   only a `core/vX.Y.Z` tag, and the GitHub Release must be named by that tag —
   both now rest on a confirmed capability. Final proof is the first Release-PR
   (step 6), inspected before merge.

1. `[x]` **Version seam — oracle first** — `core/cmd/leitwerk/integration_test.go`,
   `core/cmd/leitwerk/main.go` (**T2**), build task in `mise.toml` (**T1**).
   - *Oracle first* (tier-discipline rule): rewrite `TestIntegrationVersion`
     (`integration_test.go:252-257`, currently pins the literal `leitwerk 0.1.0`)
     to assert (a) `leitwerk version` prints one line matching `^leitwerk \S+$`,
     exit 0, and (b) a binary built with `-ldflags "-X main.version=vTEST"`
     reports exactly `leitwerk vTEST` — the injection round-trip. No pinned
     release literal remains. Confirm it fails against the current `const`.
   - `main.go:22`: `const version = "0.1.0"` → `var version = "dev"` (a `const`
     cannot be set by `-ldflags -X`; a `var` can). **`version` is also read in
     `cacheDir()` (`main.go:359`, `filepath.Join(base, "leitwerk", version)`)** —
     confirm the cache path still resolves (`.../leitwerk/dev` locally,
     `.../leitwerk/v0.1.0` in a release; version-namespaced, no regression).
     CONFIRMED: the only readers of `version` are `main.go:90` (banner) and
     `main.go:359` (cache namespace).
   - `mise.toml` `[tasks]`: the `build`/`install` tasks inject
     `-ldflags "-X main.version=${VERSION:-dev}"`. A plain `mise run build`
     reports `leitwerk dev`; `VERSION=v0.1.0 mise run build` reports
     `leitwerk v0.1.0`. Keep `CGO_ENABLED=0` (static). **Deviation (D8):** a
     Makefile was used here first, then replaced by mise tasks and removed — mise
     is the single build + toolchain entry point.
   - *Proves it:* `go test ./...` (→ `selftest §0`) with the migrated test;
     `mise run build && core/bin/leitwerk version` → `leitwerk dev`.
     `leitwerk verify --tier T2` green.
   - **Manual (T2):** eyeball that no code path treats `version` as a semver
     (only banner + cache namespace); run the built binary once and confirm
     `verify`/`init` still work with the `dev` cache dir.

2. `[x]` **release-please config** — `release-please-config.json`,
   `.release-please-manifest.json` (both **T1**, parsed by the `json` check).
   Landed; `json` check parses 6 manifests (was 4). Config keys verified against
   the release-please config schema and cross-checked with the cplace-ops-cloud
   reference (its Go packages `apps/lib`, `apps/cplace-controller`):
   - `.release-please-manifest.json`: `{ "core": "0.1.0" }` — the package keyed at
     the module dir `core`, seeded at the value the source already carried.
   - `release-please-config.json`: `release-type: go` (idiomatic for a Go module;
     Go has no version file, so nothing in source is rewritten, D3);
     `include-component-in-tag: true` + `tag-separator: "/"` (+ default
     `include-v-in-tag: true`) → tag **`core/v0.1.0`**; explicit `component: "core"`
     and `package-name: "core"`; `bump-minor-pre-major: true` so a pre-1.0 breaking
     change stays in `0.x` (D5). One package keyed at `core`.
   - *Decisions recorded:* (a) **`component`/`package-name` set explicitly to
     `core`** (== the module subdir), matching the reference; the tag prefix must
     equal the subdir for `go install` to resolve. Keying at `core` is also
     semantically right: the CLI version bumps on `core/**` changes, the actual
     released artifact. (b) **changelog** at
     the default `core/CHANGELOG.md` (release-please resolves `changelog-path`
     relative to the package; the CLI changelog next to the CLI, no `../` hack).
     (c) **first-release version** (that the very first tag is `v0.1.0`, not a bump
     past it) is release-please *runtime* behaviour — not gate-checkable; validated
     at the first Release-PR (step 6), with the manifest seed / `initial-version`
     as the lever if it proposes otherwise.
   - *Proved it:* `json` parses both (`leitwerk verify --tier T1` green).
   - *Proves it:* the `json` check parses both files (→ gate); `leitwerk verify
     --tier T1` green. (release-please itself only runs in CI — its tag output is
     validated in step 0 + the post-merge live proof, not the gate.)

3. `[x]` **CI release workflow** — `.github/workflows/release-please.yml` (**T2**).
   **Deviation from plan (recorded):** one workflow with two jobs, not two
   workflows. A tag pushed by the default `GITHUB_TOKEN` does not trigger another
   workflow (GitHub's recursion guard), so a separate tag-triggered `release.yml`
   would silently never build binaries without a long-lived PAT. Chaining the
   build as a dependent job in the same run avoids both the missed trigger and the
   secret (spec D2 updated).
   - Job `release-please` (on `push` to `main`): `googleapis/release-please-
     action@v5` with `config-file`/`manifest-file`; permissions
     `contents: write`, `pull-requests: write`; emits `core--release_created` /
     `core--tag_name` (per-path outputs, manifest mode) as job outputs. It only
     opens/updates a Release-PR until a human merges it (D2/D6).
   - Job `build-and-upload` (`needs: release-please`,
     `if: …release_created == 'true'`): `actions/checkout@v4` at the tag ref;
     `actions/setup-go@v5` (`go-version-file: core/go.mod`); build loop over the
     four `{linux,darwin}×{amd64,arm64}` targets —
     `CGO_ENABLED=0 GOOS GOARCH go build -C core -trimpath
     -ldflags "-s -w -X main.version=${TAG#core/}"`; `sha256sum` → `checksums.txt`;
     `gh release upload "$TAG" dist/leitwerk_* dist/checksums.txt --clobber`. Asset
     names `leitwerk_<os>_<arch>` + `checksums.txt` are the stable documented
     contract (spec §Behaviour Path B). `permissions: contents: write`,
     `GH_TOKEN: github.token`.
   - *Proved it:* the four-target matrix built locally with the exact command
     (Linux binaries `statically linked, stripped`; the host binary round-trips
     `leitwerk v0.1.0-test`); `actionlint` clean on both workflow files;
     `leitwerk verify --tier T2` green (no gate check parses workflow yaml, so
     green is necessary-not-sufficient — see Verification strategy).
   - **Manual (T2) — reviewed:** the `if: core--release_created` gate + `needs`
     wiring; `${TAG#core/}` yields the bare `vX.Y.Z`; least-privilege permissions
     per job; `CGO_ENABLED=0` on every target; asset names/checksums match the
     contract; no secret beyond `GITHUB_TOKEN`.

4. `[x]` **Docs + commit convention** — `README.md`, `core/README.md`,
   `docs/adoption.md`, `bindings/open/AGENTS.md`, new `CONTRIBUTING.md`,
   `.gitmessage`, `.github/workflows/semantic-pr.yml` (T2).
   - Removed the two "provisional until published" caveats that existed
     (`README.md`, `core/README.md`; `docs/adoption.md` had none). The three
     obtain paths (A `go install …@latest`, B release-binary download + checksum,
     C `LEITWERK_HOME`/checkout) are documented as working; README gained path C.
     Added the `runtime/debug` build-info fallback so a `go install` binary
     (Path A) reports its version (D3), and converted stale `M2.x` milestone refs
     to slugs.
   - `.gitmessage` is the canonical commit template (types + scopes; scopes cut to
     the three top-level areas `core`/`bindings`/`governance`). `CONTRIBUTING.md`
     carries the why (squash-merge → PR title → release-please), build/test, and
     release flow, pointing at `.gitmessage`. `bindings/open/AGENTS.md` gets a
     pointer (and its stale "roadmap human-owned" line fixed). Enforcement:
     `semantic-pr.yml` (`amannn/action-semantic-pull-request@v6`) validates the
     PR-title format (D7); scopes stay advisory.
   - *Proposed to the human* (human-owned file): the Conventional-Commits
     decision-of-record entry —
     `leitwerk/proposals/20260723_220314-conventional-commits-decision.md`.
   - *Proved it:* `leitwerk verify --tier T2` green; `actionlint` clean; no
     dangling "provisional"/`make` in live docs (`grep` clean).

5. `[ ]` **Review & land** — **T2**.
   - Update the spec's `## Anchors` to add the now-existing release files
     (`.github/workflows/release-please.yml`, `.github/workflows/semantic-pr.yml`,
     `release-please-config.json`, `.release-please-manifest.json`) —
     bidirectional refinement, same change that created them (spec §Anchors
     anticipates this). Confirm `drift` resolves every anchor.
   - `leitwerk verify --tier T2` green; spec-fidelity read; adversarial panel via
     the documented read-only fallback (roles below).
   - Landing ritual: spec `active → landed (YYYY-MM-DD)`; archive spec + this plan
     to `leitwerk/specs/archive/`; remove the `cli-publish` item from
     `roadmap.md` (it lands, so it leaves — agent edit, human confirms per the
     roadmap's ownership note); commit gate-green with a Conventional-Commits
     message (`feat: …`, the first one to dogfood D7).
   - **Human sign-offs (escalations):** the **T2 landing** sign-off; the
     **constitution decision-of-record** proposal from step 4; **whether to raise
     `mise.toml` to T2** in `tiers.conf` (human-owned) now that it carries the
     build task (D8) — vs leaving it T1 (the version contract is gated at T2 by
     `TestIntegrationVersion` and CI runs `mise run build` explicitly). Present as
     `leitwerk-review` multiple-choice.

6. `[ ]` **Live proof (post-merge, one-time — not a gate check)** — recorded like
   a plugin-live-validation / ci-live validation because it needs the network and
   a real GitHub release. The first Release-PR opens proposing `v0.1.0`; merging
   it creates the `core/v0.1.0` tag + Release; the same run's `build-and-upload`
   job appends the four binaries + `checksums.txt`. Then on a clean machine: (a) with
   Go — `go install github.com/cf-sewe/leitwerk-devkit/core/cmd/leitwerk@latest`
   resolves `core/v0.1.0`, run `leitwerk verify --tier T1` in an unrelated repo,
   `leitwerk version` prints `leitwerk v0.1.0`; (b) with no Go — download the
   release binary, verify its checksum, run `verify`; (c) the plugin launcher,
   given only a global `leitwerk`, resolves the CLI with no sibling `core/`.
   Record the transcript under `docs/reviews/`.

## Verification strategy
- **New/extended oracles:** `TestIntegrationVersion` becomes the version contract
  — format assertion + a build-with-`-ldflags` round-trip (step 1), run under
  `go test` → `selftest §0` → the gate. Verify it goes red on a mutation (revert
  `var`→`const`, or drop the ldflags from the mise `build` task) during build.
- **Gate-covered config:** the `json` check parses the release-please
  config+manifest pair automatically (step 2). The default `mise run build`
  reporting `dev` is asserted by the round-trip test's format arm plus a manual
  banner check.
- **Local cross-build check (not a gate check) — DONE:** built all four
  `{linux,darwin}×{amd64,arm64}` targets locally with the exact `build-and-upload`
  command (no GoReleaser); Linux binaries report `statically linked, stripped`,
  darwin are pure-Go Mach-O (no dynamic deps under `CGO_ENABLED=0`), and the
  host-native binary round-trips `leitwerk v0.1.0-test`.
- **Workflow lint (not a gate check) — DONE:** `actionlint` clean on
  `release-please.yml` and the existing `leitwerk.yml` (validates the `needs`/
  outputs refs and the shell run steps).
- **Deliberately *not* wired into the gate (recorded):** workflow-yaml linting and
  a release-please dry-run run **locally / inside CI**, not as per-change gate
  checks. Reason: adding them would put optional analyzers into the minimal
  open-code **parity** CI job (which ships none) and would add checks to a change
  whose whole point is to add *no* gate logic (parity invariant, spec
  §Invariants). The `build-and-upload` job building the matrix green in CI is the
  binding validation of the build path; the live proof validates the tag/release.

## Risks & rollback
- **release-please emits the wrong tag string (the linchpin).** If it cuts
  `core-v0.1.0` (dash) or `v0.1.0` (no component), `go install …/core@latest`
  can't resolve it — Path A silently broken. (The build job no longer depends on
  a tag glob — it is gated on the action's `release_created` output — so Path B is
  unaffected, but a wrong tag misnames the release and breaks `go install`.)
  Mitigation: step 0 confirmed `tag-separator: "/"` + the path-derived component
  `core` yields `core/v0.1.0`; the first Release-PR (live proof) is inspected
  before merge. *Rollback:* adjust the release-please config and re-tag; nothing
  is published until a tag exists.
- **Version misreports (the core trust risk).** A release binary that reports
  `dev`, or an empty/`0.1.0`-stale string, is an adoption/trust failure.
  Mitigation: the round-trip test pins the injection; `${TAG#core/}` is called
  out so the reported value is the bare `vX.Y.Z`; the mise `build` task's default
  `dev` is asserted. *Rollback:* revert step 1 — the binary returns to a hard-coded
  literal; no release infra depends on step 1 until step 2+ land.
- **Broken/non-static binary for some target.** A CGO leak or wrong GOOS/GOARCH
  ships a binary that won't run. Mitigation: `CGO_ENABLED=0` on every target + the
  local four-target cross-build with a static-linkage spot-check before the
  workflow lands (Verification strategy — DONE). *Rollback:* the `build-and-upload`
  job only runs when a release was created; delete/yank a bad release and revert
  the workflow — no adopter is affected until a tag exists.
- **A workflow releases something unintended.** Mitigation by construction:
  release-please only opens a PR; the `build-and-upload` job only runs when a
  release was created (`if: …release_created`); a release is only created by a
  human merging the Release-PR (D2/D6). Landing steps 2–3 opens a Release-PR but
  cuts nothing. *Rollback:* close the Release-PR; revert the workflow file.
- **Over-broad CI permissions / supply-chain.** A workflow with more scope than
  needed widens blast radius. Mitigation: least-privilege `permissions:` per job,
  pinned action majors, T2 review with the security lens (step 5). *Rollback:*
  revert the workflow file (T2 path, isolated).
- **First-tag surprise.** The first Release-PR appears as soon as
  `release-please.yml` hits `main`; that is expected (step 6), not a failure.

## Roles to wake
- `architect` — the release topology (release-please owns version/tag/changelog;
  the matrix workflow builds+attaches; D2), the subdir tag `core/v…` and its
  coupling to `go install`, and that the version seam adds no coupling to the gate.
- `test-engineer` — the version round-trip oracle and its mutation-sensitivity
  (does it actually go red when injection is removed?).
- `security-reviewer` — CI workflow permissions, action pinning, tag-trigger
  surface, and the supply-chain shape of a published binary (checksums; signing
  is out of scope per the spec but note it).
- Human — T2 landing sign-off; the constitution decision-of-record (D7).
