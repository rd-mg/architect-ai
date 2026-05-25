## Rollback Harness [All destructive operations — MANDATORY]

### When to trigger rollback

Rollback MUST be available for:
1. sdd-apply that failed mid-implementation (apply branch exists)
2. architect-ai sync that corrupted a config file
3. GGA pre-commit that left a partial state
4. Any Go installer operation that failed after partial write

### Rollback Decision Table

| Scenario | Rollback Action | Command |
|---|---|---|
| sdd-apply failed | Abandon apply branch | `git checkout {original_branch}` (branch preserved for inspection) |
| CLAUDE.md corrupted by sync | Restore from backup | `architect-ai restore CLAUDE.md` |
| opencode.json corrupted | Restore from backup | `architect-ai restore opencode.json` |
| sdd-state.yaml corrupted | Restore from backup OR re-init | `architect-ai restore sdd-state.yaml` |
| Engram SQLite corrupted | Restore from engram sync backup | `engram restore --from-git` |
| MCP config broken | Regenerate from template | `architect-ai mcp --regenerate` |

### Rollback Guide for sdd-apply

```bash
#!/usr/bin/env bash
# .atl/scripts/rollback-apply.sh {change_name}
set -euo pipefail

CHANGE="${1:?change_name required}"
APPLY_BRANCH="apply/$(echo "${CHANGE}" | tr '[:upper:] ./_' '[:lower:]----')"
ORIGINAL=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "main")

echo "Rolling back sdd-apply for change: ${CHANGE}"

# Step 1: Return to original branch
git checkout "${ORIGINAL}" 2>/dev/null || {
  echo "ERROR: could not checkout ${ORIGINAL}" >&2; exit 1
}
echo "STEP 1: Back on ${ORIGINAL}"

# Step 2: List apply branch commits for inspection
echo "STEP 2: Apply branch commits (preserved for inspection):"
git log "${APPLY_BRANCH}" --oneline 2>/dev/null || echo "  (apply branch not found)"

# Step 3: Update sdd-state.yaml — reset sdd-apply to pending
python3 -c "
import re, sys
with open('.atl/sdd-state.yaml') as f:
    content = f.read()
content = re.sub(
    r'(  sdd-apply:.*?status: )\"(running|failed)\"',
    r'\1\"pending\"',
    content, flags=re.DOTALL
)
with open('.atl/sdd-state.yaml.tmp', 'w') as f:
    f.write(content)
import os
os.rename('.atl/sdd-state.yaml.tmp', '.atl/sdd-state.yaml')
print('STEP 3: sdd-state.yaml reset to pending')
"

# Step 4: Archive apply-progress.yaml for debugging
if [ -f ".atl/apply-progress.yaml" ]; then
  TS=$(date +%Y%m%d-%H%M%S)
  mv ".atl/apply-progress.yaml" ".atl/apply-progress.yaml.${TS}.rollback"
  echo "STEP 4: apply-progress.yaml archived as .rollback"
fi

echo "ROLLBACK COMPLETE"
echo "Apply branch preserved: ${APPLY_BRANCH}"
echo "Inspect with: git log ${APPLY_BRANCH} --oneline"
echo "Discard with: git branch -D ${APPLY_BRANCH}"
echo "Resume with:  /sdd-continue"
```

### Rollback for Config Corruption

```bash
#!/usr/bin/env bash
# architect-ai restore {filename}
# Restores most recent backup from .atl/backups/
set -euo pipefail

TARGET="${1:?filename required}"
BACKUP_DIR=".atl/backups"
BASENAME=$(basename "${TARGET}")

LATEST=$(ls -t "${BACKUP_DIR}/${BASENAME}".*.bak 2>/dev/null | head -1)
if [ -z "${LATEST}" ]; then
  echo "ERROR: no backup found for ${TARGET}" >&2; exit 1
fi

echo "Restoring: ${TARGET} from ${LATEST}"
cp "${LATEST}" "${TARGET}.restore.tmp"
mv "${TARGET}.restore.tmp" "${TARGET}"
echo "RESTORED: ${TARGET}"
```
