---
name: leitwerk-spec
description: >
  Turn a request, an idea, or a bug that needs a contract change into a spec a
  change can be bound to. Use before planning or building any non-trivial change;
  a routine bugfix uses leitwerk-fix instead. Produces the anchor the gate and
  reviewers check against.
allowed-tools: "Read Grep Glob Bash Write Edit Task"
---

# Write or update a spec

A spec states the observable contract, not the implementation. It co-evolves with
the code; the gate keeps it honest.

## Steps
1. **Research first** — read every file the request mentions fully. Fan out
   `scout` subagents to locate and analyze the code the change touches —
   retrieve, don't preload: scouts return facts, not file dumps. Facts enter
   the spec as `file:line`, tagged CONFIRMED (read) or INFERRED (reasoned).
   Never ask the human what the repo can answer.
2. Copy `core/templates/spec.template.md` into `leitwerk/specs/<slug>.md`. It
   starts `draft`; only the human's approval flips it to `active` — before
   that it is not contract. If this spec realizes a roadmap item, set its
   `Roadmap: <slug>` line to that item's slug (a change record); delete the line
   for a living contract. `lifecycle` reds a present slug that is not an open
   roadmap item, so status stays derivable from the spec, not written prose.
3. **Problem** — what is wrong/missing and for whom; link the request.
4. **Behaviour** — given/when/then, including what must NOT happen and the edge
   cases. This is the part reviewers read instead of the diff.
5. **Design decisions** — decide implementation-level choices yourself and
   record them with rationale (chosen + rejected alternatives). Escalate a
   dimension to the human only when it sets intent, weakens a guarantee, or
   accepts irreversible risk — then one dimension at a time, options with
   trade-offs and a recommendation; the human decides. Do not fan design out
   to parallel agents — one context.
6. **Invariants touched** — name the constitution invariants nearby and how the
   change stays inside them.
7. **Blast radius** — set the tier (`leitwerk tier <path>` for the files you
   expect to touch) and state the worst case if it ships wrong.
8. **Acceptance checks** — the concrete oracles (tests/properties/contracts) that
   will prove correctness. These become part of `leitwerk verify` for this area.
9. **Anchors** (where the spec governs specific code) — list it under
   `## Anchors` as `` `path` `` or `` `path#symbol` `` so `drift` surfaces
   spec↔code divergence (a renamed symbol or moved code goes red).

If the change reveals the spec was incomplete, update the spec in the same change
— continuous bidirectional refinement. If code and spec disagree and you cannot
tell which is right, do NOT silently pick one: surface the drift for a human.
