# Scenarios — executable examples of the framework's guarantees

Each scenario is a self-contained script that builds a throwaway fixture repo,
performs one action an adopter (or an agent) would perform, and asserts the
observable outcome the framework promises. They are documentation that cannot
silently go stale: `run-all.sh` executes them, and the repo's own gate runs it
via the `selftest` check, so a regression in any scenario turns CI red.

Run them against a built CLI:

```bash
make -C ../../core build
./run-all.sh ../../core/bin/leitwerk
```

| # | Scenario | Claim it proves |
|---|---|---|
| s1 | tier-escalation | Paths map to blast-radius tiers: a migration is T2, docs are T0, app code is T1 (default policy as scaffolded by `leitwerk init`). |
| s2 | red-gate | A failing check turns the gate red (`exit 1`, `gate: FAIL`) — a broken change cannot land. |
| s3 | human-owned-guard | `leitwerk guard` blocks (`exit 3`) edits to human-owned files (constitution, tiers.conf) and allows ordinary files. |
| s4 | skip-honesty | A check with nothing to run reports a visible skip, never a fake pass; the gate stays green but says so. |
| s5 | local-override | A repo-local `leitwerk/checks/<name>.sh` overrides the built-in check of the same name. |
| s6 | reference-app | The bundled `examples/reference-app` runs real Go tests (not a skip) and a deliberately broken change turns the gate red. Copies the app to a throwaway dir; skips cleanly if no Go toolchain is present. |

Conventions:

- A scenario takes the CLI path as `$1`, works only inside a `mktemp -d`
  fixture, and cleans up after itself.
- It prints one `PASS: …` line on success and exits non-zero with a `FAIL: …`
  line on the first broken assertion.
- Scenarios assert the *external contract* (exit codes, output markers, file
  effects) — never internals — so they stay valid across reimplementations.
