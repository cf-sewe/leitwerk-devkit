---
name: leitwerk-spec
description: >
  Turn a request, bug report, or idea into a spec that a change can be bound to.
  Use before planning or building any non-trivial change. Produces the anchor the
  gate and reviewers check against.
allowed-tools: "Read Grep Glob Bash Write Edit"
---

# Write or update a spec

A spec states the observable contract, not the implementation. It co-evolves with
the code; the gate keeps it honest.

## Steps
1. Copy `core/templates/spec.template.md` into `leitwerk/specs/<slug>.md`.
2. **Problem** — what is wrong/missing and for whom; link the request.
3. **Behaviour** — given/when/then, including what must NOT happen and the edge
   cases. This is the part reviewers read instead of the diff.
4. **Invariants touched** — name the constitution invariants nearby and how the
   change stays inside them.
5. **Blast radius** — set the tier (`leitwerk tier <path>` for the files you
   expect to touch) and state the worst case if it ships wrong.
6. **Acceptance checks** — the concrete oracles (tests/properties/contracts) that
   will prove correctness. These become part of `leitwerk verify` for this area.

If the change reveals the spec was incomplete, update the spec in the same change
— continuous bidirectional refinement. If code and spec disagree and you cannot
tell which is right, do NOT silently pick one: surface the drift for a human.
