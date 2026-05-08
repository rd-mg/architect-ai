# Engram Topic-Key Convention

## Purpose
Establish a hierarchical and deterministic taxonomy for search keys (`topic_key`) in the Engram persistent memory system. This ensures knowledge is locatable, categorized, and reusable across different sessions and agents.

## Requirements

### Requirement: Topic-Key Structure
The system MUST use a four-segment hierarchical structure for all Engram keys.

Structure: `{domain}/{scope}/{type}/{slug}`

#### Scenario: Valid Topic-Key
- GIVEN an agent that wants to save an Odoo 19 code pattern
- WHEN building the `topic_key`
- THEN the result MUST be `knowledge/odoo-v19/pattern/owl-useservice`
- AND it complies with the 4-segment format.

### Requirement: Domain Taxonomy
The `domain` segment values MUST be limited to the following categories:

| Domain | Description |
|---|---|
| `sdd` | Spec-Driven Development flow |
| `tdd` | Test-Driven Development flow |
| `debug` | Error resolution and tracebacks |
| `knowledge` | Patterns, guides, and external knowledge |
| `arch` | Architecture decisions and contracts |

#### Scenario: Incorrect Domain Usage
- GIVEN a key with domain `random`
- WHEN the system validates the key
- THEN it MUST reject it or mark it as non-standard.

### Requirement: Global Scope
When knowledge applies universally (cross-module), the `scope` segment MUST be `_global`.

#### Scenario: Global pattern storage
- GIVEN an import error that affects all of Odoo
- WHEN saved to Engram
- THEN the key MUST be `debug/_global/error/import-error-account-move`.

### Requirement: Deterministic Slugs
Final `slugs` MUST be deterministic, lowercase, and use hyphens as separators.

#### Scenario: Slug generation
- GIVEN a title "How to use OWL hooks?"
- WHEN the slug is generated
- THEN the result MUST be `how-to-use-owl-hooks`.
