# Spec — roadmap-spec-join: derive item status from the spec lifecycle

Status: draft (2026-07-23) <!-- draft → active → landed YYYY-MM-DD → superseded by <slug> -->

Roadmap: roadmap-spec-join

Realizes the `roadmap-spec-join` roadmap item: make the roadmap↔spec link
explicit and machine-checked so an item's status is *derived*, not written.

## Problem

The roadmap no longer stores status (world-state prose rotted behind the guard —
the `ci-live` incident: it claimed "no remote repo" after the repo was public and
CI green). Status must therefore be *derivable* by linking a spec to the roadmap
item it realizes. Today that link is prose only and unchecked:

- `leitwerk/specs/cli-publish.md` says "Promotes roadmap item **M2.1 · cli-publish**"
  in prose (CONFIRMED) — a stale M-number, not resolvable.
- `leitwerk/checks/lifecycle.sh` parses `Status:` and archive placement but reads
  **no** roadmap link (CONFIRMED, `lifecycle.sh:20-82`).
- Filename-slug matching is unreliable: roadmap item `spec-lifecycle` landed as
  `leitwerk/specs/archive/lifecycle-check.md` (CONFIRMED — slug ≠ filename).
- Roadmap items are `**<slug>**` headers (CONFIRMED, `roadmap.md`).

## Behaviour (the observable contract)

- A spec **may** declare the item it realizes as a `Roadmap: <slug>` line (a
  *change record*); a *living contract* (e.g. `self.md`) omits it.
- Given an active/draft spec (not under `archive/`, not a `*.plan.md`) that
  declares a real slug, when `lifecycle` runs, then the slug must appear as an
  open item (`**<slug>**`) in the roadmap — an orphan is **red**, naming the slug.
- Archived specs are **exempt**: a landed item has left the roadmap, so its
  `Roadmap:` slug will not resolve — mirrors `drift`'s archived-exempt rule.
- An unfilled template placeholder (`Roadmap: <slug>`) or any non-slug value
  (outside `[a-z0-9-]`) is **ignored**, never red — a fresh draft must not red.
- Absent `Roadmap:` line = nothing to check (not a fake pass): living contracts
  and roadmap-less specs stay green.
- Status derivation (documented, no new command): one pass over `specs/` —
  active spec = in flight, archived = landed, no spec for a roadmap slug = not
  started.
- Must NOT: force a link on living contracts; red a placeholder; check archived
  specs; require a network or a `leitwerk status` command.

## Design decisions

- **Link lives on the spec, resolves toward the roadmap** (spec→roadmap), exactly
  as `drift` runs spec→code. Keeps the human-confirmed roadmap thin; the machine
  link sits on the agent-owned spec.
- **Slug is the key**, not the M-number (roadmap items are slug-headed; numbers
  were retired).
- **Optional, not mandatory.** Not every active spec maps to a roadmap item
  (`self.md` is a living contract). Enforcing presence would be wrong; enforcing
  *resolution when present* is right.
- **Reuse `lifecycle`** rather than a new check — it already walks every spec,
  parses `Status:`, and knows `archive/`. Cheapest, one place, one pass.
- **Placeholder-tolerant** — validate only `[a-z0-9-]` slugs, so the template can
  carry a `<slug>` placeholder without redding a draft.
- **Roadmap path is overridable** (`LEITWERK_ROADMAP`, default
  `leitwerk/roadmap.md`) so the check is testable with a fixture, and skips the
  join when the file is absent.

## Invariants touched

- **A check never fakes a pass.** An absent link is "nothing to check" (green
  honestly), not a faked pass; a present-but-orphan link is red. Stays inside.
- **The gate config is human-owned.** Unaffected — this edits a *check script*
  (`lifecycle.sh`), not the tiers config; no threshold or tier changes.

## Blast radius

**T2** — edits `leitwerk/checks/lifecycle.sh` and `leitwerk/checks/selftest.sh`
(`**/*.sh` = T2). Worst case: a false red blocks unrelated work, or a false green
lets an orphan link slip. Both are contained to the `lifecycle` check; the hard
guarantees (`go build`/`go test`, other checks) are untouched.

## Acceptance checks

- `selftest` §6 gains lifecycle-join fixtures: with `LEITWERK_ROADMAP` set to a
  fixture roadmap, an active spec whose `Roadmap:` slug **is** an item → green; a
  spec whose slug is **absent** → red; a `<slug>` placeholder → green (ignored).
- Backfill: `cli-publish.md` declares `Roadmap: cli-publish`; `leitwerk verify
  --tier T2` stays green (the real roadmap contains `**cli-publish**`).
- `core/templates/spec.template.md` and the `leitwerk-spec` skill carry the
  `Roadmap:` convention.

## Anchors
- `leitwerk/checks/lifecycle.sh`
- `core/templates/spec.template.md`
- `bindings/claude/skills/leitwerk-spec/SKILL.md`

## Out of scope
- A `leitwerk status` command (rejected — derive on demand).
- `guard-confirm-class`, and the `tiers.conf` / `constitution.md` edits it gates.
- Thinning the existing fat roadmap items (done on the touch that lands them).
- Backfilling `Roadmap:` into archived specs (they have left the roadmap).
