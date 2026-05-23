## sdd-hotfix [trigger: "/sdd-hotfix {description}" OR "hotfix {description}"]

### Eligibility Gate (ALL must pass)
- Scope: ≤ 3 files changed (rg count BEFORE starting)
- No new public API surfaces
- No schema migrations / model field changes
- D5 < 2 (not security-critical)
- User stated urgency reason

IF any gate fails → reject, recommend full sdd or manual sdd-apply.

### 5-Step Compressed Cycle

1. explore-lite: rg affected files + direct callers only (max 10 files)
2. propose-lite: one-sentence intent + one-line rollback plan
3. apply-branch: create apply/{change_name} branch, implement, commit
4. verify-lite: tests for affected files only; semantic audit SKIPPED (add RISK)
5. archive-lite: mem_save("sdd/{change_name}/hotfix", {justification, skipped_phases, files})

Audit trail MANDATORY. Note: "Recommend full sdd-verify post-urgency."
