# Spec — open-code guarantee-parity guard

## Problem
Open-code compatibility is an architectural property, not a feature: it holds as
long as the hard guarantee lives in `core/` and is reachable via `leitwerk verify`
+ CI with no agent runtime. Nothing currently *enforces* that — it survives on
discipline. As Claude-Code-native ergonomics grow (e.g. workflows), gate logic
could leak into a binding and make open code silently second-class. That erosion
is cheap to prevent continuously and expensive to retrofit.

Scope note: the target is **guarantee-parity** (the hard gate), not
**ergonomics-parity**. Claude Code's workflow orchestration is not expected to
reach open code; only the guarantee is.

## Behaviour (the observable contract)
- A structural check (`parity`) fails the gate if the guarantee/binding boundary
  erodes:
  - `core/` executables/config reference `bindings/` (core must be tool-agnostic);
  - a `checks/` dir or a tiers file appears under `bindings/` (gate logic belongs
    in `core/` or the consuming repo's `leitwerk/`);
  - a binding launcher reimplements gate logic instead of delegating to core.
- CI proves parity on every commit by running the core gate with **no agent
  runtime** present (no Claude Code, no plugin).

## Invariants touched
Adds "open-code guarantee-parity" to the constitution, tying the existing
"core never depends on an agent runtime" and "bindings never reimplement the
gate" invariants to an enforced check.

## Blast radius
T2 — this is gate/policy infrastructure (`leitwerk/checks`, tiers, CI).

## Acceptance checks
- `leitwerk/checks/parity.sh` exists, is shellcheck-clean, and is wired into the
  T1/T2 tiers.
- Introducing a `bindings/**/checks/` dir, or gate logic in the launcher, turns
  `leitwerk verify` red.
- The CI workflow has a job that runs the core gate with no agent runtime.
- `leitwerk verify --tier T2` stays green on the clean repo.

## Out of scope
Live open-code validation under Codex, AGENTS.md tuning, any open-code
orchestration shim — deferred to roadmap M3.4 (do when the framework is stable).
