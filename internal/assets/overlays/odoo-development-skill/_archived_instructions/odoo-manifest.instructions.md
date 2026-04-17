---
name: "Odoo Manifest Guidelines"
description: "Use when editing Odoo addon manifests (__manifest__.py). Covers versioning, dependencies, and data file ordering."
applyTo: "**/__manifest__.py"
---

# Manifest Guidelines

- Keep dependencies minimal and accurate.
- Ensure new XML/CSV files are included in `data`/`demo` as appropriate.
- When behavior changes materially, bump the addon `version` (follow existing conventions).
- Keep the manifest clean (no dead entries, correct sequence for security/data/views).
