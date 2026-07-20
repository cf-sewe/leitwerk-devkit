# Spec — drift-detection: anchor specs to code and surface divergence

Status: landed (2026-07-20) <!-- durable content: constitution decision of record (drift detection); T2 sign-off by the human in-session -->

## Problem
"Surface spec↔code drift, don't resolve it" is the framework's headline
principle, but the check that should enforce it is a placeholder:
`core/checks/drift.sh:12-15` (CONFIRMED) counts `*.md` files under the specs
dir and always prints "no unreconciled drift flagged" — it never reads a spec,
never looks at code, and cannot go red. The claim is therefore unproven, and
the whitepaper's drift narrative reads as aspiration. Related gap from M1.4:
`drift` has no notion of `leitwerk/specs/archive/` (CONFIRMED — the placeholder
walks the whole tree), so a landed change record would be scanned as if it were
current contract.

Two facts constrain the design:
- The golden contract for `drift` is asserted in
  `core/cmd/leitwerk/integration_test.go:212-221` (CONFIRMED): at repo root it
  exits 0 with a summary containing `spec(s) tracked`; with no specs dir it
  exits 2 (skip). This is the backward-compatible surface the change preserves.
- Checks inherit the parent environment and run from the repo root
  (`core/internal/gate/verify.go:86-91`, CONFIRMED), so an env-provided diff
  base reaches the check. CI computes such a base for *tier selection*
  (`.github/workflows/leitwerk.yml:40-46`, CONFIRMED) but does **not** yet pass
  it to the gate step — so Part 2 below is wired but not auto-fed a base in CI
  yet; that provisioning is M2.4 (a review finding corrected this; see Design
  decisions).

## Behaviour (the observable contract)
A spec declares the code it governs in an **`## Anchors`** section — a list
whose items each start with a backtick-wrapped token:

```
## Anchors
- `core/checks/drift.sh` — whole-file anchor (path must exist or a glob must match)
- `core/internal/gate/tiers.go#ChecksForTier` — symbol anchor (name must appear in the file)
```

- **Anchor token grammar:** `path` or `path#symbol`. `path` may be a literal
  path or a glob (`core/checks/*.sh`); a `symbol` anchor names a single concrete
  file (not a glob). `symbol` is a source identifier; line numbers are
  deliberately not an anchor kind (they churn without meaning). Anchor paths are
  repo-relative — an absolute or `..`-escaping path is refused (see hardening).
- **Resolution (Part 1 — always on):** for every non-archived spec that has an
  `## Anchors` section, each anchor must resolve:
  - a `path` anchor resolves iff the path exists, or (for a glob) at least one
    file matches;
  - a `path#symbol` anchor resolves iff the path is an existing file and
    `symbol` occurs as a whole word in it.
  An unresolved anchor is **red (exit 1)** with a specific line, naming the
  spec (with line number), the anchor, and why it failed:
  `drift: leitwerk/specs/self.md:42 anchor core/x.go#Foo — symbol 'Foo' not found in core/x.go`
- **One-sided change (Part 2 — when `LEITWERK_DIFF_BASE` is set):** the check
  computes the changed set `git diff --name-only "$LEITWERK_DIFF_BASE"...HEAD`
  (the same range CI uses). For each non-archived spec with anchors, if any
  anchored path is in the changed set but the spec file is not, that is
  one-sided drift — **red (exit 1)**, naming the code that moved and the spec
  that did not: `drift: core/checks/drift.sh changed since <base> but its spec
  leitwerk/specs/self.md did not — reconcile the spec or note why it is
  unchanged`. When `LEITWERK_DIFF_BASE` is unset (the ordinary per-turn gate),
  Part 2 does not run, so a working session is never blocked mid-edit.
- **Untrusted-input hardening.** A spec can carry attacker-influenced content
  (an adopter's PR), so: an anchor path that is absolute or escapes the repo via
  `..` is **red** ("escapes the repo"), never `stat`/`grep`-probed — otherwise
  the check would be a file existence/content oracle for arbitrary host paths
  (e.g. `/proc/self/environ#SECRET`). And `LEITWERK_DIFF_BASE` is validated: an
  option-looking value or an unresolvable ref is refused with a warning and
  Part 2 is skipped — never silently treated as "no changes" (a faked pass) and
  never interpolated as a git option.
- **Archived specs are ignored** (closes the M1.4 remainder): anything under
  `leitwerk/specs/archive/` is skipped by both parts — a landed record is not
  current contract and must not drift-gate.
- **Skip:** no specs dir → exit 2 (unchanged; a check never fakes a pass).
- **Green summary** still contains `spec(s) tracked` (preserves the golden
  contract) and additionally reports anchors resolved, e.g.
  `12 spec(s) tracked, 1 with anchors; 4 anchor(s) resolve; no drift`.

What must NOT happen:
- The check never edits a spec or the code — it surfaces and exits non-zero;
  the human reconciles (constitution invariant "drift is surfaced, not
  resolved").
- A spec with no `## Anchors` section is not an error — anchors are opt-in per
  spec, so existing specs keep passing until they adopt anchors.
- No dependency on any agent runtime or non-stdlib tool: `git`, `grep`, `awk`,
  `find`, `sort` only (the shell floor the other checks already assume).

## Design decisions
- **Anchor by named symbol and path, not by line number.** Line ranges churn on
  every edit and would make the check cry wolf; a symbol name is stable across
  reformatting and is what a reviewer actually means by "the spec governs this".
  Rejected: `path:line` anchors (brittle), and content hashes recorded in the
  spec (heavy, and every code edit would force a spec edit — the "forcing
  resolution" the framework rejects).
- **Textual (word-grep) symbol resolution, not semantic.** Language-agnostic and
  stdlib-only, consistent with the core's "shell present, nothing else"
  invariant. Known limitation (honest GAP): a lingering mention of a renamed
  symbol in a comment masks the rename. A real rename in code removes the token
  and goes red, which is the acceptance case. Rejected: per-language parsers
  (tree-sitter is roadmap M3.7, deferred; would break the stdlib-only floor).
- **Part 2 gated on `LEITWERK_DIFF_BASE`, red only then.** One-sided detection
  needs a range to be meaningful; running it always-on would either need a
  hard-coded base (wrong off a branch) or fire on every uncommitted edit
  (blocks the working turn). Gating on an explicit base is the pre-merge
  reconcile prompt while the local per-turn gate stays quiet. **Correcting a
  review finding:** CI computes a base for tier selection but does not yet pass
  `LEITWERK_DIFF_BASE` to the gate step, so Part 2 currently runs under
  `selftest` and any caller that sets the base — auto-provisioning it (CI env
  export / `verify --auto`) is deferred to M2.4, together with a living-contract
  exemption so a durable `active` spec like `self.md` is not nagged. Part 1
  (anchor resolution), which M1.1's acceptance actually requires, is always-on
  everywhere. Rejected: always-on Part 2 (noise); Part 2 as a mere warning (the
  roadmap wants non-zero when it fires); wiring CI in this change (it would
  activate the one-sided nag before the exemption exists).
- **Stays a core check, replacing the placeholder in place.** The roadmap pins
  it to `core/checks/drift.sh` and it is already wired at every tier for every
  adopter; fixing it in core is what makes the claim true for adopters, not just
  this repo. Rejected: a repo-local override (would leave adopters with the fake
  pass). Unlike `lifecycle` (repo-local-first by choice), `drift` was already
  core and gate-wired.
- **The repo dogfoods anchors on `self.md`.** The one living spec gains a small
  `## Anchors` section pointing at gate code it already describes in prose, so
  "consistent repo → green" is asserted against the real repo, not only
  fixtures.
- **This change crosses the "CLI handles untrusted input" line.** The
  constitution's "Roles in play" note deferred a `security-reviewer` until the
  CLI handled untrusted input; parsing spec markdown is that point. A security
  pass ran and hardened it (repo-path confinement, `git`-arg validation);
  command execution is not reachable (an anchor token cannot contain a
  backtick/newline, and `grep -wF --` neutralises option/regex injection).
  Flagged for the human: whether to record a `security-reviewer` role for
  `drift` in the constitution.

## Invariants touched
- *Drift is surfaced, not resolved* — this change is the mechanism behind the
  invariant; the check reports and exits non-zero, never edits either side.
- *A check never fakes a pass* — no specs dir still exits 2 (skip); a spec with
  no anchors is honestly "nothing to resolve", not a faked pass.
- *The core never depends on an agent runtime* — implemented in `core/checks/`
  with `git`/`grep`/`awk`/`find`/`sort` only; no runtime, no non-stdlib tool.
- *Bindings never reimplement the gate* — the logic lives in the core check; the
  `leitwerk drift` subcommand keeps running the built-in via `BuiltinScript`.
- *The gate config is human-owned* — this adds behaviour to an existing check;
  no tier/threshold/check-list change, so no `tiers.conf` edit is needed.

## Blast radius
T2 — `core/checks/drift.sh` is `*.sh` and gate behaviour, embedded in the
binary via `//go:embed`. Worst case if wrong: a false red blocks turn-end and
merges (fail-closed, annoying but safe); a false green is exactly today's
status quo (the placeholder). Because it now can go red, the risk is
false-positive noise, addressed by Part 2 being opt-in and symbol matching
favouring "resolves" on ambiguity.

## Acceptance checks
`selftest` (the executable oracle, added before the implementation) asserts
against fixtures, plus the existing Go golden suite:
- consistent fixture (all anchors resolve) → exit 0, summary contains
  `spec(s) tracked`;
- a spec anchoring a renamed/removed **symbol** → exit 1 with a readable line
  naming the spec, the anchor, and the missing symbol (the roadmap's headline
  case);
- a spec anchoring a **missing path** → exit 1;
- an **archived** spec whose anchor is broken → exit 0 (ignored);
- one-sided change with `LEITWERK_DIFF_BASE` set (git fixture: code committed,
  then changed without the spec) → exit 1;
- no specs dir → exit 2 (skip).
- `leitwerk verify --tier T2` stays green on the clean repo (self.md's anchors
  resolve), and `go test ./...` (the integration golden test) still passes.
- an anchor that is absolute or escapes via `..` → exit 1 ("escapes the repo");
  an option-like `LEITWERK_DIFF_BASE` → refused, Part 2 skipped, no file written.

## Anchors
- `core/checks/drift.sh`
- `leitwerk/checks/selftest.sh#drift`

## Out of scope
- Semantic/parsed symbol resolution and a repo symbol map (roadmap M3.7).
- Promotion of the archive-awareness pattern into a shared library used by both
  `drift` and `lifecycle` (each reads the archive convention locally for now).
- Auto-reconciliation of drift — a constitution non-goal, never in scope.
- `verify --auto` deriving `LEITWERK_DIFF_BASE` for local runs, wiring the base
  into CI's gate step, and a living-contract exemption for Part 2 — activated
  together (roadmap M2.4 / M2.3) so one-sided detection does not nag a durable
  `active` contract before the exemption exists.
