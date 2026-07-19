# Open-code binding

For agent tools that read `AGENTS.md` — Codex, GitHub Copilot, Cursor, Windsurf,
Aider, Zed, Jules, and others.

## What it ships

- **`AGENTS.md`** — the working method, roles, tiers, and the gate rule, in the
  cross-tool format. Copy to a repo root (nested per-package files are supported;
  closest wins).
- **`.codex/config.toml`** — Codex config: instruction fallbacks and a raised
  doc-size budget so the constitution + specs are not truncated.
- **`.codex/agents/*.toml`** — the specialist roles as Codex custom agents
  (`architect`, `test-engineer`, `security-reviewer`). Other tools apply the same
  roles as review lenses.
- **`ci/leitwerk-verify.yml`** — the authoritative gate as a GitHub Actions
  workflow. This is the enforcement point for open code: there is no universal
  hook, so CI runs `leitwerk verify` as a required status check and a red gate
  blocks merge.

## Enforcement note

On Claude Code the gate is enforced client-side by a Stop hook. Open-code tools
have no common hook mechanism, so enforcement moves to CI. The command is
identical — only where it is enforced differs.

Open code also has no equivalent to Claude Code's dynamic workflows, so the
multi-agent adversarial review (Layer 2) is not orchestrated automatically here —
run the roles sequentially or as review lenses. What keeps the guarantee intact
across both worlds is the same external gate: `leitwerk verify` in CI.
