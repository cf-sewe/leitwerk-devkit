---
name: leitwerk-review
description: >
  Final review before a change lands. Runs the full tier-selected gate and the
  specialist review roles, and checks the spec still matches the code. Use before
  opening a PR or merging.
allowed-tools: "Read Grep Glob Bash Task"
---

# Review a change before it lands

## Steps
1. Determine the change's highest tier across all touched files
   (`leitwerk tier <path>` for each). Review at that tier.
2. Run the authoritative gate: `leitwerk verify --tier <highest>`. It must be
   green. CI runs the same command — a local pass is a prediction, CI is the
   record.
3. **Spec fidelity** — read the spec and the plan, and confirm the behaviour
   the spec promises is what the code does. Run `leitwerk drift` and reconcile anything it surfaces
   (or escalate the reconciliation decision to a human).
4. **Run the review at the right weight for the tier:**
   - **T2 (or any large/multi-file change):** run the adversarial panel over the
     role subagents (`test-engineer`, `security-reviewer`, `architect`),
     independently refuting each finding. Prefer the saved `/leitwerk-review`
     workflow — `leitwerk init` scaffolds it into `.claude/workflows/`. If it is
     absent, or dynamic workflows are disabled, spawn the roles directly and
     sequentially instead; the panel is advisory, so the verdict is the same,
     only slower. This is *soft* verification (agents judging agents).
   - **T0/T1:** skip the panel; spawn only the roles the change warrants
     (`test-engineer` when behaviour changed; `security-reviewer` for
     auth/data/input) directly.
   - **Always:** the review is advisory. The authoritative word is the external
     gate in step 2 and the Stop hook — a red `leitwerk verify` blocks the change
     no matter how the panel voted.
   - **User-visible surface:** report what a human should eyeball (the one review
     a human still owns).
5. Summarize for the human reviewer: tier, gate result, roles run and verdicts,
   any drift surfaced, plan deviations (the `[~]` status lines and anything
   implemented that departs from the plan), and what specifically needs human
   eyes. Provenance-tag claims as CONFIRMED (checked) vs INFERRED (reasoned)
   vs GAP (unverified). Also list which decisions were escalated to the human
   and whether their answer diverged from the recommendation — a question
   class that never diverges is a pruning candidate: retire it to
   agent-decided at the next dream sweep.
6. **Lifecycle ("dreaming")** — when the change lands: set the spec's `Status:`
   to `landed YYYY-MM-DD`, merge its durable core into the area's living spec
   or the constitution's decisions of record (propose — human-owned), then
   move the spec and its plan to `leitwerk/specs/archive/`. Living specs stay
   `active`; only `active` specs are current contract, and keeping that set
   small keeps agent context relevant on a long-lived repo.
7. **Work the proposals inbox** — if `leitwerk/proposals/` is non-empty,
   present each pending decision to the human as a native multiple-choice
   question: the problem in one sentence, each option with a one-line
   description, the recommendation marked. An accepted option is the human's
   authorization — apply the change the file documents and delete the file;
   a rejected one is deleted; a deferred one stays (the gate's `lifecycle`
   warning keeps it visible).
