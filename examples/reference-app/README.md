# reference-app

A minimal repository already onboarded onto Leitwerk — the worked example the
whitepaper points to. It is a tiny Go service (order totals) with a spec, a real
test the gate runs, and a T2 database migration, so the gate demonstrates
**governance**, not just execution.

```
cd examples/reference-app
leitwerk verify --tier T1     # PASS — runs go test (real tests), go vet, gofmt, drift
leitwerk tier db/migrations/001_create_orders.sql   # -> T2 (a migration escalates the gate)
leitwerk verify --tier T2     # PASS — adds sast/erosion, which skip cleanly (no analyzer)
```

- `leitwerk/specs/orders.md` is the contract, anchored to `orders.go` so `drift`
  governs it.
- `leitwerk/checks/{lint,types,tests}.sh` are repo-local overrides of the
  built-in checks: they resolve Go via `mise` (the pinned toolchain) or a PATH
  `go`, and skip honestly if neither is present. This is the per-check override
  a consuming repo uses instead of editing installed core.
- Break `OrderTotalCents` (or the build) and the gate goes red — the test is the
  oracle. The devkit's own `selftest` and `examples/scenarios/s6-reference-app.sh`
  pin both the green run and the broken→red case.
