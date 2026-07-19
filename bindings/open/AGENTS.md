# AGENTS.md — Leitwerk

This file steers any AGENTS.md-aware coding agent (Codex, Copilot, Cursor,
Aider, Windsurf, Zed, and others). Copy it to a repo's root and edit the
project-specific parts. Nested `AGENTS.md` files are supported: put one per
package; the closest one wins on conflicts.

## The rule that never bends
Every change must pass the gate before it lands:

```
leitwerk verify --tier <T0|T1|T2>
```

There is no hook system common to all these tools, so the gate is enforced in
**CI** (see `ci/leitwerk-verify.yml`) as a required status check. A local run is
a prediction; the CI run is the record. Do not merge on a red gate.

## Working method
1. **Spec first.** Non-trivial work starts from a spec in `leitwerk/specs/`
   (template: `leitwerk/templates/spec.template.md`). It states the observable
   contract, the invariants touched, and the blast-radius tier.
2. **Plan into gated steps.** Each step leaves the gate green on its own.
3. **Write the oracle before new behaviour.** Bugs get a failing regression test
   first; untested legacy gets characterization tests before it is touched.
4. **Verify at the step's tier.** `leitwerk tier <path>` tells you the tier.
5. **Surface drift, do not resolve it.** If code and spec disagree and you can't
   tell which is right, stop and flag it for a human.

## Roles
Leitwerk defines specialist roles (architect, test-engineer, security-reviewer,
scout). Codex maps these to custom agents in `.codex/agents/*.toml`; other tools
apply them as review lenses. Wake them by signal:

- test-engineer — whenever behaviour changes.
- security-reviewer — auth, tenant/data boundaries, external input, infra (all T2).
- architect — cross-subsystem or structural change.
- scout — cheap read-only retrieval; use the smallest model.

## Blast-radius tiers
- **T0** read-only / display — light checks.
- **T1** state-mutating application code — behaviour + drift checks.
- **T2** irreversible / infra / data (migrations, IaC, billing, auth) — all
  checks plus SAST, dependency policy, and an explicit rollback.

## Constitution
Project invariants and Definition of Done live in `leitwerk/constitution.md`.
It is human-owned. Propose changes to it; do not edit it unilaterally.
