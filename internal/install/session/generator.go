package session

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// InitialSession generates the starter .atl/session.yaml content
func InitialSession(project string) string {
	return fmt.Sprintf(`# .atl/session.yaml — AUTO-MANAGED by architect-ai L0
version: "1.0"
project: %q
last_updated: %q

execution_mode: "interactive"
delivery_strategy: "ask-on-risk"
artifact_store: "hybrid"
tdd_mode: false

active_change: ""
active_phase: ""
active_agent: "architect"

mcp:
  engram: true
  context7: true
  sequential_thinking: true
  context_mode: true

history: []
`, project, time.Now().UTC().Format(time.RFC3339))
}

// WriteSession writes .atl/session.yaml atomically
func WriteSession(atDir, content string) error {
	path := filepath.Join(atDir, "session.yaml")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	return os.Rename(tmp, path)
}

// InstallRollbackScripts writes rollback scripts to .atl/scripts/
func InstallRollbackScripts(atDir string) error {
	scriptsDir := filepath.Join(atDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		return fmt.Errorf("create scripts dir: %w", err)
	}

	scripts := map[string]string{
		"rollback-apply.sh":         rollbackApplyScript,
		"resolve-task-order.py":     resolveTaskOrderScript,
		"backup-before-mutate.sh":   backupScript,
	}

	for name, content := range scripts {
		path := filepath.Join(scriptsDir, name)
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}

const rollbackApplyScript = `#!/usr/bin/env bash
set -euo pipefail
CHANGE="${1:?change_name required}"
APPLY_BRANCH="apply/$(echo "${CHANGE}" | tr '[:upper:] ./_' '[:lower:]----')"
git checkout "$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo main)" 2>/dev/null
echo "ROLLBACK: back on original branch. Apply branch preserved: ${APPLY_BRANCH}"
`

const resolveTaskOrderScript = `#!/usr/bin/env python3
import yaml, sys
from collections import defaultdict, deque
with open('.atl/apply-progress.yaml') as f:
    data = yaml.safe_load(f)
tasks = {t['id']: t for t in data.get('tasks', [])}
graph = defaultdict(list)
in_degree = {tid: 0 for tid in tasks}
for tid, task in tasks.items():
    for dep in task.get('depends_on', []):
        graph[dep].append(tid)
        in_degree[tid] += 1
queue = deque([tid for tid, deg in in_degree.items() if deg == 0])
order = []
while queue:
    curr = queue.popleft()
    order.append(curr)
    for n in graph[curr]:
        in_degree[n] -= 1
        if in_degree[n] == 0:
            queue.append(n)
if len(order) != len(tasks):
    print("ERROR: circular dependency", file=sys.stderr); sys.exit(1)
for i, tid in enumerate(order, 1):
    t = tasks[tid]
    print(f"{i}. [{tid}] {t['description']}")
`

const backupScript = `#!/usr/bin/env bash
set -euo pipefail
TARGET="${1:?target required}"
BACKUP_DIR=".atl/backups"
mkdir -p "${BACKUP_DIR}"
[ -f "${TARGET}" ] && cp "${TARGET}" "${BACKUP_DIR}/$(basename "${TARGET}").$(date +%Y%m%d-%H%M%S).bak"
ls -t "${BACKUP_DIR}/$(basename "${TARGET}")"*.bak 2>/dev/null | tail -n +6 | xargs rm -f 2>/dev/null || true
`
