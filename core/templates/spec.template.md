# Spec — <feature>

Status: draft (YYYY-MM-DD) <!-- draft → active → landed YYYY-MM-DD → superseded by <slug> -->

A spec is the anchor a change is bound to. It co-evolves with the code and is
kept honest by the executable checks the gate runs. Write it so a reviewer can
tell, without reading the diff, what "correct" means.

A spec is one of two kinds: a **living contract** (an area's durable behaviour —
it stays `active` indefinitely and absorbs what changes teach) or a **change
record** (a migration, a rework — it passes through the lifecycle and is then
consolidated). The states, their triggers, and their owners:

- `draft` — being written; not yet binding.
- `active` — the human approved the spec; it is now part of the current
  contract. Only the human's approval makes a spec binding.
- `landed YYYY-MM-DD` — the change merged. Set at the landing review, which
  also merges the durable core into the area's living spec or the
  constitution's decisions of record (propose — those are human-owned) and
  moves the file plus its plan to `leitwerk/specs/archive/` — the "dreaming"
  consolidation; a periodic sweep catches what landing missed.
- `superseded by <slug>` — replaced; the line names the successor.

Only `active` specs are current contract; keeping that set small keeps the
context agents load relevant.

## Problem
What is broken or missing, and for whom. Link the request/issue. Ground claims
in the code: cite `file:line`, tagged CONFIRMED (verified by reading) or
INFERRED (reasoned).

## Behaviour (the observable contract)
- Given … when … then …
- Edge cases and what must NOT happen.

## Design decisions
The chosen approach, the alternatives considered, and why they were rejected.
The agent decides and records — unless the choice sets intent, weakens a
guarantee, or accepts irreversible risk; those the human decides, from options
presented with evidence and a recommendation.

## Invariants touched
Which constitution invariants this change is near, and how it stays inside them.

## Blast radius
Tier (T0/T1/T2) and the reason. What is the worst case if this ships wrong?

## Acceptance checks
The concrete oracles that prove it: which tests/properties/contracts must pass.
These become part of `leitwerk verify` for this area.

## Out of scope
What this change deliberately does not do.
