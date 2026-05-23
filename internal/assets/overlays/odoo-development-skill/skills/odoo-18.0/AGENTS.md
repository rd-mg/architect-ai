# Odoo 18 — AI Agents Setup

Setup guide for Odoo 18 documentation with AI coding assistants (Cursor, Claude Code, Windsurf, Aider, etc.).

## Quick Start

### Install via skills.sh (Recommended)

```bash
# Add Odoo 18 skill to your project
npx skills add unclecatvn/agent-skills
```

Visit [https://skills.sh/](https://skills.sh/) for more install options.

### Cursor IDE — Remote Rule

Configure once in Cursor settings:
- `Settings` → `Rules` → `Add Remote Rule`
- Source: `Git Repository`
- URL: `git@github.com:unclecatvn/agent-skills.git`
- Branch: `18.0`
- Subfolder: `skills/odoo-18.0/`

---

## Documentation Structure

```
skills/odoo-18.0/
├── SKILL.md                       # Master index (all agents)
├── references/                    # Dev guides (18 files)
│   ├── odoo-18-actions-guide.md     # ir.actions.*, cron, bindings
│   ├── odoo-18-controller-guide.md  # HTTP, routing, controllers
│   ├── odoo-18-data-guide.md        # XML/CSV data files, records
│   ├── odoo-18-decorator-guide.md   # @api decorators
│   ├── odoo-18-development-guide.md # Manifest, wizards (overview)
│   ├── odoo-18-field-guide.md       # Field types, parameters
│   ├── odoo-18-manifest-guide.md    # __manifest__.py reference
│   ├── odoo-18-mixins-guide.md      # mail.thread, activities, etc.
│   ├── odoo-18-model-guide.md       # ORM, CRUD, search, domain
│   ├── odoo-18-migration-guide.md   # Migration scripts, hooks
│   ├── odoo-18-owl-guide.md         # OWL components, services
│   ├── odoo-18-performance-guide.md # N+1 prevention, optimization
│   ├── odoo-18-reports-guide.md     # QWeb reports, PDF/HTML
│   ├── odoo-18-security-guide.md    # ACL, record rules, security
│   ├── odoo-18-testing-guide.md     # Test classes, decorators
│   ├── odoo-18-transaction-guide.md # Savepoints, errors
│   ├── odoo-18-translation-guide.md # Translations, i18n
│   └── odoo-18-view-guide.md        # XML views, QWeb
├── CLAUDE.md                      # Claude Code specific
└── AGENTS.md                      # THIS FILE — setup guide
```

---

## Guide Reference

| File | Purpose | When to Use |
|------|---------|-------------|
| `SKILL.md` | Master index | Find right guide for task |
| `references/odoo-18-actions-guide.md` | Actions (window, URL, server, cron) | Create actions, menus, scheduled jobs |
| `references/odoo-18-controller-guide.md` | HTTP controllers, routing | Write endpoints |
| `references/odoo-18-data-guide.md` | XML/CSV data files, records | Create data files |
| `references/odoo-18-decorator-guide.md` | @api decorators | Use @api decorators |
| `references/odoo-18-development-guide.md` | Module structure, wizards | Create new modules |
| `references/odoo-18-field-guide.md` | Field types, parameters | Define model fields |
| `references/odoo-18-manifest-guide.md` | __manifest__.py reference | Configure module manifest |
| `references/odoo-18-mixins-guide.md` | mail.thread, activities, mixins | Add messaging, activities |
| `references/odoo-18-model-guide.md` | ORM methods, CRUD, domains | Write model methods |
| `references/odoo-18-migration-guide.md` | Migration scripts, hooks | Upgrade modules |
| `references/odoo-18-owl-guide.md` | OWL components, hooks, services | Build OWL UI |
| `references/odoo-18-performance-guide.md` | Performance optimization | Fix slow code |
| `references/odoo-18-reports-guide.md` | QWeb reports, templates | Create reports |
| `references/odoo-18-security-guide.md` | ACL, record rules, security | Configure security |
| `references/odoo-18-testing-guide.md` | Test classes, decorators, mocking | Write tests |
| `references/odoo-18-transaction-guide.md` | DB transactions, error handling | Savepoints, UniqueViolation |
| `references/odoo-18-translation-guide.md` | Translations, localization, i18n | Add translations |
| `references/odoo-18-view-guide.md` | XML views, actions, menus | Write view XML |

---

## AI Agent Configuration

### Cursor IDE

| Setting | Value |
|---------|-------|
| Source | Git Repository |
| URL | `git@github.com:unclecatvn/agent-skills.git` |
| Branch | `18.0` |
| Subfolder | `skills/odoo-18.0/` |

**Globs patterns used by Cursor:**

| File | globs Pattern |
|------|---------------|
| `SKILL.md` | `**/*.{py,xml}` |
| `references/odoo-18-actions-guide.md` | `**/*.{py,xml}` |
| `references/odoo-18-controller-guide.md` | `**/controllers/**/*.py` |
| `references/odoo-18-data-guide.md` | `**/*.{xml,csv}` |
| `references/odoo-18-decorator-guide.md` | `**/models/**/*.py` |
| `references/odoo-18-development-guide.md` | `**/*.{py,xml,csv}` |
| `references/odoo-18-field-guide.md` | `**/models/**/*.py` |
| `references/odoo-18-manifest-guide.md` | `**/__manifest__.py` |
| `references/odoo-18-mixins-guide.md` | `**/models/**/*.py` |
| `references/odoo-18-model-guide.md` | `**/models/**/*.py` |
| `references/odoo-18-migration-guide.md` | `**/migrations/**/*.py` |
| `references/odoo-18-owl-guide.md` | `static/src/**/*.{js,xml}` |
| `references/odoo-18-performance-guide.md` | `**/*.{py,xml}` |
| `references/odoo-18-reports-guide.md` | `**/report/**/*.xml` |
| `references/odoo-18-security-guide.md` | `**/security/**/*.{csv,xml}` |
| `references/odoo-18-testing-guide.md` | `**/tests/**/*.py` |
| `references/odoo-18-transaction-guide.md` | `**/models/**/*.py` |
| `references/odoo-18-translation-guide.md` | `**/*.{py,js,xml}` |
| `references/odoo-18-view-guide.md` | `**/views/**/*.xml` |

### Claude Code

```bash
# Install via skills.sh
npx skills add unclecatvn/agent-skills
```

Claude Code reads:
- `CLAUDE.md` — Project overview + quick reference
- `SKILL.md` — Master index
- Individual guides in `references/` — Detailed info

### Other Agents

| Agent | Setup |
|-------|-------|
| Windsurf | Same as Cursor (uses `.mdc` files) |
| Continue | Place `CLAUDE.md` or `SKILL.md` in root |
| Aider | Place `CLAUDE.md` or add to prompt |
| OpenCode | Copy skill folder to project — no extra config |

---

## Cursor / Claude Skills Folder

After `npx skills add unclecatvn/agent-skills`, skill placed at:

```
.cursor/skills/
└── odoo-18/
    └── SKILL.md

.claude/skills/
└── odoo-18/
 └── SKILL.md
```

---

## Key Odoo 18 Changes

| Change | Old | New |
|--------|-----|-----|
| List view tag | `<tree>` | `<list>` |
| Dynamic attributes | `attrs="{'invisible': [...]}"` | `invisible="..."` |
| Delete validation | Override `unlink()` | `@api.ondelete(at_uninstall=False)` |
| Field aggregation | `group_operator=` | `aggregator=` |
| SQL queries | `cr.execute()` | `SQL` class with `execute_query_dict()` |

---

## Repository

**URL**: `git@github.com:unclecatvn/agent-skills.git`
**Branch**: `18.0`
**License**: MIT
