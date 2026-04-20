# Engram Topic-Key Convention

## Purpose
Establecer una taxonomía jerárquica y determinista para las llaves de búsqueda (`topic_key`) en el sistema de memoria persistente Engram. Esto garantiza que el conocimiento sea localizable, categorizado y reutilizable a través de diferentes sesiones y agentes.

## Requirements

### Requirement: Topic-Key Structure
El sistema DEBE utilizar una estructura jerárquica de cuatro segmentos para todas las llaves de Engram.

Estructura: `{domain}/{scope}/{type}/{slug}`

#### Scenario: Valid Topic-Key
- GIVEN un agente que desea guardar un patrón de código Odoo 19
- WHEN construye la `topic_key`
- THEN el resultado DEBE ser `knowledge/odoo-v19/pattern/owl-useservice`
- AND cumple con el formato de 4 segmentos.

### Requirement: Domain Taxonomy
Los valores del segmento `domain` DEBEN limitarse a las siguientes categorías:

| Domain | Descripción |
|---|---|
| `sdd` | Flujo de desarrollo Spec-Driven |
| `tdd` | Flujo de desarrollo Test-Driven |
| `debug` | Resolución de errores y tracebacks |
| `knowledge` | Patrones, guías y conocimiento externo |
| `arch` | Decisiones y contratos de arquitectura |

#### Scenario: Incorrect Domain Usage
- GIVEN una llave con domain `random`
- WHEN el sistema valida la llave
- THEN DEBE rechazarla o marcarla como no estándar.

### Requirement: Global Scope
Cuando el conocimiento aplica de forma universal (cross-module), el segmento `scope` DEBE ser `_global`.

#### Scenario: Global pattern storage
- GIVEN un error de importación que afecta a todo Odoo
- WHEN se guarda en Engram
- THEN la llave DEBE ser `debug/_global/error/import-error-account-move`.

### Requirement: Deterministic Slugs
Los `slugs` finales DEBEN ser deterministas, en minúsculas, y usar guiones como separadores.

#### Scenario: Slug generation
- GIVEN un título "How to use OWL hooks?"
- WHEN se genera el slug
- THEN el resultado DEBE ser `how-to-use-owl-hooks`.
