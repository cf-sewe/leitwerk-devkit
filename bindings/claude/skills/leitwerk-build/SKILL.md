---
name: leitwerk-build
description: >
  Implement one planned step and gate it. Use to execute a step from a
  leitwerk-plan. Runs the tier-selected checks and wakes the review roles the
  step's blast radius requires.
allowed-tools: "Read Grep Glob Bash Write Edit"
---

# Build one gated step

## Steps
1. Re-read the spec and the step in the plan. Confirm the tier with
   `leitwerk tier <path>` for the files you are about to touch.
2. **Write the oracle first** where behaviour is new: the test/property/contract
   that will prove the step, so the gate has something to check.
3. Implement the smallest change that satisfies the step. Match surrounding code.
4. Run the gate at the step's tier: `leitwerk verify --tier <T0|T1|T2>`.
   Iterate until green. The gate is deterministic — do not argue with a red
   result, fix the cause.
5. **Wake the roles the tier requires:**
   - T1+: spawn the `test-engineer` if the golden suite changed.
   - T2: spawn `security-reviewer` (auth/data/infra) and, for hot paths, a
     performance check. Delegate via the `orchestrator` when several roles apply.
6. If implementing revealed the spec was wrong or incomplete, update the spec
   now (bidirectional refinement). If code and spec conflict irreconcilably,
   stop and surface the drift to a human — do not pick a winner silently.

Do not end the turn on a red gate: the Stop hook runs `leitwerk verify` and will
block. That is intentional — it is the one thing no role can talk past.
