# Spec — orders: order totals

Status: active (2026-07-20) <!-- the reference-app's living contract -->

## Problem
The reference-app needs one small, real behaviour the gate can bind a spec and
a test to. Order-total arithmetic is enough: concrete, testable, and it carries
a data migration (the `orders` table) that exercises the T2 tier.

## Behaviour (the observable contract)
- `OrderTotalCents(items)` returns the sum of `UnitCents * Qty` over the line
  items, in cents.
- An empty or nil order totals `0`.
- Inputs are assumed non-negative (the caller's contract); the function does not
  validate — validation, if added, would be a new spec'd behaviour.

## Anchors
- `orders.go#OrderTotalCents`
- `db/migrations/001_create_orders.sql`

## Blast radius
T1 for the code (`orders.go`), T2 for the migration (`db/migrations/**`).
Worst case if the total is wrong: a demonstration app miscomputes — no real
users, but the test is what keeps it honest and the gate red on a regression.

## Acceptance checks
`go test ./...` (run by the `tests` check) passes on correct code and fails on a
regression to `OrderTotalCents`; `leitwerk verify` is green at the change's tier.

## Out of scope
Validation, currencies, persistence beyond the schema, and any web/API layer —
this app is a governance demonstration, kept minimal.
