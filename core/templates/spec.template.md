# Spec — <feature>

Status: draft (YYYY-MM-DD) <!-- draft → active → landed YYYY-MM-DD → superseded by <slug> -->

A spec is the anchor a change is bound to. It co-evolves with the code and is
kept honest by the executable checks the gate runs. Write it so a reviewer can
tell, without reading the diff, what "correct" means.

A spec that describes a *change* (a migration, a rework) becomes history once it
lands — mark it `landed`, record its durable content in the constitution's
decisions of record or the area's living spec, and move it to
`leitwerk/specs/archive/` when superseded. Only `active` specs are current
contract; keeping that set small keeps the context agents load relevant.

## Problem
What is broken or missing, and for whom. Link the request/issue.

## Behaviour (the observable contract)
- Given … when … then …
- Edge cases and what must NOT happen.

## Invariants touched
Which constitution invariants this change is near, and how it stays inside them.

## Blast radius
Tier (T0/T1/T2) and the reason. What is the worst case if this ships wrong?

## Acceptance checks
The concrete oracles that prove it: which tests/properties/contracts must pass.
These become part of `leitwerk verify` for this area.

## Out of scope
What this change deliberately does not do.
