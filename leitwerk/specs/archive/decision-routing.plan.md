# Plan — decision-routing

Status: landed (2026-07-20) <!-- all steps landed -->

Spec: `leitwerk/specs/archive/decision-routing.md`. Verified references: whitepaper §10
(four rule cards, no during-work routing), `CLAUDE.md` + `CLAUDE.template.md`
(no escalation rule), `leitwerk-onboard` (already aligned), `leitwerk-plan`/
`leitwerk-build` (no human-asks present).

Step status: `[ ]` open · `[x]` done · `[~]` deviated — one line why.

## Steps
1. `[x]` **Always-on test** — `CLAUDE.md` (repo) and
   `core/templates/CLAUDE.template.md`: a compact "When to ask the human"
   block (three conditions, decide-and-record default, specialist-role rule,
   proposals inbox). Tier T0. Proves: `context` check stays green.
2. `[x]` **Skills** — `leitwerk-spec` (research: never ask what the repo can
   answer; design step: escalate only test-positive dimensions, with
   recommendation), `leitwerk-review` (summary reports escalations and
   recommendation-divergence as pruning signal). `leitwerk-onboard`, `-plan`,
   `-build` checked — no change needed. Tier T1.
3. `[x]` **Templates** — `spec.template.md` Design-decisions wording (agent
   decides and records unless the escalation test fires). Tier T0.
4. `[x]` **Open binding** — `bindings/open/AGENTS.md` working-method item
   "Escalate decisions, not questions". Tier T1.
5. `[x]` **README** — the "You own intent and judgment" bullet states the
   three escalation classes. Tier T0.
6. `[x]` **Whitepaper §10** — fifth rule card "Escalate decisions, not
   questions" + intro updated (three→four rules is wrong: count becomes five)
   + a short proposals-inbox paragraph. Tier T0.
7. `[x]` **Proposals** — constitution wording as a proposal file;
   `leitwerk/proposals/README.md` documents the inbox convention (pending
   constitution blessing). Tier T0.
8. `[x]` **Review + gate** — architect coherence pass over all touched files;
   `leitwerk verify --tier T2` green.
9. `[x]` **Proposal follow-up in the gate** — `leitwerk/checks/lifecycle.sh`
   counts open proposals in its summary and warns on files older than 30 days
   (timestamp prefix); oracle first in `selftest.sh` (output contains the
   proposals warning on a fixture). Tier T2. Added from human feedback.
10. `[x]` **SessionStart hook** — `.claude/settings.json` (repo) and the
    plugin `bindings/claude/hooks/hooks.json` surface the inbox into session
    context. Tier T1 (json).
11. `[x]` **Wizard + docs** — `leitwerk-review` step: present each open
    proposal as a native multiple-choice question (accept = authorization,
    then apply + delete); `leitwerk/proposals/README.md` (map onto
    multiple-choice; follow-up mechanics), AGENTS.md sentence, whitepaper §10
    paragraph extension. Tier T0/T1.

## Verification strategy
Mechanical: gate at T2 (`context` for the always-on additions; `lifecycle`
untouched). The routing behaviour itself has no oracle — the architect pass
checks internal consistency, and the human judges the outcome (fewer
irrelevant questions) over the following sessions.

## Risks & rollback
- Under-asking: a Class-H decision mistaken for Class-A ships without the
  human — mitigated by the conservative test wording (intent, guarantee,
  irreversibility are broad) and by recording every self-made decision in the
  spec, where review sees it. Rollback: `git revert` per file.
- Context budget creep — proven by the `context` check each run.

## Roles to wake
`architect` (coherence of the rule across seven surfaces). `security-reviewer`
n/a. `test-engineer` n/a (no code behaviour).
