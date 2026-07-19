# Spec — align Leitwerk with Claude Code steering primitives

Status: landed (2026-07-19) <!-- durable content: constitution decision of record -->

**Tier:** T2 (changes the core CLI and shell scripts).

## Problem
Anthropic's guidance on steering Claude Code ("Skills, hooks, rules, subagents,
and more") gives a precedence and a decision framework:

- *Absolute prohibitions must be hooks, not prose.* An instruction the model can
  read is not a guarantee; a hook that exits non-zero is.
- *File-specific constraints belong in path-scoped rules* (`.claude/rules/` with
  `paths:` frontmatter), not in always-on context.
- *Always-on team context belongs in CLAUDE.md* (kept small); *procedures belong
  in skills* (loaded on demand).

Leitwerk already follows the strongest of these (the gate is a Stop hook;
procedures are skills). Three gaps remained, all of them cases where Leitwerk
stated a rule as prose that the guidance says should be mechanical:

1. **Human-owned config was protected only by prose.** The constitution says an
   agent "may not edit it unilaterally" and "may not lower a threshold, remove a
   check, or downgrade a path's tier." Nothing enforced it — exactly the
   "prohibition written as an instruction" the guidance warns against.
2. **Tier discipline loaded only inside a skill body.** The behavioural reminder
   for high-blast-radius paths ("this is T2 — write the oracle first") lived only
   in the build skill, so it was absent unless that skill was invoked.
3. **Nothing surfaced the constitution to a fresh Claude session.** The
   constitution is the human-owned source of authority, but Claude Code does not
   read it unless a CLAUDE.md or rule points there.

## Target behaviour
1. **A deterministic guard on human-owned files.** The core CLI gains
   `leitwerk guard <path>`, which exits non-zero when a path is human-owned. The
   list of human-owned paths lives in the tiers file (`[human-owned]`), so it is
   itself human-owned and tool-agnostic. Claude Code enforces it with a
   `PreToolUse` hook on `Write|Edit` that blocks the edit (exit 2). Open-code has
   no universal pre-edit hook, so the equivalent is human review of those paths
   (CODEOWNERS / required reviewer); this is a guardrail, documented as such — the
   *core guarantee* remains the gate, which keeps full open-code parity.
2. **Path-scoped rules carry tier discipline.** `.claude/rules/` holds a rule
   scoped (via `paths:`) to T2 paths, so the reminder loads whenever those files
   are touched, skill or no skill. Rules are not plugin-packageable, so this ships
   as the repo's own governance plus an adopter template.
3. **A thin CLAUDE.md bridges to the constitution.** A short repo CLAUDE.md points
   at the constitution as authoritative, states the one gate rule, and defers
   procedures to skills — matching the guidance (small always-on file, procedures
   in skills). `leitwerk init` scaffolds it for adopters.

## Invariants touched
- *The gate config is human-owned* — now enforced, not just stated.
- *The core never depends on an agent runtime* — `guard` is pure shell; the
  Claude-payload parsing lives in the binding, not core.
- *Bindings never reimplement the gate* — the hook parses the Claude payload and
  delegates the decision to `leitwerk guard`; it does not fork the list.
- *Open-code guarantee-parity* — the guard is a guardrail (ergonomics-leaning),
  documented honestly; the hard guarantee (the gate) is unchanged.

## Acceptance
- `leitwerk guard leitwerk/constitution.md` exits non-zero with a human-readable
  message; `leitwerk guard src/app.py` exits 0. Covered by `selftest`.
- The `PreToolUse` hook blocks a Write/Edit to a human-owned file and allows
  others. (Verified by the binding; asserted mechanically at the CLI layer.)
- `leitwerk verify --tier T2` stays green (json/shell/drift/selftest/parity, as
  of landing; the `context` check was wired at every tier by a later change).

## Roles
`architect` (CLI surface + boundary), `test-engineer` (extend `selftest`),
`security-reviewer` not required (no untrusted input beyond a path string).
