## Traceability Auto-Injection [ALL SDD phases — mandatory]

### Rule
Every SDD artifact file (proposal.md, spec.md, design.md, tasks.md, verify-report.md)
MUST contain this block immediately after the `# Title` line:

```markdown
# {Artifact Title}

> **Source of Truth:** {absolute_path_or_engram_key}
> **Change:** {change_name}
> **Phase:** {sdd-phase-name}
> **Generated:** {ISO8601_timestamp}
> **Author:** {agent_name} (architect-ai {version})
```

### Implementation for each phase

**sdd-propose** writes to `openspec/changes/{change}/proposal.md`:
```markdown
# Proposal: {change_name}

> **Source of Truth:** {user_request_summary_or_ticket_url}
> **Change:** {change_name}
> **Phase:** sdd-propose
> **Generated:** {timestamp}
> **Author:** sdd-propose (architect-ai)
```

**sdd-spec** writes to `openspec/changes/{change}/spec.md`:
```markdown
# Specification: {change_name}

> **Source of Truth:** openspec/changes/{change}/proposal.md
> **Change:** {change_name}
> **Phase:** sdd-spec
> **Generated:** {timestamp}
> **Author:** sdd-spec (architect-ai)
```

**sdd-design** writes to `openspec/changes/{change}/design.md`:
```markdown
# Design: {change_name}

> **Source of Truth:** openspec/changes/{change}/spec.md
> **Change:** {change_name}
> **Phase:** sdd-design
> **Generated:** {timestamp}
> **Author:** sdd-design (architect-ai)
```

**sdd-tasks** writes to `openspec/changes/{change}/tasks.md`:
```markdown
# Tasks: {change_name}

> **Source of Truth:** openspec/changes/{change}/design.md
> **Change:** {change_name}
> **Phase:** sdd-tasks
> **Generated:** {timestamp}
> **Author:** sdd-tasks (architect-ai)
```

### Validation

In `sdd-verify`, before semantic audit:
```bash
# Check all artifacts have traceability headers
for artifact in openspec/changes/${CHANGE}/*.md; do
  if ! grep -q "^> \*\*Source of Truth:\*\*" "${artifact}"; then
    echo "TRACEABILITY MISSING: ${artifact}" >&2
    VIOLATIONS=$((VIOLATIONS + 1))
  fi
done
[ "${VIOLATIONS}" -gt 0 ] && exit 1
```
