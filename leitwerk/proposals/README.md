# leitwerk/proposals/ — decisions awaiting the human

The agent-writable inbox for decisions only a human may make (human-owned
files, priorities, risk acceptance). One timestamped file per decision,
containing everything needed to decide: the decision needed, the exact change
(verbatim text or command), how to accept, what rejection means, and a
recommendation — which doubles as the default when the human simply approves.

Lifecycle: **accept** → apply the change (or authorize the agent), delete the
file. **Reject** → delete the file. Proposals never carry authority by
themselves — they are requests, and this directory must trend toward empty.

Write a proposal so it maps onto a multiple-choice question: the problem in
one or two sentences, each option with a one-line description, one option
marked as the recommendation.

Follow-up is mechanical, not hoped-for: the `lifecycle` check counts open
proposals on every gate run and flags files older than 30 days (timestamp
prefix). On Claude Code, a `SessionStart` hook surfaces the inbox into every
new session, and `leitwerk-review` presents each pending decision as a native
multiple-choice question — an accepted answer authorizes the agent to apply
the documented change and delete the file.

Convention status: proposed in `20260719_211654-boundary-granularity.md`;
the escalation rule that decides *what* lands here is in
`leitwerk/specs/archive/decision-routing.md` (landed; its authority version is
the constitution's "Decision routing" section).
