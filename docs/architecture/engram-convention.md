# Engram Topic-Key Convention

## Taxonomy: Hierarchical 4-Segment Addressing

All Engram keys in `architect-ai` must follow the format:
`{domain}/{scope}/{type}/{slug}`

### Segment Definitions

1.  **Domain**: High-level lifecycle category.
    - `sdd`: Spec-Driven Development.
    - `tdd`: Test-Driven Development.
    - `debug`: Error resolution and tracebacks.
    - `knowledge`: External knowledge, patterns, and guides.
    - `arch`: Architecture contracts and decisions.

2.  **Scope**: The application context.
    - `{change-name}`: For artifacts linked to a specific change.
    - `_global`: For project-wide universal knowledge.
    - `{module-name}`: For package-specific knowledge (e.g. `internal/tui`).

3.  **Type**: The type of data stored.
    - `state`: Current state of a state machine or DAG.
    - `brief`: Summaries and explorations.
    - `decision`: Design decisions (ADRs).
    - `error`: Failure records and root causes.
    - `pattern`: Code examples and conventions.
    - `api-contract`: Interface definitions.
    - `context-pack`: Context snapshots for Context Guardian.
    - `external`: Research findings (NotebookLM/Context7).

4.  **Slug**: Unique identifier in kebab-case.

## Usage Examples

| Context | Topic Key |
|---|---|
| SDD change state | `sdd/add-auth/state/main` |
| SDD change design | `sdd/add-auth/design/v1` |
| Go import error | `debug/_global/error/import-cycle-tui` |
| Bubbletea pattern | `knowledge/tui/pattern/key-simulation` |
| NotebookLM finding | `knowledge/odoo-v19/external/sql-constraints` |

### Entity Index Convention

Archive reports MUST include an entity index block as their final section. The block uses
free-text format searchable by `mem_search`. It is not structured YAML — it is plain text
that Engram indexes as observation content.

**Required entity categories** (add more if the change touches additional concepts):
- `modules:` — Go packages modified or created
- `types:` — Go types (structs, interfaces) introduced or changed
- `functions:` — Go functions introduced or changed
- `files:` — file paths created or modified
- `services:` — external services or MCPs involved
- `concepts:` — design patterns, features, or behaviors
- `risks:` — backward compat concerns, migration notes, known issues

**Retrieval benefit**: An agent searching for any entity name will hit the archive report
via `mem_search`, even if the observation title is generic. This reduces the number of
`mem_search` calls needed to reconstruct context.
