> **IMPORTANT (WORKSPACE RULE):** This reference explicitly targets older Odoo versions. Remember that for **Odoo 19.0**, you MUST use `<list>` (not `<tree>`), `<chatter/>`, `_compute_display_name`, and avoid `attrs`. For **Odoo 18.0**, use direct attributes. See `.github/copilot-instructions.md` for strict version variations, and consult Context7 documentation (https://context7.com/...) as your PREFERRED source of truth. Use local workspace assets only as a SECONDARY fallback.


# Odoo Hooks Reference

## Table of Contents
- [Overview](#overview)
- [post_init_hook](#post_init_hook)
- [pre_init_hook](#pre_init_hook)
- [uninstall_hook](#uninstall_hook)
- [post_load](#post_load)
- [Manifest Registration](#manifest-registration)

## Overview

Hooks are functions that run at specific points in a module's lifecycle (install, upgrade, uninstall). Define them in `__init__.py` (or a dedicated file) and register in `__manifest__.py`.

## post_init_hook

Runs after module installation. Use for data migration, populating computed fields, or one-time setup that can't be done with XML data files.

```python
# __init__.py
from . import models

def post_init_hook(env):
    """Populate default data after install."""
    books = env['library.book'].search([('reference', '=', False)])
    for book in books:
        book.reference = env['ir.sequence'].next_by_code('library.book')
```

**Odoo 16 signature** (uses `cr, registry` instead of `env`):

```python
def post_init_hook(cr, registry):
    env = api.Environment(cr, SUPERUSER_ID, {})
    books = env['library.book'].search([('reference', '=', False)])
    for book in books:
        book.reference = env['ir.sequence'].next_by_code('library.book')
```

### Common Use Cases
- Populate computed stored fields for pre-existing records
- Set default values on new columns for existing data
- Create initial configuration records
- Run data migration scripts

## pre_init_hook

Runs before module installation. Use for precondition checks or preparing the database.

```python
def pre_init_hook(env):
    """Check prerequisites before install."""
    cr = env.cr
    cr.execute("SELECT count(*) FROM ir_module_module WHERE name='sale' AND state='installed'")
    if not cr.fetchone()[0]:
        raise Exception("Please install the Sales module first.")
```

**Odoo 16 signature:**

```python
def pre_init_hook(cr):
    """Run raw SQL before ORM is available."""
    cr.execute("""
        ALTER TABLE library_book ADD COLUMN IF NOT EXISTS reference VARCHAR;
    """)
```

### Common Use Cases
- Validate prerequisites
- Add columns via raw SQL before ORM expects them
- Check for incompatible modules

## uninstall_hook

Runs when the module is uninstalled. Use for cleanup.

```python
def uninstall_hook(env):
    """Clean up on uninstall."""
    env['ir.cron'].search([
        ('name', 'like', 'Library:%'),
    ]).unlink()
```

**Odoo 16 signature:**

```python
def uninstall_hook(cr, registry):
    env = api.Environment(cr, SUPERUSER_ID, {})
    env['ir.cron'].search([('name', 'like', 'Library:%')]).unlink()
```

### Common Use Cases
- Remove cron jobs or server actions
- Clean up ir.config_parameter entries
- Archive or delete records that would become orphaned

## post_load

Runs when the module is loaded into memory (before install). Very rare — used for monkey-patching or low-level framework modifications.

```python
def post_load():
    """Modify framework behavior at load time."""
    pass  # Use sparingly
```

## Manifest Registration

```python
# __manifest__.py
{
    'name': 'Library Management',
    # ...
    'pre_init_hook': 'pre_init_hook',
    'post_init_hook': 'post_init_hook',
    'uninstall_hook': 'uninstall_hook',
    'post_load': 'post_load',       # very rare
}
```

All hooks are optional. Only declare the ones you need. The function names are strings referencing functions importable from the module's `__init__.py`.
