// leitwerk review workflow — a tier-scaled adversarial panel feeding the gate.
//
// `leitwerk init` scaffolds this into a repo's .claude/workflows/. Workflows are
// a Claude Code primitive (opt-in) and cannot ship inside a plugin, so they are
// delivered as this scaffold instead. Tailor the review dimensions to your repo;
// the final phase runs `leitwerk verify`, which stays authoritative no matter how
// the panel votes.
export const meta = {
  name: 'leitwerk-review',
  description: 'Tier-scaled adversarial review of a change, then the deterministic gate. Layer 2 (soft, multi-agent) feeding Layer 3 (hard, external).',
  phases: [
    { title: 'Review', detail: 'one role per review dimension' },
    { title: 'Verify', detail: 'independently try to refute each finding' },
    { title: 'Gate', detail: 'run leitwerk verify — the authoritative oracle' },
  ],
}

// args: { tier: 'T0'|'T1'|'T2', spec?: string, scope?: string }
// The orchestrator IS this script — deterministic control flow, not a role.
const tier = (args && args.tier) || 'T2'
const scope = (args && args.scope) || 'the current change (git diff)'
const spec = (args && args.spec) || 'leitwerk/specs/'

const FINDINGS = {
  type: 'object',
  properties: {
    findings: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          summary: { type: 'string' },
          file: { type: 'string' },
          failure_scenario: { type: 'string' },
        },
        required: ['summary', 'failure_scenario'],
      },
    },
  },
  required: ['findings'],
}
const VERDICT = {
  type: 'object',
  properties: { refuted: { type: 'boolean' }, reason: { type: 'string' } },
  required: ['refuted'],
}
const GATE = {
  type: 'object',
  properties: { exitCode: { type: 'integer' }, green: { type: 'boolean' }, output: { type: 'string' } },
  required: ['exitCode', 'green'],
}

// Review dimensions, each handled by the matching role subagent (agentType
// resolves against .claude/agents + the plugin's agents/). Scale with tier:
// T0/T1 review only what the change touches; T2 runs the full adversarial panel.
const ALL_DIMS = [
  { key: 'correctness', agent: 'test-engineer', minTier: 'T0',
    prompt: `Review ${scope} for logic errors, missed edge cases, and violations of ${spec}. Prompted to REFUTE, not approve.` },
  { key: 'spec-fidelity', agent: 'architect', minTier: 'T1',
    prompt: `Check that ${scope} does what ${spec} promises, and that structural boundaries hold. Report divergences as findings.` },
  { key: 'security', agent: 'security-reviewer', minTier: 'T2',
    prompt: `Review ${scope} for auth/tenant/data-boundary/input/supply-chain issues. Each finding: a concrete failure scenario.` },
]
const rank = { T0: 0, T1: 1, T2: 2 }
const dims = ALL_DIMS.filter(d => rank[d.minTier] <= rank[tier])
log(`leitwerk-review at ${tier}: ${dims.map(d => d.key).join(', ')}`)

// Review each dimension, then adversarially verify its findings — pipelined so a
// dimension's findings get refuted while other dimensions are still reviewing.
const reviewed = await pipeline(
  dims,
  d => agent(d.prompt, { agentType: d.agent, phase: 'Review', label: `review:${d.key}`, schema: FINDINGS }),
  (review, d) => parallel((review?.findings || []).map(f => () =>
    agent(`Independently try to REFUTE this ${d.key} finding. Default to refuted=true if you cannot substantiate it.\n\n${f.summary}\n${f.failure_scenario}`,
      { agentType: 'general-purpose', phase: 'Verify', label: `verify:${d.key}`, schema: VERDICT })
      .then(v => ({ ...f, dimension: d.key, refuted: v?.refuted !== false }))
  )),
)
const confirmed = reviewed.flat().filter(Boolean).flatMap(x => x).filter(f => f && !f.refuted)

// Layer 3: the hard gate. This is an external oracle, not agent judgment — the
// workflow only REPORTS it; the Stop hook / CI ENFORCE it at turn end. A red
// gate blocks the change no matter how the review panel voted.
phase('Gate')
const gate = await agent(
  `Run exactly: leitwerk verify --tier ${tier}\nReport the integer exit code and whether it was green (exit 0). Do not soften or interpret a failure.`,
  { agentType: 'general-purpose', phase: 'Gate', label: 'gate', schema: GATE },
)

return {
  tier,
  gate,                       // authoritative
  confirmedFindings: confirmed,
  landable: gate?.green === true && confirmed.length === 0,
}
