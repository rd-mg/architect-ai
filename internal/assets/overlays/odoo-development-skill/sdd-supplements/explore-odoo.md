# SDD Explore — Odoo Context

When exploring in an Odoo project, follow this protocol IN ADDITION to the standard sdd-explore behavior.

## Version Detection (MANDATORY)

Read `__manifest__.py` → extract major version. ALL analysis must be version-specific from this point forward.

```bash
rg '"version"' __manifest__.py
# Format: "version": "18.0.1.0.0" → major version = 18
```

If multiple modules are in scope, note their versions — they might differ.

## Research Order

1. **Query NotebookLM (query-only)**:
   - Use the `mcp-notebooklm-orchestrator` skill
   - Inject the Odoo instruction: "Base answers on source code first, then technical docs, then functional docs. Match version {version}."
   - Formulate specific questions: "What does module X do in Odoo {version}?"

2. **Search local Odoo source**:
   ```bash
   rg "class ModelName" ~/gitproj/odoo/odoo/ -t py
   rg "_inherit.*'sale.order'" ~/gitproj/odoo/odoo/ -t py
   ```

3. **Search OCA repositories**:
   ```bash
   rg "class ModelName" ~/gitproj/odoo/oca/ -t py
   # Or browse: https://github.com/OCA?q={keyword}&type=repositories
   ```

4. **Use Context7 as fallback** for official Odoo documentation.

## Don't Reinvent the Wheel

Before proposing new functionality, verify it doesn't already exist:
- ✅ Odoo core (`~/gitproj/odoo/odoo/`)
- ✅ Odoo Enterprise (`~/gitproj/odoo/enterprise/`, if path configured)
- ✅ OCA repositories (`~/gitproj/odoo/oca/` or https://github.com/OCA)

If something similar exists, explore it. Decide: inherit/extend, use as reference, or combine.

## Upgrade Analysis

When the exploration involves version migration, focus on:
- Breaking API changes between source and target version
- OWL version transitions (1.x → 2.x → 3.x)
- `attrs` removal (v17+), `tree` → `list` rename (v18+)
- SQL builder requirements (v19+)
- Field type changes and default value behavior
- `@api.depends` inheritance chain changes

Refer to `skills/migration-{from}-{to}/` bundles for version-specific migration patterns.

## Deliverables

Your exploration artifact MUST include:

### Version Context
```markdown
## Version Context
- Target version: {version}
- Detected version(s) in scope: {list with module names}
- Cross-version concerns: {yes/no}
```

### Existing Solutions Found
```markdown
## Existing Solutions
| Module | Location | Fit | Notes |
|--------|----------|-----|-------|
| account.move | core | partial | Lacks multi-tier approval |
| oca/account_invoice_approval | OCA | strong | v18 compatible; uses different state machine |
```

### Domain Placement
```markdown
## Domain Placement
This feature affects: {Sales | Stock | Accounting | HR | ...}
Cross-domain dependencies: {list any}
Anti-corruption layer needed: {yes/no, where}
```
(Reference `sdd-supplements/domain-map.md` for domain definitions)

### Persist Findings

Save verified discoveries to Engram:
```
mem_save(
  title: "odoo-explore/{change-name}/{topic}",
  topic_key: "odoo-explore/{change-name}/{topic}",
  type: "discovery",
  project: "{project}",
  content: "{version-specific findings}"
)
```

## Boundaries

- Do NOT modify any files in this phase
- Do NOT generate boilerplate modules — exploration only
- Do NOT assume version compatibility across the version range
- Do NOT skip OCA search because "it's probably not there"
