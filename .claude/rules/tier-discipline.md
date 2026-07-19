---
description: High-blast-radius (T2) discipline for the gate itself
paths:
  - "core/bin/**"
  - "core/checks/**"
  - "**/*.sh"
  - ".github/**"
---

You are editing a **T2** path — the gate itself, or CI. A defect here silently
weakens every repo that adopts Leitwerk, so it is the highest blast radius.

- Write the oracle first. New behaviour in the CLI or a check gets a `selftest`
  assertion before the code, so the gate has something to prove it.
- Run `leitwerk verify --tier T2` and land only on green. The gate is
  deterministic — fix the cause of a red result, do not argue with it.
- A check never fakes a pass: nothing to run exits 2 (skip), never 0.
- Do not lower a threshold, remove a check, or downgrade a path's tier here.
  Those live in `leitwerk/tiers.conf`, which is human-owned — propose the change,
  do not make it.
