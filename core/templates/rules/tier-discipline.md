---
description: High-blast-radius (T2) discipline
paths:
  # Scope this to YOUR repo's T2 paths (see leitwerk/tiers.conf). Examples:
  - "**/db/migrations/**"
  - "**/*.sql"
  - "infra/**"
  - "**/*.tf"
---

You are editing a **T2** path — irreversible, infra, or data. A defect here is
the hardest to undo, so verification is deepest.

- Write the oracle first: the failing test/property/contract that proves the
  change before you make it.
- Run `leitwerk verify --tier T2` and land only on green.
- T2 also requires SAST, dependency policy, and an explicit rollback.
- Surface spec↔code drift; do not resolve it silently.

Rules are a Claude Code primitive and are **not** packaged by the Leitwerk
plugin. Copy this file into your repo's `.claude/rules/` and set `paths:` to your
own T2 globs. Open-code agents get the same guidance from `AGENTS.md`.
