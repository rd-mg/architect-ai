# SDD Verify — Odoo Context

When verifying in an Odoo project, apply these domain-specific checks IN ADDITION to the standard sdd-verify.

## Deterministic Checklist

Apply ALL of these checks. Each is automatable via `ripgrep` or file existence.

### Manifest Checks
- [ ] `__manifest__.py` exists in module root
- [ ] `version` field is X.Y.Z.W format
- [ ] `version` was incremented since last commit on this module
  ```bash
  git log --oneline --diff-filter=M -- __manifest__.py | head -5
  ```
- [ ] `depends` list contains all imported modules (cross-check with `_inherit` and Python imports)
- [ ] `external_dependencies` declared if new Python libraries are used
- [ ] `data` list is in correct load order (security → data → views → menus)

### Security Checks
- [ ] `security/ir.model.access.csv` exists for EVERY new model
  ```bash
  # Find new models
  rg "_name\s*=\s*['\"]" models/ --files-with-matches
  # Check each appears in ir.model.access.csv
  cat security/ir.model.access.csv
  ```
- [ ] Record rules exist for multi-company fields (`company_id`)
- [ ] No `sudo()` without documented justification
  ```bash
  rg "\.sudo\(\)" --files-with-matches
  # Each match should have a comment like "# sudo: accessing res.users as portal user"
  ```
- [ ] No raw SQL without parameterization
  ```bash
  rg "cr\.execute\(" -A 2 --files-with-matches
  # Each should use %s placeholders, not f-strings or string formatting
  ```
- [ ] No user-controlled input in XML IDs or domains

### Code Quality Checks
- [ ] `hasclass()` used in XPath (not `contains(@class, ...)`)
  ```bash
  rg "contains\(@class" views/
  # Expected: no matches
  ```
- [ ] No `attrs=` in v17+ code
  ```bash
  rg "attrs\s*=\s*\"" views/
  # Expected (v17+): no matches
  ```
- [ ] No `<tree>` in v18+ views
  ```bash
  rg "<tree" views/
  # Expected (v18+): no matches
  ```
- [ ] OWL version matches Odoo version:
  - v15: OWL 1.x (`require('@odoo/owl')`)
  - v16-18: OWL 2.x (`import { Component } from "@odoo/owl"`)
  - v19: OWL 3.x (new patterns)
- [ ] No `@api.model_create_multi` missing on `create` overrides
- [ ] All computed fields have complete `@api.depends(...)`
- [ ] No unbounded `search([])` (must have limit, offset, or narrow domain)
- [ ] No N+1 queries (browsing inside loops over large recordsets)

### Documentation Checks
- [ ] `README.md` exists in module root
- [ ] README describes: purpose, features, configuration, dependencies
- [ ] Changelog entry exists for this version (in README or CHANGELOG.md)

### Migration Checks (Conditional)
If schema changes were made in this batch:
- [ ] Pre-migrate script present in `migrations/{version}/pre-migrate.py`
- [ ] Post-migrate script present if data transformation needed
- [ ] New required fields have default values or computed defaults
- [ ] Field renames/removals have proper migration logic

### Test Checks
- [ ] Test file exists for each new capability in `tests/`
- [ ] Tests pass: `python odoo-bin -i {module} --test-enable --stop-after-init`
- [ ] Tests cover happy path AND at least one error case
- [ ] No `print()` or `import pdb` left in code
- [ ] Judgement Day Gate executed and PASSED (Required for COMPLETE status)

## Adversarial Review Focus Areas

Apply adaptive-reasoning Mode 2 (adversarial-review) with these Odoo-specific lenses:

### Pass A: Local correctness lens
- Does the code do what the spec says?
- Are all declared fields actually used?
- Do views reference fields that exist?
- Do method signatures match their invocations?

### Pass B: System impact lens
- What OTHER modules inherit from this? Did we break them?
- What happens during module upgrade from the previous version?
- What happens if the user installs this module on a database with existing data?
- What happens if the user uninstalls this module? Are there orphan records?
- Does this change affect multi-company behavior?
- Does this change interact with Enterprise features?
- Does this interact with the studio module (user customizations)?

## Human-Required Checklist (Agent Reminds, Does Not Execute)

Include these reminders in the return envelope:

- [ ] Module installs on a clean database without errors
- [ ] Module upgrades on an existing database without errors
- [ ] Feature tested with correct user roles (not only admin)
- [ ] Multi-company behavior verified (if applicable)
- [ ] Multi-currency behavior verified (if applicable)
- [ ] No JavaScript console errors in browser
- [ ] Main flows tested end-to-end on staging
- [ ] Edge cases identified by the developer tested

## Output Format

In your verify-report, include:

```markdown
## Odoo-Specific Verification

### Deterministic Checks
✅ Manifest version bumped (18.0.1.0.0 → 18.0.1.1.0)
✅ ir.model.access.csv exists for acme.approval.request
❌ README.md missing in module root — BLOCKER
⚠️ sudo() used in models/approval.py:45 without justification comment

### Adversarial Findings
- CRITICAL: Uninstall leaves orphan records in acme_approval_log
- WARNING (real): Multi-company not respected in acme.approval.request.search
- SUGGESTION: Consider adding tracking=True to state field

### Human Verification Required
- [ ] Test on clean DB
- [ ] Test upgrade from 18.0.1.0.0
- [ ] Test as non-admin user

### Verdict
NEEDS CHANGES (README missing, CRITICAL uninstall issue, sudo() needs justification)
```

## Boundaries

- Do NOT mark as APPROVED with any BLOCKER or CRITICAL finding unresolved
- Do NOT skip deterministic checks because "the code looks fine"
- Do NOT run tests in production-adjacent environments without user confirmation

## JUDGEMENT DAY GATE (ejecutar antes de marcar verify como COMPLETE)

### Cuándo activar
- El sub-agente sdd-verify está a punto de reportar "PASS" al Orchestrator
- Mode 1 activo (razonamiento expansivo disponible — contexto no saturado)
- Es la PRIMERA verificación completa del módulo (no en re-verificaciones parciales)

### Cuándo NO activar
- Mode 3 activo (contexto saturado — Judgement Day gastaría demasiado)
- Re-verificación de un sub-componente ya aprobado
- Tarea trivial (Complexity = 0-1 en el Classifier)

### Protocolo de ejecución

**INPUT:** El Brief aprobado (mem_context de sdd/{module}/brief/v{N})
**POSTURE:** +++Critical (obligatorio — Judgement Day sin Critical no tiene sentido)
**TOKEN BUDGET:** máx 600 tokens (si excede → truncar forzadamente)

**PROMPT BLOCK (inyectar en Capa 8 cuando se activa):**

> [!IMPORTANT]
> [JUDGEMENT DAY GATE — verificación final del Brief]
> Antes de cerrar sdd-verify como COMPLETE, audita el Brief activo contra los siguientes 3 criterios críticos para Odoo:
> 
> 1. INTEGRIDAD DEL MÓDULO: ¿El módulo puede desinstalarse sin dejar datos huérfanos en tablas del core? ¿Todos los campos relacionales tienen ondelete definido?
> 2. COLISIÓN CON CORE: ¿El diseño modifica directamente modelos del core (res.partner, account.move, res.users) sin usar herencia? ¿O crea duplicados de funcionalidad que OCA ya provee?
> 3. ESCALABILIDAD: ¿Hay operaciones N+1 en los computed fields? ¿Los search filters usan índices? ¿El diseño funciona con 1M registros?
> 
> Para cada criterio: PASS o FAIL + descripción de 1 oración.
> Veredicto final: PASS (los 3 OK) o FAIL (cualquier falla grave).
> 
> Si FAIL: especifica solo el cambio mínimo requerido. No reescribas el Brief.
> Si PASS: devuelve "JUDGEMENT DAY: PASS" y continúa con el cierre.

### Resultado y acción post-Judgement Day

- **PASS** → after_model hook guarda:
  `mem_save("sdd/{module}/brief/v{N}", {brief + "judgement_day": "PASS"})`
  Orchestrator avanza a la siguiente fase.
- **FAIL** → Orchestrator re-abre sdd-design con las correcciones como input.
  Se crea brief/v{N+1} (incrementa versión).
  Contador de failures NO se incrementa (es una corrección de diseño, no un error).

### Timeout y protección
Si el sub-agente tarda más de 2 ciclos en Judgement Day → asumir PASS con advertencia:
"JUDGEMENT DAY: TIMEOUT — Brief aprobado con advertencia. Revisar manualmente criterios 1-3 antes de deploy."

