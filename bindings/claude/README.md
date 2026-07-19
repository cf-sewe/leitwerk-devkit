# Claude Code binding

A Claude Code plugin. Install via the marketplace at the repo root:

```
/plugin marketplace add cf-sewe/leitwerk-devkit
/plugin install leitwerk@leitwerk
```

## What it ships

- **`bin/leitwerk`** — a launcher added to the Bash `PATH` when the plugin is
  enabled. It resolves the real `core/` CLI (see the shim for resolution order).
  This is how skills and hooks call `leitwerk` by name.
- **`skills/`** — the workflow phases, invoked by name or by the model:
  `leitwerk-onboard`, `leitwerk-spec`, `leitwerk-plan`, `leitwerk-build`,
  `leitwerk-review`.
- **`agents/`** — the specialist roles as subagents: `architect`,
  `test-engineer`, `security-reviewer`, `scout`. Models are assigned per role
  (Opus for judgment, Haiku for read-only scouting). These are spawned directly
  for small changes and reused as `agentType`s by the review workflow for larger
  ones. (There is no `orchestrator` role — a dynamic workflow is the orchestrator.)
- **`bin/leitwerk-hook-guard`** — a `PreToolUse` helper. It reads the hook
  payload, extracts the target path, and asks `leitwerk guard` whether the path
  is human-owned; if so it exits 2, blocking the edit. The decision lives in core
  (`[human-owned]` in the tiers file) — the wrapper only adapts the payload.
- **`hooks/hooks.json`** — two hooks:
  - a `Stop` hook that runs `leitwerk verify || exit 2` (exit 2 blocks the turn
    from ending, so a task cannot finish on a red gate);
  - a `PreToolUse` hook on `Write|Edit` that runs `leitwerk-hook-guard`, so an
    agent cannot silently edit the constitution, tier policy, or roadmap.

## Repo-level steering (not packaged by the plugin)

Two Claude Code primitives are repo-level only and cannot ship inside a plugin:

- **Rules** (`.claude/rules/*.md` with `paths:` frontmatter) — a path-scoped rule
  carries T2 blast-radius discipline, loading only when a matching file is
  touched.
- **`CLAUDE.md`** — a small always-on file that points a session at the
  human-owned `leitwerk/constitution.md` and leaves procedures to the skills.

Because a plugin cannot carry these, `leitwerk init` scaffolds both into a
consuming repo (templates in `core/templates/`). Edit their `paths:`/specifics to
match the repo.

## Orchestration via dynamic workflows

Layer-2 orchestration (fan-out review, adversarial verification) uses Claude
Code's **dynamic workflows**. Two things follow from the platform:

- **Plugins cannot package a workflow.** So the workflow is shipped as a core
  template that `leitwerk init` scaffolds into repo-level
  `.claude/workflows/leitwerk-review.mjs` (invocable as `/leitwerk-review`). The
  `leitwerk-build` / `leitwerk-review` skills prefer it and fall back to spawning
  the roles directly when it is absent, so review never depends on the workflow
  being installed. A workflow's `agentType` resolves against this plugin's
  `agents/`, so the roles are reused unchanged.
- **Workflows are opt-in and tier-gated.** They need Claude Code v2.1.154+, are
  triggered explicitly (or via `ultracode`), and can be turned off with
  `disableWorkflows`. Leitwerk uses them for T2 / large changes only; T0/T1 stay
  lightweight. A workflow's verification is *soft* — `leitwerk verify` (the Stop
  hook) remains the hard, authoritative gate.

## Validated against

Claude Code plugin conventions (`.claude-plugin/plugin.json`, root-level
`skills/`/`agents/`/`hooks/`/`bin/`), `Stop`-hook blocking via exit code 2, and
plugin `bin/` on the Bash `PATH`; dynamic workflows reusing the subagent
registry via `agentType`, and workflows not being packageable in plugins. Note:
plugin agents do not support per-agent `hooks`/`mcpServers`; the gate hook is
defined at plugin level here. `model:` in agent frontmatter needs Claude Code
v2.1.196+; dynamic workflows need v2.1.154+.
