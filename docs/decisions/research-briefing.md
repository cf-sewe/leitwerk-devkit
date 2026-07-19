# Research Briefing: A Framework for AI-Native Application Development

**Date:** 2026-07-19
**Status:** Research synthesis for design discussion (pre-design)
**Author:** Research compiled via web search across ~70 sources (2023–2026)

## Purpose

Collect the current state of research and industry practice for building and
maintaining larger applications with AI coding agents, as input to a design
proposal. The goal is a rigorous, scalable process for AI-native development
where:

- Applications are too complex to build one-shot and must grow over time.
- Code quality, robustness, and security must not erode as the app grows.
- Requirements must stay current and enforced, not drift.
- Humans are involved in requirements and review (functional/visual), not in
  reading or writing most of the code.

Primary target applications: operational / control-plane tools.

## Design direction agreed so far

Captured from the initial discussion; these frame the eventual proposal:

1. **Baseline: spec-anchored** (spec and code co-evolve, bound by executable
   checks). Not full spec-as-source (regenerate-only) as the default.
2. **Human-authored code must remain a first-class path.** A programming expert
   on the team can hand-write code even though the majority is generated. The
   process must not break when it encounters human-written code.
3. **Continuous, bidirectional refinement.** A spec can be enhanced on the go —
   e.g. from human feedback after seeing test results. Spec ↔ code stays in
   sync in both directions, not just spec → code.
4. **Packaging: one or more new Claude Code skills.** Existing skills (research,
   plan, implement, validate) are retired once the new approach is proven.

## The one-sentence consensus

Across the reviewed sources, the field agrees: reliable AI development does not
come from a smarter model — it comes from **structure around the model**. Four
pillars recur and map directly onto the concerns above.

| Concern | What the research supports |
|---|---|
| Too complex to one-shot | Decompose into phase-gated specs + small verifiable tasks; treat recovery from a cold context window as the default assumption |
| Quality/robustness/security erosion | Gate every change on executable, external oracles the generating agent cannot edit for itself |
| Requirements current & enforced | Enforce requirements through executable artifacts (tests, contracts, CI gates), never through prose alone |
| 100% AI coding, humans on requirements + review | Humans own the spec/constitution and the functional/visual review; everything between is machine-gated |

## Four load-bearing findings

### 1. Verification is the bottleneck, not generation

The strongest systems treat the LLM as a cheap, fallible generator and place a
hard external filter between generation and acceptance.

- **Clover** (closed-loop verifiable code generation, Stanford): reduces
  correctness to a consistency check among code, docstring, and formal
  annotations. On CloverBench it accepted up to 87% of correct programs with
  **zero false positives** on adversarial-incorrect ones. Zero false positives
  is the property a "humans don't read the code" pipeline needs.
- **AlphaCodium** (Qodo): a designed test-generate-run-fix "flow", generating
  additional AI tests before/alongside code, raised GPT-4 pass@5 on
  CodeContests from 19% → 44%. Flow engineering beats prompt engineering.
- **Agentless** (FSE 2025): a fixed three-phase pipeline (localize → repair →
  validate with regression + reproduction tests), with no autonomous
  tool-choosing agent, beat all open-source agents on SWE-bench Lite at a
  fraction of the cost. Constrain control flow; reserve the LLM for localized
  reasoning.
- **Astrogator** / **spec2code** / **Dafny**-based work: in constrained domains,
  the accepted pattern is spec → generate → formally verify against the same
  spec. General-purpose verified generation for large polyglot codebases is
  still unsolved.

### 2. Quality provably erodes without gates

Measured, not theoretical:

- **~45% of AI-generated code fails security tests** (Veracode 2025, 100+
  models, 4 languages); security performance stayed flat as models grew even as
  functional correctness improved. Java was worst (~72% failure).
- AI code carries roughly **2.7× the vulnerability density** of human code.
- Iterating with the agent makes security worse: **+37.6% critical
  vulnerabilities after 5 refinement passes** (arXiv 2506.11022). This is the
  core argument for external gates over agent self-review.
- **Code duplication rose 8.3% → 12.3%** and refactoring dropped **25% → <10%**
  of changed lines as AI authorship rose (GitClear, 211M lines). Agents strongly
  prefer copy-paste over abstraction; cloned blocks correlate with 15–50% more
  defects.
- **~1 in 5 packages an LLM suggests do not exist** ("slopsquatting", USENIX
  2025) — a live supply-chain attack surface.

Implication: without durable gates, the codebase degrades along exactly the axes
of concern, and self-refinement accelerates the decay.

### 3. Long-horizon degradation is intrinsic; bigger context windows don't fix it

- **Context rot** (Chroma, 18 frontier models): every model degrades as input
  grows, not just near the limit — "a 1M-token window still rots at 50K tokens."
  A single distractor hurts; degradation worsens as query/needle semantic
  similarity drops.
- **Lost in the Middle** (TACL 2023): U-shaped position bias — accuracy high at
  start/end, ~30% lower in the middle.
- **Self-conditioning** (arXiv 2605.02572): an agent's own past errors compound,
  so short-horizon competence does not guarantee long-horizon reliability.
- **Mitigation that works** (Anthropic long-running-agent harness): externalize
  a durable plan + progress log + a JSON feature-list where 200+ features start
  marked "failing"; treat recovery from a clean context window as the design
  default. Force end-to-end evidence (e.g. browser automation) instead of
  self-declared success. Retrieve just-in-time; never preload the repo.
- **Self-repair loops saturate quickly** (~a few iterations, diminishing
  returns); pair them with an external verifier rather than self-judging.

### 4. "Specs as source of truth" is real, but only partly production-ready

A maturity ladder (Fowler / Böckeler, "Understanding SDD"):

- **spec-first** — spec guides the initial build, then drifts (Kiro, spec-kit as
  used in practice). Shipping.
- **spec-anchored** — spec + code co-evolve, bound by executable checks. The
  production-ready sweet spot most practitioners recommend. **← chosen baseline.**
- **spec-as-source** — humans edit only the spec; code is fully regenerated and
  never hand-edited (Tessl). Private beta; wrestles with non-determinism
  (reliable regeneration needs ever more specific specs).

Even Amazon (Werner Vogels, re:Invent 2025) found **tests alone were
insufficient** to capture intent, so they re-added prose specs — but as context,
not as the enforcement mechanism. Enforcement is always the executable artifact.
The "rebuild test" (Augment Code) operationalizes the ideal: delete `src/`,
regenerate from the spec in a clean session, and check it passes the existing
test suite.

## The reference architecture the field has converged on

Nearly every credible source composes the same primitives:

1. **Durable spec + constitution** — versioned requirements as living artifacts;
   a small set of immutable architectural principles re-read every session
   (spec-kit `/constitution`, Kiro steering files, `AGENTS.md`).
2. **Plan-before-code, phase-gated** — explore → plan → decompose into small
   reviewable tasks → implement, with human checkpoints at gates (Kiro, spec-kit,
   Cursor Plan Mode, Aider architect/editor, Devin, Claude Code).
3. **Retrieve, don't preload** — AST/symbol repo-maps (Aider's tree-sitter +
   PageRank ranking) or semantic + grep indexing (Cursor); keep the window small
   and high-signal.
4. **A composite "Definition of Done" gate** the agent cannot bypass, aggregating:
   - types/contracts and the compiler (zero-cost gate, makes illegal states
     unrepresentable)
   - property-based + mutation + fuzz tests (behavioral truth)
   - SAST/DAST + dependency allowlist / SBOM (security)
   - architectural fitness functions (ArchUnit-style) + complexity / duplication
     / churn budgets (durability against erosion)
5. **Verification as external oracle + adversarial review** — LLM-as-judge and
   multi-agent review (Cloudflare runs 7 specialized reviewers returning
   structured findings) as an additional layer, never the sole gate, because
   same-model reviewers share blind spots.
6. **Recovery-first memory** — ADRs for the *why* (stops re-litigation; MAST
   attributes ~42% of multi-agent failures to specification issues), a progress
   log for the *where*, a feature-list for *done*.

## Key tensions to resolve in the design

- **Who writes the gates?** The uncomfortable finding: the agent that writes the
  code tends to write checks that pass its own code. Fitness functions and
  acceptance specs authored by the same agent encode "whatever it already built."
  The design must decide which guardrails are human-anchored vs.
  AI-generated-but-reviewed.
- **Spec completeness becomes the whole ballgame.** If humans never read code,
  the only human-anchored ground truth is the spec plus the functional/visual
  review. That puts heavy weight on spec quality and on the acceptance/golden
  tests — the one artifact everyone says humans should review.
- **Bidirectional sync is the least-solved area.** Tooling for spec↔code drift
  detection is immature; current practice is one-directional (spec → code) plus
  tests as a tripwire. Supporting human-written code + on-the-go spec refinement
  (an explicit requirement here) means we must design the reverse path
  (code/feedback → spec update) ourselves.
- **Multi-agent parallelism has a known ceiling.** Anthropic explicitly flags
  tightly-coupled coding as a poor fit for multi-agent parallelism (agents can't
  coordinate mid-run; ~15× token cost). Use sub-agents for parallelizable,
  loosely-coupled subtasks; keep a single coherent context for cross-cutting
  architectural work.

## Benchmark caution

SWE-bench is contaminated (issues predate training cutoffs) and skewed toward
trivial 1–2 line fixes; high headline scores overstate real maintenance ability.
For realistic long-horizon signal, use **SWE-bench Pro** (multi-file, harder;
performance drops sharply) or **SWE-rebench** (continuously decontaminated).

## Gaps / open research

- **Maintenance over time** (vs. one-shot issue resolution) is barely
  benchmarked; SWE-bench Pro is the closest and results are poor.
- **Formal verification** works only in verification-aware languages (Dafny,
  F*, Verus), not mainstream production stacks.
- **Requirements-as-verifiable-spec** end-to-end at scale is an open gap — the
  highest-leverage place for original work.
- Several quantitative figures (Amazon's ~50% faster, Anthropic's 90.2%
  multi-agent uplift, 15× tokens, 12.5% search gain) are vendor-internal evals —
  directionally credible, not independently reproduced.
- Several 2026 preprints cited here are recent and lightly peer-reviewed; treat
  their specific numbers as provisional.

## Sources

### Agentic SWE / multi-agent
- SWE-agent (NeurIPS 2024, arXiv 2405.15793) — agent-computer interface
- AutoCodeRover (ISSTA 2024) — AST-aware code search
- MetaGPT (ICLR 2024, arXiv 2308.00352) — SOP-encoded multi-agent
- ChatDev (ACL 2024, arXiv 2307.07924) — role-play waterfall
- MASAI (arXiv 2406.11638) — modular sub-agents with scoped context
- Agentless (FSE 2025, arXiv 2407.01489) — fixed pipeline beats autonomy
- LLM Multi-Agent SE survey (arXiv 2404.04834)
- Why Do Multi-Agent LLM Systems Fail? / MAST (arXiv 2503.13657)

### Verification / spec-to-code
- Teaching LLMs to Self-Debug (ICLR 2024, arXiv 2304.05128)
- AlphaCodium (arXiv 2401.08500)
- Clover (arXiv 2310.17807)
- Astrogator (arXiv 2507.13290)
- Rethinking Verification: From Generation to Testing (arXiv 2507.06920)
- spec2code embedded automotive (arXiv 2411.13269)
- Dafny as verification-aware IL (arXiv 2501.06283); DafnyBench
- From Prompts to Properties (ACM FSE 2025)
- TypePilot (arXiv 2510.11151); ContractEval (arXiv 2510.12047)

### Context / scaling
- Lost in the Middle (TACL 2023, arXiv 2307.03172)
- Context Rot (Chroma) — https://www.trychroma.com/research/context-rot
- Self-conditioning / long-horizon (arXiv 2605.02572)
- Reflexion (NeurIPS 2023)
- Git Context Controller (arXiv 2508.00031)
- Anthropic — Effective context engineering for AI agents
- Anthropic — Effective harnesses for long-running agents
- Anthropic — Multi-agent research system
- Aider repo map (tree-sitter + PageRank) — https://aider.chat/docs/repomap.html
- Cursor codebase indexing — https://cursor.com/docs/context/codebase-indexing

### Spec-driven development / tooling
- Fowler/Böckeler — Understanding SDD: Kiro, spec-kit, Tessl
- AWS Kiro — https://kiro.dev/docs/specs/
- GitHub Spec Kit — https://github.com/github/spec-kit
- Tessl (Podjarny) — https://tessl.io
- Spec-Driven Development: From Code to Contract (arXiv 2602.00180)
- Augment Code — Spec as Source of Truth / rebuild test
- Werner Vogels re:Invent 2025 keynote (specs as code)
- AGENTS.md standard — https://agents.md/
- Anthropic — Building Effective AI Agents
- Simon Willison — Agentic Engineering Patterns
- Addy Osmani — How to write a good spec for AI agents
- Thoughtworks — Harness engineering and agent sensors

### Quality / security / erosion
- Veracode 2025 GenAI Code Security Report
- Security Degradation in Iterative AI Code Generation (arXiv 2506.11022)
- GitClear — AI Copilot Code Quality 2025 (211M lines)
- Snyk — Slopsquatting mitigation
- ArchUnit / architectural fitness functions (InfoQ)
- LLM-as-a-Judge for SE survey (arXiv 2510.24367)
- AutoReview (ACM FSE 2025)
- Cloudflare — Orchestrating AI Code Review at scale
- CodeScene — Guardrails for AI-assisted coding
