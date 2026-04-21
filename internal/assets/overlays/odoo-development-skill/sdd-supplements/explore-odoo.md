# SDD Explore — Odoo Context

When exploring in an Odoo project, follow this protocol IN ADDITION to the standard sdd-explore behavior.

## Version Detection (MANDATORY)

Read `__manifest__.py` → extract major version. ALL analysis must be version-specific from this point forward.

```bash
rg '"version"' __manifest__.py
# Format: "version": "18.0.1.0.0" → major version = 18
```

If multiple modules are in scope, note their versions — they might differ.

## Research Fallback Chain

Priority order — stop at the first successful result. Do NOT skip steps silently.

### Step 1 — NotebookLM (MCP)
- Use the `mcp-notebooklm-orchestrator` skill
- Inject: "Base answers on source code first, then technical docs. Match version {version}."
- **If MCP unavailable**: emit `"SKIP: NotebookLM MCP offline"` → proceed to Step 2

### Step 2 — rg on community source
```bash
test -d "$ODOO_COMMUNITY_PATH" && \
  rg "class ModelName" "$ODOO_COMMUNITY_PATH/{version}/addons/" -t py || \
  echo "SKIP: ODOO_COMMUNITY_PATH not set or path not found"
```
- **If env unset or path missing**: emit the SKIP message → proceed to Step 3

### Step 3 — rg on enterprise source
```bash
test -d "$ODOO_ENTERPRISE_PATH" && \
  rg "class ModelName" "$ODOO_ENTERPRISE_PATH/{version}/" -t py || \
  echo "SKIP: ODOO_ENTERPRISE_PATH not set or path not found"
```
- **If env unset or path missing**: emit the SKIP message → proceed to Step 4

### Step 4 — Context7 MCP
- Use Context7 for official Odoo documentation and API reference.

### Step 5 — Web search (last resort)
- Include `site:github.com/odoo` or `site:github.com/OCA` filter when possible.

> ⚠️ If Steps 1–3 are ALL skipped: **declare the limitation explicitly** in your artifact
> before proceeding with Steps 4–5. Do NOT silently fall through.

### OCA Search
Before proposing new functionality, also verify OCA:
```bash
# Or browse: https://github.com/OCA?q={keyword}&type=repositories
rg "class ModelName" ~/gitproj/odoo/oca/ -t py
```


## Don't Reinvent the Wheel

Before proposing new functionality, verify it doesn't already exist:
- ✅ Odoo community (`~/gitproj/odoo/community/`)
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
