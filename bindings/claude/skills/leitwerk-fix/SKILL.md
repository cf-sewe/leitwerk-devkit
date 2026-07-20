---
name: leitwerk-fix
description: >
  Fix a defect by Workflow C: reproduce it, pin it with a failing regression
  test, then fix at the change's tier. Use when the task is a bugfix, not new
  behaviour — the lightest safe path the blast radius allows.
allowed-tools: "Read Grep Glob Bash Write Edit Task"
---

# Fix a bug (Workflow C)

The lightest safe path: optimize for speed at the least rigor the blast radius
allows, while still proving no regression. The failing test is what makes it
safe — write it before the fix, not after.

## Steps
1. **Reproduce & localize.** Confirm the defect against the code before you
   touch it — reproduce it (a failing input, a wrong output), then locate it by
   reading the area; fan out a `scout` when it is not obvious. Do not fix
   anything until you can make it fail on demand.
2. **Pin, then fail.** Wake the `test-engineer` to produce the pin — per its
   charter, a regression test that reproduces the defect plus characterization
   around the touched code. It must **fail on the current code**: run the gate
   and confirm it is *red for that reason* before you touch anything. A pin that
   is green before the fix proves nothing.
3. **Fix at the right tier.** `leitwerk tier <path>` for the files you touch
   selects the gate and which roles wake — do not assume T1. Apply the smallest
   change that makes the pin pass; match surrounding code.
4. **Gate = regression + tier checks.** `leitwerk verify --tier <T>` must go
   green: the new test passes, prior tests stay green, the tier's checks hold.
   The gate is deterministic — fix the cause of a red result, do not argue with
   it. Do not end the turn red; the Stop hook runs the same gate.
5. **Review proportional to risk.** Review at the change's tier (`leitwerk-review`
   for T2 or a multi-file change; a lighter pass for T0/T1). A T2 fix (auth,
   data, money, or the gate itself) requires human sign-off.

## Scope
- The failing test is the anchor for a T0/T1 fix — no full spec needed (lightest
  path). If the fix would **change the contract** (new behaviour, a relaxed
  guarantee), stop and go through `leitwerk-spec` first: that is a spec change,
  not a bugfix.
- If reproducing reveals the spec was wrong, update the spec in the same change
  (bidirectional refinement). If code and spec conflict and you cannot tell which
  is right, surface the drift to a human — do not pick a winner silently.
