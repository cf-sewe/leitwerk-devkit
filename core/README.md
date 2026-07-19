# core — the tool-agnostic gate

No agent runtime depends on anything in here. This is the part that CI runs and
that every binding calls.

## `bin/leitwerk`

```
leitwerk verify [--tier T0|T1|T2]   run the checks selected for a blast-radius tier
leitwerk tier <path>                print the tier for a changed path
leitwerk drift                      surface spec<->code divergence (does not resolve)
leitwerk init [dir]                 scaffold leitwerk/{constitution.md,tiers.conf}
leitwerk version
```

Exit codes: `0` gate green · `1` a check failed (gate red) · `2` usage error.
(Hooks that must *block* wrap this as `leitwerk verify || exit 2`, because a
Claude Code hook blocks only on exit code 2.)

## `checks/`

One script per check. Each:
- `exit 0` — passed,
- `exit 1` — failed (gate goes red),
- `exit 2` — nothing to run here (skips cleanly; never a fake pass).

These are generic templates that auto-detect a toolchain. A consuming repo does
**not** edit them — it drops its own `<name>.sh` into its `leitwerk/checks/`,
which overrides the built-in per check (anything not overridden falls back here).
Set `LEITWERK_CHECKS` to point elsewhere. See this repo's own `leitwerk/checks/`
for a worked example.

## `leitwerk.tiers`

Two tables: `[tiers]` maps each tier to its checks (cumulative — T2 runs T0+T1+T2
checks); `[paths]` maps path globs to tiers (first match wins). This file is
human-owned: an agent may propose a check but may not lower a threshold or remove
one here.

## `templates/`

`constitution.md`, `spec.template.md`, `plan.template.md` — copied into a target
repo by `leitwerk init` and the phase skills.
