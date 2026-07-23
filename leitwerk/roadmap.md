# Roadmap — leitwerk-devkit

The ordered backlog of work needed to take Leitwerk from a working scaffold to a
framework whose claims are all backed by running code.

## How this fits the Leitwerk process

Leitwerk defines a per-change lifecycle — spec → plan → build → review, gated at
the change's blast-radius tier. It does **not** define a separate "roadmap"
artifact, and it does not need one: a roadmap is the ordered set of *future
specs*. Each item below is a **proto-spec** — problem, target behaviour, tier,
and the acceptance condition that makes the gate green. Nothing here bypasses the
gate; the roadmap only decides order.

Lifecycle of a roadmap item:

1. **Proposed** — an entry in this file (spec vocabulary, not yet started).
2. **Active** — promoted to `leitwerk/specs/<slug>.md` via the `leitwerk-spec`
   skill; if non-trivial, `leitwerk-plan` breaks it into gated steps.
3. **Built** — `leitwerk-build` implements each step; `leitwerk verify` must be
   green at the step's tier before it lands.
4. **Reviewed** — `leitwerk-review` runs the tier's roles and confirms the spec
   still matches the code; the change merges on a green gate, the spec archives,
   and the item leaves this file.

This file is a **living worklist of open (not-yet-landed) work** — the ordered
set of future specs plus whatever is currently in flight. It holds every item
that has not yet landed; an item **leaves** only when it **lands** and its durable
record moves to `leitwerk/specs/archive/<slug>.md`. A not-started item has no spec
yet; an in-flight item has an *active* spec (`leitwerk/specs/<slug>.md`) that
carries its detail while the roadmap keeps its place in the order. Its length
tracks *open work*, not history, so it does not grow without bound.

- **Identity is the slug** — the bolded name that leads each item (`cli-publish`),
  not a number. There are no per-item numbers: a spec links to its item by slug
  (`Roadmap: <slug>`), commits scope by it, and reordering is just moving a line —
  nothing to renumber, nothing to collide. Order is the item's *position*;
  grouping is the phase section it sits under.
- **Items are thin**: one line of intent, an acceptance sketch, a tier estimate.
  Full Problem/Behaviour/Acceptance detail belongs in the spec at promotion.
- **State durable intent, never world-state.** "no remote repo yet" / "has never
  run" are status snapshots that rot behind the guard and mislead (`ci-live` did
  exactly this). Write the *gap* and the *acceptance* that closes it. **Status is
  derived** — from where an item's spec sits (none = not started, active spec = in
  flight, archived = landed and gone from here) and from git — never asserted
  here; an item's own acceptance is its status probe.

This file is **agent-editable** — deliberately not in the guard's deny-list
(`leitwerk/tiers.conf` `[human-owned]`). An agent edits it directly; the human
gates the change through the normal per-edit approval (in Claude Code the diff is
shown for confirmation) and through review — so the human still decides every
change and owns priority and scope, without a proposal round-trip.
Guarantee-bearing files (`constitution.md`, `tiers.conf`) stay **human-owned**:
hard-blocked, changed only by a human from a proposal. (A hook that *forces* the
confirmation even in auto-accept modes — `guard-confirm-class` — was considered
and deferred as unneeded.)

## Backlog

Tiers refer to `leitwerk/tiers.conf`. Order is by leverage: make the claims true
before making it broadly adoptable.

### Adoptable

**cli-publish** · tier **T1**
- *Problem:* the gate is distributed only as source; adopters and CI build it
  via `go install github.com/cf-sewe/leitwerk-devkit/core/cmd/leitwerk@latest`,
  which resolves only once the repository is public. Confirmed constraint: a
  Claude Code marketplace install sparse-copies **only** the plugin subdir, so
  the plugin's launcher cannot reach a sibling `core/` — an adopter must set
  `LEITWERK_HOME` or have a globally installed binary until publication. That
  makes publishing the real unblock for a marketplace-only adoption, not a
  nicety.
- *Behaviour:* make the module path publicly resolvable (`go install …@latest`
  works from a clean machine) and/or attach prebuilt static binaries to
  releases; document vendoring as the third path.
- *Acceptance:* a clean machine can obtain the binary (go install or release
  download) and run `leitwerk verify` in an unrelated repo.

**plugin-live-validation** · process (no tier — validation, not a code change)
- *Problem:* the plugin, skills, agents, and the review **workflow** are written
  but never run in a live Claude Code session; their composition is inferred,
  not proven.
- *Behaviour:* install the plugin from the marketplace and drive one real feature
  end to end (spec → plan → build → review) with the Stop-hook gate active; for a
  T2 change, launch `.claude/workflows/leitwerk-review.mjs` and confirm it spawns
  the role subagents as `agentType`s and its findings are advisory to the gate.
- *Acceptance:* a documented session transcript where the roles are spawned (incl.
  via the workflow), the gate blocks a red turn-end, and a green change lands.
  Findings feed back into the skills/agents/workflow (bidirectional refinement).

**ci-live** · tier **T2** (`.github/`)
- *Intent:* make the CI gate the binding **merge** constraint — running in CI is
  not enough until `main` is branch-protected with the `leitwerk gate` check
  *required*.
- *Acceptance:* a PR with a T2 change cannot merge until the gate is green (tier
  from the PR diff); a docs-only PR selects T0.

**plugin-bootstrap** · tier **T1** (optional `selftest`-backed `selfupdate` is T2)
- *Problem:* after cli-publish the gate binary is obtainable, but only by a
  **manual** `go install` / release download. A Claude Code marketplace install
  sparse-copies only the plugin dir, so the launcher cannot reach a sibling
  `core/` (`bindings/claude/bin/leitwerk` resolution order, CONFIRMED); and a
  plugin update leaves whatever binary was fetched earlier **stale**. There is no
  plugin-side path that provisions or refreshes the core binary — the adopter
  still obtains it by hand.
- *Behaviour:* provisioning is an **LLM-orchestrated onboarding step, not a hand-
  rolled cross-OS script** (a static bootstrap script rots across
  OS/arch/toolchains; the LLM adapts, a deterministic contract keeps it honest):
  - **Single path — download the matching release asset from the GitHub repo**
    (stable names + `checksums.txt`, the cli-publish contract) and **verify its
    sha256 before `chmod +x`**, showing the compare result; place it where the
    launcher already resolves it, with **no env var** required. This needs no Go
    toolchain and installs the exact CI-built, tested artifact rather than a
    locally recompiled one.
  - a `SessionStart` detection compares the **installed** version (`leitwerk
    version`, the ldflags tag) against the **expected** version — the plugin's own
    `plugin.json` `version`, read locally from the plugin dir (relative to the
    hook script, the `$SELF_DIR` technique `leitwerk-hook-guard` already uses),
    **no network**. Any mismatch (after normalising the `v` prefix) surfaces
    **missing-or-stale** and prompts a re-provision; equal is quiet. The same
    expected version also names the release **tag to download** (`v<version>`), so
    provisioning is deterministic and never queries "latest". Network happens only
    on provisioning, never on detection; **acquisition/execution is never silent**.
  - *Version coupling (what makes `plugin.json` a valid source of "expected"):*
    the plugin's version and the core release tag move **together** — the release
    process bumps `plugin.json` (release-please `extra-files`) or releases both in
    lockstep. Without that, `plugin.json` could not stand in for the expected core
    version. (This is the coupling the `architect` role owns, below.)
  - *Fallback (documented, not automated):* a platform/arch **not** in the release
    matrix has no prebuilt asset — that user builds from source (`go install
    …@<version>` or `make -C core build`). Not a primary path; just documented so
    no one is stranded.
  - *Optional, separable (tier T2, core):* a `leitwerk selfupdate` subcommand for
    the update path — it re-runs the same download+verify in-binary and cross-
    platform; only the first install needs the LLM-orchestrated fetch.
  - *Provenance (note — not required for v1):* `checksums.txt` ships **inside the
    same release** as the binary, so it proves the download arrived intact but not
    that the release itself is authentic — a compromised release/CI pipeline could
    publish a malicious binary with a matching checksum. **Signing** the release
    (cosign/minisign) and verifying against a public key shipped in the plugin
    would close that residual risk. Deferred: GitHub release + checksum over TLS
    is the widely-used bar (kubectl / gh / terraform); revisit if the threat model
    or adoption warrants it.
- *Acceptance:* on a clean machine, `/plugin install` followed by the onboarding
  step yields a working `leitwerk verify` **without a manual install step**; a
  tampered asset is refused by the integrity check; after a simulated plugin
  version bump the `SessionStart` detection reports the drift and points to the
  re-provision step.
- *Depends:* cli-publish (public releases + the stable asset/`checksums.txt`
  contract). *Relates:* plugin-live-validation (live validation would exercise
  this path).
- *Roles/skills:* `architect` (provisioning topology + plugin↔core version
  coupling), `security-reviewer` (download/verify path + no-silent-exec), 
  `test-engineer` (version-mismatch / integrity-refusal cases).

### Verification depth

**erosion-budgets** · tier **T2** — make `erosion.sh` enforce real
complexity/duplication ceilings instead of skipping when no analyzer is present;
define default budgets in the constitution.

**verification-helpers** · tier **T2** (`core` + checks)
- *Problem:* a check that *runs* is not tests that *assert*. `tests` can pass
  vacuously — `go test ./...` with no tests, or AI-written tests that reach high
  line coverage while asserting nothing meaningful. The whitepaper names a
  **mutation-score floor** as the T1+ design target ("Mutation score, not line
  coverage", §9 / Fig. 6), and the constitution invariant "the gate never
  silently under-verifies" forbids exactly this — but today check honesty is only
  a review-upheld convention (§9: "the gate cannot verify a check's own honesty").
  Nothing proves an adopter's tests are load-bearing.
- *Behaviour:* two enforcement layers, cheapest first, plus the original oracles —
  - a **non-vacuous guard** (cheap, per change): the `tests` check must actually
    execute tests exercising the changed code, not pass with zero relevant tests;
  - a **mutation-score floor** at T1+ (per the whitepaper): inject faults into the
    changed code, require the suite to catch them, red below the floor. Scoped to
    the changed set to bound cost (composes with monorepo-affected-scope), full
    score as a periodic run. Default floors live in the constitution (human-owned),
    like the erosion-budgets defaults;
  - scaffolds for property, mutation, and characterization tests so
    `leitwerk-build` has concrete oracles to reach for rather than prose.
- *Acceptance:* on a fixture, a suite that passes vacuously (delete a test / strip
  an assertion on the changed code) turns the gate red via the guard/floor; a
  load-bearing suite passes; where no mutation tooling is present the floor skips
  with a visible note until made required (required-checks).
- *Boundary:* verification-helpers = "a check that ran is *meaningful* (quality)";
  required-checks = "the check *ran* at all (presence/completeness)";
  erosion-budgets = code health (complexity/duplication) — three separate axes.
- *Roles/skills:* `test-engineer` (floor + oracle scaffolds), `architect`
  (scoping vs monorepo-affected-scope; where the periodic full run lives).

**provenance-tooling** — turn CONFIRMED/INFERRED/GAP from a review
convention into something the review output actually carries and the gate can
check for (e.g. no GAP left on a T2 change without a human sign-off).

**open-code-live-validation** — the plugin-live-validation equivalent for
Codex/AGENTS.md: drive one change under an open-code agent and confirm the CI gate
is the binding constraint.

**monorepo-affected-scope** · tier **T2** (`core/bin/leitwerk` + checks) ·
*proposed (agent-suggested, needs human priority)*
- *Problem:* the gate is diff-aware only for **tier selection**; the checks
  themselves run project-wide (`go test ./...`, `tsc`, …). On a monorepo that
  re-validates untouched code on every change — slow, and it blurs which package
  a red result belongs to. But naive "validate only touched files" is unsafe: a
  change to a shared package can break an untouched **dependent**.
- *Behaviour:* the gate validates the **affected set** — changed files plus their
  reverse-dependencies — not the whole repo and not only the literally-touched
  files. The affected set is computed deterministically from the diff plus a
  dependency graph and passed to the checks (e.g. a `$LEITWERK_CHANGED` set the
  check scripts scope themselves to). Package boundaries and per-package checks
  are human-owned policy, expressed with the existing repo-local `leitwerk/`
  override (a per-package `tiers.conf` / `checks/`). The agent may not narrow its
  own scope — scoping is part of the gate, not agent judgment.
- *Safety:* scoping trades completeness for speed, so pair a **scoped per-change
  gate** (fast, pre-merge) with a **periodic unscoped full gate** (nightly / on
  the protected branch) that catches what a stale graph or wrong exclusion missed.
  A T2 change may widen the scope.
- *Acceptance:* on a two-package monorepo fixture, editing package A runs A's
  checks (and B's iff B depends on A) and skips unrelated package C; a
  deliberately broken dependent still turns the gate red; the full nightly gate
  runs everything.
- *Roles/skills:* `architect` (affected-set model + boundary policy),
  `test-engineer` (fixture + `selftest` cases for scope correctness).

**required-checks** · tier **T2**
- *Problem:* a check that skips at T2 — `sast` when semgrep is absent, or a repo
  with a `go.mod` but no linter tool installed — leaves the gate green; "a
  skipped security check on a T2 change is a blocker" is role prose, not
  mechanism. This is the **completeness** axis: the checks a tier calls for must
  be *present and running*, not silently absent.
- *Behaviour:* `[tiers]` syntax marks checks that may not skip at a tier (e.g.
  `T2 = … sast!`); a skip of a required check turns the gate red with a message
  naming the missing tool. Onboarding additionally flags a language whose
  conventional check is unwired (marker present, tool absent) so completeness
  gaps are visible rather than silent.
- *Acceptance:* scenario/`selftest` case: empty toolchain at T2 with `sast`
  required → exit 1; with the tool installed → normal pass/fail.
- *Boundary:* required-checks = "the check *runs* (presence/completeness)"; the
  *quality* of a check that ran (vacuous / mutation-score) is verification-helpers.

**repo-map** · tier **T2** (core) — design-proposal open decision O2
- *Problem:* the research step retrieves via scouts running grep/glob; P4
  (retrieve-don't-preload) calls for a structural symbol map, and context-rot
  evidence says preloading does not scale with repo size.
- *Behaviour:* the CLI builds/refreshes a tree-sitter symbol map (Aider-style
  ranking) that scouts query for "where is X / who calls X"; a typed code
  graph stays deferred until a repo exceeds ~1k files (per O2).
- *Acceptance:* on a fixture repo, a scout answers a where-is/who-calls
  question from the map without full-text search; the map refreshes
  incrementally and stale entries are detectable.
- *Roles/skills:* `architect` (map format/CLI surface), `test-engineer`.

**roadmap-spec-join** · tier **T2** (`core` + checks)
- *Intent:* derive each item's status from the spec lifecycle, not hand-written
  prose, via an explicit machine-checked roadmap↔spec link.
- *Mechanism:* a spec declares `Roadmap: <slug>`; `lifecycle` resolves an
  **active** spec's slug against this file (orphan = red), exempting archived
  specs (as `drift` does — a landed item has left the roadmap). Status is one pass
  over `specs/`: active spec = in flight, archived = landed, no spec = not started.
  Line added to `spec.template.md` and the `leitwerk-spec` skill.
- *Acceptance:* a spec whose `Roadmap:` slug is absent here reds `lifecycle`;
  landed/active/open is derivable for every item without reading stored status.

### Measure the framework itself

**efficiency-evaluation** · process (periodic, not per-change)
- *Problem:* deterministic proxies are gated (the `context` budget check), but
  outcome efficiency — tokens-to-green-gate, wall time to a landed change,
  human interventions per change — is unmeasured, and no off-the-shelf
  benchmark measures a governance framework's overhead.
- *Behaviour:* an A/B harness: a fixed task set (the runnable scenarios in
  `examples/scenarios/` plus reference-app changes) driven with and without
  Leitwerk, several runs per task, reporting cost and outcome distributions.
  Candidate substrates: long-horizon suites (SWE-bench-Pro-class) and
  terminal-agent suites (Terminal-Bench-class) with cost reporting.
- *Acceptance:* a first dated report in `docs/reviews/` comparing the two arms
  on cost and outcome.

## Not planned (explicit non-goals for now)
- A bespoke spec DSL — specs stay Markdown.
- Auto-resolving drift — the framework surfaces, humans decide. This is a
  constitution invariant, not a backlog gap.
