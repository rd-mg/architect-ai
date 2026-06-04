---
name: work-unit-commits
description: "Plan commits as reviewable work units. Trigger: implementation, commit splitting, chained PRs, keeping tests with code."
bridge: false
on-demand: true
---

## Core Rule

**A commit = a deliverable behavior, fix, migration, or docs unit.**
NOT a file type dump (don't do "add models", "add services", "add tests" as 3 commits if none works alone).

## Mandatory Checklist Before Each Commit

- [ ] The commit has ONE clear purpose (a reviewer should understand it from the diff message)
- [ ] The repo still works after applying ONLY this commit (tests pass, app starts)
- [ ] Tests for this behavior are INCLUDED in the same commit
- [ ] Docs for user-visible changes are INCLUDED in the same commit
- [ ] Rollback is possible without reverting unrelated work

## Commit Message Format (NORMAL register MANDATORY)

**Full template (copy verbatim, fill slots in NORMAL register):**
```text
{type}({scope}): {imperative-verb} {object} {for/to/that} {user-visible outcome}  [≤72 chars]

[Optional body — include ONLY if reason is non-obvious:]
- Reason: {concrete technical or business reason}
- Impact: {what breaks or regresses if not done}
- Ref: #{issue-number} or N/A

[Optional footer:]
BREAKING CHANGE: {description} — only if API/contract changes
Closes #{issue-number}
```

**Type vocabulary (lowercase):**
feat | fix | refactor | test | docs | chore | perf | build | ci | style | revert

cudio-git format (for Odoo projects):
```text
[TAG][TASK_ID] module_name: description
```
TAG values: ADD · FIX · IMP · REF · REM · MOV · REV

**Anti-patterns (PROHIBITED in all commit messages):**
- Abbreviations from ULTRA mode: "impl", "chk", "auth drop", "no filler"
- Vague objects: "add stuff", "fix things", "update code"
- Single-word messages: "fix", "auth", "tests"
- AI attribution: "Co-authored-by: Claude", "AI-generated"
- Subject > 72 chars (use body for additional context)

**Verification before committing:**
1. Subject ≤ 72 chars? ✓
2. Starts with type(scope): ? ✓
3. Imperative mood? ✓ ("add", not "added")
4. No ULTRA abbreviations? ✓
5. BREAKING CHANGE footer if contract changed? ✓

## Split Examples

| Bad split (by file type) | Good split (by work unit) |
|---|---|
| commit: "add models" | commit: "feat(auth): add token model + validation tests" |
| commit: "add services" | commit: "feat(auth): wire token validation into login flow" |
| commit: "add tests" | ← Tests go WITH their behavior commit |
| commit: "update docs" | ← Docs go WITH the user-visible change |

## When to Create a New PR vs New Commit

- Each commit = candidate for promotion to chained PR if the change grows
- If sdd-tasks forecast says `budget_risk: high` → plan commits to be PR-ready from the start
- One commit = one PR slice when auto-chain strategy is active

## Review Budget Check

Before committing, estimate:
```bash
git diff --stat  # shows lines added + deleted
# If additions + deletions > 400 → consider splitting
```

If running under sdd-apply with `delivery_strategy = auto-chain`:
- Each commit must keep the PR within 400 lines
- Stop and report to orchestrator: "slice N complete, {lines} lines, ready for PR"
