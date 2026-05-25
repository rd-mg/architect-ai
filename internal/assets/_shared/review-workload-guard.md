## Review Workload Guard [sdd-orchestrator MANDATORY — execute after sdd-tasks, before sdd-apply]

Purpose: Protect the human reviewer from unmanageable PRs. A 400+ line PR takes 60+ minutes
to review, introduces more merge bugs, and kills team morale.

### Step 1: Read Review Workload Forecast from sdd-tasks result

The sdd-tasks Result Contract includes a `review_workload` field:
```json
{
  "review_workload": {
    "estimated_lines_changed": 250,
    "budget_risk": "low|medium|high",
    "chained_prs_recommended": false,
    "decision_needed_before_apply": false,
    "tasks_count": 8,
    "parallel_tasks": 3,
    "sequential_tasks": 5
  }
}
```

### Step 2: Apply Delivery Strategy Decision Table

Read cached `session.delivery_strategy` (set at session start).

| Budget Risk | Chained PRs? | Delivery Strategy | Action |
|---|---|---|---|
| low (≤ 400 lines) | No | Any | Proceed to sdd-apply normally |
| medium (400-800 lines) | ask-on-risk | ask-on-risk | ASK user: split or proceed? |
| medium (400-800 lines) | auto-chain | auto-chain | Auto-split into PR slices |
| high (> 800 lines) | ask-on-risk | ask-on-risk | ASK user: split strategy |
| high (> 800 lines) | auto-chain | auto-chain | Auto-split, first slice only |
| any | single-pr | single-pr | Require size:exception documentation |
| any | exception-ok | exception-ok | Record accepted exception, proceed |

### Step 3: On ask-on-risk (default)

If budget_risk = medium or high AND delivery_strategy = ask-on-risk:

Emit to user (LITE):
```
Review Workload Forecast:
  Estimated lines changed: {N}
  Budget risk: {medium|high}
  Recommended: {chained_prs_recommended}

How should I proceed?
  [1] Split into chained PRs (recommended) → I'll implement in slices, each ≤ 400 lines
  [2] Single PR with size:exception → I'll proceed but document the exception
  [3] Stop here → Let me know when ready to continue
```

Cache the user's choice as `session.current_change_delivery_decision`.

### Step 4: On auto-chain

If delivery_strategy = auto-chain AND budget_risk > low:
- Determine the first PR slice from tasks (sequential tasks first, smallest independent work unit)
- Pass `pr_boundary.slice = 1` to sdd-apply
- After apply, open PR for slice 1
- Emit: "[delivery] Slice 1/{total} applied. Continue with /sdd-apply to proceed with slice 2."

### Step 5: Pass delivery info to sdd-apply

Always include in sdd-apply prompt:
```
delivery_strategy: {strategy}
delivery_decision: {user_choice or auto_chain or exception-ok}
pr_boundary: {all | slice_N_of_M}
size_exception_rationale: {string | null}
```

### IMPORTANT: Even in Automatic execution mode, the Review Workload Guard runs.
Automatic mode does not skip reviewer protection.
