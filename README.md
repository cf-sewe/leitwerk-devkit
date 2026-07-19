# leitwerk-devkit

Leitwerk is a spec-anchored, verification-gated framework for AI-native software
development: humans own requirements and review, agents do most of the building,
and a deterministic gate — not the agent's own judgment — decides what may land.

This repository is the framework itself: the gate, the workflow skills, the
specialist roles, and the bindings that adapt them to a specific agent tool.

## The one idea

Generation is cheap; **verification is the bottleneck**. So the framework's
center of gravity is a single deterministic command the generating agent cannot
edit for itself:

```
leitwerk verify --tier <T0|T1|T2>
```

Everything else — the skills, the roles, the specs — feeds this gate. The gate
is a plain CLI on purpose (see "Why a CLI", below).

## Layout

```
leitwerk-devkit/
├── core/                 # TOOL-AGNOSTIC. The gate + templates. No agent runtime needed.
│   ├── bin/leitwerk       #   the CLI: verify · tier · drift · init
│   ├── checks/            #   one script per check (types, tests, lint, sast, drift, erosion)
│   ├── leitwerk.tiers     #   blast-radius tiers -> checks, and paths -> tiers
│   ├── templates/         #   constitution / spec / plan
│   └── package.json       #   publishes the `leitwerk` bin as @cf-sewe/leitwerk
│
├── bindings/             # thin per-tool wrappers around the SAME core
│   ├── claude/            #   a Claude Code plugin (skills + agents + Stop-hook gate + bin/)
│   └── open/              #   AGENTS.md + Codex agents + CI gate (Codex/Copilot/Cursor/Aider/…)
│
├── docs/                 # the whitepaper and the design/research record
├── examples/reference-app/  # a repo already onboarded; run the gate and watch it pass
└── .claude-plugin/marketplace.json  # this repo is also a Claude Code marketplace
```

The layout mirrors the architecture: a tool-agnostic **core** plus **thin
bindings**. A binding never reimplements the gate — it only invokes and enforces
it. That is what lets the same guarantee hold across Claude Code, Codex, and CI.

## Why a CLI? (and how it fits each tool)

A natural question: if the workflow lives in skills and prompts, why is the gate
a separate command-line tool rather than just instructions?

Because instructions are advisory and a command is not. A skill is text the model
*chooses* to follow; it can be reasoned around, skipped under context pressure,
or have its "tests pass" claim hallucinated. `leitwerk verify` returns a real
exit code from real tools — compiler, test runner, SAST. **The model cannot talk
its way past a non-zero exit.** That is the whole point of a gate.

Three properties follow from it being a plain executable:

1. **Non-bypassable.** The result comes from tools, not from the agent grading
   its own homework (which the research shows does not work).
2. **Runtime-independent.** CI has no Claude Code and no Codex — it just runs
   `leitwerk verify`. The authoritative gate therefore never depends on which
   agent (or which model) produced the change.
3. **One source of truth.** The same binary is the authority everywhere, so the
   bar cannot drift between "what the agent ran" and "what CI runs."

How each surface *invokes* and *enforces* the same command:

| Surface | How the agent runs it | How it is enforced |
|---|---|---|
| **Claude Code** | plugin ships it on `PATH` (plugin `bin/`); agents call it via the Bash tool | a `Stop` hook runs `leitwerk verify \|\| exit 2` — exit 2 blocks turn-end, so a task cannot end on a red gate |
| **Open code (Codex, Copilot, Cursor, Aider…)** | `AGENTS.md` instructs the agent to run it | no universal hook exists, so **CI** runs it as a required status check — a red gate blocks merge |
| **CI (always)** | — | `leitwerk verify` is a required check on the protected branch; this is the authoritative record |

So the CLI is not a fourth thing bolted on — it is the shared, deterministic core,
and the skills/hooks/CI are just three ways of calling and enforcing it.

## Quick start

Run the gate against the bundled example:

```bash
export PATH="$PWD/core/bin:$PATH"
cd examples/reference-app
leitwerk verify --tier T0     # -> gate: PASS
```

Adopt on a real repo:

```bash
npm install -g @cf-sewe/leitwerk       # or add core/bin to PATH
cd /path/to/your/repo
leitwerk init                          # scaffolds leitwerk/{constitution.md,tiers.conf}
# then: Claude Code -> install the plugin (below); open code -> copy bindings/open/*
```

Install the Claude Code plugin from this repo as a marketplace:

```
/plugin marketplace add cf-sewe/leitwerk-devkit
/plugin install leitwerk@leitwerk
```

## Status

v0.1 scaffold. The gate runs and the tier logic works; most `core/checks/*` are
auto-detecting stubs that skip cleanly until wired to a project's real toolchain.
See `docs/` for the whitepaper and the design rationale.
