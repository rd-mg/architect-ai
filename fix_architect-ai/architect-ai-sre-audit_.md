# Auditoría de Resiliencia Arquitectónica y SRE: `architect-ai`

> **Metodología**: Backtracking Debugging forense — cada hallazgo incluye traza de fallo reproducible, impacto en memoria LLM/tokens y mitigación de hardening.
> **Cobertura de código**: 7 batches (~7.6 MB, 218K líneas) — GEMINI.md, CLAUDE.md, assets compartidos, Go TUI/CLI, MCP, Engram, Odoo overlay, SDD protocols.
> **Fecha**: 2026-06-04 | **Versión analizada**: v0.1.x (build desde batch index)

---

## ÍNDICE DE HALLAZGOS

| ID | Severidad | Componente | Título |
|---|---|---|---|
| F-01 | 🔴 CRÍTICO | L0 / Gemini GEMINI.md | Violación de inmutabilidad L0 — Modo A permite ejecución inline |
| F-02 | 🔴 CRÍTICO | L1a↔L1b / Strict Isolation | Contaminación de contexto garantizada en Claude Code y VSCode |
| F-03 | 🔴 CRÍTICO | CLAUDE.md / Generación | Hash placeholders sin resolver rompen el runtime de producción |
| F-04 | 🔴 CRÍTICO | sdd-state.yaml / Locking | Race condition TOCTOU en escritura atómica de estado DAG |
| F-05 | 🟠 ALTO | Engram / Compactación | Amnesia de sesión por ausencia de auto-trigger de `mem_session_summary` |
| F-06 | 🟠 ALTO | Adaptive Reasoning Gate | `GATE_ERROR` sobre placeholders no inyectados produce bloqueo silencioso |
| F-07 | 🟠 ALTO | sdd-verify / Circuit Breaker | Bucle infinito de validación por desincronización Strict TDD ↔ Circuit Breaker |
| F-08 | 🟠 ALTO | MCP / Odoo | Saturación de contexto por ausencia de paginación explícita en `mcp_odoo_search_records` |
| F-09 | 🟡 MEDIO | TUI / BubbleTea | Progreso de pipeline mudo — `ProgressFunc` no usa `p.Send()` |
| F-10 | 🟡 MEDIO | TUI / BubbleTea | Goroutine huérfana en AgentBuilder ante cancelación por Esc |
| F-11 | 🟡 MEDIO | Caveman Firewall | Ausencia de validación de registro en `sdd-archive` con modo `none` |
| F-12 | 🟡 MEDIO | Engram / ByteRover | Context bloat garantizado en Level 3 (Semantic Memory) con Odoo overlay |
| F-13 | 🟡 MEDIO | Context-Guardian | Umbral de 100K chars no tiene guardia de salida — loop de compactación |
| F-14 | 🟡 MEDIO | MCP / Secrets | Credenciales Odoo referenciadas por env var resolvibles en texto plano |
| F-15 | 🟢 BAJO | TUI / Router | `ScreenUninstallConfirm` no tiene ruta forward en `linearRoutes` |
| F-16 | 🟢 BAJO | Engram Convention | Colisión de topic_key por normalización insuficiente en `ResearchTopicKey` |

---

## F-01 — 🔴 CRÍTICO: Violación de Inmutabilidad L0 — Modo A Inline en Gemini

### [COMPONENTE / VECTOR]
`internal/assets/gemini/GEMINI.md` · `.agent/GEMINI.md` · L0 Super-Orchestrator (Gemini CLI)

### [TRACE DE FALLO — Backtracking]

**Condición de activación**: El agente usa Gemini CLI. El archivo desplegado es `~/.gemini/GEMINI.md`.

```
Turno 1:  Usuario → "git status"
           L0 ejecuta ROUTING DECISION PROTOCOL
           Step 1: ningún Mandatory Trigger activo (1 archivo, no hay código)
           Step 2: NO es SDD_INTENT (no hay "/sdd-*" ni regex match)
           → Debería ir a Mode B (General Orchestrator)
           PERO: el archivo gemini/GEMINI.md define:

           ## Mode A (Gemini inline — simple tasks)
           Use bash/read/write tools directly. Do NOT use run_subagent for simple operations.

           L0 lee "git status" → lo clasifica como "simple task"
           → Ejecuta bash directamente: git status
           → VIOLACIÓN de la regla IMMUTABILITY AND DELEGATION
```

**El conflicto es estructural**: el `.agent/GEMINI.md` (fuente de verdad) dice `L0 MUST NOT perform any direct file, shell, or network operations`. Pero `internal/assets/gemini/GEMINI.md` (el template desplegado a `~/.gemini/GEMINI.md`) define explícitamente **Mode A** como ejecución inline con `bash/read/write tools directly`.

**Flujo de contaminación por número de herramientas**:
```
Turno 1:  git status     → inline (1 tool call)
Turno 2:  cat README     → inline (2 tool calls)
Turno 3:  "explora el proyecto" → inline (3+ reads sin delegación)
           → Llegamos a 5 exploratory reads SIN haber delegado
           → L0 debe PAUSAR según la regla de long-session
           → Pero Mode A no tiene contador de tool calls
           → L0 nunca pausa, acumula estado directamente
```

### [IMPACTO DE CONTEXTO / SRE]
- L0 acumula estado operacional directo en su ventana de contexto, consumiendo tokens que deberían pertenecer a L1.
- Cuando finalmente se delega a L1, el contexto de L0 ya tiene 30-50K chars de resultados de herramientas, induciendo D4 ≥ 2 artificialmente.
- La Strict Isolation Rule se rompe porque L0 ya tiene conocimiento operacional que debería estar aislado en L1b.

### [MITIGACIÓN HARDENING]

1. **Eliminar Mode A del template Gemini**. Reemplazar con:
   ```markdown
   ## GEMINI CLI — DELEGATION ONLY
   Gemini CLI soporta run_subagent. L0 NUNCA usa bash/read/write directamente.
   Simple tasks (git status, grep) → Mode C (General Orchestrator).
   No existe Mode A. Toda ejecución va a L1.
   ```

2. **Agregar chequeo de integración** en `cmd/eintegrate/main.go`:
   ```go
   if checkFile("internal/assets/gemini/GEMINI.md", "Mode A (Gemini inline") {
       errs = append(errs, "E-11: Gemini template exposes inline Mode A — L0 immutability violation")
   }
   ```

3. **Ajustar el contador de tool calls**: inyectar el token budget del Adaptive Reasoning Gate en el contexto de L0 para que el self-audit capture las llamadas directas.

---

## F-02 — 🔴 CRÍTICO: Contaminación de Contexto L1a↔L1b en Claude Code y VSCode

### [COMPONENTE / VECTOR]
`CLAUDE.md` (generado) · `internal/assets/cursor/*` · `internal/agents/ADAPTER-CONTRACT.md` · Strict Isolation Rule

### [TRACE DE FALLO — Backtracking]

En Claude Code y VSCode Copilot, L0, L1a y L1b se inyectan en el **mismo archivo de sistema**:

```
CLAUDE.md estructura real (línea 6735-6762):
  <!-- architect-ai:L0:start hash:{L0_HASH} -->
  {content from .atl/agents/architect.md}           ← L0 aquí
  <!-- architect-ai:L1a:start hash:{L1A_HASH} -->
  {content from .atl/agents/sdd-orchestrator.md}    ← L1a aquí
  <!-- architect-ai:L1b:start hash:{L1B_HASH} -->
  {content from .atl/agents/general-orchestrator.md} ← L1b aquí
  <!-- architect-ai:foundation:start -->
  {foundation.md}
```

**El modelo ve TODO en una sola ventana de contexto.** La "delegación" de L0 a L1 es lógica (cambio de rol instruccional), no física (contexto separado).

**Traza de contaminación**:
```
Sesión activa con SDD workflow:
  L0 → "ROUTING: SDD intent → L1a"
  L1a (sdd-orchestrator) ejecuta sdd-explore, sdd-propose
  → L1a emite: "Change: add-payment-gateway. Stripe API key pattern: sk-..."
  
  Usuario: "now debug the auth module" (non-SDD)
  L0 → "ROUTING: Non-SDD → L1b"
  L1b (general-orchestrator) toma control
  → L1b tiene ACCESO COMPLETO al artefacto de sdd-propose
    porque está en el mismo contexto
  → L1b puede referenciar "add-payment-gateway" aunque
    la STRICT ISOLATION RULE lo prohíbe explícitamente:
    "L1a and L1b MUST NOT know about each other"
```

**Violación en Cursor**: el archivo `internal/assets/cursor/sdd-phase-protocols/` inyecta los protocolos de fase como reglas globales en `.cursorrules`. Cualquier agente Cursor ve todos los protocolos SDD aunque no esté en una sesión SDD.

**La isolation es DECLARATIVA, no MECÁNICA** en plataformas no-run_subagent. La regla existe pero no puede ser aplicada por ningún mecanismo de enforcement en el runtime.

### [IMPACTO DE CONTEXTO / SRE]
- Artifacts de cambios SDD activos (nombres de variables, esquemas, rutas de archivos) son visibles para el general-orchestrator.
- En sesiones largas (+20 turnos mixtos), el contexto compartido entre L0/L1a/L1b acumula ~150-200K tokens de artefactos mutuamente contaminados.
- Una instrucción del SDD orchestrator puede ser "obedecida" por el general-orchestrator si el modelo no distingue el origen del rol.

### [MITIGACIÓN HARDENING]

1. **Documentar explícitamente la limitación de plataforma** en `CLAUDE.md` y `docs/architecture.md`:
   ```markdown
   ## Nota de Aislamiento en Claude Code / VSCode
   En estas plataformas, L0/L1a/L1b comparten una sola ventana de contexto.
   La STRICT ISOLATION RULE es una convención de comportamiento, no una garantía mecánica.
   Los artefactos SDD activos son visibles para el general-orchestrator.
   Mitigación: usar prefijos de contexto explícitos (`[L1a-ONLY]`) en artefactos sensibles.
   ```

2. **Agregar Context Fence Markers** en los artefactos SDD:
   ```markdown
   <!-- architect-ai:sdd-artifact:PRIVATE — do not reference from L1b -->
   ```
   E instruir al general-orchestrator a ignorar bloques con esa marca.

3. **Migrar a Claude Code con Task tool** para sub-agentes: la única solución real de aislamiento en Claude Code es usar el `Task` tool que crea contextos limpios. Documentar esto como requisito arquitectónico.

---

## F-03 — 🔴 CRÍTICO: Hash Placeholders No Resueltos en CLAUDE.md de Producción

### [COMPONENTE / VECTOR]
`CLAUDE.md` · `cmd/architect-ai/main.go` · Sistema de build `arch-hardening`

### [TRACE DE FALLO — Backtracking]

El `CLAUDE.md` en el repositorio (línea 6735) contiene literalmente:
```
<!-- AUTO-GENERATED by architect-ai sync v2 — hash:{CONTENT_HASH} -->
<!-- architect-ai:L0:start hash:{L0_HASH} -->
{content from .atl/agents/architect.md}
<!-- architect-ai:L1a:start hash:{L1A_HASH} -->
{content from .atl/agents/sdd-orchestrator.md}
```

**Estos son placeholders sin resolver**. El contenido real (`{content from ...}`) es una referencia que debe ser materializada por `architect-ai build`. El archivo `CLAUDE.md` en el repositorio **no contiene el prompt del sistema real**.

**Traza de fallo en deployment**:
```
git clone architect-ai → cd myproject → architect-ai init
→ Si `architect-ai build` falla o se omite:
  CLAUDE.md queda con placeholders literales como "{content from .atl/agents/architect.md}"
→ Claude Code lee CLAUDE.md
→ El modelo interpreta "{content from ...}" como texto, no como inclusión
→ L0 tiene instrucciones vacías: "You are ARCHITECT" + nada
→ El modelo alucia el comportamiento de L0 sin guardrails
→ TODAS las reglas de delegación, inmutabilidad y circuit breaker están ausentes
```

**Fallo del hash**: los comentarios `hash:{L0_HASH}` no son verificados en runtime. Si el archivo materializado se corrompe o queda desactualizado, no hay mecanismo de detección. El archivo `cmd/eintegrate/main.go` verifica solo 10 condiciones textuales pero no verifica la ausencia de placeholders.

### [IMPACTO DE CONTEXTO / SRE]
- Un CLAUDE.md con placeholders da instrucciones vacías al modelo, eliminando todas las salvaguardas de L0.
- La ausencia de "NEVER write code inline" lleva a que Claude Code ejecute inline en producción.
- El modelo sin contexto de GEMINI.md o CLAUDE.md real es funcional pero **sin protocolo de hardening**.

### [MITIGACIÓN HARDENING]

1. **Agregar check de placeholder en `architect-ai check all`**:
   ```go
   // internal/verify/checks.go — nuevo check
   {
       ID: "placeholder-resolved",
       Description: "CLAUDE.md contains no unresolved template placeholders",
       Run: func(ctx context.Context) error {
           data, err := os.ReadFile("CLAUDE.md")
           if err != nil { return err }
           if strings.Contains(string(data), "{content from") ||
              strings.Contains(string(data), "{L0_HASH}") ||
              strings.Contains(string(data), "{CONTENT_HASH}") {
               return fmt.Errorf("CLAUDE.md contains unresolved placeholders — run: architect-ai build")
           }
           return nil
       },
       FixHint: "architect-ai build",
   }
   ```

2. **Agregar E-12 al eintegrate** con verificación de placeholders.

3. **Incluir validación post-sync** en el workflow `arch-hardening` (Phase A Step A3) que lea el CLAUDE.md materializado y rechace si encuentra `{content from` o `{.*_HASH}`.

---

## F-04 — 🔴 CRÍTICO: Race Condition TOCTOU en Escritura Atómica de sdd-state.yaml

### [COMPONENTE / VECTOR]
`internal/components/sdd/` · `WriteSddState()` · Protocolo de locking de `.atl/sdd-state.yaml`

### [TRACE DE FALLO — Backtracking]

El código en `WriteSddState()` (batch 6, línea ~30346):

```go
// ① Stat para verificar si el lock existe
if info, err := os.Stat(lockFile); err == nil {
    if time.Since(info.ModTime()) > 30*time.Second {
        os.Remove(lockFile)
    } else {
        return fmt.Errorf("state file is locked — another process is writing")
    }
}
// ② Escribir el lockfile (NO es atómico)
if err := os.WriteFile(lockFile, []byte(pid), 0644); err != nil { ... }
defer os.Remove(lockFile)
```

**El patrón Stat → WriteFile NO es atómico**. Hay una ventana de race entre ① y ②:

```
Proceso A (sdd-apply turno N):              Proceso B (context-guardian compactación):
  ① Stat(lockFile) → not found             ① Stat(lockFile) → not found
                                              ② WriteFile(lockFile, pidB)
  ② WriteFile(lockFile, pidA)  ← RACE
  ③ Write .tmp + rename
  ④ defer Remove(lockFile)                  ③ Write .tmp + rename ← SOBRESCRIBE el de A
```

**La corrección requiere `O_EXCL`** (crear-o-fallar atómico del kernel). El código actual usa `os.WriteFile` que es equivalente a `O_WRONLY|O_CREATE|O_TRUNC` — sin exclusión.

**Escenario de fallo en paralelo**: el sdd-orchestrator lanza `sdd-explore` y `sdd-verify` en paralelo (Parallel Dispatch Table, `sdd-verify: YES parallelizable`). Ambas fases actualizan `sdd-state.yaml`. Con la ventana de race activa, el estado final puede tener `sdd-explore: completed` sobrescrito por `sdd-verify: running` del mismo campo, corrompiendo el DAG.

**Consecuencia del estado corrupto**:
```
sdd-spec ve sdd-explore.status = "running" (en lugar de "completed")
→ CheckPrerequisites("sdd-spec") → BLOCKED
→ Circuit Breaker incrementa attempt_count
→ attempt_count >= 3 → phase abandoned
→ sdd-spec ABANDONADO por corrupción de estado, no por fallo real
```

### [IMPACTO DE CONTEXTO / SRE]
- Corrupción silenciosa del DAG en escenarios de paralelismo (estadísticamente probable con sdd-verify).
- El circuit breaker puede eliminar fases correctas que aparecen como prerequisito no completado.
- Recuperación requiere intervención manual: `architect-ai restore sdd-state.yaml`.

### [MITIGACIÓN HARDENING]

Reemplazar el lockfile basado en Stat con creación exclusiva atómica:

```go
func acquireLock(lockFile string) error {
    f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
    if err != nil {
        if os.IsExist(err) {
            // Verificar stale lock
            info, statErr := os.Stat(lockFile)
            if statErr == nil && time.Since(info.ModTime()) > 30*time.Second {
                os.Remove(lockFile)
                return acquireLock(lockFile) // retry una vez
            }
            return fmt.Errorf("state file locked by another process")
        }
        return err
    }
    defer f.Close()
    fmt.Fprintf(f, "%d", os.Getpid())
    return nil
}
```

`O_CREATE|O_EXCL` garantiza que **solo un proceso** puede crear el archivo — es atómico a nivel de kernel en Linux y macOS. Agregar test de concurrencia en `internal/components/sdd/` con `go test -race`.

---

## F-05 — 🟠 ALTO: Amnesia de Sesión por Ausencia de Auto-Trigger de `mem_session_summary`

### [COMPONENTE / VECTOR]
`.agent/GEMINI.md` · `internal/assets/_shared/engram-convention.md` · L0 SESSION MEMORY

### [TRACE DE FALLO — Backtracking]

El protocolo de cierre de sesión en L0 requiere invocación **manual**:

```markdown
On session end ("wrap up", "done", "close session"):
1. mem_save("session-config/{project}", ...)
2. mem_session_summary(goal, accomplished, next_steps, key_files) → persist
```

**El trigger depende de que el usuario diga palabras clave**. La lista está en el GEMINI.md (Claude): `"done" / "that's it" / "listo"`.

**Traza de fallo — Compactación no anunciada**:
```
Sesión turno 45:  sdd-design completado, artefactos en Engram
Turno 46:  Usuario escribe prompt largo
           → Claude/Gemini detecta context > threshold
           → Se ejecuta /compact (Claude) o /compress (Gemini) INTERNAMENTE
           → El modelo pierde estado de conversación
           → L0 nunca dijo "done" → mem_session_summary NUNCA fue llamado
           → El estado sdd/{change}/state en Engram tiene:
              sdd-design: completed
              sdd-tasks: pending   ← última fase antes de compaction
           → PERO el apply-progress que estaba en contexto se pierde
           → La siguiente sesión: mem_search("sdd/{change}/state") encuentra el state
           → sdd-orchestrator asume "continuar desde sdd-tasks"
           → sdd-tasks se ejecuta de nuevo (redundancia) o se asume sin base
```

**Riesgo adicional**: el Engram solo guarda la última versión (upsert sin historial). Si mem_session_summary no se llamó, `next_steps` y `key_files` de la sesión anterior están perdidos permanentemente.

La `caveman-firewall.md` especifica que el `context-guardian` debe preservar `caveman_firewall_active: true` pero no incluye el estado del pipeline SDD como `protected_fact`.

### [IMPACTO DE CONTEXTO / SRE]
- Pérdida de continuidad en sesiones largas con compactación automática.
- Riesgo de re-ejecución de fases ya completadas (desperdicio de tokens).
- En modo `engram` puro, los artefactos persisten pero el contexto orquestal (qué decisión se tomó y por qué) se pierde.

### [MITIGACIÓN HARDENING]

1. **Auto-trigger de `mem_session_summary` en el hook de compactación**. En el GEMINI.md (Gemini) y CLAUDE.md, agregar:
   ```markdown
   ## COMPACTION HOOK (MANDATORY — ejecutar ANTES de /compact o /compress)
   Antes de responder con /compact o cuando D4 alcanza nivel 3:
   1. mem_save("sdd/{change}/state", {estado actual completo})
   2. mem_session_summary(...)
   3. Entonces ejecutar compactación
   ```

2. **Elevar estado SDD a `protected_fact` en context-guardian**:
   ```yaml
   protected_facts:
     - key: caveman_firewall_active
       value: true
     - key: active_sdd_change      # NUEVO
       value: "{change-name}"
     - key: last_completed_phase   # NUEVO
       value: "{phase}"
   ```

3. **Verificación post-compactación**: agregar un `mem_search("sdd/{change}/state")` como primera acción al detectar que el contexto fue compactado (D4 subió desde 0).

---

## F-06 — 🟠 ALTO: GATE_ERROR Silencioso por Placeholders No Inyectados

### [COMPONENTE / VECTOR]
`internal/assets/_shared/adaptive-reasoning-gate-v2.md` · L2 Sub-agents · Orquestador L1

### [TRACE DE FALLO — Backtracking]

El Adaptive Reasoning Gate v2 requiere que el orquestador inyecte:
```
[MODE {INJECTED_MODE} | D1={D1}, D2={D2}, D3={D3}, D4={D4}] {rationale}
```

Cuando el orquestador falla en la inyección (por error de template, truncamiento de prompt o pérdida de contexto post-compactación):

```
L2 sub-agent recibe prompt con literales: "[MODE {INJECTED_MODE} | D1={D1}, ...]"
→ Gate v2 rule: "IF any fields contain unfilled placeholders → GATE_ERROR"
→ Sub-agent emite: [GATE_ERROR: orchestrator did not inject mode]
→ Sub-agent hace status: blocked y STOPS

→ Orquestador recibe status: blocked
→ Incrementa attempt_count en sdd-state.yaml
→ attempt_count = 1 → "Update approach, retry"
→ Reintenta con MISMO prompt (el bug está en el template, no en el contenido)
→ attempt_count = 2 → "Request user context"  
→ attempt_count = 3 → CIRCUIT BREAKER TRIPS
→ Phase ABANDONED por error de infraestructura, no por error de dominio
```

**El gate error es indistinguible de un fallo de fase real para el circuit breaker.** Ambos incrementan `attempt_count`. El circuit breaker no tiene categoría "infrastructure_error" separada de "domain_error".

**Condición de activación**: ocurre cuando el L1 orquestador pierde el D1-D4 computation por compactación o cuando el Progressive Phase Loading falla al cargar el protocolo de fase desde disk (si `context-mode` no está disponible).

### [IMPACTO DE CONTEXTO / SRE]
- Fases abandonadas por errores de infraestructura MCP/context, no por lógica de dominio.
- El usuario recibe un mensaje de "Phase abandoned" sin indicación de causa raíz.
- El circuit breaker queda contaminado con intentos de infrastructure failures.

### [MITIGACIÓN HARDENING]

1. **Separar contadores de errores en `sdd-state.yaml`**:
   ```yaml
   circuit_breaker:
     attempt_counts:
       sdd-spec: 1
     infrastructure_error_counts:  # NUEVO
       sdd-spec: 2  # no bloquea el CB, solo alerta
   ```

2. **Gate v2: fallback a auto-score en lugar de STOP**:
   ```markdown
   IF placeholders unfilled AND attempt_count < 2:
     → Auto-compute D1-D4 from task description
     → Log: [GATE_FALLBACK: self-scored D1=X,D2=X,D3=X,D4=X]
     → Proceed with MODE 2 (safe default under uncertainty)
   IF placeholders unfilled AND attempt_count >= 2:
     → GATE_ERROR → status: infrastructure_blocked
   ```

3. **Añadir validación de template en L1** antes de hacer `Task()`:
   ```python
   if "{INJECTED_MODE}" in prompt or "{D1}" in prompt:
       raise InfrastructureError("mode template not materialized")
   ```

---

## F-07 — 🟠 ALTO: Bucle Infinito en sdd-verify con Strict TDD

### [COMPONENTE / VECTOR]
`.agent/skills/sdd-verify/strict-tdd-verify.md` · `circuit-breaker.md` · `sdd-orchestrator.md` — Dependency Graph

### [TRACE DE FALLO — Backtracking]

El Dependency Graph del sdd-orchestrator define:

```
proposal → specs → tasks → apply → verify → archive
                                       |
                            FAIL (Judgement Day)
            design ←────────────────────┘
```

Cuando `sdd-verify` en modo Strict TDD detecta un fallo CRÍTICO:

```
sdd-verify Step 5f: "expect(true).toBe(true)" — Assertion Quality: CRITICAL
→ sdd-verify status: failed
→ sdd-orchestrator: detecta FAIL → rutea a "Judgement Day" o redesign

Caso 1 — sdd-orchestrator redirige a sdd-apply para fix:
  sdd-apply arregla el test → sdd-verify se ejecuta de nuevo
  IF sdd-apply no puede resolver (test es estructuralmente trivial):
    sdd-verify CRÍTICO de nuevo → FAIL de nuevo → sdd-apply de nuevo
    → attempt_count++ (circuit breaker cuenta CADA run de sdd-verify)
    → attempt_count = 3 → sdd-verify ABANDONED
    → El cambio queda en estado apply:completed / verify:abandoned
    → sdd-archive no puede ejecutar (prerequisito verify no está "completed")
    → DEADLOCK de cierre: el cambio no puede archivarse ni revertirse limpiamente
```

**El Dependency Graph del sdd-orchestrator no incluye una ruta de salida** para `verify:abandoned`. El `sdd-archive` requiere `verify:completed` pero no hay rama para `verify:abandoned → force-archive`.

**Caso 2 — Strict TDD activo, runner de tests no disponible**:
```
sdd-verify intenta ejecutar test runner
→ runner no disponible (D3 >= 2 → Mode 3-ERR → no web, no tools)
→ sdd-verify no puede confirmar "GREEN confirmed" del TDD cycle
→ Step 5b: "CRITICAL — test file missing" (porque no pudo leer)
→ Mismo deadlock que Caso 1
```

### [IMPACTO DE CONTEXTO / SRE]
- Cambios legítimos pueden quedar en limbo permanente por un único test trivial.
- El estado `verify:abandoned` no tiene ruta forward documentada en el DAG.
- Recuperación solo disponible vía `/sdd-ff` (force-forward) que salta verificación — anulando el propósito de Strict TDD.

### [MITIGACIÓN HARDENING]

1. **Agregar rama `verify:abandoned` en el Dependency Graph**:
   ```
   verify:abandoned → sdd-orchestrator emite:
     [VERIFY-DEADLOCK] Phase abandoned after {N} attempts.
     Options:
       [1] /sdd-hotfix — Apply targeted test fix (scope: test files only)
       [2] /sdd-archive --status=abandoned — Close change as abandoned
       [3] /sdd-ff verify --force — Skip verification (documents exception)
   ```

2. **Separar conteo de circuit breaker para Strict TDD failures vs domain failures**:
   El circuit breaker debe tener `max_attempts: 5` para `sdd-verify` en modo Strict TDD (vs 3 por defecto) dado el ciclo TDD iterativo.

3. **Fallback de test runner en Mode 3-ERR**: si el test runner es inaccesible por restricciones de contexto, el sdd-verify debe reportar `verify:degraded` (warning, no CRITICAL) en lugar de fallar.

---

## F-08 — 🟠 ALTO: Saturación de Contexto por Payloads Masivos de Odoo MCP

### [COMPONENTE / VECTOR]
`internal/assets/_shared/odoo-overlay-routing.md` · `internal/assets/overlays/odoo-development-skill/` · MCP `mcp-server-odoo`

### [TRACE DE FALLO — Backtracking]

El Odoo Research Order define:
```
3. rg en Odoo Community local (~/gitproj/odoo/community/{version}/addons/)
```

Y el `odoo-overlay-routing.md` indica llamadas a:
- `mcp_odoo_search_records`
- `odoo-database-query` sub-agent

**Sin límite explícito en los resultados de búsqueda**:

```
sdd-explore delegado a odoo-context-gatherer:
  → odoo-context-gatherer llama mcp_odoo_search_records(model="sale.order", domain=[], limit=None)
  → Entorno de producción: 50.000 registros
  → mcp-server-odoo devuelve JSON con 50.000 registros
  → El payload entra al contexto del sub-agent
  → D4 sube de 0 → 3 en un solo tool call
  → Context-guardian se activa (pero la data ya está en el contexto)
  → Sub-agent entra en Mode 3-CTX (+++Pragmatic, 50% output compression)
  → El análisis es superficial por saturación del contexto
  → apply-progress se guarda con análisis incompleto
```

**El Odoo Guide menciona** `limit` como opción pero no lo establece como obligatorio:
```
# Línea ~7272: Con limit y order
records = self.search([], order='date DESC', limit=10)
```

Las guías son documentación de referencia, no restricciones forzadas por el MCP server.

**YOLO mode adicional**: cuando `ODOO_YOLO=true` (configurable en `generator.go`), el MCP server permite operaciones de escritura sin confirmación. Un sub-agent en Mode 3-CTX (context saturado, reasoning degradado) + YOLO puede ejecutar mutaciones sin la verificación adecuada.

### [IMPACTO DE CONTEXTO / SRE]
- Un único `mcp_odoo_search_records` sin limit puede consumir el 80% del context window.
- Análisis degradado por saturación lleva a specs y designs incompletos.
- En YOLO mode, mutaciones ejecutadas con contexto degradado pueden corromper datos de producción Odoo.

### [MITIGACIÓN HARDENING]

1. **Añadir `max_records` global en odoo-overlay-routing.md**:
   ```markdown
   ## Regla de Paginación Obligatoria (MANDATORY)
   TODA llamada a mcp_odoo_search_records DEBE incluir limit <= 50.
   Para análisis de volumen, usar mcp_odoo_aggregate_records (solo metadatos).
   PROHIBIDO: domain=[] sin limit.
   ```

2. **Proteger YOLO mode** con D-score check:
   ```markdown
   ## YOLO Mode Guard
   YOLO=true permite mutaciones. PROHIBIDO ejecutar mutaciones cuando D3 >= 2 OR D4 >= 2.
   Ante D3/D4 elevado: solicitar confirmación explícita al usuario antes de cualquier write.
   ```

3. **Context-guardian Drop Priority para Odoo payloads**:
   ```
   Priority 0 (drop PRIMERO): raw odoo record payloads (json > 10KB)
   → Reemplazar con resumen: "Fetched N records from sale.order [truncated]"
   ```

---

## F-09 — 🟡 MEDIO: Progreso de Pipeline Mudo en TUI BubbleTea

### [COMPONENTE / VECTOR]
`internal/tui/model.go:startInstalling()` · `pipeline.ProgressFunc` · `ScreenInstalling`

### [TRACE DE FALLO — Backtracking]

El código en `startInstalling()` (línea ~12940):

```go
return m, tea.Batch(tickCmd(), func() tea.Msg {
    onProgress := func(event pipeline.ProgressEvent) {
        // NOTE: ProgressFunc is called synchronously from the pipeline goroutine.
        // We cannot use p.Send() here because we don't have a reference to the
        // tea.Program. Instead, these events are collected in the ExecutionResult
        // and the PipelineDoneMsg handles the final state.
    }
    result := executeFn(selection, resolved, detection, onProgress)
    return PipelineDoneMsg{Result: result}
})
```

**`onProgress` es un no-op**. La función es invocada pero no envía nada al canal de mensajes del TUI. Los `StepProgressMsg` que el model.go define (`case StepProgressMsg`) **nunca son producidos por ningún emisor real**.

**Consecuencia observable**:
```
ScreenInstalling renderiza:
  [ ] Install dependencies    ← spinner girando
  [ ] Configure agents        ← spinner girando
  [ ] Inject components       ← spinner girando

El usuario espera N segundos/minutos sin feedback de progreso.
PipelineDoneMsg llega → todos los pasos se marcan de golpe.
Si el pipeline falla en el paso 1, el usuario no lo sabe hasta el final.
```

El `ProgressFromExecution()` reconstruye el progreso desde el `ExecutionResult` final — útil para el resultado final pero elimina la experiencia de progreso en tiempo real.

### [IMPACTO DE CONTEXTO / SRE]
- UX degradada: instalaciones largas (30-120s) sin feedback intermedio.
- En fallos de pipeline, el usuario no puede estimar cuándo intervenir.

### [MITIGACIÓN HARDENING]

Capturar referencia al `tea.Program` en el closure:

```go
func (m Model) startInstalling(p *tea.Program) (tea.Model, tea.Cmd) {
    // ...
    return m, tea.Batch(tickCmd(), func() tea.Msg {
        onProgress := func(event pipeline.ProgressEvent) {
            p.Send(StepProgressMsg{
                StepID: event.StepID,
                Status: pipeline.StepStatus(event.Status),
                Err:    event.Err,
            })
        }
        result := executeFn(selection, resolved, detection, onProgress)
        return PipelineDoneMsg{Result: result}
    })
}
```

Requiere pasar `*tea.Program` al `Model` en `main.go` (patrón estándar en BubbleTea avanzado).

---

## F-10 — 🟡 MEDIO: Goroutine Huérfana en AgentBuilder ante Cancelación por Esc

### [COMPONENTE / VECTOR]
`internal/tui/model.go:startGeneration()` · `AgentBuilderState.GenerationCancel` · `context.WithTimeout`

### [TRACE DE FALLO — Backtracking]

```go
// startGeneration (línea ~14329)
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
m.AgentBuilder.GenerationCancel = cancel

return m, tea.Batch(tickCmd(), func() tea.Msg {
    defer cancel()
    engine := agentbuilder.NewEngine(engineID)
    raw, err := engine.Generate(ctx, prompt)  // llamada bloqueante a API externa
    // ...
    return AgentBuilderGeneratedMsg{Agent: agent}
})
```

**Al presionar Esc durante la generación** (línea ~13217):
```go
if m.Screen == ScreenAgentBuilderGenerating {
    if m.AgentBuilder.GenerationCancel != nil {
        m.AgentBuilder.GenerationCancel()  // cancela el context
    }
    m.AgentBuilder.Generating = false
    m.setScreen(ScreenAgentBuilderPrompt)
}
```

El context es cancelado correctamente. Pero si el `engine.Generate()` no respeta la cancelación del context (por ejemplo, si usa `http.DefaultClient` sin el context):

```
Esc → cancel() llamado
→ context cancelado
→ goroutine continúa ejecutando (API call ignora context)
→ goroutine eventualmente retorna AgentBuilderGeneratedMsg
→ Model.Update() recibe el mensaje:
    case AgentBuilderGeneratedMsg:
        if !m.AgentBuilder.Generating {
            return m, nil  // ← mensaje ignorado correctamente
        }
```

El mensaje huérfano es ignorado correctamente en el handler. **El problema es el timeout de 5 minutos**: si el engine no respeta el context, la goroutine vive 5 minutos consumiendo recursos, conexiones HTTP y potencialmente incurriendo en costos de API.

**Caso más grave**: si el usuario lanza una nueva generación después de cancelar, hay dos goroutines concurrentes llamando a la misma API. La segunda puede completar antes que la primera (que fue cancelada), ambos mensajes llegan, el primero es ignorado pero la carrera puede causar state inconsistente si el engine tiene estado compartido.

### [IMPACTO DE CONTEXTO / SRE]
- Fuga de goroutines en sesiones interactivas largas con múltiples cancelaciones.
- Potencial doble facturación de API si el engine no respeta context cancellation.
- Race condition entre generaciones concurrentes con el mismo `SelectedEngine`.

### [MITIGACIÓN HARDENING]

1. **Verificar que el engine respeta context** en tests:
   ```go
   func TestGeneration_RespectsContextCancellation(t *testing.T) {
       ctx, cancel := context.WithCancel(context.Background())
       cancel() // cancelar inmediatamente
       _, err := engine.Generate(ctx, "test prompt")
       if !errors.Is(err, context.Canceled) {
           t.Error("engine must respect context cancellation")
       }
   }
   ```

2. **Guardar goroutine ID o usar `sync.WaitGroup`** para garantizar que no hay dos goroutines de generación activas simultáneamente:
   ```go
   // En startGeneration: verificar y esperar la goroutine anterior
   if m.AgentBuilder.Generating {
       m.AgentBuilder.GenerationCancel() // cancelar la anterior
       // esperar a que complete (con timeout corto)
   }
   ```

---

## F-11 — 🟡 MEDIO: Caveman Firewall Indetectable en Modo `none`

### [COMPONENTE / VECTOR]
`internal/assets/_shared/caveman-firewall.md` · Artifact store mode `none` · `sdd-archive`

### [TRACE DE FALLO — Backtracking]

Cuando el artifact store es `none`:
```markdown
persistence-contract.md:
  none → do NOT create or modify project files; return results inline only
```

El `sdd-archive` en modo `none`:
```
sdd-archive recibe: artifact_store = none
→ No persiste nada
→ No verifica apply-progress porque no hay acceso a Engram ni filesystem
→ Emite result contract: status: completed (inline)
→ Regresa al orchestrator con archive "exitoso"
```

**El Caveman Firewall** requiere logging explícito de transiciones de registro:
```
[REGISTER→NORMAL] Entering code zone for task T-03
[REGISTER→ULTRA] Exiting code zone
```

En modo `none`, el `apply-progress` es inline. Si el agente usó ULTRA en código (violación del firewall), no hay artifact persistido que pueda ser auditado retroactivamente. El `sdd-verify` tampoco puede verificar TDD compliance porque no hay apply-progress en Engram ni en filesystem.

**Consecuencia**: en modo `none`, el caveman firewall es no verificable, y el sdd-verify puede pasar (o directamente no ejecutarse) sin evidencia de cumplimiento.

### [MITIGACIÓN HARDENING]

1. **Agregar advertencia explícita al seleccionar modo `none`**:
   ```markdown
   ## Modo `none` — Advertencia de Audit Trail
   WARN: En modo `none`, el caveman firewall y la evidencia TDD NO son verificables.
   El sdd-verify puede no tener apply-progress para auditar.
   Recomendado solo para exploración/prototipado. NO usar en producción.
   ```

2. **sdd-verify con artifact_store=none**: emitir `verify:degraded` en lugar de `verify:completed` cuando no hay apply-progress disponible. Documentar explícitamente en el result contract.

---

## F-12 — 🟡 MEDIO: Context Bloat Garantizado en Level 3 Semantic Memory con Odoo

### [COMPONENTE / VECTOR]
`engram-convention.md` · ByteRover Loading Order · Odoo overlay skills

### [TRACE DE FALLO — Backtracking]

ByteRover define:
```
Level 3 (lazy): knowledge/odoo-*/reference/* → semantic memory (search, don't preload)
```

La instrucción es "lazy" (no precargar). Pero el sdd-orchestrator inyecta en cada sub-agent:

```markdown
## Project Standards (auto-resolved)
{matching compact rules}
```

Y el `odoo-overlay-routing.md` define que se inyectan reglas de:
- `coding-style.md` (compact rules)
- `security.md` (compact rules)  
- `CAUTION_POLICY.md` (**full** — critical)
- `cudio-git.md` (compact rules)

**`CAUTION_POLICY.md` se inyecta completo** en CADA prompt de sub-agent en proyectos Odoo. Si este archivo tiene 5-10KB (típico para una política de cautela), se suma a cada llamada a sdd-apply, sdd-verify, etc.

Con un pipeline completo (8 fases × 2-3 batchs de sdd-apply):
```
~15 llamadas a sub-agents × 8KB de CAUTION_POLICY = ~120KB de repetición
```

Esto presiona D4 hacia valores altos artificialmente, activando Mode 3-CTX incluso en tareas simples.

### [MITIGACIÓN HARDENING]

1. **Mover CAUTION_POLICY a Engram** como knowledge semántico, no inyección de prompt:
   ```
   mem_search("knowledge/odoo/caution-policy") → solo cuando hay riesgo detectado
   ```

2. **Limitar CAUTION_POLICY a compact summary** en la inyección regular. Versión completa solo en `sdd-apply` y `sdd-verify`, no en fases de lectura (sdd-explore, sdd-propose).

---

## F-13 — 🟡 MEDIO: Loop de Compactación por Umbral sin Guardia de Salida

### [COMPONENTE / VECTOR]
`.atl/skill-registry.md` · context-guardian · `ctx_stats`

### [TRACE DE FALLO — Backtracking]

El context-guardian define: `char_count(context_history) >= 100_000 → invoke`.

**Traza de loop**:
```
Turno N:   context_history = 95.000 chars
           L1 ejecuta ctx_execute("shell", "rg -C 3 pattern")
           → resultado: 8.000 chars
           → context_history = 103.000 chars → umbral superado
           Context-guardian invocado:
             Crea context pack: 15.000 chars (summary de 103K)
             mem_save(context_pack) → 800 chars de respuesta Engram
             Total nuevo contexto: ~16.000 chars
           → OK, debajo del umbral

Turno N+1: orquestador carga skill registry (ALWAYS injected — Tier 1)
           → skill registry: ~20.000 chars
           → context_history = 36.000 chars (base) + 20K (skills) = 56K → OK

Turno N+2: sdd-orchestrator carga protocolo de fase (Progressive Phase Loading)
           → protocolo sdd-apply.md: ~15.000 chars
           → context_history = 71K

Turno N+3: sub-agent sdd-apply retorna resultado + apply-progress
           → +40.000 chars de resultado
           → context_history = 111K → UMBRAL SUPERADO OTRA VEZ
           Context-guardian invocado de nuevo
           → crea pack desde 111K...
```

**El context-guardian no tiene "cooldown"**. Si cada turno de trabajo sube el contexto por encima del umbral, el guardian se invoca cada turno, consumiendo tokens adicionales para las operaciones de compactación.

### [MITIGACIÓN HARDENING]

1. **Agregar cooldown mínimo**: no invocar context-guardian si fue invocado en los últimos N=3 turnos.
2. **Umbral dinámico**: después de una compactación, el umbral sube temporalmente al 120% para permitir que el trabajo continúe sin interrupción inmediata.

---

## F-14 — 🟡 MEDIO: Credenciales Odoo Accesibles en Texto Plano vía env var

### [COMPONENTE / VECTOR]
`internal/components/mcp/secrets.go` · `WriteSecretsEnv()` · `.env.mcp`

### [TRACE DE FALLO — Backtracking]

```go
// secrets.go: WriteSecretsEnv escribe las credenciales a .env.mcp
func WriteSecretsEnv(projectDir string, secrets map[string]string) error {
    envPath := filepath.Join(projectDir, ".env.mcp")
    ensureGitignored(filepath.Join(projectDir, ".gitignore"), ".env.mcp")
    // ... escribe credenciales en texto plano
    return os.WriteFile(envPath, []byte(...), 0600)
}
```

Las credenciales se almacenan en `.env.mcp` con permisos `0600` (correcto). Pero la función `ensureGitignored` usa `O_APPEND` sin verificar si el `.gitignore` ya tiene la entrada de forma diferente (ej: `*.env` o `!.env.mcp` en una regla de negación previa).

```bash
# Escenario: proyecto con .gitignore que contiene:
*.env      ← cubre .env pero no .env.mcp si la extensión es .mcp
!.env.production  ← negación que podría interferir
```

`ensureGitignored` solo verifica `strings.Contains(string(data), ".env.mcp")`. Si `.gitignore` tiene `*.env` pero no `.env.mcp` explícito, el check pasa (no agrega nada) pero `.env.mcp` podría no estar ignorado dependiendo de las reglas gitignore.

**Además**: el `generator.go` incluye `ODOO_PASSWORD: "${ODOO_PASSWORD}"` para Antigravity, pero `ODOO_PASSWORD: "${input:odoo-password}"` para VSCode. En Antigravity, la contraseña proviene de una variable de entorno del shell — potencialmente visible en el historial de shell o en `ps aux`.

### [MITIGACIÓN HARDENING]

1. **Verificar reglas de gitignore con la librería de gitignore** en lugar de `strings.Contains`:
   ```go
   import "github.com/go-git/go-git/v5/plumbing/format/gitignore"
   // Verificar que .env.mcp realmente está ignorado según las reglas combinadas
   ```

2. **Agregar `.env.mcp` al check de secrets en el code-reviewer** (sección A1 del reviewer que ya detecta API keys), para que git pre-commit capture cualquier staged `.env.mcp`.

---

## F-15 — 🟢 BAJO: `ScreenUninstallConfirm` sin Ruta Forward en Router

### [COMPONENTE / VECTOR]
`internal/tui/router.go` · `linearRoutes`

### [TRACE DE FALLO]

```go
// router.go — linearRoutes
ScreenUninstallConfirm no aparece en el mapa como key
// Solo aparece como valor en:
// ScreenUninstallProfiles: {Backward: ScreenUninstallComponents}
// (no hay Forward definido para UninstallConfirm)
```

Si la lógica de navegación llama `NextScreen(ScreenUninstallConfirm)`, retorna `ScreenUnknown, false`. El handler de esta condición en `handleKeyPress` debe verificar `ok` para no navegar a pantalla 0.

**Probabilidad de impacto**: baja — el flujo de uninstall usa `m.setScreen()` directamente en `handlePipelineDone`, bypaseando el router. Pero si alguna tecla de "continuar" se procesa en ScreenUninstallConfirm via el router genérico, podría navegar a ScreenUnknown.

### [MITIGACIÓN HARDENING]

Agregar la ruta explícita:
```go
ScreenUninstallConfirm: {Forward: ScreenUninstall, Backward: ScreenUninstallProfiles},
```

O asegurarse de que `ScreenUninstallConfirm` nunca llama `NextScreen` — documentar el invariante.

---

## F-16 — 🟢 BAJO: Colisión de topic_key en `ResearchTopicKey` por Truncamiento

### [COMPONENTE / VECTOR]
`internal/components/engram/engramkeys/keys.go` · `ResearchTopicKey()`

### [TRACE DE FALLO]

```go
func ResearchTopicKey(tool, query string) string {
    // ...
    if len(cleaned) > 50 {
        cleaned = strings.Trim(cleaned[:50], "-")
    }
    return fmt.Sprintf("research/%s/%s-len%d", tool, cleaned, len(query))
}
```

El sufijo `-len{N}` intenta diferenciar queries que producen el mismo slug después de truncar a 50 chars. Pero:

```
Query 1: "odoo sale order workflow configuration for enterprise"
  → slug(50): "odoo-sale-order-workflow-configuration-for-enterp"
  → len: 53
  → key: research/context7/odoo-sale-order-workflow-configuration-for-enterp-len53

Query 2: "odoo sale order workflow configuration for enterprise (v18)"
  → slug(50): "odoo-sale-order-workflow-configuration-for-enterp"  ← mismo slug
  → len: 61  ← diferente len
  → key: research/context7/odoo-sale-order-workflow-configuration-for-enterp-len61
```

Las queries son distintas y tienen keys distintos (diferente len). **Sin colisión en este caso.**

**Colisión real**: dos queries con mismo contenido hasta el char 50 Y misma longitud total producirán la misma key → upsert silencioso (sobrescritura del resultado anterior). En producción con Odoo y muchas queries de investigación similares, esto puede causar que un resultado de investigación anterior sea sobrescrito por uno nuevo.

### [MITIGACIÓN HARDENING]

Agregar hash del contenido completo al key:
```go
import "crypto/sha256"
func ResearchTopicKey(tool, query string) string {
    // ... slug generation ...
    h := sha256.Sum256([]byte(query))
    return fmt.Sprintf("research/%s/%s-%x", tool, cleaned, h[:4]) // 8 hex chars
}
```

Esto garantiza unicidad sin depender solo del truncamiento.

---

## RESUMEN EJECUTIVO SRE

### Riesgos Críticos (acción inmediata requerida)

| ID | Riesgo | Resolución |
|---|---|---|
| F-01 | L0 ejecuta código inline en Gemini — elimina todas las garantías de delegación | Eliminar Mode A del template Gemini |
| F-02 | L1a/L1b comparten contexto en Claude Code/VSCode — Strict Isolation es declarativa, no mecánica | Documentar y mitigar con Context Fence Markers |
| F-03 | CLAUDE.md desplegado con placeholders sin resolver — LLM opera sin guardrails | Agregar check de placeholder en `architect-ai check all` |
| F-04 | TOCTOU en lockfile de sdd-state.yaml — corrupción de DAG en modo paralelo | Reemplazar con `O_CREATE|O_EXCL` |

### Riesgo Sistémico Más Relevante

**F-02 combinado con F-05**: en sesiones largas con compactación, L1a pierde su estado (F-05) y al reanudarse, L1b (que comparte ventana de contexto) puede "heredar" fragmentos del estado SDD que el modelo no filtró correctamente. Esto crea escenarios donde el general-orchestrator actúa sobre artefactos SDD que no debería ver, produciendo acciones fuera de scope con efectos en el filesystem.

### Arquitectura de Defensa Recomendada

```
Prioridad 1 (semana 1):
  F-04: O_EXCL en WriteSddState (30min de desarrollo)
  F-03: placeholder check en verify (1h)
  F-01: eliminar Mode A (15min)

Prioridad 2 (sprint 1):
  F-05: auto-trigger de mem_session_summary en compaction hook
  F-06: gate v2 fallback a auto-score
  F-07: verify:abandoned → ruta documentada en DAG

Prioridad 3 (sprint 2):
  F-08: paginación obligatoria Odoo + YOLO D-score guard
  F-09: p.Send() en ProgressFunc
  F-02: Context Fence Markers + documentación de limitación

Monitoreo continuo:
  F-12, F-13: métricas de contexto por sesión (D4 rate)
  F-16: collision rate en Engram topic_keys (log de upserts inesperados)
```

---

*Reporte generado mediante análisis forense de 7.6MB de código fuente, 218K líneas, 7 batches. Metodología: Backtracking Debugging con Sequential Thinking aplicado a cada vector de análisis.*
