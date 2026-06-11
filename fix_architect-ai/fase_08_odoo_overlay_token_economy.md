# Fase 8: Odoo Overlay — Token Economy y YOLO Guard

**Objetivo:** Resolver F-08 (saturación de contexto por payloads masivos de Odoo MCP), F-12 (CAUTION_POLICY.md inyectado completo en cada sub-agent), y F-08-B (mutaciones con D4>=2 + YOLO activo). Modifica únicamente archivos Markdown de assets y el script de validación de result-contract.

---

## Paso 1: Reestructurar `§7 Odoo Rules Injection` con tiers por fase

**Archivo a modificar:** `internal/assets/_shared/odoo-overlay-routing.md`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```markdown
### 7. Odoo Rules Injection (alongside base guardrails)
```
.atl/overlays/odoo-development-skill/rules/coding-style.md     → compact rules
.atl/overlays/odoo-development-skill/rules/security.md          → compact rules
.atl/overlays/odoo-development-skill/rules/CAUTION_POLICY.md    → full (critical)
.atl/overlays/odoo-development-skill/rules/cudio-git.md         → compact rules
```
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```markdown
### 7. Odoo Rules Injection (alongside base guardrails)

Rules are injected in two tiers to minimize token overhead. The orchestrator
reads `current_phase` from sdd-state.yaml and selects the appropriate tier.

#### Tier A — Read-only phases (sdd-explore, sdd-propose, sdd-spec, sdd-design, sdd-tasks, sdd-archive)

```
.atl/overlays/odoo-development-skill/rules/coding-style.md → compact rules  (~1.5KB)
.atl/overlays/odoo-development-skill/rules/security.md     → compact rules  (~2KB)
.atl/overlays/odoo-development-skill/rules/CAUTION_POLICY.md → SUMMARY ONLY (~500 chars)
  Full CAUTION_POLICY available via: mem_search("knowledge/odoo/caution-policy")
  Load full version only when: task involves write ops OR D1 >= 2 OR security risk detected
```

**Tier A total injection: ~4KB per sub-agent** (vs ~13KB before — 69% reduction)

#### Tier B — Write phases (sdd-apply, sdd-verify)

```
.atl/overlays/odoo-development-skill/rules/coding-style.md     → compact rules  (~1.5KB)
.atl/overlays/odoo-development-skill/rules/security.md          → compact rules  (~2KB)
.atl/overlays/odoo-development-skill/rules/CAUTION_POLICY.md    → FULL           (~8KB)
.atl/overlays/odoo-development-skill/rules/cudio-git.md         → compact rules  (~1KB)
```

**Tier B total injection: ~14KB per sub-agent** (same as before — no quality regression on write phases)

#### Tier Selection Logic (inject at orchestrator launch prompt construction)

```
IF current_phase IN [sdd-explore, sdd-propose, sdd-spec, sdd-design, sdd-tasks, sdd-archive]:
  → USE Tier A injection
ELIF current_phase IN [sdd-apply, sdd-verify]:
  → USE Tier B injection
ELSE (general-orchestrator non-SDD tasks with IS_ODOO=true):
  → USE Tier A injection (default to minimal)
```

### 8. Odoo MCP Pagination Guard (MANDATORY for ALL mcp_odoo_search_records calls)

Every call to `mcp_odoo_search_records` MUST include an explicit `limit` parameter.

**Maximum limit per call: 50 records.**

PROHIBITED patterns (will exceed context budget and trigger D4 saturation):
```
mcp_odoo_search_records(model="sale.order", domain=[])
mcp_odoo_search_records(model="account.move", domain=[], limit=None)
mcp_odoo_search_records(model="stock.move", fields=["all"], domain=[])
```

REQUIRED pattern:
```
mcp_odoo_search_records(model="sale.order", domain=[("state","=","sale")], limit=50)
```

For volume/count analysis: use `mcp_odoo_aggregate_records` (returns metadata only, no payload).
For specific record exploration: filter `domain` to target ≤ 50 records before calling.

If more than 50 records are needed: paginate with `offset` in sequential calls, processing
one page at a time and summarizing to masked_evidence before fetching the next page.

### 9. YOLO Mode Guard (MANDATORY when ODOO_YOLO=true)

When `ODOO_YOLO=true` is set in the MCP environment:

**CHECK before ANY write operation (create, write, delete, unlink):**

```
IF D3 >= 2 OR D4 >= 2:
  BLOCK the write operation.
  EMIT: "[YOLO_GUARD] Context degraded (D3={D3}, D4={D4}). Write operation suspended."
  EMIT: "Resolve context pressure before executing mutations. Run context-guardian or /compact."
  DO NOT proceed with the write — return status: "blocked", blocked_reason: "yolo_guard_d_score"

IF D3 < 2 AND D4 < 2:
  Proceed with YOLO write (normal behavior).
```

**CHECK for large-volume writes:**

```
IF records_to_modify > 100 AND ODOO_YOLO=true:
  STOP and emit to user:
  "[YOLO_GUARD] About to modify {N} records in {model}.
   This operation cannot be undone automatically.
   Confirm? Type: confirm-yolo-{model}-{N}"
  Wait for exact confirmation string before proceeding.
```
```

---

## Paso 2: Actualizar `result-contract.md` con los nuevos status values y campos obligatorios

**Archivo a modificar:** `internal/assets/_shared/result-contract.md`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```markdown
## Validation Rules

1. Valid JSON syntax required.
2. All keys above required.
3. `status` field must be one of: `completed`, `failed`, `blocked`, `abandoned`.
4. Orchestrator validates via `.atl/scripts/validate-result-contract.sh`. On failure: increment attempt, retry phase.
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```markdown
## Validation Rules

1. Valid JSON syntax required.
2. All keys above required (including `estimated_tokens`, `attempt_number`, `error_type`).
3. `status` field must be one of:
   `completed`, `failed`, `blocked`, `abandoned`,
   `infrastructure_blocked`, `partially_completed`
4. `error_type` field must be one of: `domain`, `infrastructure`, `none`
5. `blocked_reason` is REQUIRED (non-empty string) when `status` is `blocked` or `infrastructure_blocked`.
6. `artifacts` must be non-empty when `status` is `completed` AND `phase` is one of:
   `sdd-spec`, `sdd-design`, `sdd-apply`, `sdd-verify`, `sdd-archive`
7. `next_recommended` MUST NOT be `sdd-archive` when `status` is `failed` or `blocked`.
8. `attempt_number` MUST match `circuit_breaker.attempt_counts[phase]` from `sdd-state.yaml`.
9. Orchestrator validates via `.atl/scripts/validate-result-contract.sh`.
   On domain failure: increment `attempt_counts[phase]`, retry phase.
   On infrastructure failure (`error_type: infrastructure`): increment `infra_attempt_counts[phase]`,
   retry WITHOUT incrementing `attempt_counts[phase]`.
```

---

## Paso 3: Añadir campos nuevos al JSON Schema del result-contract

**Archivo a modificar:** `internal/assets/_shared/result-contract.md`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```json
{
  "status": "completed|failed|blocked|abandoned",
  "phase": "sdd-explore",
  "change_name": "string",
  "executive_summary": "string",
  "artifacts": ["string"],
  "next_recommended": "string",
  "risks": ["string"],
  "skill_resolution": {
    "status": "paths-injected|fallback-registry|fallback-path|none",
    "skills_used": ["string"],
    "fallback_reason": null
  },
  "attempt_number": 1,
  "blocked_reason": null
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```json
{
  "status": "completed|failed|blocked|abandoned|infrastructure_blocked|partially_completed",
  "phase": "sdd-explore",
  "change_name": "string",
  "executive_summary": "string (min 50 chars)",
  "artifacts": ["string (path or engram key)"],
  "next_recommended": "string",
  "risks": ["string"],
  "skill_resolution": {
    "status": "paths-injected|fallback-registry|fallback-path|none",
    "skills_used": ["string"],
    "fallback_reason": null
  },
  "attempt_number": 1,
  "blocked_reason": null,
  "error_type": "domain|infrastructure|none",
  "estimated_tokens": 0,
  "audit_mode": "normal|degraded"
}
```

---

## Paso 4: Actualizar el script de validación de result-contract

**Archivo a modificar:** `.atl/scripts/validate-result-contract.sh`

**Acción:** Modificar (reemplazar contenido completo)

```bash
#!/usr/bin/env python3
# .atl/scripts/validate-result-contract.sh
# Validates a result contract JSON block emitted by SDD phase agents.
# Usage: echo '{...json...}' | python3 .atl/scripts/validate-result-contract.sh
# Exit 0 = valid. Exit 1 = validation error (stderr has details).

import json
import sys


VALID_STATUSES = {
    "completed", "failed", "blocked", "abandoned",
    "infrastructure_blocked", "partially_completed",
}

VALID_ERROR_TYPES = {"domain", "infrastructure", "none"}

ARTIFACT_REQUIRED_PHASES = {
    "sdd-spec", "sdd-design", "sdd-apply", "sdd-verify", "sdd-archive",
}

INVALID_NEXT_FOR_STATUS = {
    "failed":    {"sdd-archive"},
    "blocked":   {"sdd-archive"},
    "abandoned": {"sdd-archive", "sdd-verify", "sdd-apply"},
}

REQUIRED_KEYS = {
    "status", "phase", "change_name", "executive_summary",
    "artifacts", "next_recommended", "risks", "skill_resolution",
    "attempt_number", "blocked_reason", "error_type", "estimated_tokens",
}


def validate(data: dict) -> list[str]:
    errors = []

    # 1. Required keys
    for key in REQUIRED_KEYS:
        if key not in data:
            errors.append(f"missing required key: '{key}'")

    if errors:
        return errors  # cannot proceed without required keys

    status = data["status"]
    phase = data["phase"]
    artifacts = data.get("artifacts", [])
    blocked_reason = data.get("blocked_reason") or ""
    next_rec = data.get("next_recommended", "")
    attempt_number = data.get("attempt_number", 0)
    error_type = data.get("error_type", "")
    executive_summary = data.get("executive_summary", "")

    # 2. status enum
    if status not in VALID_STATUSES:
        errors.append(
            f"invalid status '{status}'; valid values: {sorted(VALID_STATUSES)}"
        )

    # 3. error_type enum
    if error_type not in VALID_ERROR_TYPES:
        errors.append(
            f"invalid error_type '{error_type}'; valid values: {sorted(VALID_ERROR_TYPES)}"
        )

    # 4. blocked_reason required when blocked/infrastructure_blocked
    if status in ("blocked", "infrastructure_blocked"):
        if not blocked_reason.strip():
            errors.append(
                f"status='{status}' requires a non-empty 'blocked_reason'"
            )

    # 5. artifacts non-empty on completed for artifact-producing phases
    if status == "completed" and phase in ARTIFACT_REQUIRED_PHASES:
        if not artifacts:
            errors.append(
                f"phase='{phase}' with status='completed' must have at least one artifact"
            )

    # 6. invalid next_recommended transitions
    invalid_nexts = INVALID_NEXT_FOR_STATUS.get(status, set())
    if next_rec in invalid_nexts:
        errors.append(
            f"status='{status}' cannot recommend next='{next_rec}'"
        )

    # 7. executive_summary minimum length
    if len(executive_summary.strip()) < 20:
        errors.append(
            f"executive_summary too short ({len(executive_summary)} chars); minimum 20 chars"
        )

    # 8. attempt_number type check
    if not isinstance(attempt_number, int) or attempt_number < 0:
        errors.append(
            f"attempt_number must be a non-negative integer, got {attempt_number!r}"
        )

    # 9. skill_resolution shape check
    sr = data.get("skill_resolution", {})
    if not isinstance(sr, dict):
        errors.append("skill_resolution must be a JSON object")
    else:
        sr_status = sr.get("status", "")
        valid_sr = {"paths-injected", "fallback-registry", "fallback-path", "none"}
        if sr_status not in valid_sr:
            errors.append(
                f"skill_resolution.status='{sr_status}'; valid values: {sorted(valid_sr)}"
            )

    return errors


def main():
    raw = sys.stdin.read().strip()
    if not raw:
        print("ERROR: empty input", file=sys.stderr)
        sys.exit(1)

    try:
        data = json.loads(raw)
    except json.JSONDecodeError as e:
        print(f"ERROR: invalid JSON: {e}", file=sys.stderr)
        sys.exit(1)

    errors = validate(data)
    if errors:
        print("RESULT CONTRACT VALIDATION FAILED:", file=sys.stderr)
        for err in errors:
            print(f"  - {err}", file=sys.stderr)
        sys.exit(1)

    print(f"OK: result contract valid (status={data['status']}, phase={data['phase']})")
    sys.exit(0)


if __name__ == "__main__":
    main()
```

**Comando para hacerlo ejecutable:**
```bash
chmod +x .atl/scripts/validate-result-contract.sh
```

---

## Verificación de Fase

```bash
# 1. Verificar sintaxis del script de validación
python3 -m py_compile .atl/scripts/validate-result-contract.sh && echo "Python syntax OK"

# 2. Test: contrato válido
echo '{
  "status": "completed",
  "phase": "sdd-explore",
  "change_name": "add-payment",
  "executive_summary": "Explored the payment module structure and identified integration points.",
  "artifacts": ["sdd/add-payment/explore-report"],
  "next_recommended": "sdd-propose",
  "risks": [],
  "skill_resolution": {"status": "paths-injected", "skills_used": ["sdd-explore"], "fallback_reason": null},
  "attempt_number": 1,
  "blocked_reason": null,
  "error_type": "none",
  "estimated_tokens": 4200,
  "audit_mode": "normal"
}' | python3 .atl/scripts/validate-result-contract.sh

# 3. Test: contrato con status=completed y artifacts vacíos para sdd-spec (debe fallar)
echo '{
  "status": "completed",
  "phase": "sdd-spec",
  "change_name": "add-payment",
  "executive_summary": "Spec written.",
  "artifacts": [],
  "next_recommended": "sdd-design",
  "risks": [],
  "skill_resolution": {"status": "paths-injected", "skills_used": [], "fallback_reason": null},
  "attempt_number": 1,
  "blocked_reason": null,
  "error_type": "none",
  "estimated_tokens": 3000,
  "audit_mode": "normal"
}' | python3 .atl/scripts/validate-result-contract.sh || echo "EXPECTED FAILURE: artifacts empty for sdd-spec"

# 4. Test: status=failed recomienda sdd-archive (debe fallar)
echo '{
  "status": "failed",
  "phase": "sdd-apply",
  "change_name": "add-payment",
  "executive_summary": "Apply failed due to import error in models.py.",
  "artifacts": [],
  "next_recommended": "sdd-archive",
  "risks": ["import error"],
  "skill_resolution": {"status": "paths-injected", "skills_used": [], "fallback_reason": null},
  "attempt_number": 2,
  "blocked_reason": null,
  "error_type": "domain",
  "estimated_tokens": 8000,
  "audit_mode": "normal"
}' | python3 .atl/scripts/validate-result-contract.sh || echo "EXPECTED FAILURE: failed cannot recommend sdd-archive"

# 5. Test: infrastructure_blocked sin blocked_reason (debe fallar)
echo '{
  "status": "infrastructure_blocked",
  "phase": "sdd-spec",
  "change_name": "test",
  "executive_summary": "Gate template not resolved.",
  "artifacts": [],
  "next_recommended": "sdd-spec",
  "risks": [],
  "skill_resolution": {"status": "none", "skills_used": [], "fallback_reason": null},
  "attempt_number": 1,
  "blocked_reason": "",
  "error_type": "infrastructure",
  "estimated_tokens": 500,
  "audit_mode": "normal"
}' | python3 .atl/scripts/validate-result-contract.sh || echo "EXPECTED FAILURE: infrastructure_blocked needs blocked_reason"

# 6. Verificar que odoo-overlay-routing.md tiene las secciones nuevas
python3 -c "
with open('internal/assets/_shared/odoo-overlay-routing.md') as f:
    content = f.read()
required = ['Tier A', 'Tier B', 'Tier Selection Logic', 'Pagination Guard', 'YOLO Mode Guard', '50 records']
for r in required:
    assert r in content, f'MISSING: {r}'
print('odoo-overlay-routing.md has all required sections')
"

# 7. Verificar result-contract.md actualizado
python3 -c "
with open('internal/assets/_shared/result-contract.md') as f:
    content = f.read()
for r in ['infrastructure_blocked', 'estimated_tokens', 'error_type', 'audit_mode']:
    assert r in content, f'MISSING: {r}'
print('result-contract.md has all new fields')
"

# 8. Compilar todo (ningún archivo Go fue modificado)
go build ./...
```

