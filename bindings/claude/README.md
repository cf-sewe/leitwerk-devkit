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
- **`agents/`** — the specialist roles as subagents the orchestrator spawns:
  `orchestrator`, `architect`, `test-engineer`, `security-reviewer`, `scout`.
  Models are assigned per role (Opus for judgment, Haiku for read-only scouting).
- **`hooks/hooks.json`** — a `Stop` hook that runs `leitwerk verify || exit 2`.
  Exit 2 blocks the turn from ending, so a task cannot finish on a red gate.

## Validated against

Claude Code plugin conventions (`.claude-plugin/plugin.json`, root-level
`skills/`/`agents/`/`hooks/`/`bin/`), `Stop`-hook blocking via exit code 2, and
plugin `bin/` on the Bash `PATH`. Note: plugin agents do not support per-agent
`hooks`/`mcpServers`; the gate hook is defined at plugin level here. `model:` in
agent frontmatter needs Claude Code v2.1.196+.
