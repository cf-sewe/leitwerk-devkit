# Proposal — adopt two patterns from obra/superpowers into the roadmap

Decision needed: **you own roadmap priority** (`leitwerk/tiers.conf [human-owned]`),
so an agent may not add backlog items. This proposes adding **two** items —
already selected by you on 2026-07-21 — to `leitwerk/roadmap.md`, and asks you to
accept the placement/priority.

Source studied: `https://github.com/obra/superpowers` (Jesse Vincent's Claude
Code / multi-harness skills plugin). The two adopted patterns and the two
deliberately-skipped ones are recorded at the bottom for the record.

## The exact change

Insert the following two proto-spec entries into `leitwerk/roadmap.md`, in
**Milestone 3 — verification depth**, after `M3.7 · repo-map` and before
`### Milestone 4`:

---

```markdown
**M3.8 · skill-authoring-standard** · tier **T1**
- *Problem:* the `leitwerk-*` skills (`bindings/claude/skills/*/SKILL.md`) are
  written ad hoc — no shared standard for their frontmatter, description
  wording, size, or how they cross-reference each other. superpowers'
  `writing-skills` skill shows the value of a single authoring standard:
  triggering-focused descriptions, size budgets, explicit sub-skill markers.
- *Behaviour:* a `leitwerk-authoring` reference (skill or doc) states the
  standard and the shipped skills are brought into line with it:
  - descriptions in the third person, starting "Use when …", stating *when* to
    invoke (triggers/symptoms) not summarizing the workflow — a description that
    summarizes the workflow lets an agent follow the description instead of
    reading the skill;
  - size budgets (frequently-loaded skills stay small; heavy reference material
    moves to a co-located file rather than bloating SKILL.md);
  - explicit cross-reference markers between skills (e.g. a "REQUIRED SUB-SKILL"
    line) instead of context-burning `@`-includes;
  - required frontmatter (`name`, `description`, `allowed-tools`) validated.
- *Acceptance:* the standard exists as a referenceable artifact; every shipped
  `leitwerk-*` skill conforms (checkable — description prefix, presence of
  required frontmatter, size ceiling); a lint or `selftest` case fails a skill
  that violates the standard, so conformance is mechanized, not prose.
- *Roles/skills:* `architect` (standard + which rules are machine-checkable).

**M3.9 · agent-eval-harness** · tier **T2** (`core`/`selftest` + fixtures)
- *Problem:* `selftest` and `examples/scenarios/` prove **the gate** (the Go
  binary and checks). Nothing proves the **agent layer** — whether the skills
  trigger, whether the roles/workflow run in the right order, whether the
  bindings actually compose. That layer is currently inferred, the same gap
  M2.2 names for a single live run. superpowers treats skill triggering and
  ordering as testable: prompt fixtures driven at real subagents, asserting
  content and step order (`assert_contains`, `assert_order`) against responses.
- *Behaviour:* an eval harness that drives fixture prompts at the plugin's
  skills/agents and asserts observable behaviour — a skill triggers on its
  intended prompt (and not on an unrelated one), a workflow spawns its role
  `agentType`s, a role's output precedes another's where the process requires
  it. Evals are advisory to the gate (they need an agent runtime), mirroring how
  workflow verification is soft while the external gate stays authoritative;
  they run on demand and in the live-validation lane, not as a per-change block.
- *Acceptance:* a first eval suite with at least a trigger case and an
  ordering case that pass against the shipped plugin; a deliberately mis-worded
  skill description fails its trigger eval (proving the eval discriminates).
- *Roles/skills:* `test-engineer` (harness + fixtures), `architect` (advisory
  vs authoritative boundary). Pairs with M3.8 (the standard is *what good looks
  like*; the eval *proves it*) and feeds M2.2 / M4.1.
```

---

## How to accept
Approve, and the agent inserts the block above verbatim (or with any
placement/priority change you note) into `leitwerk/roadmap.md`, then deletes
this proposal. Promotion of either item to an active spec follows separately via
`leitwerk-spec`; the tiers above are proto-spec estimates and are re-derived at
spec time from the files actually touched (cf. cli-publish, where the estimate
moved T1→T2).

## What rejection means
Delete this file; nothing is added to the roadmap. The analysis stays in the
session transcript only.

## Recommendation (default on a plain approval)
**Accept both into Milestone 3, M3.8 before M3.9.** M3.8 is low-effort (T1,
mostly convention + a lint) and defines the target the M3.9 eval enforces, so
sequencing the standard first is cheap and makes the eval's assertions concrete.
Both raise verification depth of the agent layer, which is currently unproven —
consistent with the framework's "make the claim true with running code" thesis.

Maps to the question: *Add M3.8 + M3.9 to Milestone 3 as written?* →
(a) yes, both, M3.8 first **[recommended]**; (b) yes, but reprioritize / different
milestone; (c) only M3.8 (standard) now, defer the eval; (d) reject.

## For the record — patterns considered, not adopted
- **Discovery-by-coercion** (`<EXTREMELY_IMPORTANT>` bootstrap, "1% chance →
  you MUST invoke the skill", mandatory "Using [skill] to …" announcement).
  Skipped: Leitwerk already has a hard guarantee (the gate binary + Stop hook);
  forcing skill use through prose is redundant and off-thesis (mechanism over
  prose).
- **Own version-bump script + `.version-bump.json`.** Skipped: the release topic
  chose release-please + GoReleaser (see `leitwerk/specs/cli-publish.md` D2),
  which gives changelog/version discipline superpowers' hand-rolled bump does
  not.
- **Dated spec/plan filenames** (`YYYY-MM-DD-<slug>.md`). Skipped: Leitwerk's
  stable slug names carry the `lifecycle`/`drift` machinery (anchors, terminal
  states); dated names would churn those. (Proposals *do* use the timestamp
  prefix — that is this directory's own convention.)
- **Harness-neutral skill bodies + thin per-harness bindings** (kill the
  `bindings/claude` skills ↔ `bindings/open/AGENTS.md` method duplication).
  Not adopted *now* — you did not select it — but noted as the highest-leverage
  structural idea if the binding duplication becomes a maintenance problem.
