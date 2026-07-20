# Spec — bugfix-workflow: reproduce → pin → fix as a first-class entry path

Status: landed (2026-07-20) <!-- change record; landed at the T2 review 2026-07-20 -->

## Problem
The whitepaper defines three workflows that share the core loop but enter it
differently; **Workflow C — Bugfix** (`docs/whitepaper.html:673-681`, CONFIRMED)
is the lightest-rigor path: *reproduce & localize → pin with a failing test →
fix at the tier → gate (regression + tier checks) → review proportional to
risk*. No skill implements that entry path (CONFIRMED — a repo-wide grep for
bugfix/reproduce/pin/workflow C finds only prose in the whitepaper, the
`test-engineer` role, and one line of `bindings/open/AGENTS.md`). A defect today
runs through `leitwerk-spec` + `leitwerk-build`, which never make the
reproduce-first / pin-before-fix discipline explicit — so the one thing that
makes a fix safe (a test that fails *before* the fix and passes *after*, plus a
characterization net around the touched code) is left to memory.

The `test-engineer` agent already carries the rule
(`bindings/claude/agents/test-engineer.md:13-16`, CONFIRMED: "bugs get a
regression test that reproduces the bug first" + characterization). What is
missing is the *entry path* that composes it — a skill a human or agent reaches
for when the task is "fix this bug", and an executable demonstration that the
pin actually reds the gate and the fix greens it.

M1.5 (`leitwerk/roadmap.md`) authorizes this and leaves the shape open ("own
skill or a `leitwerk-build` variant"); acceptance is a documented run on
`examples/reference-app` (the M1.2 substrate, now in place).

## Behaviour (the observable contract)
A new skill `leitwerk-fix` is the bugfix entry path. Given a defect report, the
skill drives Workflow C:

- **Reproduce & localize.** Confirm the defect against the code before touching
  it (read the touched area; a `scout` locates it when it is not obvious).
  Nothing is fixed until the defect is reproduced.
- **Pin, then fail.** Write a regression test that captures the defect and
  *fails on the current code* — plus characterization tests around the touched
  code where it is otherwise unpinned. The failing test is the anchor; for a
  T0/T1 fix it stands in for a full spec (lightest safe path, whitepaper P6). A
  fix that changes the contract escalates to `leitwerk-spec`.
- **Fix at the tier.** Apply the minimal change. `leitwerk tier <path>` selects
  the gate and which roles wake — the skill does not assume T1.
- **Gate & review.** `leitwerk verify` at the change's tier must go green (the
  new regression test passes, prior tests stay green); review is proportional to
  the tier, and a T2 fix requires human sign-off.

The skill **composes** the `test-engineer` role for the pin rather than
restating its rules, and mirrors `leitwerk-build`'s gate discipline for the fix
and `leitwerk-review` for landing (a skill cannot invoke a skill — the method is
mirrored, not delegated). `bindings/open/AGENTS.md` mirrors
the same entry path in its working method (open-code parity of *method*, not of
the skill artifact).

Edge cases / what must NOT happen:
- The skill must not fix before a failing test exists — a green-from-the-start
  "regression test" proves nothing. The demonstration exercises this: a bug that
  the *existing* suite does not cover stays green until the pin is added.
- It must not silently widen scope to a spec change; a contract change is an
  escalation to `leitwerk-spec`, recorded, not assumed.
- It adds no new always-on context beyond one skill frontmatter description
  (the `context` budget check still holds).

## Design decisions
- **Own skill `leitwerk-fix`, not a `leitwerk-build` variant.** Workflow C's
  distinguishing part is its *prefix* — reproduce, localize, pin — which fits
  neither `leitwerk-spec` (writes the contract) nor `leitwerk-build` ("implement
  one planned step … re-read the step in the plan"; a bugfix on the lightest
  path has no plan). A dedicated skill is discoverable, matches the five
  existing `leitwerk-<verb>` lifecycle skills and the whitepaper's named
  workflow, and its always-on cost is ~30 words of frontmatter (negligible
  against the 2000-token budget). *Rejected:* a `leitwerk-build` branch — buries
  the method, muddies build's single responsibility, and forces a plan a
  light-path fix does not have.
- **Name the pin, source it from the `test-engineer` charter.** The regression +
  characterization taxonomy lives in the agent
  (`bindings/claude/agents/test-engineer.md:13-16`); the skill wakes the role and
  names the pin "per its charter" rather than copying the rules, so the taxonomy
  has one source of truth. The skill adds only the Workflow-C sequencing the role
  does not own — pin *before* the fix, gate red *for that reason*. *Rejected:*
  restating the test-design rules in the skill (two sources of truth that drift).
- **Executable acceptance via `examples/scenarios/s7-bugfix.sh`, not a prose
  transcript.** A scenario is re-runnable, auto-discovered by
  `examples/scenarios/run-all.sh` (`:14`), and run by the repo's `selftest`
  check — so it is CI-verified documentation that cannot rot. It mechanically
  proves the Workflow-C claim on the reference-app. *Rejected:* a Markdown
  transcript in `docs/reviews/` — unverified, rots, not gated.
- **Demonstrate a bug the existing suite misses.** The scenario injects a latent
  defect that the reference-app's current tests (Qty 2, 3, nil) do not exercise,
  shows the gate stays green (the bug escaped — this is *why* Workflow C pins),
  then adds the failing regression test (gate red), then fixes (gate green).
  This shows all four gated Workflow-C steps, not just "break → red" (which s6
  already covers). *Rejected:* re-breaking the covered path — would not
  demonstrate the pin step.

## Invariants touched
- *Bindings never reimplement the gate.* `leitwerk-fix` is ergonomics — it wakes
  roles and sequences method; the authoritative gate is unchanged. The skill and
  its AGENTS.md mirror add no gate logic.
- *A check never fakes a pass.* The scenario skips honestly when no Go toolchain
  is present (mirrors s6) and never asserts a green it did not observe.
- *Open-code guarantee-parity.* The *guarantee* (gate + regression) is already in
  core; only the *method* is mirrored to AGENTS.md — ergonomics-parity, not a new
  guarantee, so parity is unaffected.

## Blast radius
Mixed, reviewed at the **highest** tier touched:
- `bindings/claude/skills/leitwerk-fix/SKILL.md`, `bindings/open/AGENTS.md` — T1
  (`bindings/**/*.md`).
- `examples/scenarios/s7-bugfix.sh` — **T2** (`**/*.sh`): a scenario is gate
  behaviour (it runs under `selftest`).
- spec + plan — T0 (`**/*.md`).

Highest tier = **T2**, so the change reviews at T2 and lands only on a T2
sign-off. Worst case if it ships wrong: a scenario that fakes its own pass would
let a broken bugfix demonstration look green (masking a regression in the
example), or the skill misleads an agent into fixing before pinning. Mitigated by
the scenario asserting both the red (pin) and the green (fix) transitions, and by
the T2 gate + panel.

## Acceptance checks
- `examples/scenarios/s7-bugfix.sh <cli>` run on a throwaway copy of
  `examples/reference-app`: baseline green → inject a latent bug (gate still
  green, bug uncovered) → add failing regression test (gate **red**, `gate:
  FAIL`) → apply fix (gate **green**, `go test green`). It SKIPs honestly with no
  Go toolchain.
- `examples/scenarios/run-all.sh` (hence `selftest`) discovers and passes s7.
- `leitwerk verify --tier T2` on the devkit stays green (all checks, incl.
  `context` budget within limit and `shell` clean on the new scenario).

## Anchors
- `bindings/claude/skills/leitwerk-fix/SKILL.md`
- `examples/scenarios/s7-bugfix.sh`

## Out of scope
- Auto-landing a T0 fix without review (whitepaper C step 5 mentions sampling) —
  the gate + Stop hook stay authoritative; sampling policy is not built here.
- A `.codex/` prompt/skill equivalent beyond the AGENTS.md working-method mirror.
- Changing the reference-app's committed code or its living spec — the scenario
  works on a throwaway copy only.
- Code-graph / structural localization (whitepaper C step 1 "localize via the
  code graph") — that is M3.7 (repo-map); here localization is read-the-area.
