# Plan — lifecycle-check

Status: landed (2026-07-20) <!-- all steps landed -->

Spec: `leitwerk/specs/archive/lifecycle-check.md`. Verified references: the `drift`
placeholder (`core/checks/drift.sh:12-15`), the selftest structure
(`leitwerk/checks/selftest.sh`, sections 1–5), and a grep confirming
`examples/scenarios/` hardcodes no check lists.

Step status: `[ ]` open · `[x]` done · `[~]` deviated — one line why.

## Steps
1. `[x]` **Oracle** — `leitwerk/checks/selftest.sh` section 6: fixture
   assertions (consistent fixture → 0; landed outside `archive/` → 1; missing
   `Status:` line → 1; no specs dir → 2). Tier T2 (`*.sh`); proven by the
   gate run itself.
2. `[x]` **Check** — `leitwerk/checks/lifecycle.sh` implementing the spec's
   red/warn/skip rules. Tier T2. Manual check: run once against a broken
   fixture and read the messages for clarity.
3. `[x]` **Wiring** — `leitwerk/tiers.conf` `[tiers]`: append `lifecycle` to
   T0/T1/T2. Human-owned; the constitution permits adding a check; applied via
   staged copy under the human's in-session directive.
4. `[x]` **Docs** — `self.md` (check list + acceptance), README (tree comment,
   enforcement sentence in "How specs age"), whitepaper ("Keeping artifacts
   current" passage + enforcement table after Figure 3; §9.2 enumeration and
   refreshed gate excerpt; §13 lifecycle bullet updated honestly). Tier T0/T1.

## Verification strategy
The oracle lands with the check (selftest is the executable contract). The
red path is proven first against a fixture, then the gate runs green on the
clean repo.

## Risks & rollback
- False reds mid-session — mitigated by the warn tier for the legitimate
  complete-but-unlanded window; rollback: remove `lifecycle` from `[tiers]`
  (human-owned, one line).
- `date` portability for the aging warning — dual BSD/GNU fallback, silently
  skipped when unavailable.

## Roles to wake
`test-engineer` (fixtures/oracle). `architect` not woken — a single additive
check, no boundary moved. `security-reviewer` n/a (no untrusted input).
