---
name: verification-before-completion
description: Use when about to claim work is complete, fixed, or passing, before committing or creating PRs - requires running verification commands and confirming output before making any success claims; evidence before assertions always
---

# Verification Before Completion

**Core principle:** Evidence before claims, always.
**Violating letter = violating spirit.**

## The Iron Law

```
NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE
```

If you haven't run verification command in this message, you cannot claim it passes.

## The Gate Function

```
BEFORE claiming any status or expressing satisfaction:
1. IDENTIFY: What command proves this claim?
2. RUN: Execute FULL command (fresh, complete)
3. READ: Full output, check exit code, count failures
4. VERIFY: Does output confirm claim?
   - NO: State actual status with evidence
   - YES: State claim WITH evidence
5. ONLY THEN: Make claim

Skip any step = lying, not verifying
```

## Common Failures

| Claim | Requires | Not Sufficient |
|-------|----------|----------------|
| Tests pass | Test output: 0 failures | Previous run, "should pass" |
| Linter clean | Linter output: 0 errors | Partial check, extrapolation |
| Build succeeds | Build: exit 0 | Linter passing, logs look good |
| Bug fixed | Test original symptom: passes | Code changed, assumed fixed |
| Regression test works | Red-green cycle verified | Test passes once |
| Agent completed | VCS diff shows changes | Agent reports "success" |
| Requirements met | Line-by-line checklist | Tests passing |

## Red Flags — STOP

"should"/"probably"/"seems to", satisfaction before verification ("Great!"/"Perfect!"/"Done!"), commit/push/PR without verify, trust agent success reports, partial verification, "just this once", tired wanting work over, ANY wording implying success without running verification

## Rationalization Prevention

| Excuse | Reality |
|--------|---------|
| "Should work now" | RUN verification |
| "I'm confident" | Confidence ≠ evidence |
| "Just this once" | No exceptions |
| "Linter passed" | Linter ≠ compiler |
| "Agent said success" | Verify independently |
| "I'm tired" | Exhaustion ≠ excuse |
| "Partial check enough" | Partial proves nothing |
| "Different words so rule doesn't apply" | Spirit over letter |

## Key Patterns

**Tests:**
```
 [Run test command] [See: 34/34 pass] "All tests pass"
 "Should pass now" / "Looks correct"
```

**Regression (TDD Red-Green):**
```
 Write → Run (pass) → Revert fix → Run (MUST FAIL) → Restore → Run (pass)
 "I've written regression test" (without red-green verification)
```

**Build:**
```
 [Run build] [See: exit 0] "Build passes"
 "Linter passed" (linter doesn't check compilation)
```

**Requirements:**
```
 Re-read plan → Create checklist → Verify each → Report gaps or completion
 "Tests pass, phase complete"
```

**Agent delegation:**
```
 Agent reports success → Check VCS diff → Verify changes → Report actual state
 Trust agent report
```

## Why This Matters

From 24 failures: trust broken ("I don't believe you"), undefined functions shipped (crash), missing requirements shipped (incomplete), time wasted on false completion → rework. Violates: "Honesty core value. Lie → replaced."

## When To Apply

**ALWAYS before:** Any success/completion claims, expression of satisfaction, positive statement about work state, commit/PR/task completion, moving to next task, delegating to agents.

**Rule applies to:** Exact phrases, paraphrases, synonyms, implications of success, ANY communication suggesting completion/correctness.

## Bottom Line

**No shortcuts for verification.** Run command. Read output. THEN claim result. Non-negotiable.
