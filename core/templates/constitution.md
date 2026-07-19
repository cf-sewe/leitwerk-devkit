# Constitution — <project name>

The constitution is the human-owned, durable contract for this codebase. Agents
read it every session; agents may *propose* changes to it but may not change it
without human approval. Keep it short and non-obvious — do not restate what the
code or the README already says.

## Purpose & scope
What this system is for, and the boundary of what it may do.

## Invariants (never violate)
- <e.g. every state-mutating endpoint is idempotent or explicitly guarded>
- <e.g. no tenant may read another tenant's data>
- <e.g. destructive operations require an explicit confirmation token>

## Blast-radius policy
Which areas are T0 / T1 / T2 and why (see `leitwerk/tiers.conf`). Name the paths
that are irreversible (migrations, infra, billing) so the gate escalates on them.

## Definition of Done
A change is done when `leitwerk verify` is green at its tier AND the woken review
roles have signed off. Nothing merges on a red gate.

## Roles in play
Which specialist roles this project uses and what wakes them (see the trigger
table in the whitepaper). Trim the ensemble to what this project needs.

## Decisions of record
Link the design decisions that constrain implementation. New decisions are
appended here, not scattered in code comments.
