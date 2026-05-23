---
name: code-review
description: Use when receiving code review feedback (especially if unclear or technically questionable), when completing tasks or major features requiring review before proceeding, or before making any completion/success claims. Covers three practices - receiving feedback with technical rigor over performative agreement, requesting reviews via code-reviewer subagent, and verification gates requiring evidence before any status claims. Essential for subagent-driven development, pull requests, and preventing false completion claims.
---

# Code Review

Three practices: Receiving feedback (tech rigor > performative), Requesting reviews (code-reviewer subagent), Verification gates (evidence before claims).

## Core Principle

**Tech correctness > social comfort.** Verify before implement. Ask before assume. Evidence before claims.

## When to Use

### Receiving Feedback
- Review comments from any source, unclear/questionable feedback, multiple items need prioritization, external reviewer lacks context, suggestion conflicts with decisions
**Ref:** `references/code-review-reception.md`

### Requesting Review
- After EACH task in subagent-driven dev, major feature/refactor done, before merge to main, stuck need fresh perspective, after complex bug fix
**Ref:** `references/requesting-code-review.md`

### Verification Gates
- About to claim tests pass/build succeeds/work complete, before commit/push/PR, moving to next task, any success/completion statement, expressing satisfaction
**Ref:** `references/verification-before-completion.md`

## Quick Decision Tree

```
Received feedback?
  Unclear? → STOP ask clarification first
  Human partner? → Understand then implement
  External? → Verify tech before implement
Completed work?
  Major feature/task? → Request code-reviewer
  Before merge? → Request code-reviewer
About to claim status?
  Fresh verification? → Claim WITH evidence
  No verification? → RUN verification first
```

## Receiving Feedback Protocol

**Pattern:** READ → UNDERSTAND → VERIFY → EVALUATE → RESPOND → IMPLEMENT

**Rules:**
- NO performative agreement ("You're absolutely right!", "Great point!", "Thanks")
- NO implement before verification
- Restate requirement, ask questions, push back with tech reasoning, or just act
- If unclear: STOP, clarify ALL unclear items first
- YAGNI: grep usage before implementing suggested "proper" features

**Source Handling:**
- **Human partner:** Trusted — implement after understanding, no performative
- **External:** Verify tech correct, check breakage, push back if wrong

## Requesting Review Protocol

**Process:**
1. Git SHAs: `BASE_SHA=$(git rev-parse HEAD~1)`, `HEAD_SHA=$(git rev-parse HEAD)`
2. Dispatch code-reviewer via Task tool: WHAT_WAS_IMPLEMENTED, PLAN_OR_REQUIREMENTS, BASE_SHA, HEAD_SHA, DESCRIPTION
3. Act: Fix Critical immediately, Important before proceeding, Minor for later

## Verification Gates Protocol

**Iron Law:** NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE

**Gate:** IDENTIFY command → RUN full command → READ output → VERIFY claim → THEN claim. Skip any step = lying.

**Requirements:**
- Tests pass: 0 failures
- Build: exit 0
- Bug fixed: test original symptom passes
- Requirements met: line-by-line checklist verified

**Red Flags:** "should"/"probably"/"seems to", satisfaction before verification, commit without verify, trust agent reports, ANY wording implying success without running verification

## Integration

- **Subagent-Driven:** Review after EACH task, verify before next
- **PRs:** Verify tests pass, request code-reviewer before merge
- **General:** Verification gates before any status claims, push back on invalid feedback

## Bottom Line

1. Tech rigor > social performance — No performative agreement
2. Systematic review — Use code-reviewer subagent
3. Evidence before claims — Verification gates always

Verify. Question. Then implement. Evidence. Then claim.
