# Engram Topic-Key Convention

## Taxonomy: Hierarchical 4-Segment Addressing

Todas las llaves de Engram en `architect-ai` deben seguir el formato:
`{domain}/{scope}/{type}/{slug}`

### Segment Definitions

1.  **Domain**: Categoría de alto nivel del ciclo de vida.
    - `sdd`: Spec-Driven Development.
    - `tdd`: Test-Driven Development.
    - `debug`: Resolución de errores y tracebacks.
    - `knowledge`: Conocimiento externo, patrones y guías.
    - `arch`: Contratos de arquitectura y decisiones.

2.  **Scope**: El contexto de aplicación.
    - `{change-name}`: Para artefactos vinculados a un cambio específico.
    - `_global`: Para conocimiento universal del proyecto.
    - `{module-name}`: Para conocimiento específico de un paquete (ej. `internal/tui`).

3.  **Type**: El tipo de dato almacenado.
    - `state`: Estado actual de una máquina de estados o DAG.
    - `brief`: Resúmenes y exploraciones.
    - `decision`: Decisiones de diseño (ADRs).
    - `error`: Registro de fallos y causas raíz.
    - `pattern`: Ejemplos de código y convenciones.
    - `api-contract`: Definiciones de interfaces.
    - `context-pack`: Snapshots de contexto para el Context Guardian.
    - `external`: Hallazgos de investigación (NotebookLM/Context7).

4.  **Slug**: Identificador único en kebab-case.

## Usage Examples

| Contexto | Topic Key |
|---|---|
| Estado de un cambio SDD | `sdd/add-auth/state/main` |
| Diseño de un cambio SDD | `sdd/add-auth/design/v1` |
| Error de importación en Go | `debug/_global/error/import-cycle-tui` |
| Patrón de Bubbletea | `knowledge/tui/pattern/key-simulation` |
| Hallazgo de NotebookLM | `knowledge/odoo-v19/external/sql-constraints` |

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
