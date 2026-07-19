# Plan — <feature>

The plan turns a spec into a sequence of gated steps. Each step is small enough
to verify on its own and leaves the gate green.

## Steps
1. <step> — files touched, tier, which checks prove it.
2. …

## Verification strategy
Which oracles are added or extended (tests, properties, contracts, characterization
tests for existing behaviour) and at which tier they run.

## Risks & rollback
What could go wrong per step and how it is reverted. For T2 steps, the explicit
rollback procedure.

## Roles to wake
Which specialists review which steps, and on what signal.
