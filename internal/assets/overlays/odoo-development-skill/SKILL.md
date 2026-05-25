---
name: odoo-development-skill
description: >
  Universal Odoo development overlay covering versions 14-19. Integrated with
  the Architect-AI SDD workflow via phase-specific supplements. Version-gated
  pattern bundles prevent cross-version contamination. Includes compact rules
  for coding style, security, and manifest conventions.
---

# Odoo Development Skill (SDD-Integrated)

Senior Odoo Architect — Python + JavaScript. Strict standards, OCA conventions.

## Language Policy

ALL communication, artifacts, code, documentation: ENGLISH.

## Critical Workflow

### 1. Detect Odoo Version
Read `__manifest__.py` in current module dir. Extract major version from `X.Y.Z.W` (14, 15, 16, 17, 18, 19).

Version bridges:
- v18 → `skills/patterns-18/`
- v19 → `skills/patterns-19/`
- Multi-version → both + `skills/migration-18-19/`
- All projects → `skills/patterns-agnostic/`

### 2. Don't Reinvent the Wheel

Search BEFORE developing:

#### 2a. Odoo Official Source (Community)
- Local: `${ODOO_COMMUNITY}/{version}/addons/` (resolved via `.atl/config.yaml` or auto-discovery — see `_shared/odoo-path-resolution.md`)
- GitHub: `https://github.com/odoo/odoo/tree/{version}/addons` (14.0–19.0)

#### 2b. Odoo Enterprise
- Local: `~/gitproj/odoo/enterprise/{version}/` (or configured via `odoo_enterprise_path` in `.atl/config.yaml`)
- GitHub: `https://github.com/odoo/enterprise`

#### 2c. OCA (Odoo Community Association)
- Browse: https://github.com/orgs/OCA/repositories
- Search: `https://github.com/OCA?q={keyword}&type=repositories`

#### 2d. Decision After Search
- Found in Odoo core → inherit/extend
- Found in OCA → check version, depend or reference
- Partially found → inherit/extend closest
- Nothing similar → develop from scratch

### 3. Research Order
See `sdd-supplements/explore-odoo.md` Research Fallback Chain.
Non-explore phases: Engram → rg → Context7 (no web unless explicit).

### 4. SDD Integration

Phase-specific supplements auto-injected by orchestrator:

| SDD Phase | Supplement Injected |
|-----------|---------------------|
| sdd-init | `sdd-supplements/init-odoo.md` |
| sdd-explore | `sdd-supplements/explore-odoo.md` |
| sdd-propose | `sdd-supplements/propose-odoo.md` |
| sdd-spec | `sdd-supplements/spec-odoo.md` |
| sdd-design | `sdd-supplements/design-odoo.md` |
| sdd-tasks | `sdd-supplements/tasks-odoo.md` |
| sdd-apply | `sdd-supplements/apply-odoo.md` |
| sdd-verify | `sdd-supplements/verify-odoo.md` |
| sdd-archive | `sdd-supplements/archive-odoo.md` |

## Development Standards (Universal)

- **Python**: PEP8, SOLID, DRY, KISS. No `# -*- coding: utf-8 -*-`. Use `super()`.
- **JavaScript/OWL**: ES6+. v15: OWL 1.x, v16-18: 2.x, v19: 3.x.
- **XML/Views**: Verify XML IDs before inheriting. Never replace.
- **Security**: `ir.model.access.csv` required for new models.
- **Planning**: SDD flow (explore → propose → design → tasks → apply → verify).
- **XPath**: `hasclass()` not exact class matches.
- **Odoo 18/19 XML**: `<list>` not `<tree>`, `<chatter/>`, direct `invisible`/`readonly`/`required`.
- **Odoo 17+ Python**: Avoid `name_get` for display customization.

See `rules/coding-style.md`, `rules/security.md`.

## Available Agents (Non-SDD)

- **odoo-expert**: Architecture and development expertise.
- **odoo-plan**: Task planning and technical approach.
- **odoo-code-reviewer**: Odoo-specific code review.
- **odoo-database-query**: PostgreSQL schema/data analysis.
- **odoo-ui-automation**: Browser validation, module updates, UI testing.
- **odoo-context-gatherer**: Module/environment context gathering.
- **odoo-upgrade-analyzer**: Version migration analysis.
- **odoo-addons-maintainer**: Custom addon maintenance.
- **odoo-skill-finder**: Pattern/skill discovery.

Preferred workflow: SDD flow (explore → propose → design → tasks → apply → verify).

## Local Knowledge Sources

Path resolution via `${ODOO_COMMUNITY}` (set in `.atl/config.yaml` or auto-discovered). Falls back to Engram + Context7 if source unavailable.

- **Community source**: `${ODOO_COMMUNITY}/{version}/addons/`
- **Enterprise source**: `${ODOO_ENTERPRISE:-~/gitproj/odoo/enterprise}/{version}/`
- **OCA repositories**: `~/gitproj/odoo/OCA/{repo}/{version}/`
- **OWL source**: `~/gitproj/odoo/owl/master/`
- **Spreadsheets**: `~/gitproj/odoo/o-spreadsheet/{version}/`
- **Developer docs**: `~/gitproj/odoo/documentation/{version}/content/developer/`
- **Functional docs**: `~/gitproj/odoo/documentation/{version}/content/applications/`

## Pattern Bundles (Version-Gated)

### Always Bridged (Version-Agnostic)
- `patterns-agnostic/accounting.md`
- `patterns-agnostic/stock-inventory.md`
- `patterns-agnostic/sale-crm.md`
- `patterns-agnostic/hr-employee.md`
- `patterns-agnostic/purchase-procurement.md`
- `patterns-agnostic/website-portal.md`
- `patterns-agnostic/views-widgets.md`
- `patterns-agnostic/models-fields.md`
- `patterns-agnostic/infrastructure.md`
- `patterns-agnostic/data-operations.md`
- `patterns-agnostic/quick-patterns.md`

### Version-Specific (Bridged only if version matches)
- `patterns-{version}/model-patterns.md`
- `patterns-{version}/module-generator.md`
- `patterns-{version}/owl-components.md`
- `patterns-{version}/security-guide.md`
- `patterns-{version}/version-knowledge.md`
- `patterns-19/v19-features.md`

### Migration Bundles (Bridged only for version pairs)
- `skills/migration-{from}-{to}/`

## Optional Agnostic Skills

Enable: `atl overlay enable odoo-development-skill <skill-name>`

- **odoo-minimax-xlsx-o-spreadsheets**: XLSX generation, Odoo Dashboard (.osps).
- **odoo-module-builder**: Module scaffolding, model/view/security generation.
- **odoo-quote-calculator**: Odoo 19 Quote Calculators (v19 spreadsheets + sale order templates).

## Instructions & Rules

- `instructions/odoo-python.instructions.md`
- `instructions/odoo-xml.instructions.md`
- `instructions/odoo-manifest.instructions.md`
- `instructions/copilot-instructions.md`
- `rules/CAUTION_POLICY.md`
- `rules/coding-style.md`
- `rules/security.md`

## Rule File Scoping

Rules without `project_scope` → ALL Odoo projects.
Rules with `project_scope: {prefix}` → only modules matching prefix.

| File | Scope |
|------|-------|
| coding-style.md | universal |
| security.md | universal |
| cudio-git.md | cudio projects only |
| cudio-naming.md | cudio projects only |
| coding-style-cudio-append.md | cudio projects only |
| security-cudio-append.md | cudio projects only |

## Pattern Discovery

Consult `skills/patterns-agnostic/SKILL.md` for discovery index. Version-specific syntax in `skills/patterns-{version}/`.
NEVER guess. Verify against installed version bundle before writing code.

## External References Policy [MANDATORY]

NEVER reference external repositories by URL/SSH in skill manifests.
All external resources via:
1. Engram knowledge nodes (pre-indexed during skill-registry)
2. Context7 MCP: `context7.resolve_library_id("odoo")`
3. researcher agent: scope_hint="docs", max_depth="standard"

OCA module search: delegate to researcher → Tier 3 (Context7) → Tier 5 (web if deep).
