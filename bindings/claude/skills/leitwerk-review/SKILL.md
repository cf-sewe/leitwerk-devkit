---
name: leitwerk-review
description: >
  Final review before a change lands. Runs the full tier-selected gate and the
  specialist review roles, and checks the spec still matches the code. Use before
  opening a PR or merging.
allowed-tools: "Read Grep Glob Bash"
---

# Review a change before it lands

## Steps
1. Determine the change's highest tier across all touched files
   (`leitwerk tier <path>` for each). Review at that tier.
2. Run the authoritative gate: `leitwerk verify --tier <highest>`. It must be
   green. CI runs the same command — a local pass is a prediction, CI is the
   record.
3. **Spec fidelity** — read the spec and confirm the behaviour it promises is
   what the code does. Run `leitwerk drift` and reconcile anything it surfaces
   (or escalate the reconciliation decision to a human).
4. **Wake reviewers by signal, not ritual:**
   - always: `test-engineer` confirms the oracle actually exercises the change.
   - auth / data / external input / infra: `security-reviewer`.
   - user-visible surface: report what a human should eyeball (the one review a
     human still owns).
5. Summarize for the human reviewer: tier, gate result, roles run and verdicts,
   any drift surfaced, and what specifically needs human eyes. Provenance-tag
   claims as CONFIRMED (checked) vs INFERRED (reasoned) vs GAP (unverified).
