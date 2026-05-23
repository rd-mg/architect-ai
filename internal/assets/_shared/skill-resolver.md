## Skill Resolver Protocol v3.0 [Orchestrators L1a + L1b — before EVERY delegation]

Purpose: Inject the right skills into each sub-agent without overloading context.
Rule: Load Tier 1 always. Load Tier 2 only if condition matches. Never load Tier 3 inline.

### Step 1: Load Tier 1 (Foundation)
```
foundation_block = read(".atl/_generated/foundation.md")
# This file is pre-generated at install time — no runtime reading of 6 separate files.
# Inject as ## Project Foundation Standards block in sub-agent prompt.
```

### Step 2: Detect Tier 2 Conditions
```bash
DIFF_FILES=$(git diff --cached --name-only 2>/dev/null || echo "")
PROJECT_FILES=$(find . -maxdepth 3 -name "*.go" -o -name "__manifest__.py" | head -20)

# Check conditions
HAS_GO=$(echo "${PROJECT_FILES}" | grep -c "\.go$" || echo 0)
HAS_ODOO=$([ -f "$(find . -name '__manifest__.py' -maxdepth 5 | head -1)" ] && echo 1 || echo 0)
TASK_HAS_PR=$(echo "${TASK_DESCRIPTION}" | grep -ci "PR\|pull request\|push\|review" || echo 0)
TASK_HAS_COMMIT=$(echo "${TASK_DESCRIPTION}" | grep -ci "commit\|sdd-apply" || echo 0)
```

### Step 3: Build Skill Injection List
```
SKILLS_TO_INJECT = [foundation_block]  # Always starts with Tier 1

if HAS_GO > 0 AND task involves testing:
    SKILLS_TO_INJECT += compact_rules("go-testing")

if HAS_ODOO > 0:
    SKILLS_TO_INJECT += compact_rules("odoo-development-skill")

if TASK_HAS_PR > 0:
    SKILLS_TO_INJECT += compact_rules("branch-pr")

if TASK_HAS_COMMIT > 0:
    SKILLS_TO_INJECT += compact_rules("work-unit-commits")

# Hard ceiling: max 3 additional skills beyond foundation
SKILLS_TO_INJECT = SKILLS_TO_INJECT[:4]  # foundation + max 3
```

### Step 4: Record Skill Resolution in Delegation
Include in sub-agent task description:
```
## Project Standards (auto-resolved)
[foundation block content]
[tier-2 compact rules if applicable]

skill_resolution: {
  "status": "paths-injected",  # paths-injected | fallback-registry | none
  "skills_used": ["foundation", "go-testing"],
  "tier_2_activated": ["go-testing"],
  "fallback_reason": null
}
```

### Step 5: Post-Delegation Feedback Check
After sub-agent returns Result Contract:
IF result.skill_resolution.status == "fallback-registry" OR "none":
  → Re-read .atl/skill-manifest.yaml
  → Rebuild skill injection for that agent
  → Retry phase once (circuit breaker allows it)
```