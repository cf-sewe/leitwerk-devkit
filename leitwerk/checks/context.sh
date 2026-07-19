#!/usr/bin/env bash
# Repo check: context budget. The framework's efficiency claim rests on keeping
# the ALWAYS-ON context surface small: CLAUDE.md loads into every session, and
# every skill/agent frontmatter description is listed in every session's system
# prompt. This check measures that surface and fails when it exceeds budget, so
# a regression (someone inlining procedure text into CLAUDE.md, a description
# bloating) turns the gate red instead of silently taxing every future turn.
#
# Budgets (token estimate = bytes/4; proposed for the constitution's record):
#   CLAUDE.md                      <= 200 lines  (the limit CLAUDE.template.md states)
#   each .claude/rules/*.md        <= 100 lines  (loads whenever a matching path is touched)
#   each skill/agent description   <=  80 words  (frontmatter; always in the system prompt)
#   always-on total                <= 2000 estimated tokens
#
# Exit 0 = within budget, 1 = a budget exceeded, 2 = nothing to measure.
set -euo pipefail

fail=0
over() { echo "context budget exceeded: $*" >&2; fail=1; }

[ -f CLAUDE.md ] || { echo "no CLAUDE.md to budget"; exit 2; }

total_bytes=0

# 1. CLAUDE.md — always-on in every session.
lines="$(wc -l < CLAUDE.md | tr -d ' ')"
[ "$lines" -le 200 ] || over "CLAUDE.md is $lines lines (budget 200)"
total_bytes=$(( total_bytes + $(wc -c < CLAUDE.md) ))

# 2. Path-scoped rules — loaded whenever a matching path is touched.
for f in .claude/rules/*.md; do
  [ -e "$f" ] || continue
  lines="$(wc -l < "$f" | tr -d ' ')"
  [ "$lines" -le 100 ] || over "$f is $lines lines (budget 100)"
done

# 3. Skill/agent frontmatter — the name+description block between the first two
#    `---` lines is what every session's system prompt carries.
frontmatter() { awk '/^---$/{c++; next} c==1' "$1"; }
for f in bindings/claude/skills/*/SKILL.md bindings/claude/agents/*.md; do
  [ -e "$f" ] || continue
  words="$(frontmatter "$f" | wc -w | tr -d ' ')"
  [ "$words" -le 80 ] || over "$f frontmatter is $words words (budget 80)"
  total_bytes=$(( total_bytes + $(frontmatter "$f" | wc -c) ))
done

est_tokens=$(( total_bytes / 4 ))
[ "$est_tokens" -le 2000 ] || over "always-on surface ~${est_tokens} estimated tokens (budget 2000)"

[ "$fail" -eq 0 ] && echo "always-on context ~${est_tokens} est. tokens (budget 2000), files within budget"
exit "$fail"
