---
name: requesting-code-review
description: Use when completing tasks, implementing major features, or before merging to verify work meets requirements - dispatches code-reviewer subagent to review implementation against plan or requirements before proceeding
---

# Requesting Code Review

**Core principle:** Review early, review often.

## When to Request

**Mandatory:** After each task in subagent-driven dev, after major feature, before merge to main.
**Optional:** When stuck (fresh perspective), before refactoring (baseline check), after complex bug fix.

## How to Request

**1. Get git SHAs:**
```bash
BASE_SHA=$(git rev-parse HEAD~1)  # or origin/main
HEAD_SHA=$(git rev-parse HEAD)
```

**2. Dispatch code-reviewer subagent:** Task tool with code-reviewer type, fill template.

**Placeholders:** `{WHAT_WAS_IMPLEMENTED}`, `{PLAN_OR_REQUIREMENTS}`, `{BASE_SHA}`, `{HEAD_SHA}`, `{DESCRIPTION}`

**3. Act on feedback:** Fix Critical immediately, Important before proceeding, Minor for later. Push back if wrong (with reasoning).

## Example

```
[Task 2 done: Add verification function]

BASE_SHA=$(git log --oneline | grep "Task 1" | head -1 | awk '{print $1}')
HEAD_SHA=$(git rev-parse HEAD)

[Dispatched code-reviewer]
  WHAT: Verification and repair functions for conversation index
  PLAN: Task 2 from docs/plans/deployment-plan.md
  BASE: a7981ec
  HEAD: 3df7661
  DESC: Added verifyIndex() and repairIndex() with 4 issue types

[Returns]:
  Strengths: Clean architecture, real tests
  Issues:
    Important: Missing progress indicators
    Minor: Magic number (100) for reporting interval
  Assessment: Ready to proceed

[Fix progress indicators] → [Continue to Task 3]
```

## Integration

**Subagent-Driven Dev:** Review after EACH task. Catch issues before they compound.
**Executing Plans:** Review after batch (3 tasks). Get feedback, apply, continue.
**Ad-Hoc:** Review before merge. Review when stuck.

## Red Flags

**Never:** Skip review ("it's simple"), ignore Critical issues, proceed with unfixed Important issues, argue with valid tech feedback.

**If reviewer wrong:** Push back with tech reasoning. Show code/tests proving it works. Request clarification.

See template: requesting-code-review/code-reviewer.md
