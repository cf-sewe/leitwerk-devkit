# leitwerk/checks/ — this repo's own gate checks

Repo-local checks the gate runs when it verifies **this** repository. Each is a
shell script `<check>.sh` with the standard contract: `exit 0` pass, `exit 1`
fail (gate red), `exit 2` skip (nothing to run — never a fake pass).

## Resolution — why these override the built-ins

The CLI resolves a check by name in order (first match wins;
`core/internal/gate/resolve.go`, mechanism in `core/README.md`):

**repo-local `leitwerk/checks/` → built-in on disk (`core/checks/`) → embedded in the binary.**

So a script here overrides the generic built-in of the same name. `LEITWERK_CHECKS`
points the local dir elsewhere.

## This repo is not a typical app — its check set differs from the default

The generic default an adopter gets (`core/leitwerk.tiers`) is an application
check set: `lint types tests drift sast erosion`. This repo builds a *framework*
(shell + Go + plugin manifests + governance docs), so it **replaces** that set
with checks that prove the framework's own integrity — keeping only **`drift`**
from `core/`. The authoritative tier→check map is `leitwerk/tiers.conf`; this
file only narrates it (do not treat this README as the source of truth).

## Tier → checks (cumulative — T2 runs T0+T1+T2)

| Tier | Checks |
|---|---|
| **T0** | `json` · `context` · `lifecycle` |
| **T1** | + `shell` · `drift` · `parity` |
| **T2** | + `selftest` |

## The checks

| Check | File | What it proves |
|---|---|---|
| `json` | `json.sh` | every JSON manifest (plugin, marketplace) parses — a broken one otherwise fails silently at install time |
| `context` | `context.sh` | the always-on context stays within budget (`CLAUDE.md` ≤ 200 lines, each rule ≤ 100, each skill/agent description ≤ 80 words, ≤ ~2000 est. tokens total) |
| `lifecycle` | `lifecycle.sh` | spec/plan/proposal lifecycle states are valid — a `landed`/`superseded` record lives in `archive/`, a plan doesn't outlive its spec, open proposals stay visible and overdue ones are flagged |
| `shell` | `shell.sh` | every shell script is `bash -n`- and shellcheck-clean — the scripts *are* the gate, so they get the strictest check |
| `drift` | `core/checks/drift.sh` *(built-in)* | spec↔code divergence: a spec's `## Anchors` resolve, and (with a diff base) anchored code isn't changed one-sidedly. Surfaces; never resolves |
| `parity` | `parity.sh` | the hard guarantee stays in `core/` — no gate logic leaks into a binding (open-code guarantee-parity) |
| `selftest` | `selftest.sh` | the CLI's golden behaviour holds — the executable oracle (glob/tier tables, gate output, scenarios, lifecycle, drift, bash portability, workflow syntax) |

## Adding or changing a check

An agent may **add** a check: drop `<name>.sh` here (following the exit contract)
and **propose** its wiring in `leitwerk/tiers.conf`. That file and
`leitwerk/constitution.md` are human-owned — lowering a threshold, removing a
check, or downgrading a path's tier is a human decision, not an agent edit
(enforced by `leitwerk guard` + the `PreToolUse` hook). See the constitution's
blast-radius policy and Definition of Done.
