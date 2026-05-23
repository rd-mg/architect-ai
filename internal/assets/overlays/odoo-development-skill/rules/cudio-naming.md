---
name: cudio-naming
project_scope: cudio
---

# Cudio Module Naming & Manifest Convention

Organization-specific rules for Cudio Inc. modules. Extends `odoo-development-skill` overlay.

## Module Naming

### Client Modules
Format: `{client_prefix}_{core_app}_{descriptive_name}`

Rules:
- `client_prefix`: client name lowercase, underscore suffix
- `core_app`: if extending Odoo core app, include technical name (sale, stock, account, hr, etc.)
- `descriptive_name`: short functionality description
- Characters: lowercase letters, numbers, underscores ONLY
- Avoid generic names — MUST reflect client AND purpose

Examples:
- `acme_account_invoice_report`
- `acme_google_drive_import`
- `mega_corp_stock_customization`
- `acme_sale_custom_approval`

### Internal Cudio Modules
Non-client-specific modules:

Format: `cudio_{core_app}_{descriptive_name}`

Examples:
- `cudio_google_drive_import`
- `cudio_api_connector`
- `cudio_stock_customization`

### Notes
- **Folder name** (technical name) MUST follow convention
- **`__manifest__.py` name field** can be user-friendly: `"Acme | Invoice Report"`

## __manifest__.py Required Fields

Follow official Odoo Manifest guidelines, plus Cudio-specific values:

### Required Fields

| Field | Value |
|-------|-------|
| `name` | `"{Customer Name} \| {Module Title}"` or `"Cudio \| {Module Title}"` |
| `category` | Clear category aligned with Odoo's existing categories |
| `version` | `X.Y.Z.W` where `X.Y` = Odoo major version |
| `summary` | Short one-line summary |
| `description` | Multi-line string describing purpose, features, main components |
| `author` | `"Cudio Inc."` |
| `company` | `"Cudio Inc."` |
| `maintainer` | `"Cudio Inc."` |
| `website` | `"https://www.cudio.com"` |
| `license` | `"OPL-1"` unless otherwise agreed |
| `installable` | `True` |

### Example Manifest

```python
{
    "name": "Cudio | Google Drive Import",
    "category": "Hidden/Tools",
    "version": "18.0.1.0.0",
    "summary": "Common functionality for importing files from Google Drive",
    "description": """
        This module provides a mixin class and common functionality for importing files from Google Drive.
        It includes:
        - Google Drive file operations (search, read, move, archive)
        - Access token validation and refresh
        - Email notifications for import results
        - Error handling and logging
        - Cron job management for automated imports
    """,
    "author": "Cudio Inc.",
    "company": "Cudio Inc.",
    "maintainer": "Cudio Inc.",
    "website": "https://www.cudio.com",
    "depends": [
        "base",
        "mail",
        "google_api_credentials",
    ],
    "external_dependencies": {
        "python": ["pandas"]
    },
    "data": [
        "security/ir.model.access.csv",
        "views/google_drive_import_mixin_views.xml",
        "data/mail_template_data.xml",
    ],
    "installable": True,
    "application": False,
    "auto_install": False,
    "license": "OPL-1",
}
```

## Version Bump Rule

Every behavior-modifying change MUST increment version:

- `X.Y` = Odoo major version (constant: 18.0, 19.0, etc.)
- `Z` = incremented for major updates (new features, model/view changes)
- `W` = incremented for minor updates (bug fixes, small improvements)

Agent enforcement: verify-odoo phase flags code change without version increment.

## Module Icon

- `icon.png` MUST be present in `static/description/`
- Use standard Cudio icon when applicable
- verify-odoo flags if missing

## Documentation Language

All `__manifest__.py` descriptions and code comments: **English only**.

`README.rst` and `index.html` can be client's preferred language (agent defaults to English).

## Validation Regex (for automation)

- Client module name: `^[a-z][a-z0-9_]*$` (folder name)
- Manifest name field: `^(Cudio|[A-Z][a-zA-Z0-9 ]*) \| .+$`
- Version field: `^\d+\.\d+\.\d+\.\d+$` with first two matching Odoo version

## Compact Rule Summary (for skill registry)

```
### cudio-naming
- Client modules: `{client}_{core_app}_{description}` (lowercase, underscores only)
- Cudio modules: `cudio_{core_app}_{description}`
- Manifest `name`: "{Customer} | {Title}" or "Cudio | {Title}"
- Manifest `version`: X.Y.Z.W (X.Y = Odoo version, Z = major, W = minor)
- Author/company/maintainer: "Cudio Inc.", website: "https://www.cudio.com"
- License: "OPL-1" (unless otherwise agreed)
- icon.png required in static/description/
- EVERY code change MUST bump version (Z for features, W for fixes)
```
