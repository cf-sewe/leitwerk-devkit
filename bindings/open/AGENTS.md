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
   (template: `core/templates/spec.template.md`, shipped with the CLI). It states the observable
   contract, the invariants touched, and the blast-radius tier. Research before
   writing: read the files the task mentions fully, verify facts in the code
   (`file:line`, tagged CONFIRMED/INFERRED), and record design decisions —
   chosen approach and rejected alternatives — in the spec; escalate a design
   dimension only when the test in item 6 fires.
2. **Plan into gated steps.** Each step leaves the gate green on its own.
   Verify every file/symbol reference against the code before it enters the
   plan; keep each step's status line current (`[x]` done, `[~]` deviated —
   one line why) so a cold session can resume from the plan.
3. **Write the oracle before new behaviour.** Bugs get a failing regression test
   first; untested legacy gets characterization tests before it is touched.
4. **Verify at the step's tier.** `leitwerk tier <path>` tells you the tier.
5. **Surface drift, do not resolve it.** Declare the code a spec governs under
   its `## Anchors` section (`path` / `path#symbol`); the `drift` check goes red
   when an anchor stops resolving, or (given a diff base) when anchored code
   changed while its spec did not. If code and spec disagree and you can't tell
   which is right, stop and flag it for a human.
6. **Escalate decisions, not questions.** Ask the human only for intent, scope,
   and priorities; for anything that would weaken a guarantee (thresholds,
   checks, tiers); or to accept irreversible risk (T2 sign-off). Everything
   else: decide, record it in the spec's Design decisions, keep it reversible;
   use the specialist roles for domain judgment. Escalations carry options,
   evidence, and a recommendation — durable ones as `leitwerk/proposals/` files.
   The gate's `lifecycle` warning re-surfaces open proposals on every run;
   work the inbox off at review.

## Roles
Leitwerk defines specialist roles (architect, test-engineer, security-reviewer,
scout). Codex maps these to custom agents in `.codex/agents/*.toml`; other tools
apply them as review lenses. Wake them by signal:

- test-engineer — whenever behaviour changes.
- security-reviewer — auth, tenant/data boundaries, external input, infra (all T2).
- architect — cross-subsystem or structural change.
- scout — cheap read-only retrieval; use the smallest model.

Note: unlike Claude Code, open-code tools have no built-in workflow engine to
orchestrate these roles in parallel with adversarial verification. Run them
sequentially (or as review lenses) and rely on the CI gate as the hard,
authoritative check. The roles are advisory; `leitwerk verify` in CI is binding.

## Blast-radius tiers
- **T0** read-only / display — light checks.
- **T1** state-mutating application code — behaviour + drift checks.
- **T2** irreversible / infra / data (migrations, IaC, billing, auth) — all
  checks plus SAST, dependency policy, and an explicit rollback.

## Constitution and human-owned files
Project invariants and Definition of Done live in `leitwerk/constitution.md`.
It, `leitwerk/tiers.conf`, and `leitwerk/roadmap.md` are **human-owned**: propose
changes, do not edit them unilaterally. `leitwerk guard <path>` reports whether a
path is human-owned (the list is `[human-owned]` in the tiers file).

Claude Code blocks these edits with a `PreToolUse` hook; open-code tools have no
universal pre-edit hook, so enforce the same boundary with **required review** —
a CODEOWNERS entry (or equivalent) so a human must approve any change to those
paths. This is a guardrail. The hard guarantee is still the CI gate, which every
tool shares equally.
