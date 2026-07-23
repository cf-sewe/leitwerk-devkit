---
name: leitwerk-onboard
description: >
  Bring an existing (brownfield) or new repository under Leitwerk. Establishes
  the constitution, blast-radius tiers, and the verify gate. Use at the start of
  adopting the framework on a codebase, or when onboarding a new area of it.
allowed-tools: "Read Grep Glob Bash Write Edit"
---

# Onboard a repository onto Leitwerk

Goal: leave the repo with a working `leitwerk verify`, a human-owned
constitution, and a tier map that reflects real blast radius. Do not invent
facts — read the code and ask the human only where intent is genuinely unknowable
from the repository. Label every fact derived from the code CONFIRMED (verified
by reading/running) / INFERRED (reasoned) / GAP (unknown); INFERRED and GAP
need human validation before they gain authority.

**New/empty repo vs brownfield — pick the path first.** The numbered steps are
the *brownfield* survey path (there is existing code to understand). On a **new or
empty repo there is nothing to survey** — do not invent a structure. Go
intent-first: run `leitwerk init`, then hold the *constitution interview*
(whitepaper §8.1 step 1 — stack, principles, tier rules, boundaries) and set the
`[paths]` tiers from that declared intent. Path globs match files that do not
exist yet (`**/db/migrations/** = T2` fires the moment the first migration lands),
and checks skip cleanly until a toolchain is wired — so the concrete tier map,
real checks, and characterization tests **accrete per change** through the normal
spec→plan→build loop, not up front. Steps 2 and 4 below then apply as code arrives.

## Steps
1. **Scaffold** if absent: run `leitwerk init` to create `leitwerk/constitution.md`
   and `leitwerk/tiers.conf` from the templates.
2. **Map blast radius.** Survey the tree. Identify irreversible/infra/data paths
   (migrations, IaC, billing, auth) and record them as T2 in `leitwerk/tiers.conf`.
   Everything state-mutating is T1; read-only/display is T0.
3. **Wire real checks.** Add project checks in the repo's own `leitwerk/checks/`
   (one `<name>.sh` per check; repo-local overrides the built-in per check, so
   never edit the installed `core/checks/`). Wire them to the project's actual
   toolchain (build, test, lint, SAST). A check that has nothing to run must
   `exit 2` (skip), never fake a pass.
4. **Characterize existing behaviour.** For brownfield code with no tests around
   a risky area, add characterization tests so the gate has an oracle before any
   change is made.
5. **Draft the constitution.** Fill invariants, DoD, and the role ensemble this
   project needs. Keep it to non-obvious facts. The human reviews and owns it.
6. **Prove it.** Run `leitwerk verify --tier T2` and confirm it executes end to
   end. Report what is enforced vs. still skipped.

**A pre-existing `specs/` (or other docs) in another structure.** `leitwerk/` is
a *namespaced, additive* layer: `leitwerk init` never migrates or touches an
existing top-level `specs/`, and `drift` only tracks anchors under
`leitwerk/specs/`. Do not bulk-convert. Keep the project's own specs as they are
and use `leitwerk/specs/` for new work; if the human wants an existing spec under
the gate, convert *that one* to the Leitwerk shape (Problem / Behaviour / Anchors)
when a change next touches its area — strangler-fig, never big-bang.

Hand back a short summary: what is now gated, what tiers map where, and which
checks are still stubs needing real tooling.
