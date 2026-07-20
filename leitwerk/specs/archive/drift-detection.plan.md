# Plan — drift-detection

Status: landed (2026-07-20) <!-- all steps landed; T2 sign-off in-session -->

Spec: `leitwerk/specs/archive/drift-detection.md`. Verified references (read before
planning): the placeholder `core/checks/drift.sh:12-15`; the golden contract
`core/cmd/leitwerk/integration_test.go:212-221` (exit 0 + `spec(s) tracked` at
repo root, exit 2 with no specs); the selftest fixture pattern
`leitwerk/checks/selftest.sh:80-120` (lifecycle section 6, the model to copy);
the CI diff base `.github/workflows/leitwerk.yml:40-46`; check env/cwd
inheritance `core/internal/gate/verify.go:86-91`; resolver precedence puts
on-disk `core/checks/` ahead of the embedded copy for this repo
(`core/cmd/leitwerk/main.go:129` + `resolveSelf`), so an edited `drift.sh` is
live without a rebuild, and `selftest` rebuilds to re-embed for `go install`.

Every step leaves `leitwerk verify --tier T2` green on its own. Oracle-first is
applied *within* each step (the fixture is written before the code it proves);
a new check plus its fixtures land together so the gate is never red at a step
boundary.

## Steps
Step status: `[ ]` open · `[x]` done · `[~]` deviated — one line why.

1. `[x]` **Part 1 — anchor resolution + oracle.** Rewrite `core/checks/drift.sh`
   to: skip `archive/`; parse each spec's `## Anchors` section; resolve `path`
   and `path#symbol` anchors (path exists / glob matches; symbol is a whole word
   in the file); red (exit 1) with a readable per-anchor line on failure; keep
   exit 2 when no specs dir; keep a green summary containing `spec(s) tracked`.
   Add `selftest.sh` section 7 fixtures FIRST (consistent→0, missing-symbol→1,
   missing-path→1, archived-broken→0, no-specs→2), then implement until green.
   Files: `core/checks/drift.sh`, `leitwerk/checks/selftest.sh`. Tier T2.
   Checks: `selftest` (fixtures + rebuilt binary + Go golden test), `shell`.
   Manual check (T2): run drift by hand against a temp spec whose symbol was
   renamed and confirm the message names spec:line, anchor, and missing symbol.

2. `[x]` **Part 2 — one-sided change + oracle.** Add `LEITWERK_DIFF_BASE`
   handling: when set, `git diff --name-only "$base"...HEAD`; for each
   non-archived spec with anchors, red if an anchored path is in the changed set
   but the spec file is not. Unset → Part 2 skipped. Add a selftest git fixture
   (init repo, commit spec+code, change code only, run with the base) asserting
   exit 1. Files: `core/checks/drift.sh`, `leitwerk/checks/selftest.sh`. Tier T2.
   Checks: `selftest`, `shell`. Manual check (T2): the range matches CI's
   `base...HEAD`; unset-base run is unaffected.

3. `[x]` **Dogfood — anchor `self.md`.** Add a small `## Anchors` section to
   `leitwerk/specs/self.md` pointing at gate code it already describes (anchors
   that resolve), and update its one-line `drift (…)` behaviour description to
   state real drift detection instead of "specs tracked". Files:
   `leitwerk/specs/self.md`. Tier T2 (the change's tier governs). Checks:
   `drift` (must stay green — anchors resolve), `lifecycle`, `selftest`.
   Manual check (T2): repo-root `leitwerk drift` exits 0.

4. `[x]` **Docs — the Anchors convention.** `core/templates/spec.template.md`
   gains an optional `## Anchors` section (so adopters learn it); the
   `leitwerk-spec` skill gains one line to declare anchors; `bindings/open/
   AGENTS.md` mirrors it if it enumerates spec sections. Files: template (T0),
   `bindings/claude/skills/leitwerk-spec/SKILL.md` (T1), `bindings/open/AGENTS.md`
   (T1). Tier T2 (bundled under the change's tier). Checks: `context` (skill
   frontmatter budget unchanged), `selftest` (template re-embedded), `json`.
   Manual check: the template wording matches the spec's grammar.

## Verification strategy
The executable oracle is `leitwerk/checks/selftest.sh` section 7 — shell
fixtures covering each acceptance bullet in the spec, added before the code
they prove. The existing Go golden test
(`core/cmd/leitwerk/integration_test.go`) is preserved (green path stays
backward-compatible); it is extended only if a Go-level assertion is clearer
than a shell fixture. New behaviour never lands without a failing→passing
fixture. Runs at T2 (the tier `*.sh` selects).

## Risks & rollback
- **False-positive red on the real repo** (an anchor that does not resolve or a
  Part-2 range firing spuriously). Mitigation: symbol matching favours
  "resolves" on ambiguity; Part 2 is opt-in via the env base; self.md anchors
  are verified green in step 3 before landing.
- **Stale embedded copy** — the on-disk `core/checks/drift.sh` is live for this
  repo, but a `go install` user gets the embedded copy; `selftest` rebuilds the
  binary each run, re-embedding, so the two never diverge silently.
- **Rollback (T2, per step):** `git checkout -- core/checks/drift.sh
  leitwerk/checks/selftest.sh` and, if the binary was rebuilt,
  `make -C core build`; the gate returns to the placeholder behaviour, which is
  green. No data or external state is touched.

## Roles to wake
- `architect` — the anchor format and the Part-2 range semantics (recorded in
  the spec's Design decisions). Done.
- `test-engineer` — owns section 7 fixtures and confirms which fixture exercises
  each acceptance bullet (no green-but-untouched pass).
- Review panel at T2 (`/leitwerk-review` workflow or roles spawned directly) at
  the landing review; the human gives the T2 sign-off before it lands.

## Review findings addressed (T2 adversarial panel, 2026-07-20)
Three lenses (correctness, architecture, security) reviewed the change; all
verified findings were fixed before sign-off:
- **security:** anchor paths confined to the repo (absolute/`..` refused, not
  probed — closed an arbitrary-file existence/content oracle);
  `LEITWERK_DIFF_BASE` validated (option-like/unresolvable refused, not
  interpolated into `git`). Fixtures added.
- **correctness:** the Part-2 fixture no longer builds inside the asserted
  subshell, so a setup failure cannot fake a pass; exit code captured and
  asserted `== 1`; the message and summary enrichment are now asserted.
- **architecture:** corrected the false "CI already passes the base" claim
  (it does not; deferred to M2.4); failure lines now print the repo-relative
  spec path (`$f`); the shipped `drift.sh` header no longer cites this repo's
  spec path; plan files excluded from the spec scan; heading match tolerates
  trailing text; `sort` added to the stated tool floor.
- **deferred (noted in spec Out of scope):** CI base wiring + a living-contract
  exemption for Part 2 → M2.4/M2.3; a `security-reviewer` role entry in the
  constitution → flagged to the human.
