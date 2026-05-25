## Skill Digestion Harness [L1 Orchestrators L1a + L1b — before EVERY delegation]

Purpose: Prevent sub-agents from being overwhelmed by full SKILL.md files.
The orchestrator "digests" complex skills into actionable compact rules.
Rule: Load Tier 1 always. Load Tier 2 only if condition matches. Never load Tier 3 inline.

### Protocol

**Step 1: Identify required skills for this delegation**
```
From internal/assets/skills/ (skill-manifest.yaml when available):
  ALWAYS include: foundation block (Tier 1 — project conventions)

  Tier 2 check:
    IF task involves .go files → include go-testing compact rules
    IF project is Odoo → include odoo-development-skill compact rules
    IF task involves PR → include branch-pr compact rules
    IF sdd-apply task → include work-unit-commits compact rules
```

Detection script for Tier 2 conditions:
```bash
DIFF_FILES=$(git diff --cached --name-only 2>/dev/null || echo "")
PROJECT_FILES=$(find . -maxdepth 3 -name "*.go" -o -name "__manifest__.py" | head -20)

HAS_GO=$(echo "${PROJECT_FILES}" | grep -c "\.go$" || echo 0)
HAS_ODOO=$([ -f "$(find . -name '__manifest__.py' -maxdepth 5 | head -1)" ] && echo 1 || echo 0)
TASK_HAS_PR=$(echo "${TASK_DESCRIPTION}" | grep -ci "PR\|pull request\|push\|review" || echo 0)
TASK_HAS_COMMIT=$(echo "${TASK_DESCRIPTION}" | grep -ci "commit\|sdd-apply" || echo 0)
```

**Step 2: Extract compact rules (NOT full SKILL.md)**
```bash
# Extract only the ## Compact Rules section from a SKILL.md
SKILL_PATH="internal/assets/skills/${SKILL_NAME}/SKILL.md"
COMPACT=$(awk '/^## Compact Rules/,/^## [^C]/' "${SKILL_PATH}" 2>/dev/null | head -20)

# If no Compact Rules section: take first 15 lines after frontmatter
if [ -z "${COMPACT}" ]; then
  COMPACT=$(sed '/^---$/,/^---$/d' "${SKILL_PATH}" 2>/dev/null | head -15)
fi

# If skill file doesn't exist at path, fall back to skill name reference
if [ ! -f "${SKILL_PATH}" ]; then
  COMPACT="[${SKILL_NAME} — skill file not found at ${SKILL_PATH}. Reference directly.]"
fi
```

**Step 3: Build injection list and inject into delegation**
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

# Hard limit: 4 total blocks (1 foundation + 3 tier-2)
# Never inject full SKILL.md files — only compact rules
SKILLS_TO_INJECT = SKILLS_TO_INJECT[:4]
```

Inject into sub-agent task description:
```
## Project Standards (auto-resolved)

{foundation block — always first}

{go-testing compact rules — if applicable}
  Example: "Use table-driven tests. Run: go test ./... -race -count=1"

{odoo compact rules — if Odoo project}
  Example: "v18: use <list> not <tree>. Use invisible= not attrs=."

{work-unit-commits compact rules — if sdd-apply}
  Example: "1 commit = 1 deliverable behavior. Tests in same commit as implementation."

skill_resolution: {
  "status": "paths-injected",  # paths-injected | fallback-registry | none
  "skills_used": ["foundation", "go-testing"],
  "tier_2_activated": ["go-testing"],
  "fallback_reason": null
}
```

**Step 4: Skill Resolution Feedback — auto-correction**
After sub-agent returns Result Contract, inspect `result.skill_resolution.status`:

```
IF "fallback-registry":
  → "Sub-agent fell back to generic behavior — skill wasn't applied"
  → Re-read skill-manifest.yaml (or scan internal/assets/skills/)
  → Rebuild skill injection
  → Retry phase (circuit breaker permits 1 retry for skill failure)

IF "none":
  → "Sub-agent ran without any skill injection — serious gap"
  → Log to Engram: topic_key="skill-resolution-failure/{phase}/{timestamp}"
  → Escalate to human if D5 >= 2 (adaptive reasoning severity)

IF "paths-injected":
  → "Skills correctly applied" — proceed
```
