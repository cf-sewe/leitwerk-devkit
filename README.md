# leitwerk-devkit

**Leitwerk is a way of building software with AI agents in which a deterministic
check — not the agent's own say-so — decides what is allowed to ship.**

## The problem

AI agents produce code faster than anyone can review it, so the bottleneck moves
from *writing* code to *knowing whether it is correct*. Two facts make that hard:

- An agent asked to grade its own work will claim a success it cannot back up.
- Reading every diff by hand does not scale to the volume agents generate.

The accept/reject decision therefore has to come from something outside the agent
that it cannot talk its way past.

## The idea

Put one deterministic command at the center; everything an agent produces must
pass it before it lands:

    leitwerk verify --tier <T0|T1|T2>

It returns an exit code from real tools — compiler, tests, SAST, structural
checks. Exit 0 lands; anything else is blocked (by a Claude Code hook while you
work, and by CI before merge). The agent cannot edit the gate to pass itself.

- **You own intent and judgment** — the requirements, the spec, and the review
  of what a human must eyeball. These are human-owned; the agent proposes changes
  to them, it does not make them. Only three kinds of decisions are escalated
  to you: intent and priorities, anything that would weaken a guarantee, and
  accepting irreversible risk — everything else the agent decides and records.
- **The agent owns generation** — most of the code, guided by the spec.
- **Spec and code co-evolve, bound together** — a change that makes them disagree
  is surfaced as drift, not silently resolved.
- **Checks scale with blast radius** — display code light (T0), state-mutating
  more (T1), irreversible / infra / data the most (T2).

```
     you -- intent, spec, review        agent -- most of the code
              \                               /
               v                             v
        +-------------------------------------+
        |   spec  <->  code  (co-evolve)      |
        +-------------------------------------+
                          |
                          v
        +-------------------------------------+
        |           leitwerk verify           |  deterministic + external:
        |    checks scale by blast radius     |  the agent cannot edit it
        +-------------------------------------+
                 |                     |
           red: exit != 0         green: exit 0
                 |                     |
                 v                     v
          back to the agent         it lands
```

## Why this is not a "research -> design -> plan" pipeline

A phase pipeline organizes *how the agent works* — research, design, planning,
each producing prose — and the outcome is judged, by the agent or a reviewer, at
the end. That structures generation; it does not constrain correctness, so a
convincing plan and a broken result can coexist.

Leitwerk keeps lightweight phases (spec -> plan -> build -> review) but puts the
external check at the center instead of the sequence of thinking steps. Every
change must pass the gate, every time, at its blast radius. Intent stays human,
generation is the agent's, and the decision to accept belongs to a tool rather
than to a judgment call.

This repository is the framework itself — the gate, the phase skills, the
specialist roles, and the bindings that adapt them to Claude Code or open-code
tools. The gate is a plain CLI on purpose (see "Why a CLI", below).

## Layout

```
leitwerk-devkit/
├── core/                    # TOOL-AGNOSTIC. The gate itself — runs with only a shell.
│   ├── bin/leitwerk          #   the compiled CLI (Go): verify · tier · guard · drift · init
│   ├── cmd/ · internal/      #   Go source: entrypoint + the unit-tested gate library
│   ├── checks/               #   one script per check (lint, types, tests, sast, drift, erosion)
│   ├── leitwerk.tiers        #   default policy: tiers→checks, paths→tiers, and [human-owned] files
│   ├── templates/            #   what `init` drops into a repo: constitution, spec, plan,
│   │                         #     CLAUDE.md, .claude/rules/, .claude/workflows/ — pieces a plugin can't carry
│   └── assets.go · Makefile  #   embed checks+templates into the binary; `make build`
│
├── bindings/                # thin per-tool adapters — they INVOKE core, never reimplement it
│   ├── claude/               #   Claude Code plugin: skills, role agents, bin/ (launcher + guard),
│   │                         #     hooks.json (Stop-hook gate + PreToolUse human-owned guard)
│   └── open/                 #   AGENTS.md + Codex agents + CI gate (Codex/Copilot/Cursor/Aider/…)
│
├── examples/reference-app/  # a repo already onboarded — run the gate and watch it pass
├── docs/                    # whitepaper + adoption guide + design/research record
│
│  # ── the devkit governing itself (dogfooding) — not shipped to adopters ──
├── leitwerk/                # THIS repo's own governance: constitution, tiers, specs, roadmap,
│                            #   and repo-local checks (json, shell, selftest, parity, context, lifecycle)
├── .claude/                 # THIS repo's Claude config: settings (hooks), rules, review workflow
├── .github/workflows/       # THIS repo's CI — the gate gating its own development
└── .claude-plugin/          # marketplace.json — this repo is also a Claude Code marketplace
```

The layout mirrors the architecture: a tool-agnostic **core** plus **thin
bindings**. A binding never reimplements the gate — it only invokes and enforces
it. That is what lets the same guarantee hold across Claude Code, Codex, and CI.

The bottom group is the repo applying Leitwerk to **itself**: `leitwerk/`,
`.claude/`, and `.github/` govern this repo's own development. They are the
reference for what an adopter's repo ends up looking like — not part of what gets
installed.

## How specs age

A spec's filename is its topic — `leitwerk/specs/<slug>.md`, with an optional
`<slug>.plan.md` alongside. Its age lives in the `Status:` line inside the file
(`draft → active → landed → superseded`), never in the filename; dated filenames
are reserved for frozen snapshots such as review reports. Change records are
perishable: when a change lands, the durable part merges into the area's living
spec or the constitution's decisions of record, and the file moves to
`leitwerk/specs/archive/` — the consolidation pass nicknamed "dreaming". Only
`active` specs are current contract, which keeps what an agent loads small and
relevant as a repository ages. The states are enforced, not just convention: a
`lifecycle` check turns the gate red when a `landed` record sits outside
`archive/`, a `Status:` line is missing, or a spec and its plan disagree — and
it flags plans that are ready to land and records overdue for the dreaming pass.

```
roadmap.md          leitwerk/specs/<slug>.md (+ <slug>.plan.md)      specs/archive/
(human-owned)

proposed --promote--> draft --build--> active --land--> landed --move--> archived
                                         |                |
                     the only state that is               | durable core merges into
                     current contract for agents          v the living spec / decisions
                                                            of record ("dreaming")
```

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

## Try it in 30 seconds

Run the gate against the bundled example:

```bash
make -C core build            # build the gate binary (Go toolchain pinned in mise.toml)
export PATH="$PWD/core/bin:$PATH"
cd examples/reference-app
leitwerk verify --tier T0     # -> gate: PASS
```

## Adding Leitwerk to your project

Adoption is layered: the **gate** comes first and is tool-independent; the agent
tooling sits on top and is optional. (`docs/adoption.md` has the full phased
version.)

### 1 · Make the gate available — required

The gate is a single static Go binary. Get it one of these ways:

```bash
# A · build from a checkout (Go toolchain pinned in mise.toml)
git clone https://github.com/cf-sewe/leitwerk-devkit && cd leitwerk-devkit
make -C core build
export LEITWERK_HOME="$PWD/core"; export PATH="$LEITWERK_HOME/bin:$PATH"

# B · go install — one self-contained binary (carries its checks/templates via embed)
go install github.com/cf-sewe/leitwerk-devkit/core/cmd/leitwerk@latest   # path provisional until published

# C · a prebuilt release binary placed on your PATH
```

Scaffold your repo and set your policy:

```bash
cd /path/to/your/repo
leitwerk init      # writes leitwerk/{constitution.md,tiers.conf}, CLAUDE.md, .claude/{rules,workflows}/
```

Edit `leitwerk/tiers.conf` so your irreversible paths (migrations, IaC, auth) are
T2, wire `core/checks/*` to your real toolchain, and add the CI gate
(`bindings/open/ci/leitwerk-verify.yml`) as a required check. **This alone gives
the guarantee — no agent involved.**

### 2 · Claude Code — install the plugin (optional ergonomics)

```
/plugin marketplace add cf-sewe/leitwerk-devkit
/plugin install leitwerk@leitwerk
```

**You do not hand-edit any `.claude` hooks — the plugin carries them.** Enabling
it puts `leitwerk` on the Bash `PATH` and activates two hooks automatically: a
`Stop` hook that runs the gate before a turn can end, and a `PreToolUse` guard
that blocks edits to human-owned files. It also provides the phase skills and the
role subagents.

Three things a plugin *cannot* package — Claude Code loads them only from your
repo — which is exactly why step 1's `leitwerk init` scaffolds them:

- **`CLAUDE.md`** — the small always-on file pointing at your constitution.
- **`.claude/rules/`** — path-scoped rules (e.g. T2 discipline).
- **`.claude/workflows/leitwerk-review.mjs`** — the fan-out review used on T2
  changes (opt-in; tailor its dimensions to your repo).

Note the plugin's launcher still calls the core CLI from step 1 — a marketplace
install copies only the plugin, not `core/`, so `LEITWERK_HOME` (or a
`go install`ed binary on PATH) must be present.

### 3 · Open code (Codex, Copilot, Cursor, Aider…) — optional

Copy `bindings/open/AGENTS.md` to your repo root and edit the project-specific
parts; add `.codex/` if you use Codex. There is no universal hook, so the CI gate
from step 1 is what enforces the bar.

### What ships where

| Piece | Delivered by | Auto-active in a session? |
|---|---|---|
| The gate (`leitwerk verify`) | core CLI (`make build` / `go install` / `LEITWERK_HOME`) | it *is* the guarantee |
| Governance: constitution, `tiers.conf` | `leitwerk init` | — (human-owned) |
| `CLAUDE.md`, `.claude/rules/` | `leitwerk init` | yes — Claude reads them |
| Phase skills, role agents | Claude plugin | on enable |
| Stop-hook gate + PreToolUse guard | Claude plugin (`hooks.json`) | **yes — no manual setup** |
| Review workflow (`.mjs`) | `leitwerk init` → `.claude/workflows/` | opt-in (ultracode) |
| CI gate, `AGENTS.md`, Codex agents | `bindings/open/*` (copy) | CI: yes |

## Status

v0.1 scaffold. The gate is a compiled, unit-tested Go binary and the tier logic
works; most `core/checks/*` are auto-detecting stubs that skip cleanly until wired
to a project's real toolchain. See `docs/` for the whitepaper and the design
rationale.
