# Leitwerk concept review — 2026-07-19

Scope: a full review of the Leitwerk concept against its reference
implementation in this repository — value, completeness, token/context
efficiency, CI-measurability of efficiency, long-running-project support, and a
line-by-line cross-check of `docs/whitepaper.html` against the code. Method:
manual survey of every framework artifact, measurements on this repository, and
an independent fact-check pass over the whitepaper (every checkable claim
verified against a file). Small issues found were fixed in the same pass;
everything requiring a human-owned edit is listed as a proposal at the end.

## 1 · Verdict in brief

The core idea holds together: one deterministic command at the center, policy in
human-owned files, procedures loaded on demand, and enforcement that does not
depend on the agent's cooperation. The layering (always-on ≈ 1k tokens, skills
and specs on demand) is genuinely context-efficient by design, not by accident.

The main weaknesses found: (a) the whitepaper overstated three things the
implementation does not deliver (drift sensing, "the agent cannot weaken the
gate", automatic signal-driven triggers) — now corrected; (b) the framework had
no self-cost accounting and no way to catch context-efficiency regressions in
CI — a budget check now exists, pending policy wiring; (c) no artifact
lifecycle for long-lived repositories — specs/plans accumulated with no status,
archive, or consolidation convention — a lifecycle convention is now in the
templates and skills; (d) several enforcement gaps remain open and are listed
in §6 below.

## 2 · Token efficiency

What a session pays, measured on this repository (2026-07-19):

| Surface | When loaded | Measured |
|---|---|---|
| `CLAUDE.md` | every session | 197 words / 1 423 bytes ≈ 360 est. tokens |
| skill + agent frontmatter (9 files) | every session (system prompt listing) | 36–48 words each; total ≈ 630 est. tokens |
| **always-on total** | | **≈ 990 est. tokens** |
| `.claude/rules/tier-discipline.md` | only when a T2 path is touched | 149 words |
| skill bodies | only on invocation | 171–315 words each |
| role charters | only when a role is spawned | 115–166 words each |
| constitution | linked from CLAUDE.md, read on demand | 634 words |
| specs/plans | read when relevant to the change | 246–1 143 words each |

Gate execution (wall time, not tokens; warm build cache): T0 ≈ 0.1 s, T1 ≈
0.5 s, T2 ≈ 1–2 s. Verify output is ≤ ~420 bytes, so a gate run costs an agent
almost nothing to read back; a red gate surfaces only the failing check's last
line.

The one real token sink is the T2 review workflow: up to 3 role reviews, plus
one refutation agent per finding, plus a gate runner. The findings array was
uncapped — a chatty reviewer could spawn unbounded refuters. **Fixed:** the
findings schema now carries `maxItems: 8` per dimension (both the template and
this repo's copy), bounding the worst case at 3 + 24 + 1 agents, paid only at
T2.

## 3 · Context efficiency

The design already follows "retrieve, don't preload" for its own artifacts:
CLAUDE.md is a bridge, not a manual; procedures live in skills; the constitution
is linked, not inlined; the tier-discipline rule is path-scoped. Measured
always-on cost ≈ 990 tokens is small.

The risk is erosion, not the current state: nothing stopped CLAUDE.md from
growing, a skill description from bloating, or procedures from creeping into
always-on files. That is exactly the class of regression the framework's own
erosion principle targets, so it should be gated (next section).

## 4 · Can efficiency be tested in CI? (yes — for the deterministic part)

Two measurement classes, different homes:

**Deterministic proxies — CI-gateable.** Implemented:
`leitwerk/checks/context.sh` measures the always-on surface and fails red when
a budget is exceeded: CLAUDE.md ≤ 200 lines (the limit the template itself
states), each path-scoped rule ≤ 100 lines, each frontmatter description ≤ 80
words, always-on total ≤ 2 000 estimated tokens (bytes/4 heuristic — crude but
monotone, which is all a regression gate needs). Current reading: ~990 tokens,
comfortably within budget. Gate wall-time and verify-output-size budgets could
be added later, but wall time is noisy in CI and better observed than gated.

Caveat the framework itself imposes: an agent may *write* a check but cannot
*wire* it — `[tiers]` lives in human-owned `leitwerk/tiers.conf`. The check is
therefore dormant until the one-line proposal in §8 is applied. (Note in
passing: the constitution says "an agent may add a check", but mechanically it
can only propose the wiring. That is the guard working as designed, but the
constitution's wording could acknowledge it.)

**Outcome efficiency — not CI-gateable.** Tokens-to-green-gate, wall time to a
landed change, human interventions per change, pass rate: stochastic and
model-bound, so a per-PR gate would flake. There is no off-the-shelf benchmark
that measures a governance framework's overhead. The honest protocol is an A/B
harness: a fixed task set (the runnable scenarios plus reference-app changes)
driven with and without Leitwerk, several runs per task, reporting cost and
outcome distributions. Closest usable substrates: long-horizon suites
(SWE-bench-Pro-class) and terminal-agent suites (Terminal-Bench-class) with
cost reporting, run against a Leitwerk-governed copy of the target repo.
Proposed as roadmap item M4.1 (§8) — periodic, not per-change.

The whitepaper now contains both the self-cost accounting (§9.4) and the
efficiency-measurement split (§12).

## 5 · Long-running projects: naming, retention, "dreaming"

Findings:

- **The framework conflated two document kinds.** A *living contract* (what the
  system must do — `self.md`) and a *change record* (how we got here —
  `go-cli.md`, `steering-alignment.md`). Both lived in `leitwerk/specs/` with
  no status, so after one day of development, 4 of 6 spec files were already
  history presented as if current. Over years this becomes a context liability:
  agents told to "read the spec" face a growing pile where staleness is
  invisible.
- **Naming.** Specs are slug-named, which is right for living documents (their
  identity is the area, not the date). Perishable artifacts — plans, review
  reports, analyses like this one — should carry the date; this document
  follows the repo owner's `<date>_<time>-` prefix convention. What was missing
  was not timestamps in spec names but a *status* marker and a place for
  history to go.
- **Retention/pruning ("dreaming").** Nothing is ever deleted (git keeps
  history), but the *active set* must stay small. The needed cycle is
  consolidation: when a change lands, its durable content moves to the
  constitution's decisions-of-record or the area's living spec; the change-spec
  is marked landed; finished plans and superseded specs move to
  `leitwerk/specs/archive/`.

**Fixed now:**
- Spec and plan templates carry a `Status:` line
  (`draft → active → landed YYYY-MM-DD → superseded by <slug>`) and state the
  archive convention.
- All six existing spec/plan files are stamped (`self.md` active as the living
  contract; three change-specs landed; `go-cli` active until merged).
- `leitwerk-review` gained a lifecycle step: on landing, set the status and
  archive finished plans/superseded specs.

**Still open (proposals in §8):** a consolidation cadence (who runs the
"dreaming" pass and when — suggested: part of review at landing, plus a
periodic grooming pass), `drift`/tooling awareness of `archive/` once M1.1
lands, and the fact that the constitution's decisions-of-record also grow
monotonically (acceptable while entries stay one paragraph; worth a budget
eventually).

## 6 · Completeness — weak spots that remain open

Ordered by how much of the framework's promise rests on them:

1. **The drift sensor is a placeholder** (counts spec files, always reports no
   drift) while "surface drift, never resolve it" is the headline principle.
   Already roadmap M1.1; this review adds urgency: the stale `self.md` found by
   the whitepaper cross-check is precisely the drift the sensor should have
   caught, and it sat unnoticed in the gate's own repository.
2. **The Stop hook verifies at a static tier** (`$LEITWERK_TIER`, default T1 in
   the plugin; T2 in this repo). It cannot see the change's real blast radius —
   a T2 change can end a turn having passed only T1 locally, and a docs-only
   turn pays the full configured tier. CI derives the tier from the diff; the
   CLI should too: a `leitwerk verify --auto` that computes the highest tier
   from `git diff` would make local and CI enforcement identical. (T2 change →
   needs a spec; proposed in §8.)
3. **A green gate with skips can over-reassure.** At T2, `sast` skipping
   because semgrep is absent is visible but non-blocking; the constitution says
   a skipped security check on T2 is a blocker, but only the security-reviewer
   role enforces that, not the gate. A `!`-marked required check in `[tiers]`
   (red if it *skips* at that tier) would move this from convention to
   mechanism. (Proposed in §8.)
4. **Check honesty is unverifiable by the gate.** A repo-local override that
   unconditionally exits 0 satisfies the gate. This is inherent (the gate
   cannot audit its own oracles) — the mitigations are review of `leitwerk/checks/`
   diffs and CI running on a protected branch. The whitepaper now states this
   limit honestly instead of implying the opposite.
5. **Signal-driven triggers are design, not mechanism.** Only path→tier is
   mechanical; "touches auth → wake security-reviewer" is prose in skills. The
   whitepaper now says so. A diff-scanning trigger table would be a natural
   `verify --auto` companion but is unproven.
6. **Never-executed paths.** CI has never run remotely (M2.3) and the
   plugin/workflow composition has never been driven in a live session (M2.2).
   The whitepaper now carries a dated implementation-status note. These two
   validations are the highest-leverage next steps after publication.

## 7 · Example scenarios — defined and CI-tested

`examples/scenarios/` now exists: five self-contained scripts, each building a
throwaway fixture repo and asserting one observable guarantee, with a
`run-all.sh` runner. They are executable documentation — wired into the gate
via the `selftest` check, so a regression in any scenario turns CI red:

| Scenario | Guarantee proven |
|---|---|
| s1 tier-escalation | migration/infra → T2, docs → T0, app → T1 under the scaffolded default policy |
| s2 red-gate | a failing check ⇒ exit 1 + `gate: FAIL` — a broken change cannot land |
| s3 human-owned-guard | `guard` exits 3 on policy files (incl. absolute paths and `//` spellings), 0 on app code |
| s4 skip-honesty | empty toolchain ⇒ green gate with visible `(skipped)`, never a fake pass |
| s5 local-override | repo-local `leitwerk/checks/lint.sh` shadows the built-in |

Candidates for later scenarios: drift detection (blocked on M1.1), tier
escalation across a multi-file diff (pairs with `verify --auto`), and a
required-check skip turning T2 red (pairs with proposal 3).

## 8 · Proposals requiring human-owned edits

> **Update (2026-07-19, later the same day):** A–D were applied with explicit
> human authorization; the T2 gate now runs six checks including `context`.

The guard blocks an agent from applying these; they are ready to paste.

**A · `leitwerk/tiers.conf` — wire the context check** (makes the efficiency
budget actually gate; `context` is cheap, so it can run at every tier):

```
[tiers]
T0 = json context
T1 = json shell drift parity context
T2 = json shell drift selftest parity context
```

**B · `leitwerk/tiers.conf` — tier the Go gate source as T2** (carried over
from the Go-migration review; without it a `.go`-only change is caught by the
`*` catch-all at T1 and skips `selftest`):

```
core/cmd/**       = T2
core/internal/**  = T2
core/assets.go    = T2
core/go.mod       = T2
core/Makefile     = T2
```

**C · `leitwerk/constitution.md` — decision of record for the Go CLI**
(carried over; text proposed in the Go-migration summary).

**D · `leitwerk/roadmap.md` — updates:**
- M2.1 (cli-publish): npm wording → Go distribution
  (`go install github.com/cf-sewe/leitwerk-devkit/core/cmd/leitwerk@latest`,
  prebuilt release binaries, or vendoring). The M2.1 problem statement still
  cites the removed npm package.
- New **M1.4 · spec-lifecycle**: `leitwerk/specs/archive/`, drift/tooling
  awareness of archived specs, and a consolidation ("dreaming") cadence —
  merge landed change-specs into living specs/decisions, prune the active set;
  define whether grooming is part of review, a periodic skill, or both.
- New **M2.4 · verify --auto**: derive the tier from the git diff in the CLI
  (shared by the Stop hook and CI); optionally extend toward diff-signal
  triggers (§6.5). T2, spec-first.
- New **M3.6 · required checks**: `[tiers]` syntax for checks that must not
  skip at a tier (e.g. `sast` at T2), turning the constitution's "skipped SAST
  is a blocker" from prose into mechanism.
- New **M4.1 · efficiency evaluation**: the A/B harness of §4 — fixed task
  set, with/without framework, cost + outcome distributions; report
  tokens-to-green, wall time, interventions. Periodic, not per-change.
- Record budgets: the `context.sh` budgets (200/100/80 lines-words, 2 000
  est. tokens) belong in the constitution once accepted, since a check script
  is agent-editable while the numbers should be policy.

## 9 · Fixes applied in this pass (gate green at T2 throughout)

Whitepaper (`docs/whitepaper.html`):
- Corrected the misquoted `self.md` excerpt (now matches the file, which was
  itself modernized — parity + Go-era selftest, npm remnant removed).
- Softened/scoped the overstated claims: "gate the agent cannot weaken" →
  policy-scoped wording; drift sensor marked as placeholder (M1.1); trigger
  table marked design-target vs. mechanical path→tier; Stop-hook tier default
  documented; check-honesty limit stated; Fig. 1 and Fig. 5 annotations
  aligned with shipped behavior.
- Added §9.4 "Cost of the framework itself" (measured numbers, dated), an
  implementation-status note (CI/workflow never executed remotely/live), an
  efficiency-measurement paragraph in §12, and two new §13 limitations
  (threat model; artifact lifecycle).
- Neutralized non-neutral phrasing ("unambiguous", "not stylistic", "cheap
  insurance", "pay for themselves", "decisively", "honest signal", "by
  construction, not hope", …) and fixed stale pointers (footer companion-doc
  path, `core/bin/leitwerk` link → source dir, `.codex/config.toml` routing
  claim → per-agent `model_reasoning_effort`).
- Updated the sample verify output to the current run (19 scripts, scenarios).
- QA: re-rendered light + dark (Edge headless); new sections verified in both.

Repo:
- `examples/scenarios/` (5 scenarios + runner) wired into `selftest`.
- `leitwerk/checks/context.sh` (dormant until proposal A).
- Review-workflow findings cap (`maxItems: 8`, template + repo copy).
- Spec/plan templates: `Status:` lifecycle line + archive convention; all six
  existing specs/plans stamped; `leitwerk-review` skill gained the lifecycle
  step.
- `bindings/open/.codex/agents/scout.toml` added (AGENTS.md promised all four
  roles; only three existed).
- Stale npm references removed from `README.md`, `docs/adoption.md`,
  `leitwerk/specs/self.md`.
