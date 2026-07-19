# Spec — <feature>

A spec is the anchor a change is bound to. It co-evolves with the code and is
kept honest by the executable checks the gate runs. Write it so a reviewer can
tell, without reading the diff, what "correct" means.

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
