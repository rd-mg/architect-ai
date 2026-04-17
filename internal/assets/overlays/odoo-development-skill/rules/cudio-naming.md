# Cudio Module Naming & Manifest Convention

Organization-specific rules for modules developed at Cudio Inc. Extends the
generic `odoo-development-skill` overlay.

## Module Naming

### Client Modules
Format: `{client_prefix}_{core_app}_{descriptive_name}`

Rules:
- `client_prefix`: client's name in lowercase, followed by underscore
- `core_app`: if the module extends an Odoo core application, include its technical name (sale, stock, account, hr, etc.)
- `descriptive_name`: short description of the functionality
- Characters: lowercase letters, numbers, and underscores ONLY
- Avoid generic names — names MUST reflect both the client AND the purpose

Examples:
- `acme_account_invoice_report`
- `acme_google_drive_import`
- `mega_corp_stock_customization`
- `acme_sale_custom_approval`

### Internal Cudio Modules
For modules that are not client-specific:

Format: `cudio_{core_app}_{descriptive_name}`

Examples:
- `cudio_google_drive_import`
- `cudio_api_connector`
- `cudio_stock_customization`

### Notes
- The **folder name** (technical name) MUST follow this convention
- The **name field** in `__manifest__.py` can be a user-friendly title: `"Acme | Invoice Report"`

## __manifest__.py Required Fields

Follow the official Odoo Manifest guidelines, plus these Cudio-specific values:

### Required Fields

| Field | Value |
|-------|-------|
| `name` | `"{Customer Name} \| {Module Title}"` or `"Cudio \| {Module Title}"` |
| `category` | Clear category aligned with Odoo's existing categories |
| `version` | `X.Y.Z.W` where `X.Y` = Odoo major version |
| `summary` | Short one-line summary |
| `description` | Multi-line string describing purpose, features, and main components |
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

Every change that modifies behavior MUST increment the version:

- `X.Y` = Odoo major version (stays constant: 18.0, 19.0, etc.)
- `Z` = incremented for major updates (new features, model/view changes)
- `W` = incremented for minor updates (bug fixes, small improvements)

Agent enforcement: the verify-odoo phase will flag if code changed but version didn't increment.

## Module Icon

- `icon.png` MUST be present in `static/description/` directory
- Use the standard Cudio icon when applicable
- Verify-odoo will flag if missing

## Documentation Language

All descriptions in `__manifest__.py` and code comments: **English only**.

`README.rst` and `index.html` can be in the client's preferred language (but the agent defaults to English).

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
