# Spec — workflow-syntax-check: the gate parses the review workflow

Status: landed (2026-07-20) <!-- closes the workflow-orchestration JS-syntax-check follow-up; durable content stays here -->

## Problem
`.claude/workflows/leitwerk-review.mjs` (and its shipped template
`core/templates/workflows/leitwerk-review.mjs`) is JavaScript that no check
parses. A syntax error would surface only when a human invokes `/leitwerk-review`
— long after the change that broke it landed green. The workflow-orchestration
decision (constitution, 2026-07-19) named this as an open follow-up. `selftest`
already validates the workflow (it diffs the template against the scaffolded
copy, §3), so it is the natural home for a syntax check too.

Constraint (CONFIRMED by running `node --check`): a workflow script is not
standalone-parseable JS. It has one module-level `export const meta = …` AND
top-level `await`/`return` (the body runs inside Claude Code's async wrapper).
`node --check` in script mode reports "Illegal return statement"; module mode
rejects the top-level `return`. So a plain `node --check` would red the gate on
a *valid* workflow.

## Behaviour (the observable contract)
- `selftest` syntax-checks every `*.mjs` under `.claude/workflows/` and
  `core/templates/workflows/` by mirroring the runtime: strip a leading `export`
  from the `meta` declaration, wrap the body in an `async function`, and run
  `node --check` on that. A real syntax error anywhere in the body fails the
  gate (exit 1) naming the file; a valid workflow passes.
- Node is resolved via `mise` (the pinned toolchain) then a PATH `node`. If no
  node is present the check prints a note and is skipped — it never fakes a
  failure (nor a pass: it makes no claim it could not verify).

## Design decisions
- **Mirror the runtime (strip `export`, wrap in async fn), don't use a plain
  `node --check`.** The workflow dialect (module export + top-level
  await/return) is what Claude Code executes; a naive check gives a false
  positive on valid workflows. The transform is a small heuristic matched to the
  documented workflow contract ("begins with `export const meta`", "runs in an
  async context"). Rejected: `node --check` as-is (false red), and executing the
  workflow (side effects, needs a live agent runtime).
- **Extend `selftest`, not a new gate check.** A dedicated `workflows` check
  would need wiring into the human-owned `tiers.conf` (a proposal + human edit);
  `selftest` is already wired at every tier and already owns workflow-template
  validity (§3), so the check lands now without touching human-owned policy.
  Rejected/deferred: a standalone `workflows` check (a future refinement, via a
  tiers.conf proposal).
- **Soft-skip when node is absent.** The core's floor is "a shell present";
  node is not guaranteed everywhere. A skip-with-note keeps the honest-skip
  invariant. Rejected: hard-failing when node is missing.

## Invariants touched
- *A check never fakes a pass* — absent node → a visible skip note, not a green
  claim; present node → a real parse that can fail.
- *Bindings never reimplement the gate* — this validates the workflow (an
  ergonomics layer); the authoritative gate is unchanged.

## Blast radius
T2 (`selftest.sh`, gate behaviour). Worst case if wrong: the transform diverges
from the runtime and either misses an error (false green) or flags a valid
workflow (false red); mitigated by testing both the real workflow (passes) and a
deliberately broken one (caught).

## Acceptance checks
- `selftest` passes with the current (valid) workflow and its template; a
  deliberately broken `.mjs` (e.g. `const y = ;`) makes it exit 1 naming the
  file. Verified for both.
- `leitwerk verify --tier T2` stays green.

## Anchors
- `leitwerk/checks/selftest.sh`
- `core/templates/workflows/leitwerk-review.mjs`

## Out of scope
- Promoting this into a standalone `workflows` gate check (needs a tiers.conf
  proposal) — a later refinement.
- Linting/formatting workflow scripts, or type-checking them — syntax only.
