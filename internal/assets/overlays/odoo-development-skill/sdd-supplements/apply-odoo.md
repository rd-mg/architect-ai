# SDD Apply — Odoo Context

When applying changes in an Odoo project, follow this protocol IN ADDITION to the standard sdd-apply behavior.

## Odoo Version
{See shared preamble — version already cached from sdd-init.}

## File Order Within a Module

When creating/modifying module files, respect Odoo's import order:

```
my_module/
├── __init__.py              # Imports models/, controllers/, wizards/
├── __manifest__.py          # Declares data files in load order
├── models/
│   ├── __init__.py
│   ├── {model_name}.py      # Inherit order: base first, then mixins, then new
├── controllers/
│   ├── __init__.py
│   └── {controller}.py
├── views/
│   └── {model_name}_views.xml    # Namespace by model
├── security/
│   ├── ir.model.access.csv  # MUST exist for every new model
│   └── {group}_security.xml
├── data/
│   └── {data_type}_data.xml
├── demo/
│   └── {data_type}_demo.xml
├── static/
│   └── description/
│       ├── icon.png         # REQUIRED
│       └── index.html       # Optional
├── migrations/
│   └── {version}/
│       ├── pre-migrate.py
│       └── post-migrate.py
├── tests/
│   └── test_{feature}.py
└── README.md
```

## Manifest data[] Load Order

The order of entries in `data` is CRITICAL. Follow this sequence:

1. Security definitions (groups, categories)
2. `security/ir.model.access.csv`
3. Record rules (security/record_rules.xml)
4. Base data (required by views)
5. Views (tree/list, form, kanban, search — typically in one file per model)
6. Actions (actions referenced by menus)
7. Menus (menus reference actions)
8. Reports
9. Wizards' views
10. Cron jobs / scheduled actions
11. Templates (mail templates, QWeb reports)

Wrong order = install failure. Double-check before committing.

## Migration Script Scaffolding

When the change involves schema changes, scaffold migrations:

### Adding a required field to an existing model
```python
# {module}/migrations/{new-version}/post-migrate.py
from odoo import api, SUPERUSER_ID

def migrate(cr, version):
    if not version:
        return  # Fresh install; no migration needed

    env = api.Environment(cr, SUPERUSER_ID, {})
    # Populate the new required field for existing records
    env['my.model'].search([]).write({'new_field': 'default_value'})
```

### Renaming a field
```python
# {module}/migrations/{new-version}/pre-migrate.py
def migrate(cr, version):
    if not version:
        return

    # Rename column before ORM load
    cr.execute("""
        ALTER TABLE my_model
        RENAME COLUMN old_name TO new_name
    """)
```

### Removing a model
```python
# {module}/migrations/{new-version}/pre-migrate.py
def migrate(cr, version):
    if not version:
        return

    # Preserve data if needed
    cr.execute("SELECT * FROM deprecated_model")
    # Store, log, or archive before removal

    cr.execute("DROP TABLE IF EXISTS deprecated_model CASCADE")
```

## Security Files

For EVERY new model, create `security/ir.model.access.csv` BEFORE writing the model file. The model won't be accessible otherwise.

```csv
id,name,model_id:id,group_id:id,perm_read,perm_write,perm_create,perm_unlink
access_my_model_user,my.model.user,model_my_model,base.group_user,1,0,0,0
access_my_model_manager,my.model.manager,model_my_model,base.group_system,1,1,1,1
```

## Code Patterns to ENFORCE

From `rules/coding-style.md` and Odoo conventions:

- `hasclass('o_state_button')` NOT `contains(@class, 'o_state_button')`
- `super().method()` NOT `super(MyClass, self).method()`
- `@api.depends('field1', 'field2')` with ALL dependencies explicit
- `tracking=True` on fields that need audit trail
- `<list>` NOT `<tree>` in v18+
- `invisible="state == 'done'"` NOT `attrs="{'invisible': [('state', '=', 'done')]}"` in v17+

## DDD Tactical Implementation

Reference `skills/patterns-ddd/SKILL.md` for implementation details of:
- **Invariants**: Always use `@api.constrains` for rules that must persist regardless of UI.
- **Service Orchestration**: Use `models.AbstractModel` to group logic that doesn't fit in a single model.
- **Reusable Filters**: Implement Specifications as `@api.model` methods returning domains.

## README Auto-Generation

When a module is created or significantly modified, generate/update README.md:

```markdown
# {Module Name}

{description from manifest}

## Features
- {bullet per capability}

## Configuration
- {list ir.config_parameter entries}
- {list required groups/permissions}
- {list required external APIs}

## Dependencies
- {from manifest depends}

## Changelog
## [{current-version}] - {today's date}
### Added / Changed / Fixed
- {description of this change}
```

## Test Scaffolding

For each capability in the spec, create at minimum:
- Unit test in `tests/test_{capability}.py`
- Test class inherits from `odoo.tests.TransactionCase` (for ORM tests) or `HttpCase` (for HTTP tests)

## Size Budget

Respect the 400-word progress report limit. Code changes themselves are separate artifacts.

## Boundaries

- Do NOT modify files outside the scope of assigned tasks
- Do NOT silently skip the manifest version bump
- Do NOT commit code without `ir.model.access.csv` for new models
- Do NOT use `attrs=` in v17+
- Do NOT use `<tree>` in v18+

## Branch & PR Checklist (from branch-pr-odoo.md)

### Branch Naming
Format: `{type}/{odoo-version}/{ticket-id}-{slug}`
Examples:
- `feat/18.0/OCA-42-approval-workflow`
- `fix/18.0/CUDIO-101-manifest-version-bump`

Types: `feat`, `fix`, `refactor`, `chore`, `docs`

### Pre-PR Verification
Before opening a pull request, confirm:
- [ ] Manifest version bumped
- [ ] All rg-based verify checks pass locally
- [ ] `ir.model.access.csv` is complete
- [ ] No debug print() or _logger.warning("DEBUG") statements
- [ ] Migration scripts present if schema changed
- [ ] CHANGELOG.md or README updated with this version's changes

### PR Description Template
```
## What
{1-2 sentence summary}

## Odoo Version(s)
{v18 | v19 | v18+v19}

## Changes
- {module}: {what changed}

## Testing
- [ ] Unit tests pass
- [ ] Installed on clean DB
- [ ] Upgraded from previous version (if migration scripts exist)

## Migration Required
{YES — describe | NO}
```
