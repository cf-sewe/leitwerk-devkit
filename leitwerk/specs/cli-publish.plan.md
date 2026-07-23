# Plan — cli-publish (build & release the gate binary)

Status: active (2026-07-23) <!-- per-change and perishable: when every step has landed, mark it landed and move it to leitwerk/specs/archive/ -->

Turns `leitwerk/specs/cli-publish.md` (M2.1) into gated steps. Strategy: land the
mechanism inward-out so the gate stays green and `main` stays shippable at every
step — first the **version seam** in the binary (it reports `dev` honestly until
a tag exists), then the **build config** (GoReleaser), then the **release
config** (release-please), then the **CI workflows** that wire them, then the
**docs/convention**, then **review & land**. No release is cut by landing this;
the first release is the post-merge live proof (spec §Acceptance).

The reported version is a real behaviour change on `main`: `leitwerk version`
goes from `leitwerk 0.1.0` to `leitwerk dev` the moment step 1 lands, and stays
`dev` until the first `vX.Y.Z` tag is cut by merging the first Release-PR. That
is the intended contract (spec D3): a non-release build must not claim a release
tag.

## Steps
Step status: `[ ]` open · `[x]` done · `[~]` deviated — one line why. Kept current
by whoever executes the steps; a cold session resumes from these lines.

1. `[ ]` **Version seam — oracle first** — `core/cmd/leitwerk/integration_test.go`,
   `core/cmd/leitwerk/main.go`, `core/Makefile` (all **T2**).
   - *Oracle first* (tier-discipline rule): rewrite `TestIntegrationVersion`
     (`integration_test.go:252-257`, currently pins the literal `leitwerk 0.1.0`)
     to assert (a) `leitwerk version` prints one line matching `^leitwerk \S+$`,
     exit 0, and (b) a binary built with `-ldflags "-X main.version=vTEST"`
     reports exactly `leitwerk vTEST` — the injection round-trip. No pinned
     release literal remains. Confirm it fails against the current `const`.
   - `main.go:22`: `const version = "0.1.0"` → `var version = "dev"` (a `const`
     cannot be set by `-ldflags -X`; a `var` can). **`version` is also used in
     `cacheDir()` (`main.go:359`, `filepath.Join(base, "leitwerk", version)`)** —
     confirm the cache path still resolves (`.../leitwerk/dev` locally,
     `.../leitwerk/v0.1.0` in a release; version-namespaced, no regression). No
     other reader of `version` beyond `main.go:90` and `:359`.
   - `Makefile`: add `VERSION ?= dev` and
     `-ldflags "-X main.version=$(VERSION)"` to `build` and `install`. A plain
     `make -C core build` reports `leitwerk dev`; `make build VERSION=$(git
     describe --tags)` reports the tag. Keep `CGO_ENABLED=0` (static, unchanged).
   - *Proves it:* `go test ./...` (→ `selftest §0`) with the migrated test;
     `make -C core build && core/bin/leitwerk version` → `leitwerk dev`.
     `leitwerk verify --tier T2` green.
   - **Manual (T2):** eyeball that no code path treats `version` as a semver
     (only banner + cache namespace); run the built binary once and confirm
     `verify`/`init` still work with the `dev` cache dir.

2. `[ ]` **GoReleaser build config + pin the tool** — `.goreleaser.yaml` (**T1**),
   `mise.toml` (**T0**).
   - `.goreleaser.yaml` (GoReleaser v2): one `builds` entry for the module in the
     subdir — `dir: core`, `main: ./cmd/leitwerk`, `binary: leitwerk`,
     `env: [CGO_ENABLED=0]`, `goos: [linux, darwin, windows]`,
     `goarch: [amd64, arm64]`. **ldflags inject `{{.Tag}}` not `{{.Version}}`** —
     GoReleaser strips the leading `v` from `.Version`, and the contract (D3) is
     that a release reports its full tag `vX.Y.Z`; so
     `-ldflags "-s -w -X main.version={{.Tag}}"`. Archives with a **stable
     `name_template`** + a `checksums.txt` — a documented contract a future
     plugin bootstrap relies on (spec §Behaviour Path B / §Acceptance);
     `changelog.disable: true` and `release.mode: append` (release-please owns
     the notes and creates the release — GoReleaser only appends artifacts, spec
     D2). Resolve the monorepo `gomod` pipe empirically (module lives in `core/`,
     not root) — set `gomod.dir: core` or leave the pipe unused; the `builds.dir`
     is what matters for the compile.
   - `mise.toml`: add `goreleaser = "2.17.0"` so `goreleaser check` /
     `build --snapshot` are reproducible locally (matches the repo's pinned-
     toolchain rule; CI uses the goreleaser-action, not mise).
   - *Proves it (local acceptance, not a gate check):*
     `mise exec -- goreleaser check` passes; `mise exec -- goreleaser build
     --snapshot --clean` produces the full `{linux,darwin,windows}×{amd64,arm64}`
     matrix; spot-check one snapshot binary reports its snapshot version and is
     static (`file` says "statically linked" / no dynamic deps). `leitwerk verify
     --tier T1` green (the yaml itself adds no gate check — see the deviation note
     in Verification strategy on why goreleaser validation is not wired into the
     gate).

3. `[ ]` **release-please config** — `release-please-config.json`,
   `.release-please-manifest.json` (both **T1**, parsed by the `json` check).
   - `.release-please-manifest.json`: `{ ".": "0.1.0" }` — seed at the value the
     source already carried, so the first Release-PR proposes `v0.1.0` (spec D5).
   - `release-please-config.json`: `release-type: simple` (not `go`) — the version
     lives only in the manifest + tag, never in a source file (D3), and `simple`
     touches no language files; it manages `CHANGELOG.md` and cuts the tag.
     Single package at the repo root → `include-component-in-tag: false` so the
     tag is `v0.1.0`, not `pkg-v0.1.0`; keep the default `v` prefix (Go-module
     convention required by `go install …@vX.Y.Z`).
   - *Proves it:* the `json` check parses both files (→ gate); `leitwerk verify
     --tier T1` green. (release-please itself only runs in CI — its behaviour is
     validated in the post-merge live proof, not the gate.)

4. `[ ]` **CI workflows** — `.github/workflows/release-please.yml`,
   `.github/workflows/release.yml` (both **T2**).
   - `release-please.yml` (on `push` to `main`): `googleapis/release-please-
     action@v4` with `config-file`/`manifest-file`; permissions
     `contents: write`, `pull-requests: write`. It only ever opens/updates a
     Release-PR — merging that PR is the per-release human gate (D2/D6); landing
     the workflow does **not** release anything.
   - `release.yml` (on `push` tag `v*` — the tag release-please creates on merge):
     checkout with `fetch-depth: 0` (GoReleaser needs tags/history),
     `actions/setup-go@v5` `go-version-file: core/go.mod`,
     `goreleaser/goreleaser-action@v6` `args: release --clean`; permissions
     `contents: write`; `GITHUB_TOKEN` in env. Uploads to the release
     release-please already created (`release.mode: append`, step 2).
   - *Proves it:* `leitwerk verify --tier T2` green (no gate check parses workflow
     yaml, so green is necessary-not-sufficient here).
   - **Manual (T2) — the load-bearing eyeball for this step:** the tag glob
     (`v*`) matches release-please's output; least-privilege permissions on both
     jobs; the tag trigger → GoReleaser wiring; `fetch-depth: 0`; no secret beyond
     `GITHUB_TOKEN`. Optionally run `actionlint` by hand (not a gate check).

5. `[ ]` **Docs + commit convention** — `README.md`, `core/README.md`,
   `docs/adoption.md` (**T0**), `bindings/open/AGENTS.md` (**T1**), new
   `CONTRIBUTING.md` (**T0**).
   - Remove the three "provisional until published" caveats (`README.md:74`,
     `core/README.md:69`, `docs/adoption.md:10`) and document the three obtain
     paths (A `go install`, B release-binary download + checksum, C
     `LEITWERK_HOME` checkout/vendoring) as working (spec §Behaviour).
   - `CONTRIBUTING.md`: the Conventional Commits convention and the release flow
     (Release-PR → merge → tag → binaries), with the note that a non-conforming
     commit simply doesn't contribute a changelog entry/bump (D7). Add a one-line
     commit-convention pointer to `bindings/open/AGENTS.md` (open-code mirror).
   - *Propose to the human* (human-owned files — proposal only, not an edit): add
     "Conventional Commits, going forward" and "roadmap = ordered backlog of
     future specs" to `leitwerk/constitution.md` decisions of record (the latter
     is the roadmap's own long-standing proposed decision).
   - *Proves it:* `leitwerk verify --tier T1` green; the removed caveats leave no
     dangling "provisional" (`grep` clean).

6. `[ ]` **Review & land** — **T2**.
   - Update the spec's `## Anchors` to add the now-existing release files
     (`.goreleaser.yaml`, `release-please-config.json`,
     `.release-please-manifest.json`, both workflows) — bidirectional refinement,
     same change that created them (spec §Anchors anticipates this). Confirm
     `drift` resolves every anchor.
   - `leitwerk verify --tier T2` green; spec-fidelity read; adversarial panel via
     the documented read-only fallback (roles below).
   - Landing ritual: spec `draft→active` (done at plan start) `→ landed
     (2026-07-23)`; archive spec + this plan to `leitwerk/specs/archive/`; close
     **M2.1** in `roadmap.md` (staged-copy proposal — human-owned); commit
     gate-green with a Conventional-Commits message (`feat: …`, the first one to
     dogfood D7).
   - **Human sign-offs (escalations):** the **T2 landing** sign-off; the two
     **constitution decisions-of-record** proposals from step 5. Present as
     `leitwerk-review` multiple-choice.

7. `[ ]` **Live proof (post-merge, one-time — not a gate check)** — recorded like
   an M2.2/M2.3 validation because it needs the network and a real GitHub release.
   The first Release-PR opens proposing `v0.1.0`; merging it creates the tag +
   Release; the tag fires GoReleaser, which appends the binaries + checksums. Then
   on a clean machine: (a) with Go — `go install
   github.com/cf-sewe/leitwerk-devkit/core/cmd/leitwerk@latest`, run `leitwerk
   verify --tier T1` in an unrelated repo, `leitwerk version` prints the tag; (b)
   with no Go — download the release binary, verify its checksum, run `verify`;
   (c) the plugin launcher, given only a global `leitwerk`, resolves the CLI with
   no sibling `core/`. Record the transcript under `docs/reviews/`.

## Verification strategy
- **New/extended oracles:** `TestIntegrationVersion` becomes the version contract
  — format assertion + a build-with-`-ldflags` round-trip (step 1), run under
  `go test` → `selftest §0` → the gate. Verify it goes red on a mutation (revert
  `var`→`const`, or drop the ldflags from the Makefile) during build.
- **Gate-covered config:** the `json` check parses the release-please
  config+manifest pair automatically (step 3). The default `make build` reporting
  `dev` is asserted by the round-trip test's format arm plus a manual banner check.
- **Deliberately *not* wired into the gate (recorded deviation from the spec's
  optimistic "gate-verifiable" wording):** `goreleaser check` /
  `build --snapshot` and workflow-yaml linting run **locally (via pinned mise)
  and inside the release workflow**, not as per-change gate checks. Reason: adding
  them would put GoReleaser/actionlint into the minimal open-code **parity** CI
  job (which deliberately ships no optional analyzers) and would add checks to a
  change whose whole point is to add *no* gate logic (parity invariant, spec
  §Invariants). The release workflow's own `goreleaser release` is the binding
  validation of the config; the snapshot build is the pre-merge local acceptance.

## Risks & rollback
- **Version misreports (the core trust risk).** A release binary that reports
  `dev`, or an empty/`0.1.0`-stale string, is an adoption/trust failure.
  Mitigation: the round-trip test pins the injection; `{{.Tag}}` (not `.Version`)
  is called out so the reported value carries the `v`; the Makefile default `dev`
  is asserted. *Rollback:* revert step 1 — the binary returns to a hard-coded
  literal; no release infra depends on step 1 until step 2+ land.
- **Broken/non-static binary for some platform.** A bad `.goreleaser.yaml`
  (CGO leak, wrong `dir`, missing goarch) ships a binary that won't run.
  Mitigation: `goreleaser check` + a full `--snapshot` matrix build with a
  static-linkage spot-check before the config lands (step 2). *Rollback:* the
  release workflow only runs on a tag; delete/yank a bad release and revert
  `.goreleaser.yaml` — no adopter is affected until a tag exists.
- **A workflow releases something unintended.** Mitigation by construction:
  release-please only opens a PR; GoReleaser only fires on a `v*` tag; the tag is
  only created by a human merging the Release-PR (D2/D6). Landing steps 3–4 opens
  a Release-PR but cuts nothing. *Rollback:* close the Release-PR; revert the
  workflow files.
- **Over-broad CI permissions / supply-chain.** A workflow with more scope than
  needed widens blast radius. Mitigation: least-privilege `permissions:` per job,
  pinned action majors, T2 review with the security lens (step 6). *Rollback:*
  revert the workflow file (T2 path, isolated).
- **First-tag surprise.** The first Release-PR appears as soon as
  `release-please.yml` hits `main`; that is expected (step 7), not a failure.

## Roles to wake
- `architect` — the release topology (release-please ↔ GoReleaser division of
  labour, D2), the monorepo `dir: core` build, and that the version seam adds no
  coupling to the gate.
- `test-engineer` — the version round-trip oracle and its mutation-sensitivity
  (does it actually go red when injection is removed?).
- `security-reviewer` — CI workflow permissions, action pinning, tag-trigger
  surface, and the supply-chain shape of a published binary (checksums; signing
  is out of scope per the spec but note it).
- Human — T2 landing sign-off; the two constitution decisions-of-record.
