# Plan — phase depth: research, design, and progress in the phase skills

Status: landed (2026-07-20) <!-- all steps landed -->

Spec: `leitwerk/specs/archive/phase-depth.md`. All steps are Markdown-only — templates
and AGENTS-adjacent docs at T0, binding skills at T1 after the human raised
`bindings/**/*.md` to T1 mid-change. Every step leaves the gate green on its
own; steps 1–7 are independent and can land in any order, but the templates
(step 1) come first so the skills can point at sections that exist.

Step status: `[ ]` open · `[x]` done · `[~]` deviated — one line why.
(This plan uses the convention it introduces.)

## Steps
1. `[~]` (provenance note landed in Problem only, not Problem+Behaviour — one
   place suffices; flagged by the architect review) **Templates** — `core/templates/spec.template.md`: add a *Design
   decisions* section (chosen approach, alternatives, why rejected) and a note
   in Problem/Behaviour that facts carry `file:line` + CONFIRMED/INFERRED tags.
   `core/templates/plan.template.md`: document the per-step status convention
   and a manual-verification line for T2 steps. Tier T0. Templates are
   embedded via go:embed — run `make -C core build`; `selftest` re-embeds and
   proves the build. Checks: json/shell/selftest.
2. `[x]` **leitwerk-spec skill** — prepend a research step (read mentioned
   files fully; fan out `scout` for locate/analyze; facts enter the spec
   tagged) and add a design step (dimensions one at a time, human decides,
   outcome into *Design decisions*). Keep the addition ≤ 15 lines. Tier T0.
3. `[x]` **leitwerk-plan skill** — add the verification rule (no file/symbol
   reference enters the plan unverified against the code) and manual criteria
   for T2 steps. Tier T0.
4. `[x]` **leitwerk-build skill** — after each landed step, update the step's
   status line in the plan; record deviations in one line. Tier T0.
5. `[x]` **leitwerk-review skill** — extend the summary step: report plan↔
   implementation deviations (from the plan's status lines). Tier T0.
6. `[x]` **leitwerk-onboard skill** — label reverse-derived facts
   CONFIRMED/INFERRED/GAP; INFERRED/GAP need human validation (aligns with
   whitepaper §8.2). Tier T0.
7. `[x]` **bindings/open/AGENTS.md** — mirror the working-method additions
   (research before spec; design decisions recorded in the spec; plan status
   maintained) so process descriptions stay equivalent across bindings.
   Tier T0.
8. `[x]` **Roadmap proposals (human-owned — proposal only, no edit)** — hand
   the human two backlog entries to accept or reject: *bugfix-workflow*
   (whitepaper §8.3 reproduce→pin→fix as a skill; relates to M1.2) and
   *repo-map* (design-proposal O2; retrieve-don't-preload substrate).
9. `[x]` **Lifecycle definition + first dream pass** — added after review
   feedback that the lifecycle was described only fragmentarily: normative
   states/triggers/owners in `spec.template.md` (binding-neutral), the
   draft→active transition in `leitwerk-spec`, the landing consolidation in
   `leitwerk-review`; first pass executed (`go-cli{,.plan}` set to landed;
   five landed records moved to `leitwerk/specs/archive/`; references in
   editable files updated — the constitution/roadmap pointers are proposed
   to the human, who owns those files). Tier T0/T1.

## Verification strategy
No new mechanical oracles: the changes are procedure prose, and the spec's
acceptance is honest about that. Per step: `leitwerk verify --tier T2` green
(the `selftest` in step 1 proves the re-embedded templates build; `context`
proves always-on budgets are untouched). The advisory check is the review
panel plus the human reading the skill diffs — here the prose *is* the product.

## Risks & rollback
- **Skill bloat** — every line added loads whenever the skill is invoked.
  Mitigation: ≤ 15 lines per skill, no duplicated template content. Rollback:
  `git revert` per file; steps are independent.
- **Template/skill mismatch** — a skill pointing at a section the template
  lacks confuses adopters. Mitigation: step 1 lands first.
- **Process-description drift between bindings** — mitigated by step 7 in the
  same change.

## Roles to wake
- `architect` — coherence of the process changes across skills/templates and
  with the whitepaper's workflow prefixes (§8).
- `test-engineer` / `security-reviewer` — not required (no behaviour change in
  code, no untrusted input); the T0 tier matches.
- Human — owns the roadmap decision (step 8) and reviews the skill prose.
