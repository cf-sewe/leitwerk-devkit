# Adopting Leitwerk on a repository

The recommended order — highest leverage first, matching the whitepaper's
Phase 0 → 3 rollout.

## Phase 0 — the gate first (do this before anything else)
1. Make the CLI resolvable. The gate is a single static Go binary — either build
   it from a checkout (`mise run build`, then set
   `LEITWERK_HOME=/path/to/leitwerk-devkit/core` and put `$LEITWERK_HOME/bin` on
   PATH), `go install github.com/cf-sewe/leitwerk-devkit/core/cmd/leitwerk@latest`, or
   drop a prebuilt release binary on PATH (after verifying its checksum). The binary embeds its checks/templates,
   so it does not depend on the repo layout.
2. `leitwerk init` in the target repo → `leitwerk/{constitution.md,tiers.conf}`
   plus the Claude steering files `CLAUDE.md`, `.claude/rules/tier-discipline.md`,
   and `.claude/workflows/leitwerk-review.mjs` (these are repo-level; a plugin
   cannot carry them).
3. Edit `tiers.conf` so real irreversible paths (migrations, IaC, auth, billing)
   are T2, and confirm the `[human-owned]` list names your policy files. On a
   new/empty repo, set these from the *intended* architecture — path globs match
   files that do not exist yet, so the policy is ready before the code is.
4. Wire `core/checks/*` to the project's real build/test/lint (SAST for T2).
5. Add `bindings/open/ci/leitwerk-verify.yml` to `.github/workflows/` and make
   "leitwerk gate" a required status check.

At this point the guarantee exists independently of any agent. Everything below
adds agent ergonomics on top.

## Phase 1 — the constitution + specs
Fill `leitwerk/constitution.md` (invariants, DoD, roles). Start writing specs in
`leitwerk/specs/` for new work. For brownfield code, add characterization tests
around risky areas before changing them. A repo that already has its own `specs/`
(or docs) in another format is fine — `leitwerk/` is a separate, additive
namespace, nothing migrates it, and `drift` only tracks `leitwerk/specs/`. Convert
an existing spec to the Leitwerk shape only when a change next touches its area
(strangler-fig), never in bulk.

## Phase 2 — Claude Code binding
`/plugin marketplace add cf-sewe/leitwerk-devkit` then
`/plugin install leitwerk@leitwerk`. Agents now get the phase skills, the roles,
and — activated automatically, with no manual `.claude` hook setup — the
Stop-hook gate and the `PreToolUse` guard that blocks edits to human-owned files.
The plugin's `leitwerk` launcher still calls the core CLI from Phase 0 (a
marketplace install copies only the plugin, not `core/`), so `LEITWERK_HOME` or
a `go install`ed binary on PATH must be present. The `CLAUDE.md`, `.claude/rules/`, and the review
workflow `.claude/workflows/leitwerk-review.mjs` scaffolded in Phase 0 are what
steer the session; the workflow drives T2 fan-out review (opt-in).

## Phase 3 — open-code binding
Copy `bindings/open/AGENTS.md` to the repo root (edit the project-specific
parts); add `.codex/` if the team uses Codex. The CI gate from Phase 0 already
enforces the same command.

## What stays human-owned
Requirements, the constitution, tier thresholds, spec↔code drift reconciliation,
and review of user-visible surfaces. The framework surfaces these; it does not
decide them.
