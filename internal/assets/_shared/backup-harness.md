## Backup Harness [architect-ai install/sync operations — MANDATORY]

### Purpose
Prevent data loss when the installer mutates critical configuration files.
Create a timestamped backup BEFORE any mutation. Restore if mutation fails.

### Files requiring backup before mutation
- `opencode.json`
- `CLAUDE.md`
- `GEMINI.md`
- `.github/copilot-instructions.md`
- `.antigravity/agent.md`
- `.vscode/mcp.json`
- `.claude/settings.json`
- `.gemini/settings.json`
- `.atl/sdd-state.yaml`
- `.atl/session.yaml`
- `.atl/skill-manifest.yaml`

### Backup Protocol
```bash
#!/usr/bin/env bash
# backup-before-mutate.sh
set -euo pipefail

TARGET="${1:?target file required}"
BACKUP_DIR=".atl/backups"
mkdir -p "${BACKUP_DIR}"

if [ -f "${TARGET}" ]; then
  TIMESTAMP=$(date +%Y%m%d-%H%M%S)
  BASENAME=$(basename "${TARGET}")
  BACKUP="${BACKUP_DIR}/${BASENAME}.${TIMESTAMP}.bak"
  cp "${TARGET}" "${BACKUP}"
  echo "BACKUP: ${TARGET} → ${BACKUP}"

  # Keep only last 5 backups per file
  ls -t "${BACKUP_DIR}/${BASENAME}".*.bak 2>/dev/null | tail -n +6 | xargs rm -f 2>/dev/null || true
fi
```

### Retention Policy
- Max 5 backups per file
- Auto-purge older backups on new backup creation
- Backup location: `.atl/backups/` (gitignored)
- Backup format: `{filename}.{YYYYMMDD-HHMMSS}.bak`
