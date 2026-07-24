# Proposal — record "Conventional Commits + PR conventions" in the constitution

Status: open (awaiting human decision)
Date: 2026-07-23
Relates: leitwerk/specs/cli-publish.md (D7)

## Decision to record
The cli-publish work (D2) automates versioning and the changelog with
`release-please`, which derives both from commit history. That makes **Conventional
Commits** a standing workflow rule, not a detail of one change — so it belongs in
the constitution's *Decisions of record*. `constitution.md` is human-owned, hence
this proposal rather than a direct edit.

## Evidence
- release-please cannot derive a version or changelog from ad-hoc messages
  (cli-publish D7).
- PRs are squash-merged, so the PR **title** is the commit release-please reads; a
  `semantic-pr` check (`amannn/action-semantic-pull-request`) enforces the
  Conventional-Commit format on PR titles. Types and scopes are documented in
  `.gitmessage` (the commit template) and `CONTRIBUTING.md`; scopes are advisory.
- Matches the cplace-ops-cloud reference (release-please + `.gitmessage` + a
  semantic-PR check).

## Options
- **A (recommended):** add the decision-of-record entry below.
- **B:** leave it recorded only in the cli-publish spec, which is a change record
  and archives on landing — but a standing workflow rule is easy to lose once the
  spec leaves the active set.

## Recommendation
**A.** Proposed entry (verbatim, for `constitution.md` → Decisions of record; date
filled when landed alongside cli-publish):

> - 2026-07-2X: Conventional Commits adopted repo-wide. `release-please` derives
>   the version and changelog from commit history; because PRs are squash-merged,
>   the PR **title** is the commit it reads, so a `semantic-pr` check enforces the
>   Conventional-Commit format on PR titles. Allowed types and scopes live in
>   `.gitmessage` (the commit template) and `CONTRIBUTING.md`; scopes are advisory.
>   No leitwerk gate check parses commit messages — this is a PR-gating CI check,
>   so the gate's open-code parity is unaffected. See
>   `leitwerk/specs/archive/cli-publish.md`.
