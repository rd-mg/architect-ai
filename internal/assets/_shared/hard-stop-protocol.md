## Hard Stop Protocol [sdd-apply + sdd-spec + sdd-design]

### When to Hard Stop (MANDATORY halt)

Apply this check for each item being implemented:

```
IF source_of_truth is null or empty:
  → HARD STOP: "Source of Truth missing for this task"

IF required_function_not_in_design:
  → HARD STOP: "Function {name} referenced in tasks not defined in design"

IF field_conflict_detected:
  → HARD STOP: "Field {name} in spec conflicts with existing codebase field"

IF task_dependency_not_met:
  → HARD STOP: "Task {T2} requires {T1} completion, {T1} is still pending"

IF spec_has_contradictory_requirements:
  → HARD STOP: "Spec line {N} contradicts line {M} — cannot resolve"
```

### Hard Stop Output Format

```json
{
  "status": "blocked",
  "phase": "sdd-apply",
  "executive_summary": "Hard stop: {reason}. User clarification required.",
  "artifacts": [],
  "next_recommended": "clarification_needed",
  "risks": ["Implementation cannot proceed without clarification"],
  "skill_resolution": {"status": "paths-injected"},
  "blocked_reason": "{specific reason with file:line if applicable}",
  "clarification_needed": "{specific question for user}",
  "attempt_number": 1
}
```

### --force-assume Flag

If the user wants to bypass hard stops for minor ambiguities:
```
User: "use sdd-apply --force-assume"
```
Then:
- Document the assumption explicitly in apply-progress.yaml
- Add it to the risks section of the Result Contract
- Continue with the assumption, but mark the task as "assumed"
- At sdd-verify, flag all "assumed" tasks for extra scrutiny

DO NOT use --force-assume for: missing Source of Truth, field conflicts, broken dependencies.
These require actual clarification.

### After Clarification

User provides clarification → store in Engram:
```
mem_save(
  topic_key: "sdd/{change}/clarification/{topic}",
  content: {question, answer, timestamp, affects_tasks: [T1, T2]}
)
```
Then resume sdd-apply from the blocked task.
