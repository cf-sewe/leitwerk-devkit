# Adopting Leitwerk on a repository

The recommended order — highest leverage first, matching the whitepaper's
Phase 0 → 3 rollout.

## Phase 0 — the gate first (do this before anything else)
1. Install the CLI: `npm install -g @cf-sewe/leitwerk` (or put `core/bin` on PATH).
2. `leitwerk init` in the target repo → `leitwerk/{constitution.md,tiers.conf}`.
3. Edit `tiers.conf` so real irreversible paths (migrations, IaC, auth, billing)
   are T2.
4. Wire `core/checks/*` to the project's real build/test/lint (SAST for T2).
5. Add `bindings/open/ci/leitwerk-verify.yml` to `.github/workflows/` and make
   "leitwerk gate" a required status check.

At this point the guarantee exists independently of any agent. Everything below
adds agent ergonomics on top.

## Phase 1 — the constitution + specs
Fill `leitwerk/constitution.md` (invariants, DoD, roles). Start writing specs in
`leitwerk/specs/` for new work. For brownfield code, add characterization tests
around risky areas before changing them.

## Phase 2 — Claude Code binding
`/plugin marketplace add cf-sewe/leitwerk-devkit` then
`/plugin install leitwerk@leitwerk`. Agents now get the phase skills, the roles,
and the Stop-hook gate.

## Phase 3 — open-code binding
Copy `bindings/open/AGENTS.md` to the repo root (edit the project-specific
parts); add `.codex/` if the team uses Codex. The CI gate from Phase 0 already
enforces the same command.

## What stays human-owned
Requirements, the constitution, tier thresholds, spec↔code drift reconciliation,
and review of user-visible surfaces. The framework surfaces these; it does not
decide them.
