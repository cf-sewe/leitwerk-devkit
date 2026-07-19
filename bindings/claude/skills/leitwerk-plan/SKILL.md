---
name: leitwerk-plan
description: >
  Turn an approved spec into a sequence of small, individually gated steps. Use
  after leitwerk-spec and before leitwerk-build for anything larger than a
  one-file change.
allowed-tools: "Read Grep Glob Bash Write Edit"
---

# Plan a change into gated steps

## Steps
1. Copy `core/templates/plan.template.md` into `leitwerk/specs/<slug>.plan.md`.
2. Break the work into steps small enough that each **leaves the gate green** on
   its own. Prefer a sequence that keeps the system shippable throughout
   (strangler-fig for replacements).
3. For each step record: files touched, tier (`leitwerk tier <path>`), and the
   checks that prove it.
4. **Verification strategy** — which oracles are added/extended and at which
   tier. New behaviour needs a new test before it is built.
5. **Risks & rollback** — per step; for T2 steps write the explicit rollback.
6. **Roles to wake** — which specialists review which steps and on what signal
   (see the trigger table in the whitepaper / constitution).

Keep the plan in the repo. It is a durable artifact, not chat scrollback.
