# core — the tool-agnostic gate

No agent runtime depends on anything in here. This is the part that CI runs and
that every binding calls. The gate is a single static **Go binary**; the checks it
orchestrates stay as shell scripts that shell out to real toolchains.

## `bin/leitwerk`

```
leitwerk verify [--tier T0|T1|T2]   run the checks selected for a blast-radius tier
leitwerk tier <path>                print the tier for a changed path
leitwerk guard <path>               exit 3 if a path is human-owned (else 0)
leitwerk drift                      surface spec<->code divergence (does not resolve)
leitwerk init [dir]                 scaffold leitwerk/{constitution.md,tiers.conf} + steering
leitwerk version
```

Exit codes: `0` gate green / path editable · `1` a check failed (gate red) · `2`
usage error · `3` path is human-owned (guard). A check's own `exit 2` means
*skip* (nothing to run) and never fails the gate. (Hooks that must *block* wrap
this as `leitwerk verify || exit 2`, because a Claude Code hook blocks only on
exit code 2.)

`bin/leitwerk` is a build artifact and is not committed — build it with `make
build` (below).

## Source layout

```
cmd/leitwerk/main.go     CLI entrypoint: subcommand dispatch, path/env resolution
internal/gate/           the gate library, unit-tested and dependency-free:
  glob.go                  glob→regex (the exact **/ , ** , * , catch-all rules)
  tiers.go                 parse the tiers file; tier-for-path; checks-for-tier; guard match
  resolve.go               check resolution (repo-local → built-in on disk → embedded)
  verify.go                run the checks and format the gate output
assets.go                //go:embed checks templates leitwerk.tiers
```

The binary depends only on the Go standard library (no third-party modules), so it
builds offline and reproducibly.

## Building

The toolchain is pinned with [mise](https://mise.jdx.dev) (`mise.toml` at the repo
root: Go + Node LTS), which also defines the build tasks. From the repo root
(after `mise trust` on a fresh clone):

```
mise run build      # -> core/bin/leitwerk (static; CGO disabled)
mise run test       # unit + integration tests
mise run install    # `go install` into GOBIN
mise run clean
```

The tasks run under mise, so `go` is always the pinned toolchain (`mise.toml`);
CI installs it with `jdx/mise-action`. A fresh clone needs `mise trust` once.

## Distribution

The checks, templates, and the default `leitwerk.tiers` are **embedded** in the
binary (`assets.go`), so a copy of `bin/leitwerk` alone is self-contained: it can
`init` a repo and run the full gate from its embedded assets, independent of the
repo layout. On-disk siblings (`checks/`, `templates/`, `leitwerk.tiers`) win when
present, so a full checkout or a release tarball uses the on-disk copies.

Ways to make the gate resolvable for adopters and CI:

- **`go install`** — `go install github.com/cf-sewe/leitwerk-devkit/core/cmd/leitwerk@latest`
  installs a single `leitwerk` binary into `GOBIN` (needs a Go toolchain; no
  registry account required). It resolves the `core/vX.Y.Z` release tags, and the
  installed binary reports its version from the module build info.
- **Prebuilt binary** — download the static binary for your platform from a release
  and put it on `PATH`. Cross-compiles from source are `GOOS=… GOARCH=… go build`
  (the release workflow builds the full matrix).
- **`LEITWERK_HOME`** — for a full checkout, set `LEITWERK_HOME=/path/to/…/core`
  and put `$LEITWERK_HOME/bin` on `PATH`. The Claude plugin launcher resolves the
  core CLI this way (a marketplace install copies only the plugin, not `core/`).
- **Vendoring** — copy the built `bin/leitwerk` into the consuming repo (it carries
  its own checks/templates).

## `checks/`

One script per check. Each:
- `exit 0` — passed,
- `exit 1` — failed (gate goes red),
- `exit 2` — nothing to run here (skips cleanly; never a fake pass).

These are generic templates that auto-detect a toolchain. A consuming repo does
**not** edit them — it drops its own `<name>.sh` into its `leitwerk/checks/`,
which overrides the built-in per check (anything not overridden falls back here,
or to the embedded copy). Set `LEITWERK_CHECKS` to point elsewhere. See this
repo's own `leitwerk/checks/` for a worked example.

## `leitwerk.tiers`

Two tables plus a list: `[tiers]` maps each tier to its checks (cumulative — T2
runs T0+T1+T2 checks); `[paths]` maps path globs to tiers (first match wins);
`[human-owned]` lists the files an agent may propose but not edit. This file is
human-owned: an agent may propose a check but may not lower a threshold or remove
one here. `LEITWERK_TIERS` overrides which file is read.

## `templates/`

Starting scaffolds, kept single-source here and embedded into the binary. `leitwerk
init` instantiates the ones a repo owns: `constitution.template.md` →
`leitwerk/constitution.md`, `CLAUDE.template.md` → `CLAUDE.md`,
`rules/tier-discipline.md` → `.claude/rules/`, and `workflows/leitwerk-review.mjs`
→ `.claude/workflows/` (the review workflow, which a plugin cannot package). The
`spec.template.md` and `plan.template.md` are referenced in place by the phase
skills each time a spec or plan is written.

`workflows/leitwerk-review.mjs` is advisory orchestration, so per-repo copies may
be tailored; the repo keeps its own copy identical to this template, enforced by
`selftest`.
