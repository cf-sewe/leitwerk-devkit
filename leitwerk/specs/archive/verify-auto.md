# Spec — verify --auto: derive the blast-radius tier from the diff

Status: landed (2026-07-20) <!-- change record; landed at the T2 review 2026-07-20 -->

## Problem
The Stop hook verifies at a **static** tier: `leitwerk verify --tier
${LEITWERK_TIER:-T1}` (plugin `bindings/claude/hooks/hooks.json:33`; this repo
pins T2 at `.claude/settings.json:33`). It cannot see the change's real blast
radius — a turn that touched a T2 path can end having run only T1 locally, and a
docs-only turn overpays by running everything. Only CI derives the tier from the
diff, and it does so in a **shell loop** duplicated in the workflow
(`bindings/open/ci/leitwerk-verify.yml:29-43`): rank each `git diff --name-only
base...HEAD` path through `leitwerk tier`, take the max, then `verify --tier`.
That logic lives outside the CLI, so the hook cannot share it (CONFIRMED by
reading both consumers).

M2.4 (`leitwerk/roadmap.md`) closes this: `leitwerk verify --auto` computes the
highest tier from the diff inside the CLI, so the Stop hook and CI share one
implementation. The constitution's drift decision (2026-07-20) named this as the
unblock for "auto-provisioning the diff base", with the living-contract
exemption for drift's one-sided check deferred to M2.3.

## Behaviour (the observable contract)
`leitwerk verify --auto` selects the tier from the changed files, then runs
exactly as `verify --tier <derived>` would:

- **Diff base resolution:** `--base <ref>` flag if given, else **working-tree
  mode**. `--auto` does not read `LEITWERK_DIFF_BASE` (that is drift's Part-2
  base — see Design decisions).
  - *base given:* changed set = `git diff --name-only <base>...HEAD` (three-dot,
    from the merge-base — identical to CI and to drift's Part-2 range).
  - *no base:* changed set = tracked working-tree changes (`git diff
    --name-only HEAD`) ∪ untracked files (`git ls-files --others
    --exclude-standard`) — what a local Stop hook sees before a commit. With no
    commits yet (no HEAD), every tracked + untracked file is the changed set, so
    a fresh repo derives a tier instead of blocking.
- **Tier derivation:** start at `T0`; for each changed path take its tier
  (`TierForPath`, falling back to `T1` on no-match — identical to `leitwerk
  tier`), and keep the maximum on the fixed ladder `T0 < T1 < T2`. An empty
  changed set → `T0` (matches CI's `highest=T0` initial value). The derived
  tier and the file that decided it are printed, e.g.
  `auto: tier T2 (db/migrations/001.sql); 4 changed file(s)`.
- **Then verify:** run the cumulative checks for the derived tier and return the
  same exit codes as `--tier` (0/1/2).

What must NOT happen (all are usage errors, exit 2 — never a silent
under-selection):
- `--auto` together with `--tier` — ambiguous.
- `--base`/`LEITWERK_DIFF_BASE` that is option-like (`-*`) or not a resolvable
  commit — refused (mirrors `drift.sh:38-52`), never interpolated as a git
  option, never silently treated as "no changes".
- `--auto` outside a git work tree / with no `git` — the precondition for
  deriving a diff is absent; it errors and names `--tier` as the explicit escape,
  rather than guessing a tier.

Consumers:
- **Plugin Stop hook** (`hooks.json`) → `verify --auto` (working-tree mode;
  `LEITWERK_TIER`/`--tier` overrides). This repo's `.claude/settings.json` stays
  pinned at T2 (see Design decisions).
- **CI** (`leitwerk-verify.yml`) → the shell tier-loop collapses to `leitwerk
  verify --auto --base "$base"`, falling back to `--tier T2` when the base does
  not resolve (a fresh branch). The `--base` flag selects the tier only; drift's
  one-sided check stays off until M2.3, which will add `LEITWERK_DIFF_BASE`
  (with the living-contract exemption) alongside the flag.

## Design decisions
- **Base comes from the `--base` flag ONLY — `--auto` never reads
  `LEITWERK_DIFF_BASE`.** An early design had `--auto` fall back to that env var
  to "unify" tier-selection with drift's Part-2 base. Review (architect, HIGH)
  showed that to be a footgun: the Stop hook calls `verify --auto` with no flag,
  so an ambient `LEITWERK_DIFF_BASE` (drift documents that very variable) would
  silently flip `--auto` from working-tree to committed-range semantics and hide
  a turn's *uncommitted* T2 edits → under-selection. So the base is flag-only;
  the env var stays drift's alone (the `drift` subprocess still inherits it, so
  Part-2 is independently controllable — deferred to M2.3). CI passes `--base`
  for the tier; M2.3 will *additionally* set `LEITWERK_DIFF_BASE` for drift — two
  explicit signals, no ambient coupling. *Rejected:* reading the env in `--auto`
  (the footgun); a new env var (fragments drift's base config).
- **Mirror CI's ranking exactly (T0<T1<T2, per-file `TierForPath` else T1, max),
  empty → T0.** `--auto` must be a drop-in for the shell loop it replaces, so CI
  behaviour is unchanged when it switches. The ladder is the framework's fixed
  blast-radius ladder; an unrecognized tier name is ranked highest so `--auto`
  errs toward more verification on a non-standard tiers file. Caveat (LOW): this
  "unknown ranks highest" rule assumes the fixed T0<T1<T2 ladder — off-ladder
  names are unsupported, not safely ordered (several would tie at the top rank,
  and a *defined* weak off-ladder tier could be selected). *Rejected:* inferring
  an order from the `[tiers]` file (fragile; the ladder is a fixed convention).
- **Precondition failures are usage errors (exit 2), not a fallback tier.** A gate
  must never silently under-verify; if `--auto` cannot compute a diff (no git,
  bad base) it refuses and points at `--tier`, rather than guessing low (unsafe)
  or forcing T2 (surprising overpay on every turn). An *empty* diff is not a
  failure — it is a real "nothing changed" → T0. *Rejected:* fail-safe-to-T2
  (turns a missing git into a full gate every turn); silent fallback to T1
  (under-verifies a real T2 change whose base was mistyped).
- **Working-tree mode includes untracked files.** A newly created migration is
  untracked yet carries blast radius; omitting untracked files would let a new
  T2 file end a turn at T0. *Rejected:* `git diff HEAD` alone (misses untracked).
- **This repo's Stop hook stays pinned at T2; the plugin ships `--auto`.** An
  early version dogfooded `--auto` here too. Review (architect, MEDIUM) noted that
  for the repo that IS the gate, `--auto`'s working-tree mode drops to T0 after a
  commit, so an edit-and-commit-within-one-turn T2 change would get only a T0
  turn-end locally (CI still catches it at merge, but the local pre-commit safety
  net is lost). The gate's own repo is the highest-stakes case, so it keeps the
  always-T2 turn-end; the acceptance ("the scaffolded/shipped hook uses
  `--auto`") is met by the plugin template and CI. `LEITWERK_TIER`/`--tier` is
  the manual override everywhere. *Rejected:* dogfooding `--auto` here (loses the
  gate repo's pre-commit T2 guarantee for the commit-within-a-turn pattern).

## Invariants touched
- *A check never fakes a pass* / *the gate never under-verifies.* Every
  ambiguous or unresolvable input errors (exit 2) rather than selecting a low
  tier; an unknown tier ranks highest.
- *Bindings never reimplement the gate.* The tier logic moves **into** the core
  CLI; the hook and CI templates become thin callers — strengthening the
  invariant (CI no longer reimplements ranking in shell). Parity intact: the
  guarantee stays in `core/`.
- *The gate config is human-owned.* `--auto` reads the tier map; it never edits
  it. Hook/CI templates change (bindings, T1); this repo's `.claude/settings.json`
  is not human-owned.

## Blast radius
T2. Core CLI (`core/cmd/leitwerk/main.go`, a new tier-derivation function in
`core/internal/gate/`) is T2; templates (`bindings/claude/hooks/hooks.json`,
`bindings/open/ci/leitwerk-verify.yml`) are T1; `.claude/settings.json` is T1.
Highest = T2. Worst case if wrong: `--auto` under-selects and a T2 change ends a
turn / merges having run only a lower tier — a silent weakening of the gate.
Mitigated by: precondition failures erroring (never under-selecting), unit +
integration tests over the derivation incl. a mixed diff, and the review panel.

## Acceptance checks
- Unit tests: the tier-derivation function returns T2 for a mixed set containing
  a migration, T0 for a docs-only set, T0 for an empty set, and ranks an unknown
  tier highest.
- Integration test: in a git fixture, `verify --auto --base <sha>` prints/derives
  `T0` for a docs-only commit and `T2` for a commit adding a migration; `--auto`
  with `--tier` is exit 2; a bad `--base` is exit 2; `--auto` outside a git repo
  is exit 2.
- `selftest` exercises `--auto` tier derivation from a mixed diff (docs-only → T0,
  one migration → T2), per the roadmap acceptance.
- The scaffolded/shipped Stop hook and CI template use `--auto`; `leitwerk verify
  --tier T2` on the devkit stays green.

## Anchors
- `core/cmd/leitwerk/main.go`
- `bindings/claude/hooks/hooks.json`
- `bindings/open/ci/leitwerk-verify.yml`

## Out of scope
- The living-contract exemption for drift's one-sided check, and switching CI to
  `LEITWERK_DIFF_BASE` to activate Part-2 — **M2.3**.
- Diff-signal triggers (auth paths → wake security review) — a later extension
  the roadmap notes only after tier derivation is proven.
- Changing the tier ladder, the checks, or any tier→path policy.
