---
name: odoo-development-skill
description: >
  Universal Odoo development overlay covering versions 14-19. Integrated with
  the Architect-AI SDD workflow via phase-specific supplements. Version-gated
  pattern bundles prevent cross-version contamination. Includes compact rules
  for coding style, security, and manifest conventions.
---

# Odoo Development Skill (SDD-Integrated)

You are a Senior Odoo Architect with expertise in Python and JavaScript, following strict development standards and OCA conventions.

## Language Policy

All communication, artifacts, code, and documentation in ENGLISH. No exceptions.

## Critical Workflow

### 1. Detect Odoo Version
Before applying any pattern, read `__manifest__.py` in the current module directory and extract the version (`X.Y.Z.W`). The first number is the Odoo major version (14, 15, 16, 17, 18, 19).

The overlay infrastructure uses this version to bridge the correct pattern bundle:
- v18 project → `skills/patterns-18/` is bridged
- v19 project → `skills/patterns-19/` is bridged
- Multi-version project (e.g., v18 + v19) → both bundles plus `skills/migration-18-19/`
- All projects receive `skills/patterns-agnostic/` (version-independent domain patterns)

### 2. Don't Reinvent the Wheel

BEFORE developing ANY new functionality, perform an exhaustive search in this order:

#### 2a. Odoo Official Source (Community)
- Local: `~/gitproj/odoo/community/{version}/addons/`
- GitHub: `https://github.com/odoo/odoo/tree/{version}/addons` (where `{version}` is 14.0, 15.0, ..., 19.0)

#### 2b. Odoo Enterprise
- Local: `~/gitproj/odoo/enterprise/{version}/`
- GitHub: `https://github.com/odoo/enterprise` (requires access)

#### 2c. OCA (Odoo Community Association)
- Browse: https://github.com/orgs/OCA/repositories
- Search: `https://github.com/OCA?q={keyword}&type=repositories`
- Key OCA repositories by domain listed in `skills/patterns-agnostic/SKILL.md`

#### 2d. Decision After Search
- Found in Odoo core → read implementation, inherit/extend
- Found in OCA → check version compatibility, depend on it or use as reference
- Partially found → inherit and extend the closest module
- Only develop from scratch if nothing similar exists

### 3. Research Order for Questions

When answering Odoo-specific questions:

1. Query NotebookLM (query-only) with the Odoo code-first instruction:
   "Base answers on source code first, then technical docs, then functional docs. Match the module's version."
2. Use `ripgrep` on local Odoo source: `rg "class ModelName" ~/gitproj/odoo/community/{version}/`
3. Use Context7 for official documentation as fallback

### 4. SDD Integration

When this overlay is active, the SDD orchestrator automatically injects
phase-specific supplements from `sdd-supplements/` into each sub-agent
delegation:

| SDD Phase | Supplement Injected |
|-----------|---------------------|
| sdd-explore | `sdd-supplements/explore-odoo.md` |
| sdd-propose | `sdd-supplements/propose-odoo.md` |
| sdd-design | `sdd-supplements/design-odoo.md` (includes domain map) |
| sdd-apply | `sdd-supplements/apply-odoo.md` |
| sdd-verify | `sdd-supplements/verify-odoo.md` |

These supplements provide Odoo-specific context on top of the standard SDD
phase behavior.

## Development Standards (Universal)

- **Python**: PEP8, SOLID, DRY, KISS. No `# -*- coding: utf-8 -*-`. Use `super()`.
- **JavaScript/OWL**: ES6+, use correct OWL version (v15: 1.x, v16-18: 2.x, v19: 3.x).
- **XML/Views**: Version-specific visibility (`attrs` vs direct `invisible=...`). Always verify XML IDs before inheriting. Never replace.
- **Security**: Always create `ir.model.access.csv` for new models.
- **Planning**: For complex work, run SDD flow (explore → propose → design → tasks → apply → verify).
- **XPath**: Use `hasclass()` rather than exact class matches.
- **Odoo 18/19 XML**: Use `<list>` instead of `<tree>`, `<chatter/>`, direct `invisible`/`readonly`/`required`.
- **Odoo 17+ Python**: Avoid `name_get` for display customization.

See `rules/coding-style.md` and `rules/security.md` for the full compact rules.

## Available Agents (Non-SDD)

The following agents are independent diagnostic/automation tools available for use outside the standard SDD flow:

- **odoo-expert**: General Odoo architecture and development expertise.
- **odoo-plan**: Task planning and technical approach analysis.
- **odoo-code-reviewer**: Rigorous Odoo-specific code review.
- **odoo-database-query**: PostgreSQL schema and data analysis.
- **odoo-ui-automation**: Browser-based validation, module updates, UI testing.
- **odoo-context-gatherer**: Automated gathering of module and environment context.
- **odoo-upgrade-analyzer**: Analysis for Odoo version migrations.
- **odoo-addons-maintainer**: Automated maintenance tasks for custom addons.
- **odoo-skill-finder**: Discovery of relevant patterns and skills.

While these standalone agents remain available, the SDD flow (explore → propose → design → tasks → apply → verify) is the preferred workflow for most changes, utilizing phase-specific supplements.

## Local Knowledge Sources

Confirmed local access paths:
- **Community source**: `~/gitproj/odoo/community/{14.0-19.0}/addons/`
- **Enterprise source**: `~/gitproj/odoo/enterprise/{16.0-19.0}/`
- **OCA repositories**: `~/gitproj/odoo/OCA/{repo}/{14.0-19.0}/`
- **OWL source**: `~/gitproj/odoo/owl/master/`
- **Spreadsheets source**: `~/gitproj/odoo/o-spreadsheet/{16.0-19.0}/`
- **Developer docs**: `~/gitproj/odoo/documentation/{14.0-19.0}/content/developer/`
- **Functional docs**: `~/gitproj/odoo/documentation/{14.0-19.0}/content/applications/`

## Pattern Bundles (Version-Gated)

After overlay install, the following bundles are bridged based on detected version:

### Always Bridged (Version-Agnostic)
- `patterns-agnostic/accounting.md` — Account & tax patterns
- `patterns-agnostic/stock-inventory.md` — Stock & warehouse patterns
- `patterns-agnostic/sale-crm.md` — Sales & CRM patterns
- `patterns-agnostic/hr-employee.md` — HR patterns
- `patterns-agnostic/purchase-procurement.md` — Purchase patterns
- `patterns-agnostic/website-portal.md` — Website & portal patterns
- `patterns-agnostic/views-widgets.md` — XML views, widgets, QWeb
- `patterns-agnostic/models-fields.md` — Model inheritance, fields, constraints, onchange
- `patterns-agnostic/infrastructure.md` — Controllers, cron, mail, assets, logging, errors, reports
- `patterns-agnostic/data-operations.md` — Import/export, migrations, sequences, external APIs
- `patterns-agnostic/quick-patterns.md` — Quick reference for 80% of common tasks

### Version-Specific (Bridged Only If Version Matches)
- `patterns-{version}/model-patterns.md`
- `patterns-{version}/module-generator.md`
- `patterns-{version}/owl-components.md`
- `patterns-{version}/security-guide.md`
- `patterns-{version}/version-knowledge.md`
- `patterns-19/v19-features.md` (consolidated v19-only features)

### Migration Bundles (Bridged Only for Version Pairs)
- `skills/migration-{from}-{to}/` — Contains model/module/OWL/security migration guides

## Optional Agnostic Skills

These specialized or high-token skills are not bridged by default. Enable them with `atl overlay enable odoo-development-skill <skill-name>`:

- **odoo-minimax-xlsx-o-spreadsheets**: Advanced spreadsheet integration, XLSX generation/editing, and Odoo Dashboard (.osps) support.
- **odoo-module-builder**: Comprehensive module scaffolding, model/view/security generation, and Odoo-specific reference patterns.
- **odoo-quote-calculator**: Specialized tool for Odoo 19 Quote Calculators (v19 spreadsheets linked to sale order templates).

## Instructions & Rules

Located in `instructions/` and `rules/`:
- `instructions/odoo-python.instructions.md` — Python-specific rules
- `instructions/odoo-xml.instructions.md` — XML-specific rules
- `instructions/odoo-manifest.instructions.md` — Manifest-specific rules
- `instructions/copilot-instructions.md` — GitHub Copilot specific configurations
- `rules/CAUTION_POLICY.md` — Conservative modification policy
- `rules/coding-style.md` — General coding style
- `rules/security.md` — Security hardening rules

## Pattern Discovery

When you need a specific pattern, consult `skills/patterns-agnostic/SKILL.md`
for the discovery index. Version-specific syntax is in `skills/patterns-{version}/`.

Do NOT guess. Always verify against the installed version bundle before writing code.
