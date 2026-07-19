# docs

- **`whitepaper.html`** — the framework paper: motivation, evidence, roles,
  workflows, tiered gate, and adoption bindings. Open in a browser.
- **`decisions/design-proposal.md`** — the design: principles, decisions, the
  three-layer architecture, lifecycle, and rollout phases.
- **`decisions/research-briefing.md`** — the evidence base the design is built
  on (synthesis of the source research across agentic SWE, verification, context
  scaling, spec-driven dev, quality/security erosion, and brownfield adoption).

The reference implementation of what the whitepaper describes is this repository:
`core/` is the gate, `bindings/` are the per-tool adapters, `examples/` runs it.
