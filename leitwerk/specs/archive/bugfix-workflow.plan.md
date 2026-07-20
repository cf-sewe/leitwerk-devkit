# Plan — bugfix-workflow

Status: landed (2026-07-20) <!-- landed with its spec at the T2 review -->

Turns `leitwerk/specs/bugfix-workflow.md` into gated steps. The spec's `##
Anchors` forward-reference `leitwerk-fix/SKILL.md` and `s7-bugfix.sh`, so `drift`
stays red until **both** exist — the first green checkpoint is at the end of
step 2. The whole milestone lands as one gate-green commit (one strand), not one
commit per step.

## Steps
Step status: `[ ]` open · `[x]` done · `[~]` deviated — one line why.

1. `[x]` **Create the skill** `bindings/claude/skills/leitwerk-fix/SKILL.md` —
   T1 (`bindings/**/*.md`). Frontmatter (`name: leitwerk-fix`, folded
   `description:` ≤80 words / ~30 to match the other five, `allowed-tools`
   including `Task` for waking `test-engineer`). Body = Workflow C
   (`docs/whitepaper.html:673-681`): reproduce & localize → pin with a failing
   test (compose `test-engineer`, do not restate) → fix at `leitwerk tier` →
   gate + review; light path uses the failing test as the anchor, a contract
   change escalates to `leitwerk-spec`. Proves it: `context` budget (frontmatter
   word count, always-on total). *Gate not yet green here — scenario anchor still
   unresolved; do not end the turn.*

2. `[x]` **Create the scenario** `examples/scenarios/s7-bugfix.sh` — **T2**
   (`**/*.sh`). Modeled on `s6-reference-app.sh`: throwaway copy of
   `examples/reference-app` + `mise.toml`; SKIP honestly with no Go toolchain.
   Sequence proving all four gated Workflow-C steps:
   - baseline `verify --tier T1` green (`go test green`);
   - inject the latent bug `s/it.UnitCents \* it.Qty/it.UnitCents * (it.Qty % 100)/`
     in `orders.go` → `verify --tier T1` still **green** (existing tests use
     Qty 2/3; bug only bites Qty≥100 — it escaped the suite);
   - append a failing regression pin (Qty 100 → want 50000) to `orders_test.go`
     → `verify --tier T1` **red**, `gate: FAIL`;
   - revert the injection (the fix) → `verify --tier T1` **green**, `go test
     green` (pin passes, prior tests stay green).
   Injected + pin code must be gofmt- and vet-clean (lint/types run at T1) so
   only `tests` catches it — verified by running the scenario. Proves it:
   the scenario itself asserts red then green; `run-all.sh` discovers it;
   `shell` check clean. **After this step both anchors resolve** → run `leitwerk
   verify --tier T2` on the devkit and confirm green (first checkpoint).
   Manual eyeball (T2): read s7 and confirm every asserted transition is the one
   claimed and no branch can print PASS without observing it.

3. `[x]` **Mirror the method** into `bindings/open/AGENTS.md` — T1. Extend the
   working method (currently item 3, `:31-32`) so the bugfix entry path
   (reproduce → pin → fix at tier) is explicit for open-code, building on the
   existing "bugs get a failing regression test first" line — method parity, no
   new guarantee. Proves it: `context` budget, `verify --tier T1` green.

4. `[x]` **Review & land** — T2. `leitwerk verify --tier T2` green; `leitwerk
   drift`; spec fidelity; adversarial panel (roles via the documented fallback —
   `test-engineer`, `architect`, `security-reviewer` — as read-only agents,
   since the plugin's agent-types are not registered this session). Then the
   landing ritual: T2 sign-off (human), set spec `landed`, archive spec + plan,
   propose the roadmap move (M1.5 → done) + any durable core, commit gate-green.

## Verification strategy
- **New oracle:** `examples/scenarios/s7-bugfix.sh` — the executable acceptance;
  it is itself a red-then-green assertion (the pin fails the gate, the fix passes
  it). Runs under `run-all.sh` → `selftest` → the devkit gate at T1/T2.
- **Existing oracles unchanged:** `s6` still proves reference-app real; `context`
  still bounds the always-on surface; `shell` lints the new scenario.
- New behaviour (the skill's method) is documentation of an agent workflow; its
  mechanical core (pin reds, fix greens) is what s7 proves.

## Risks & rollback
- **Step 2, gofmt/vet false-red:** if the injected/pin code is not gofmt-clean,
  lint (not tests) reds the "still green" step for the wrong reason. Mitigation:
  chosen defect is single-line and gofmt/vet-stable; verified by running s7.
  Rollback: revert s7 (T2 file) — no product code changes.
- **Step 2, drift window:** between step 1 and step 2 the gate is red (scenario
  anchor unresolved). Mitigation: do not end the turn until step 2 lands; the
  two artifacts are one commit.
- **Coupling to reference-app tests:** if the reference-app later gains a Qty≥100
  test, s7's "still green after inject" step breaks loudly. Accepted: s7 works on
  a copy and would be updated with such a change; documented in the spec.
- **T2 rollback (whole strand):** `git revert` the single M1.5 commit — the skill,
  scenario, and AGENTS.md edit are additive; no human-owned or core files change
  except via the landing proposal.

## Roles to wake
- `test-engineer` — step 2 (the golden/scenario suite grows) and step 4 (the pin
  is a test-design question).
- `architect` — step 1 (skill shape / entry-path composition) and step 4.
- `security-reviewer` — step 4: light. The scenario runs `go test` on a temp copy
  and the skill wakes roles; no untrusted-input parsing is added, so the lens is
  "does s7 shell-inject or fake a pass", not a deep review.

## Review outcome (2026-07-20)
T2 gate green; panel run via the documented fallback (read-only role agents).
- *security-reviewer:* clean — mutation confined to the temp copy, no
  injection/eval, SKIP can only under-claim.
- *architect:* 2 findings, both fixed — (1) routing ambiguity: `leitwerk-spec`'s
  description narrowed to contract-changing bugs and pointed at `leitwerk-fix`
  (deviation: this added `leitwerk-spec/SKILL.md` to the change); (2) the pin
  taxonomy was deferred to the `test-engineer` charter and the spec's design
  decision reconciled; minor "hands to build" → "mirrors".
- *test-engineer:* sound (ran s7 + replayed stages); 1 LOW fixed — the pin stage
  now also asserts `go test green` is absent, proving the red came from `tests`.
Human: T2 sign-off granted; roadmap move + constitution decision-of-record
approved (staged-copy). No escalated decision diverged from the recommendation.
