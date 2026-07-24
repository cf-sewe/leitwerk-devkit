# Adopting Leitwerk on a Swift / macOS app — validated proposal

Motivating project: a local-AI team-meeting assistant (Swift, macOS app,
on-device inference). The findings and artifacts below generalise to any
Swift/SPM or Xcode project.

Status: proposal. Every mechanism here was run against a real Swift Package
fixture on 2026-07-24 with Swift 6.3.3 / Xcode 26.6; results are tagged
**CONFIRMED** (executed here) or **INFERRED** (reasoned, not executed).

## The problem this addresses

Leitwerk's built-in checks auto-detect the toolchain, and they only recognise
Node/TS (`package.json`, `tsconfig.json`) and Go (`go.mod`). **Swift matches no
detector.** On a Swift repo the generic `lint`, `types`, `tests`, and `erosion`
checks all take the "nothing to run here" path (`exit 2` = skip), which is not a
failure — so the gate reports **green while verifying essentially no Swift
code**. `required-checks` (the roadmap item that would make a silent skip red at
a tier) is not built yet, so nothing surfaces the gap on its own.

CONFIRMED — with the default policy on a Swift tree, `lint/types/tests/erosion`
skip; only `drift` (spec-anchor tracking, language-agnostic) and `sast` (semgrep,
whose Swift ruleset is thin) do anything.

The two highest-leverage moves, in order:

1. Make Swift actually verifiable — override the checks with the Swift
   toolchain. Without this, the gate is a formality on this project.
2. Encode the domain invariants (on-device privacy, macOS permissions, the model
   boundary, the dependency graph) as blast-radius policy — this is where the
   gate earns its keep on a local-AI meeting app, well beyond generic linting.

## Part 1 — wire the Swift toolchain (Phase 0)

An adopter never edits the built-in checks. Drop per-check overrides into the
target repo's `leitwerk/checks/`; each overrides the built-in of the same name.

### Design decisions (validated)

- **`swift format lint` must run with `--strict`.** CONFIRMED: without `--strict`
  it reports violations as *warnings* and exits `0` (a vacuous pass); with
  `--strict` the same violations are errors and it exits `1`. The `lint` override
  uses `--strict`, and prefers SwiftLint when it is installed.
- **On a Swift project, a missing linter is a red gate (`exit 1`), not a skip.**
  CONFIRMED. The built-in skip contract (`exit 2` = "nothing to run") is correct
  when the repo is *not* Swift, but on a Swift project a missing tool is a
  misconfiguration, not "nothing to run". Failing loud closes the silent-skip
  hole locally, before `required-checks` lands upstream.
- **The compiler is the type oracle.** `swift build` (SPM) / `xcodebuild build`
  (Xcode) is the `types` check — the cheapest external oracle for a typed
  language.

### `leitwerk/checks/lint.sh` (CONFIRMED on SPM; SwiftLint branch INFERRED)

```bash
#!/usr/bin/env bash
# Swift lint override. Prefers SwiftLint; falls back to Apple swift-format.
# Contract: exit 0 pass, 1 fail, 2 skip. On a Swift project a *missing* linter
# is a red gate, not a skip. swift-format's `lint` only fails with --strict.
set -euo pipefail

is_swift_project() {
  [ -f Package.swift ] || ls -d ./*.xcodeproj ./*.xcworkspace >/dev/null 2>&1
}

if ! is_swift_project; then
  echo "no Swift project (no Package.swift / .xcodeproj / .xcworkspace)"
  exit 2
fi

paths=()
[ -d Sources ] && paths+=(Sources)
[ -d Tests ] && paths+=(Tests)
[ ${#paths[@]} -eq 0 ] && paths+=(.)

if command -v swiftlint >/dev/null 2>&1; then
  swiftlint lint --quiet --strict
  echo "swiftlint: clean (${paths[*]})"
elif swift format --version >/dev/null 2>&1; then
  swift format lint --strict --recursive "${paths[@]}"
  echo "swift-format: clean (${paths[*]})"
else
  echo "Swift project but no linter (swiftlint / swift-format) present — cannot lint" >&2
  exit 1
fi
```

### `leitwerk/checks/types.sh` (SPM CONFIRMED; xcodebuild INFERRED)

```bash
#!/usr/bin/env bash
# Swift type/contract oracle: the compiler.
set -euo pipefail

if [ -f Package.swift ]; then
  swift build
  echo "swift build: type-check ok"
elif ls -d ./*.xcworkspace >/dev/null 2>&1; then
  ws=$(ls -d ./*.xcworkspace | head -1)
  xcodebuild -workspace "$ws" -scheme "${LEITWERK_SWIFT_SCHEME:?set LEITWERK_SWIFT_SCHEME for the xcodebuild path}" build
  echo "xcodebuild build: ok ($ws)"
elif ls -d ./*.xcodeproj >/dev/null 2>&1; then
  proj=$(ls -d ./*.xcodeproj | head -1)
  xcodebuild -project "$proj" -scheme "${LEITWERK_SWIFT_SCHEME:?set LEITWERK_SWIFT_SCHEME for the xcodebuild path}" build
  echo "xcodebuild build: ok ($proj)"
else
  echo "no Swift project"
  exit 2
fi
```

### `leitwerk/checks/tests.sh` (SPM CONFIRMED; xcodebuild INFERRED)

```bash
#!/usr/bin/env bash
# Swift test oracle: swift test (SPM) or xcodebuild test (Xcode project).
# Runs the suite; does not yet prove it is non-vacuous (that is the roadmap's
# verification-helpers axis).
set -euo pipefail

if [ -f Package.swift ]; then
  swift test
  echo "swift test: passed"
elif ls -d ./*.xcworkspace >/dev/null 2>&1; then
  ws=$(ls -d ./*.xcworkspace | head -1)
  xcodebuild test -workspace "$ws" \
    -scheme "${LEITWERK_SWIFT_SCHEME:?set LEITWERK_SWIFT_SCHEME for the xcodebuild path}" \
    -destination "${LEITWERK_SWIFT_DESTINATION:-platform=macOS}"
  echo "xcodebuild test: passed ($ws)"
elif ls -d ./*.xcodeproj >/dev/null 2>&1; then
  proj=$(ls -d ./*.xcodeproj | head -1)
  xcodebuild test -project "$proj" \
    -scheme "${LEITWERK_SWIFT_SCHEME:?set LEITWERK_SWIFT_SCHEME for the xcodebuild path}" \
    -destination "${LEITWERK_SWIFT_DESTINATION:-platform=macOS}"
  echo "xcodebuild test: passed ($proj)"
else
  echo "no Swift project"
  exit 2
fi
```

For the Xcode (non-SPM) path, set `LEITWERK_SWIFT_SCHEME` (and optionally
`LEITWERK_SWIFT_DESTINATION`, default `platform=macOS`) in the CI environment.

## Part 2 — blast-radius policy (`leitwerk/tiers.conf`)

`[paths]` is first-match-wins, so the T2 privacy/security globs are listed
before the broad `Sources/**` rule. This encodes the "local AI" promise as
enforceable policy: the code paths that could move meeting content off-device,
or that drive the model, are the highest tier.

```ini
[tiers]
T0 = lint drift
T1 = lint types tests drift
T2 = lint types tests drift sast

[paths]
# --- T2: irreversible / security / privacy / supply-chain surface ---
**/*.entitlements = T2
**/Info.plist = T2
**/*.xcconfig = T2
fastlane/** = T2
Package.swift = T2
Package.resolved = T2
Sources/**/Networking/** = T2
Sources/**/Telemetry/** = T2
Sources/**/Inference/** = T2
Sources/**/Persistence/** = T2

# --- T1: normal Swift code and tests ---
Sources/** = T1
Tests/** = T1

# --- T0: docs ---
**/*.md = T0
docs/** = T0

[human-owned]
leitwerk/constitution.md
leitwerk/tiers.conf
```

CONFIRMED — `leitwerk tier <path>` against this policy:

| Path | Tier | Why |
|---|---|---|
| `Sources/App/Networking/Client.swift` | **T2** | can move data off-device |
| `Sources/App/Inference/WhisperEngine.swift` | **T2** | local-model boundary |
| `Sources/App/Persistence/TranscriptStore.swift` | **T2** | stores meeting content |
| `Sources/App/MeetingView.swift` | T1 | normal UI code |
| `Tests/AppTests/MeetingTests.swift` | T1 | test code |
| `App/Info.plist` | **T2** | permission usage descriptions |
| `App/App.entitlements` | **T2** | mic / screen-recording / sandbox |
| `Package.resolved` | **T2** | pinned dependency graph |
| `Config/Release.xcconfig` | **T2** | signing / release config |
| `README.md` | T0 | docs |

The policy targets egress/model/persistence code by *path*, so the project must
keep that code in named boundary directories (`Networking/`, `Telemetry/`,
`Inference/`, `Persistence/`). State that convention in the constitution — it is
what makes the glob-based policy meaningful.

## Part 3 — constitution invariants (proposed text for the new repo)

Add to the new project's `leitwerk/constitution.md` (its own, human-owned file):

- **On-device by default.** Meeting content — audio, transcripts, derived
  summaries — does not leave the device. Any code that opens a network
  connection, sends telemetry, or persists meeting content lives under the named
  boundary directories and is tier T2; a change there requires the T2 gate and an
  explicit sign-off.
- **Least-privilege permissions.** Every entitlement and `Info.plist` usage
  description is justified in the spec that introduces it. Entitlement and
  `Info.plist` changes are T2.
- **Pinned local model.** The inference model has a pinned version and a recorded
  source + checksum; a model change is T2 and states the provenance.

## Validation record

Harness: a minimal SPM package (`MeetingKit`) with one clean source file and one
Swift Testing test, plus the three overrides and the policy above. Run with the
gate binary built from this repo (`mise run build`).

| Case | Expectation | Result |
|---|---|---|
| Clean fixture, `verify --tier T1` | PASS; lint/types/tests all run | CONFIRMED PASS |
| Clean fixture, `verify --tier T2` | PASS; `sast` skips (no semgrep) | CONFIRMED PASS |
| Tier mapping (10 paths) | as table above | CONFIRMED |
| `swift format lint` without `--strict`, on a violation | exits 0 (vacuous) | CONFIRMED |
| `swift format lint --strict`, on a violation | exits 1 | CONFIRMED |
| Lint violation in `Sources` | gate RED at `lint` | CONFIRMED |
| Failing test | gate RED at `tests` | CONFIRMED |
| Non-Swift directory | `lint` skips (exit 2), no false red | CONFIRMED |
| Swift project, no linter reachable | `lint` fails loud (exit 1) | CONFIRMED |
| Xcode (non-SPM) `xcodebuild` branches | build/test via scheme | INFERRED (not run) |

## Gaps this surfaces (feed back to the leitwerk roadmap)

- **`sast` skips silently at T2** when semgrep is absent, and semgrep's Swift
  coverage is thin regardless — CONFIRMED live in the T2 run above. This is the
  `required-checks` roadmap item (mark a check non-skippable at a tier) and the
  dependency/SAST gap (`bindings/open/AGENTS.md:72` promises a "dependency policy"
  at T2 with no check behind it). A Swift-aware security/dependency check has no
  off-the-shelf answer here yet.
- **`erosion` has no Swift analyzer** (built-in expects `jscpd`). `periphery`
  (dead-code) is the closest Swift tool; not installed on the test machine, so an
  `erosion` override is INFERRED, not provided.
- **Language coverage** — this is the concrete instance of the roadmap's
  language-coverage gap. These overrides could become a shipped Swift preset
  rather than per-repo copy-paste.
- **The Xcode (non-SPM) path is unvalidated** — it needs a real `.xcodeproj` +
  scheme + destination and a signing environment.

## What Leitwerk does not cover here

The Go release pipeline in this repo (release-please + a build matrix producing
static binaries) does not apply to a macOS app. Distribution means code signing,
notarization, and a DMG/Sparkle channel — a separate pipeline. Leitwerk's gate is
language-agnostic; its release automation is not.

## Next steps

1. Create the new repo; `leitwerk init`; drop in the three overrides and the
   tiers policy; write the constitution invariants above.
2. Confirm each check reports "runs", not "skipped" — on Swift, green is only
   meaningful once the overrides are active.
3. Install `swiftlint` and `periphery` in dev + CI to raise lint depth and add an
   `erosion` override.
4. Validate the `xcodebuild` path once the app is a real Xcode project.
