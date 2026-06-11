# Fase 7: Context-Guardian — Cooldown, Threshold Dinámico y Versionado de Packs

**Objetivo:** Resolver F-13 (loop de compactación por ausencia de cooldown), SM-06 (D4 no se recalibra post-compactación), y EPI-02 (topic_key de packs sin historial). Modifica únicamente el archivo Markdown del skill; no toca código Go.

---

## Paso 1: Añadir reglas de Cooldown y Threshold Dinámico en `context-guardian/SKILL.md`

**Archivo a modificar:** `internal/assets/skills/context-guardian/SKILL.md`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```markdown
## Auto-Trigger Rules (Orchestrator evaluates BEFORE each delegation)

Orchestrator MUST invoke when ANY condition holds:

1. `char_count(context_history) >= 100_000` → invoke
2. Sub-agent `skill_resolution: none` in last 2 turns → invoke
3. D4 >= 2 in current reasoning evaluation → invoke
4. 3+ file reads in same context window without compaction → invoke
5. User says "compact", "reset context", "what's my state" → invoke
6. `attempt_count >= 2` for current phase → invoke (context may be corrupted)
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```markdown
## Auto-Trigger Rules (Orchestrator evaluates BEFORE each delegation)

Orchestrator MUST invoke when ANY condition holds:

1. `char_count(context_history) >= active_threshold` → invoke
2. Sub-agent `skill_resolution: none` in last 2 turns → invoke
3. D4 >= 2 in current reasoning evaluation → invoke
4. 3+ file reads in same context window without compaction → invoke
5. User says "compact", "reset context", "what's my state" → invoke
6. `attempt_count >= 2` for current phase → invoke (context may be corrupted)

## Cooldown Rule (MANDATORY — prevents compaction loops)

After ANY compaction event:
1. SET `protected_fact: last_compaction_delegation={N}` where N = current delegation count.
2. For the next 3 delegations after compaction: SKIP triggers 1, 2, 3, 4, and 6.
   Only trigger 5 (explicit user request) fires during cooldown.
3. CHECK before evaluating triggers: if `current_delegation - last_compaction_delegation < 3` → skip.

Rationale: skill reload (foundation + compact rules + protocol) adds ~20KB per delegation
after compaction. Without cooldown, the reloaded skills immediately re-trigger the guardian.

## Dynamic Threshold (prevents immediate re-trigger during skill reload)

`active_threshold` is NOT a fixed 100_000 chars. It adjusts:

- **Base threshold**: 100_000 chars
- **Post-compaction threshold** (applies for 5 delegations after compaction):
  `active_threshold = 150_000 chars`
  Accommodates mandatory skill/protocol reload without immediate re-trigger.
- **Return to base**: after 5 delegations post-compaction, `active_threshold = 100_000`.

Track via `protected_fact: post_compact_delegations_remaining={N}` (starts at 5, decrements each delegation).

## Post-Compaction D4 Re-evaluation (MANDATORY)

After `/compact` or `/compress` completes:
1. Re-estimate D4 from char_count of rebuilt context (post-pack context, NOT pre-compact).
2. If `new_D4 < pre_compact_D4`: EMIT `[MODE_RECALIBRATED: Mode{old}→Mode{new} | D4:{old}→{new}]`
3. Update the injected mode header for the NEXT delegation to reflect the new D4.
4. DO NOT continue in the old mode — context pressure was relieved by compaction.

Example:
```
Pre-compact:  D4=3 → Mode 3 (+++Pragmatic)
Post-compact: D4=0 → [MODE_RECALIBRATED: Mode3→Mode1 | D4:3→0]
Next delegation: [MODE 1 | D1=2, D2=1, D3=0, D4=0] Normal task complexity
```
```

---

## Paso 2: Actualizar la sección de Persistencia con versionado de packs

**Archivo a modificar:** `internal/assets/skills/context-guardian/SKILL.md`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```markdown
## Persistence

After assembling Context Pack, persist to Engram:

```
mem_save(
  title: "context-pack/{project}/{change-or-session-id}",
  topic_key: "context-pack/{project}/{change-or-session-id}",
  type: "architecture",
  project: "{project}",
  content: "{full markdown Context Pack}"
)
```

Enables recovery after compaction: orchestrator retrieves via `mem_search` + `mem_get_observation`.
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```markdown
## Persistence

After assembling Context Pack, persist to Engram with versioned topic_key:

```
# Increment compaction_count protected_fact before saving
compaction_count = protected_facts.compaction_count + 1
SET protected_fact: compaction_count={compaction_count}

# Save versioned pack (preserves full history)
mem_save(
  title: "context-pack/{project}/{session-id}/pack-{compaction_count}",
  topic_key: "context-pack/{project}/{session-id}/pack-{compaction_count}",
  type: "architecture",
  project: "{project}",
  content: "{full markdown Context Pack}"
)

# Update the "current" pointer (always points to latest pack)
mem_save(
  title: "context-pack/{project}/{session-id}/current",
  topic_key: "context-pack/{project}/{session-id}/current",
  type: "architecture",
  project: "{project}",
  content: "→ pack-{compaction_count}"
)
```

### Retrieval Protocol

On session resume or post-compaction reload:
1. `mem_search("context-pack/{project}/{session-id}/current")` → get pointer entry
2. `mem_get_observation(pointer_id)` → read "→ pack-N"
3. `mem_search("context-pack/{project}/{session-id}/pack-N")` → get pack entry
4. `mem_get_observation(pack_id)` → read full Context Pack

### Pack Versioning Benefits

- Full compaction history preserved: pack-1, pack-2, … pack-N
- Debug: `architect-ai ctx-history {session-id}` lists all packs with timestamps
- Regression: if pack-3 is corrupt, orchestrator can fall back to pack-2
- DO NOT delete prior packs — Engram upserts by topic_key, prior versions remain accessible

### If Engram is Unavailable

Write emergency pack to filesystem:
```bash
# Written atomically by context-guardian when Engram mem_save fails
.atl/emergency-checkpoint.yaml
```
Content: YAML dump of protected_facts + active_tasks only (no masked_evidence).
Orchestrator reads this file at next session start if Engram probe fails.
```

---

## Paso 3: Añadir regla de Anti-Patterns actualizada

**Archivo a modificar:** `internal/assets/skills/context-guardian/SKILL.md`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```markdown
## Anti-Patterns

- Treating this skill as optional for long sessions — auto-trigger makes it mandatory at threshold
- Rebuilding pack from scratch every turn when only one fact changed (use `mem_update` on existing pack)
- Including irrelevant historical exploration ("we tried X, didn't work") unless it informs active constraint
- Mixing suppression decisions into `masked_evidence` — different sections, different purposes
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```markdown
## Anti-Patterns

- Treating this skill as optional for long sessions — auto-trigger makes it mandatory at threshold
- Rebuilding pack from scratch every turn when only one fact changed (use `mem_update` on existing pack)
- Including irrelevant historical exploration ("we tried X, didn't work") unless it informs active constraint
- Mixing suppression decisions into `masked_evidence` — different sections, different purposes
- Ignoring the Cooldown Rule — triggering compaction on the very next delegation after a compaction (loop)
- Using a fixed 100_000 threshold immediately after compaction — skill reload alone consumes ~20KB
- Continuing in old Mode (e.g. Mode 3) after compaction when D4 has dropped — mode drift wastes tokens
- Saving to a fixed topic_key without incrementing pack version — overwrites history, prevents rollback
- Omitting `compaction_count` from protected_facts — makes version tracking impossible across sessions
```

---

## Paso 4: Crear el script de hook de compactación para Claude Code

**Archivo a crear:** `.atl/hooks/pre_compact_hook.sh`

> **Nota**: Este archivo es generado por `architect-ai install` en el proyecto del usuario. Se crea en el directorio `.atl/hooks/` del proyecto. Es un artefacto de estado, no código fuente.

**Acción:** Crear

```bash
#!/usr/bin/env bash
# .atl/hooks/pre_compact_hook.sh
# Pre-compact hook for architect-ai projects.
# Called by Claude Code hooks system (PreToolUse: Compact) before context compaction.
# Saves an emergency checkpoint to .atl/ when Engram is unavailable.
# Safe to run multiple times (idempotent).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
ATL_DIR="${PROJECT_ROOT}/.atl"
CHECKPOINT_FILE="${ATL_DIR}/emergency-checkpoint.yaml"
TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

mkdir -p "${ATL_DIR}"

# Try Engram checkpoint first (best-effort — never block the compact)
ENGRAM_BIN="${ENGRAM_BIN:-engram}"
if command -v "${ENGRAM_BIN}" > /dev/null 2>&1; then
  "${ENGRAM_BIN}" checkpoint save \
    --project "${ARCHITECT_PROJECT:-unknown}" \
    --timestamp "${TIMESTAMP}" \
    2>/dev/null \
    && echo "[pre_compact_hook] Engram checkpoint saved at ${TIMESTAMP}" \
    || echo "[pre_compact_hook] WARN: Engram checkpoint failed — writing filesystem fallback"
else
  echo "[pre_compact_hook] WARN: engram binary not found — writing filesystem fallback"
fi

# Write filesystem fallback checkpoint (always — belt and suspenders)
CHECKPOINT_TMP="${CHECKPOINT_FILE}.tmp"
cat > "${CHECKPOINT_TMP}" << YAML
# Emergency checkpoint — written by pre_compact_hook.sh
# Read by architect-ai on next session start if Engram probe fails.
timestamp: "${TIMESTAMP}"
hook_version: "1.0"
project: "${ARCHITECT_PROJECT:-unknown}"
platform: "${ARCHITECT_PLATFORM:-unknown}"
note: "Full context pack may be available in Engram under context-pack/{project}/{session-id}/current"
YAML

# Atomic write
mv "${CHECKPOINT_TMP}" "${CHECKPOINT_FILE}"
echo "[pre_compact_hook] Filesystem checkpoint written: ${CHECKPOINT_FILE}"
```

**Comando para hacerlo ejecutable:**
```bash
chmod +x .atl/hooks/pre_compact_hook.sh
```

---

## Paso 5: Registrar el hook en `.claude/settings.json` del proyecto

**Archivo a modificar:** `.claude/settings.json`

**Acción:** Modificar — agregar sección `hooks` si no existe, o extender la existente

**Comando previo:**
```bash
cat .claude/settings.json 2>/dev/null || echo "{}"
```

**Código a insertar/reemplazar dentro del JSON existente** — añadir la clave `"hooks"` al nivel raíz del objeto JSON:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Compact",
        "hooks": [
          {
            "type": "command",
            "command": "bash .atl/hooks/pre_compact_hook.sh"
          }
        ]
      }
    ]
  }
}
```

> **Nota de merge**: si `.claude/settings.json` ya tiene otras claves, añadir solo el bloque `"hooks"` sin reemplazar las demás. Usar el siguiente comando para merge seguro:

```bash
python3 - << 'PYEOF'
import json, os

settings_path = ".claude/settings.json"
hook_entry = {
    "hooks": {
        "PreToolUse": [
            {
                "matcher": "Compact",
                "hooks": [
                    {
                        "type": "command",
                        "command": "bash .atl/hooks/pre_compact_hook.sh"
                    }
                ]
            }
        ]
    }
}

existing = {}
if os.path.exists(settings_path):
    with open(settings_path, "r") as f:
        try:
            existing = json.load(f)
        except json.JSONDecodeError:
            pass

# Deep merge hooks only
if "hooks" not in existing:
    existing["hooks"] = {}
if "PreToolUse" not in existing["hooks"]:
    existing["hooks"]["PreToolUse"] = []

# Check if Compact hook already present
has_compact = any(
    h.get("matcher") == "Compact"
    for h in existing["hooks"]["PreToolUse"]
)
if not has_compact:
    existing["hooks"]["PreToolUse"].append(
        hook_entry["hooks"]["PreToolUse"][0]
    )

tmp_path = settings_path + ".tmp"
with open(tmp_path, "w") as f:
    json.dump(existing, f, indent=2)
    f.write("\n")
os.rename(tmp_path, settings_path)
print(f"Updated {settings_path}")
PYEOF
```

---

## Verificación de Fase

```bash
# 1. Verificar que los cambios en SKILL.md son sintácticamente válidos (YAML frontmatter)
python3 -c "
import re, sys
with open('internal/assets/skills/context-guardian/SKILL.md') as f:
    content = f.read()
# Verify key sections exist
required = [
    'Cooldown Rule',
    'Dynamic Threshold',
    'Post-Compaction D4 Re-evaluation',
    'pack-{compaction_count}',
    'last_compaction_delegation',
    'post_compact_delegations_remaining',
]
for r in required:
    if r not in content:
        print(f'MISSING: {r}')
        sys.exit(1)
print('All required sections present')
"

# 2. Verificar que el hook es ejecutable
chmod +x .atl/hooks/pre_compact_hook.sh
bash -n .atl/hooks/pre_compact_hook.sh && echo "Hook syntax OK"

# 3. Verificar que .claude/settings.json es JSON válido después del merge
python3 -c "
import json
with open('.claude/settings.json') as f:
    data = json.load(f)
hooks = data.get('hooks', {}).get('PreToolUse', [])
has_compact = any(h.get('matcher') == 'Compact' for h in hooks)
assert has_compact, 'Compact hook not found in PreToolUse'
print('Claude settings.json OK — Compact hook registered')
"

# 4. Verificar que las nuevas palabras clave no rompieron ninguna referencia en tests Go
grep -rn "context-guardian" internal/ --include="*.go" | grep -v "_test.go" | head -5

# 5. Compilar todo (ningún archivo Go fue modificado en esta fase — debe compilar limpio)
go build ./...
```

