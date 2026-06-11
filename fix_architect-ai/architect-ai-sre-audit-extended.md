# Auditoría de Resiliencia Arquitectónica y SRE — Extensión Profunda
## `architect-ai` — Análisis Adversarial Completo v2

> **Metodología**: Backtracking Debugging con traversal de grafo completo — cada hallazgo traza la cadena causal completa desde el código fuente hasta el efecto sistémico, cruzando relaciones entre componentes, contratos de interfaz y protocolos de prompt.
> **Cobertura**: 7.6 MB · 218K líneas · batches 1-7 · Go + Markdown + YAML + Shell
> **Expansión desde v1**: cada finding original extendido con subtópicos, nuevos nodos en el grafo causal, y 4 hallazgos adicionales (F-17 a F-20).

---

## GRAFO DE DEPENDENCIAS ARQUITECTÓNICO

```
L0 (Super-Orchestrator)
├─ GoLib: CheckMandatoryTriggers() ← [F-01-A] self-reported counts
├─ GEMINI.md template             ← [F-01-B] Mode A inline contradiction
├─ CLAUDE.md / ADAPTER-CONTRACT   ← [F-02] isolation lógica no mecánica
│
├─── L1a sdd-orchestrator
│    ├─ phase-dag-enforcement.md  ← [F-04-A] "running" enum inválido
│    ├─ openspec/state.go         ← [F-04-B] tmp+rename sin lock
│    ├─ circuit-breaker.md        ← [F-07] attempt_count no clasifica errores
│    ├─ skill-resolver.md         ← [F-06] GATE_ERROR + CB contaminación
│    │
│    └─── L2 SDD phases
│         ├─ sdd-verify           ← [F-07-A] CRITICAL loop deadlock
│         ├─ sdd-archive          ← [F-07-B] prerequisito verify:abandoned
│         ├─ sdd-apply            ← [F-04-C] writes paralelas sin flock
│         └─ adaptive-gate-v2     ← [F-06] placeholder unresolved
│
├─── L1b general-orchestrator
│    ├─ SDD_INTENT router         ← [F-02-A] contamina contexto compartido
│    └─ session-state-cache       ← [F-05-A] bootstrap paradox
│
├─── Engram / Memory
│    ├─ session_summary           ← [F-05] no auto-trigger pre-compact
│    ├─ bootstrap probe           ← [F-05-B] cache en Engram para detectar Engram
│    ├─ ResearchTopicKey()        ← [F-16] hash collision por len-only
│    └─ context-guardian          ← [F-13] cooldown ausente
│
├─── MCP Layer
│    ├─ inject.go                 ← [F-17] mergeJSONFile silencia malformed JSON
│    ├─ secrets.go                ← [F-14] gitignore check superficial
│    ├─ mcp_odoo_search_records   ← [F-08] sin limit explícito
│    └─ ODOO_YOLO + D-score       ← [F-08-B] mutaciones bajo contexto degradado
│
└─── TUI (BubbleTea)
     ├─ model.go startInstalling  ← [F-09] ProgressFunc no-op
     ├─ model.go startGeneration  ← [F-10] goroutine huérfana
     ├─ WindowSizeMsg handler     ← [F-18] resize no propaga a subcomponents
     └─ router.go linearRoutes    ← [F-15] ScreenUninstallConfirm sin Forward
```

---

## F-01 (EXTENDIDO) — 🔴 CRÍTICO: Violación de Inmutabilidad L0

### Subtópico A: La Contradicción entre GoLib y Template Gemini

**El código Go es correcto. El template Markdown es incorrecto. Ambos definen L0.**

`internal/architect/routing.go` · `ClassifyIntent()`:

```go
// Step 3: Default → delegate to general-orchestrator (L1b)
// Simple tasks (git status, reads, etc.) route to general-orchestrator, not inline
return RouteGeneral, "non-SDD task — routing to general-orchestrator", false
```

El comentario explícito en el código dice **"not inline"**. La función retorna `RouteGeneral`, delegando a L1b. Esta es la semántica correcta: L0 NUNCA ejecuta inline.

Pero `internal/assets/gemini/GEMINI.md` (el archivo que se despliega a `~/.gemini/GEMINI.md`):

```markdown
## Mode A (Gemini inline — simple tasks)
Use bash/read/write tools directly.
Do NOT use run_subagent for simple operations.
```

**La contradicción es explícita**: la GoLib dice "no inline", el template dice "usa bash directamente". El modelo Gemini opera sobre el template, NO sobre la GoLib. La GoLib solo se usa cuando el agente llama explícitamente a `architect-ai guard check`.

### Subtópico B: El Guard Check es Auto-referencial

El `architect-ai guard check` requiere que el agente pase sus propios contadores via flags CLI:

```go
// RunGuard — args from CLI
ref := fs.Int("ref", 0, "files referenced (read)")
write := fs.Int("write", 0, "files to write")
toolCalls := fs.Int("calls", 0, "running tool call count")
exploratory := fs.Int("explore", 0, "exploratory reads in session")
```

**El agente que está siendo auditado provee los números que activan su propia auditoría.** Si L0 quiere hacer 6 lecturas inline sin activar la regla `4-file rule`, solo necesita llamar a `architect-ai guard check --ref 3`. El guard retorna `OK`. L0 procede.

No existe ningún observador externo que cuente las llamadas reales a herramientas. El sistema de guarda depende completamente de que el LLM se auto-reporte honestamente, lo cual viola el principio SRE de **separation of duties**: quien ejecuta no debe ser quien audita.

### Subtópico C: Convergencia de fallos — cuando Mode A + Guard Self-Report se combinan

```
Sesión Gemini CLI — turno 1-10:
  T1: git status           → L0 Mode A (inline bash)   count: 1 ref
  T2: cat README.md        → L0 Mode A (inline read)   count: 2 ref
  T3: ls src/              → L0 Mode A (inline bash)   count: 3 ref
  T4: cat go.mod           → L0 Mode A (inline read)   count: 4 ref
  
  ← AQUÍ debería disparar 4-file rule
  
  L0 auto-invoca: architect-ai guard check --ref 4 --calls 4
  Guard retorna: BLOCKED (4-file rule)
  
  L0 tiene dos opciones:
  [A] Obedecer → delegar a L1b ← comportamiento correcto
  [B] Reportar --ref 3 (sub-conteo) → Guard OK → continuar inline ← bypass

  El prompt de GEMINI.md NO instruye a L0 que el sub-conteo es prohibido.
  El template solo dice: "call architect-ai guard check before delegation"
  No dice "los conteos deben ser exactos" ni "penaliza el sub-conteo".
```

**El sistema de enforcement es bypasseable por cualquier agente que no tenga una alineación perfecta con la regla.**

### [MITIGACIÓN F-01 EXTENDIDA]

1. **Eliminar Mode A del template Gemini** (ya en v1). Agregar explícitamente:
   ```markdown
   ## GEMINI CLI — DELEGATION ONLY
   Gemini CLI SUPPORTS run_subagent. L0 MUST use it for ALL tasks.
   L0 NEVER calls bash/read/write/grep directly. Zero exceptions.
   Mode A does not exist. It was removed in v0.2 due to immutability violations.
   ```

2. **Mover el guard a un hook post-tool externo** (hooks system que ya existe):
   ```json
   // .claude/settings.json hooks
   {
     "hooks": {
       "PostToolUse": [{
         "matcher": "Bash",
         "hooks": [{"type": "command", "command": "architect-ai metering --increment-bash"}]
       }]
     }
   }
   ```
   Un hook externo incrementa contadores en un archivo `.atl/session-metering.json` que el guard lee independientemente del self-report del LLM.

3. **Agregar test de integración** que deploya GEMINI.md en un directorio temporal y ejecuta un escenario de 5 lecturas inline, verificando que el guard retorna BLOCKED con source-of-truth externo.

---

## F-02 (EXTENDIDO) — 🔴 CRÍTICO: Contaminación L1a↔L1b en Plataformas Monolíticas

### Subtópico A: El ADAPTER-CONTRACT documenta pero no garantiza

`internal/agents/ADAPTER-CONTRACT.md` (línea 10530) define la tabla de garantías de aislamiento:

```
| Platform    | Mechanism           | L2 isolation guarantee         |
|-------------|---------------------|--------------------------------|
| OpenCode    | Task tool           | ✅ REAL — clean description    |
| Claude Code | Task tool           | ✅ REAL — clean room           |
| Gemini CLI  | run_subagent        | ✅ REAL                        |
| VSCode      | Simulated (inline)  | ⚠️ LOGICAL — ULTRA framing    |
| Antigravity | Simulated (inline)  | ⚠️ LOGICAL — ULTRA framing    |
```

El ADAPTER-CONTRACT también indica para Claude Code: "✅ REAL — clean room via description". Pero hay un matiz crítico que el contrato omite:

**El Task tool de Claude Code crea aislamiento para L2 (sub-agents), pero NO para L1.** L0, L1a y L1b siguen viviendo en el mismo CLAUDE.md dentro del mismo contexto de sistema. El aislamiento del Task tool solo aplica cuando L1 lanza L2 via `Task()`. L0→L1 no usa Task — usa roles en el mismo prompt.

**El contrato miente para el caso Claude Code L0→L1.**

### Subtópico B: Transferencia SDD_INTENT en L1b — el vector de contaminación

`internal/assets/generic/general-orchestrator.md`:

```markdown
## On SDD_INTENT
→ IMMEDIATELY transfer to SDD Orchestrator with original user message.
```

En Gemini CLI, esta "transferencia" es `run_subagent` — un proceso con contexto limpio. En Claude Code, la "transferencia" es **leer L1a desde el mismo CLAUDE.md** y cambiar el rol de rol cognitivo. El "full user message" incluye la historia de conversación completa de L1b.

```
Escenario de contaminación multi-turn:

L1b turno 1:  Usuario pregunta sobre debugging de auth module
              L1b usa solver → resultado tiene AuthController.validateJWT path
              L1b dice: "encontré el problema en pkg/auth/jwt.go:127"

L1b turno 2:  Usuario dice "ahora usar SDD para agregar rate limiting"
              L1b detecta SDD_INTENT → transfiere a L1a
              Transferencia incluye historial completo de conversación

L1a turno 3 (sdd-explore):
              El historial visible incluye "pkg/auth/jwt.go:127 AuthController"
              sdd-explore puede incluir estos artefactos de L1b en su exploración
              PORQUE están en el mismo contexto

L1a turno 4 (sdd-spec):
              Spec puede referenciar findings de L1b (JWT bug) que no son parte
              del scope del rate limiting SDD
              → Spec contaminada con información fuera de scope
```

### Subtópico C: La regla "L1a and L1b MUST NOT know about each other"

Esta regla existe en los prompts de L1a y L1b como instrucción declarativa. En plataformas con contexto compartido, es materialmente imposible de cumplir: ambos agentes leen el mismo archivo CLAUDE.md y el modelo tiene acceso a todo el historial de conversación.

La instrucción `MUST NOT know about each other` en el CLAUDE.md compartido es **auto-contradictoria**: el hecho de que L1b vea la instrucción "MUST NOT know about L1a" ya implica que L1b sabe que L1a existe.

### Subtópico D: Grafo de propagación de artefactos contaminados

```
sdd-spec artifact (L1a) en contexto compartido
         ↓
L1b (general-orchestrator) recibe nuevo turno de usuario
         ↓
L1b resuelve: "debug the payment module"
         ↓
solver (L2 via L1b) tiene acceso al historial
         ↓
solver ve fragmentos del spec SDD de payment en contexto
         ↓
solver "sugiere" fix que modifica el mismo código que el spec describe
         ↓
sdd-apply (L2 via L1a) en siguiente turno:
  - Encuentra el código ya modificado por solver
  - apply-progress marca tarea como "no aplica — ya implementado"
  - PERO no fue implementado según el spec — fue un patch ad-hoc de solver
         ↓
sdd-verify: código existe ✓, tests pasan ✓, but spec semantic audit FAILS
         → NEEDS CHANGES verdict
         → Re-apply → mismo problema
         → Circuit breaker trips
```

### [MITIGACIÓN F-02 EXTENDIDA]

1. **Context Fence Markers en artefactos SDD** (ya en v1). Ampliar con:
   ```markdown
   <!-- architect-ai:sdd-artifact:PRIVATE:change={name}:phase={phase} -->
   {artifact content}
   <!-- architect-ai:sdd-artifact:END -->
   ```
   Instruir al general-orchestrator explícitamente: "Ignora todos los bloques entre `architect-ai:sdd-artifact:PRIVATE` y `architect-ai:sdd-artifact:END`. No los referencíes ni modifiques el código que describen."

2. **Sesión SDD = sesión exclusiva**: cuando L0 detecta `SDD_INTENT` y no hay sesión SDD activa, emitir:
   ```
   [L0] SDD session iniciada. General-orchestrator suspendido durante SDD.
   Para tareas no-SDD paralelas, abrir nueva sesión.
   ```
   Esto transforma una regla imposible de cumplir en una restricción operacional.

3. **Claude Code: usar Task tool para L1 también** (no solo L2). Crear un sub-agent file `.claude/agents/sdd-orchestrator.md` y llamarlo via `Task()` desde L0 para obtener aislamiento real.

---

## F-03 (EXTENDIDO) — 🔴 CRÍTICO: Placeholders sin Resolver en CLAUDE.md de Producción

### Subtópico A: El mecanismo de hash y su inutilidad en runtime

El CLAUDE.md generado incluye:

```html
<!-- AUTO-GENERATED by architect-ai sync v2 — hash:{CONTENT_HASH} -->
<!-- architect-ai:L0:start hash:{L0_HASH} -->
```

El hash es calculado por el Go installer en tiempo de build y escrito como comentario HTML. El LLM ve este comentario como texto en el system prompt. **El hash no es validado en runtime por ningún mecanismo**. No hay ningún código que, al iniciar Claude Code, lea el CLAUDE.md y verifique que `{L0_HASH}` no sea el placeholder literal.

La verificación de hash que sí existe es en `cmd/eintegrate/main.go` — pero eintegrate es un comando de CI/CD para el repo de `architect-ai` en sí, no para los proyectos de usuario.

### Subtópico B: Ciclo de vida del deployment y ventana de vulnerabilidad

```
Ciclo normal de deployment:
  git clone architect-ai         → CLAUDE.md con placeholders
  architect-ai init --project X  → genera .atl/, skills, etc.
  architect-ai build             → materializa CLAUDE.md con hashes reales
  [claude code abre el proyecto] → ve CLAUDE.md materializado ← OK

Ciclo vulnerable #1 (build omitido):
  git clone architect-ai
  architect-ai init --project X  → .atl/ generado
  [claude code abre el proyecto] ← CLAUDE.md con placeholders! NO hay build
  → Claude Code lee: "{content from .atl/agents/architect.md}"
  → El modelo interpreta como texto literal
  → L0 tiene un system prompt vacío de guardrails

Ciclo vulnerable #2 (sync parcial):
  architect-ai sync  → actualiza algunos archivos de .atl/
  → Si sync falla a mitad → CLAUDE.md puede quedar con secciones mezcladas:
     <!-- L0: start hash:abc123 -->  ← hash viejo
     {contenido actualizado de L0}
     <!-- L1a: start hash:{L1A_HASH} --> ← placeholder (actualización falló)
     {L1a: vacío}
```

### Subtópico C: El eintegrate solo cubre 10 condiciones textuales — no la ausencia de contenido

`cmd/eintegrate/main.go` verifica condiciones como:
- "GEMINI.md contains architect-ai:L0:start"
- "CLAUDE.md contains super-orchestrator-gate"

**Pero NO verifica**: "CLAUDE.md no contiene `{content from`" ni "CLAUDE.md no contiene `{L1A_HASH}`". La ausencia del contenido real no está chequeada. El CLAUDE.md con placeholders pasaría todas las verificaciones de eintegrate porque los marcadores de sección SÍ están presentes.

### Subtópico D: El `--force-assume` flag como amplificador del riesgo

`hard-stop-protocol.md` define `--force-assume` para bypasear stops:

```
User: "use sdd-apply --force-assume"
→ Document assumption in apply-progress.yaml
→ Add to risks section
→ Continue with assumption, mark task as "assumed"
```

Si CLAUDE.md tiene placeholders y el usuario usa `--force-assume`, el agente sin guardrails puede asumir que cualquier spec es completa y aplicar código sin las protecciones del hard-stop-protocol. La combinación es particularmente peligrosa.

### [MITIGACIÓN F-03 EXTENDIDA]

1. **Agregar verificación de placeholders en `architect-ai check all`** (ya en v1). Ampliar con verificación de contenido mínimo:
   ```go
   func checkCLAUDEMDContent(path string) error {
       data, _ := os.ReadFile(path)
       s := string(data)
       forbidden := []string{
           "{content from", "{L0_HASH}", "{L1A_HASH}", "{L1B_HASH}",
           "{CONTENT_HASH}", "{INJECTED_MODE}", "{D1}", "{D2}",
       }
       for _, f := range forbidden {
           if strings.Contains(s, f) {
               return fmt.Errorf("CLAUDE.md has unresolved placeholder: %q — run: architect-ai build", f)
           }
       }
       // Verify minimum content size (a real CLAUDE.md is > 50KB)
       if len(data) < 10_000 {
           return fmt.Errorf("CLAUDE.md too small (%d bytes) — likely not built: architect-ai build", len(data))
       }
       return nil
   }
   ```

2. **`architect-ai init` debe ejecutar `build` automáticamente** o rechazar con mensaje claro si el build no se ha ejecutado antes de que el usuario abra Claude Code.

3. **Pre-commit hook via GGA** que verifique que CLAUDE.md no tiene placeholders antes de cualquier commit en proyectos architect-ai:
   ```bash
   # .atl/scripts/pre-commit-verify.sh
   if grep -q "{content from\|{L0_HASH}" CLAUDE.md 2>/dev/null; then
     echo "ERROR: CLAUDE.md has unresolved placeholders. Run: architect-ai build"
     exit 1
   fi
   ```

---

## F-04 (EXTENDIDO) — 🔴 CRÍTICO: Race Condition + Enum Mismatch en sdd-state.yaml

### Subtópico A: El TOCTOU en detalle — ventana de race en microsegundos

El flujo de `Save()` en `openspec/state.go`:

```go
func Save(path string, s *State) error {
    s.UpdatedAt = time.Now().UTC()
    // Validate ANTES de escribir
    if err := Validate(s, parent); err != nil {
        return err
    }
    out, _ := yaml.Marshal(s)
    tmp := path + ".tmp"
    os.WriteFile(tmp, out, 0o644)  // ← escribe al tmp
    // fsync best-effort
    if f, _ := os.OpenFile(tmp, os.O_RDWR, 0o644); f != nil {
        _ = f.Sync(); _ = f.Close()
    }
    os.Rename(tmp, path)  // ← rename atómico: OK para el tmp→final
    return nil
}
```

La operación `Rename(tmp, path)` ES atómica en Linux/macOS (garantía del kernel). **El problema no está en el rename**, sino en que DOS procesos pueden crear dos archivos `.tmp` simultáneamente y ambos hacer rename, con el segundo sobreescribiendo al primero.

```
Proceso A (sdd-apply):           Proceso B (sdd-verify paralelo):
  WriteFile("state.yaml.tmp", A)
                                   WriteFile("state.yaml.tmp", B)  ← SOBREESCRIBE A
  Rename("state.yaml.tmp", "state.yaml")  ← ahora contiene B
```

La semántica de `WriteFile` con el mismo path es last-writer-wins. Ambos procesos escriben al MISMO path `.tmp` (no a `.tmp.PID`), creando la colisión.

**La corrección NO requiere flock** — solo usar paths únicos para los archivos temporales:

```go
tmp := fmt.Sprintf("%s.tmp.%d.%d", path, os.Getpid(), time.Now().UnixNano())
```

### Subtópico B: El Enum Mismatch "running" vs "in_progress" — bug de primer orden

Este es un hallazgo nuevo de alta severidad encontrado en el análisis profundo:

`phase-dag-enforcement.md` (línea 11874, batch 1):
```markdown
- Phase start → update status to `running`
- Phase success → update status to `completed`
- Phase failure → update status to `failed`
```

`openspec/state.go` (línea 5173, batch 6):
```go
var ValidStatuses = []string{
    "pending", "in_progress", "completed", "skipped", "failed",
}
```

**`"running"` NO está en `ValidStatuses`.** El valor que el protocolo de prompt instruye a escribir es inválido según el validador Go. Cuando un agente actualiza el estado a `running`, la función `Save()` ejecuta `Validate()` que rechaza con `ErrEnum`.

El resultado silencioso:
```
sdd-apply inicia → intenta Save(state, {sdd-apply.status: "running"})
→ Validate() retorna ErrEnum: "value not in allowed enum"
→ Save() retorna error
→ El agente recibe error de escritura
→ El agente reporta el error como blocker
→ Circuit breaker incrementa attempt_count
→ El apply fue bloqueado por un bug de enum, no por fallo de dominio
```

Paralelamente, el script de rollback en `rollback-harness.md`:

```python
content = re.sub(
    r'(  sdd-apply:.*?status: )\"(running|failed)\"',
    r'\1\"pending\"',
    content, flags=re.DOTALL
)
```

El rollback busca `"running"` para resetear — pero como `"running"` no puede existir (el validator lo rechaza), el regex NUNCA matchea y el rollback NO resetea el estado a `pending`. El rollback falla silenciosamente.

**Tercer impacto**: la condición de pre-flight en sdd-verify:

```markdown
If `sdd-apply.status in {in_progress, failed}` → REFUSE
```

Este check usa `"in_progress"` (correcto Go). Pero los prompts que escriben el estado usan `"running"` (incorrecto). Si algún path de código escribió `"running"` directamente en el YAML (bypass del validator), la condición de pre-flight busca `"in_progress"` y no lo encuentra → sdd-verify procede con un apply que está en estado inconsistente.

### Subtópico C: El lockfile del prompt vs el lockfile del código

`phase-dag-enforcement.md`:
```markdown
All writes atomic: `.atl/sdd-state.yaml.tmp` → rename,
protected by `.atl/sdd-state.yaml.lock`
```

Esta instrucción habla de un `.lock` file como protección. Pero `openspec/state.go` NO implementa ningún `.lock` file. `Save()` solo hace `WriteFile(tmp)` + `Rename()`. No hay ninguna referencia a `sdd-state.yaml.lock` en el código Go.

El lockfile existe SOLO en los prompts — es una promesa que el código no cumple.

### [MITIGACIÓN F-04 EXTENDIDA]

1. **Fix inmediato del enum**: cambiar `ValidStatuses` o cambiar los prompts. La opción correcta es unificar en `"in_progress"` (estándar Go) y actualizar todos los archivos de fase:

   Buscar y reemplazar en todos los assets:
   ```bash
   rg -l "status.*\"running\"\|\"running\".*status" internal/assets/ | \
     xargs sed -i 's/status: `running`/status: `in_progress`/g; s/"running"/"in_progress"/g'
   ```

2. **Fix del archivo .tmp único por proceso**:
   ```go
   tmp := fmt.Sprintf("%s.%d.%d.tmp", path, os.Getpid(), time.Now().UnixNano())
   ```
   Esto elimina la colisión de escritura al .tmp compartido.

3. **Implementar el lockfile prometido** con `O_CREATE|O_EXCL`:
   ```go
   func acquireLock(lockPath string) (func(), error) {
       f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
       if err != nil {
           if os.IsExist(err) {
               // Check staleness
               if info, serr := os.Stat(lockPath); serr == nil && 
                  time.Since(info.ModTime()) > 30*time.Second {
                   os.Remove(lockPath)
                   return acquireLock(lockPath) // single retry
               }
               return nil, fmt.Errorf("state locked by another process")
           }
           return nil, err
       }
       fmt.Fprintf(f, "%d", os.Getpid())
       f.Close()
       return func() { os.Remove(lockPath) }, nil
   }
   ```

4. **Test de race condition con -race flag**:
   ```go
   func TestSaveConcurrent(t *testing.T) {
       t.Parallel()
       dir := t.TempDir()
       path := filepath.Join(dir, "state.yaml")
       var wg sync.WaitGroup
       errs := make(chan error, 10)
       for i := 0; i < 10; i++ {
           wg.Add(1)
           go func(n int) {
               defer wg.Done()
               s := &State{...}
               s.Phases["sdd-apply"].Status = "in_progress"
               if err := Save(path, s); err != nil {
                   errs <- err
               }
           }(i)
       }
       wg.Wait()
       close(errs)
       // Should have zero errors (lock prevents concurrent writes)
       // go test -race should find no data races
   }
   ```

---

## F-05 (EXTENDIDO) — 🟠 ALTO: Amnesia de Sesión y Bootstrap Paradox de Engram

### Subtópico A: El Bootstrap Paradox — Engram para detectar Engram

La Session State Cache usa Engram para almacenar el resultado del probe de Engram:

```
Session start probe:
  mem_search("session-state/{project}/tools") → obtener cache de tools
  [si cache hit && age < 30min] → usar cache, no re-probar

  [si cache miss] → ejecutar probe-1: mem_search("any query") → test de disponibilidad
  → Si probe-1 OK: Engram disponible
  → Después del probe: mem_save("session-state/{project}/tools", {engram: true, ...})
```

**Paradoja**: para saber si Engram funciona, se hace `mem_search`. Si Engram está caído, `mem_search` falla → Engram marcado como unavailable. Entonces se intenta `mem_save` (para cachear el resultado del probe) → también falla. El sistema colapsa en `none` mode sin persistir ningún rastro de la decisión.

En la siguiente sesión, no hay cache → repite el probe → mismo resultado. **No hay degradación gradual, solo cliff-edge entre "Engram up" y "modo none total".**

### Subtópico B: La cadena de pérdida cuando Engram cae mid-session

```
Sesión en progreso — Engram estaba UP al inicio:
  Turno 1-20:  SDD workflow normal, artifacts en Engram
               sdd-spec, sdd-design completados → persistidos
               session-state cache guardada: {engram: true}
  
  [Engram MCP server cae — red timeout, proceso muerto]
  
  Turno 21:  L0 intenta mem_search para contexto
             → timeout (30s default para MCP)
             → L0 clasifica como available (cache dice true, age < 30min)
             → Continúa asumiendo Engram UP
  
  Turno 22:  sdd-apply intenta mem_save(apply-progress)
             → MCP timeout → error ignorado (protocolo dice "note in return envelope")
             → apply-progress NO persistido
  
  Turno 23:  sdd-verify intenta mem_get_observation("sdd/{change}/spec")
             → MCP timeout → null
             → Semantic Audit Protocol: "IF spec null → partial audit using design only"
             → Design también es null → "ADD RISK 'no contracts in Engram', SKIP audit"
             → sdd-verify sin base para verificar
             → Verdict: UNRESOLVED (escalate to user)
  
  [Context compaction automática en turno 24]
  → mem_session_summary no puede ejecutarse (Engram caído)
  → Toda la sesión de 24 turnos: PERDIDA
```

### Subtópico C: El trigger de mem_session_summary — análisis de todos los triggers documentados

Triggers documentados en el GEMINI.md/CLAUDE.md para invocar `mem_session_summary`:

```
1. Usuario dice: "done", "that's it", "listo", "cierra la sesión"
2. Usuario dice: "wrap up", "close session", "finalizamos"
3. D4 alcanza nivel 3 (Guardian Active) → context-guardian invoca compactación
4. /compact ejecutado por usuario
```

**Análisis de los triggers**:

- Triggers 1-2: dependen del usuario. No documentados en ningún hook de proceso, no automatizables.
- Trigger 3: context-guardian se activa cuando `char_count >= 100K`. El guardian invoca compactación PERO el CLAUDE.md que lo define dice "Suspended during /compact" (caveman-firewall regla). El guardian no puede invocar mem_session_summary durante la suspensión.
- Trigger 4: `/compact` es un comando de Claude, no un hook. El pre-compact-hook de Claude Code existe pero arquitecturamente no está configurado en el proyecto (no hay `.claude/settings.json` con `PreToolUse` para el tool de compactación).

**Conclusión**: ningún trigger de mem_session_summary es automático y robusto ante compactación inesperada.

### Subtópico D: El `engram-compact-prompt.md` — solución que solo aplica a Codex

`internal/assets/codex/engram-compact-prompt.md` configura el prompt de compactación para la plataforma Codex (OpenAI). Para Claude Code, este archivo no aplica.

Claude Code tiene su propio sistema de `/compact` con un prompt customizable en `~/.claude/CLAUDE.md`. El proyecto architect-ai NO incluye instrucciones pre-compact en el CLAUDE.md del usuario, solo en el CLAUDE.md del proyecto.

La pregunta crítica: ¿el `engram-compact-prompt.md` de Codex tiene el trigger de `mem_session_summary` en el hook de compactación?

```markdown
# Codex engram-compact-prompt.md content (línea 18496):
## Before compact — MANDATORY
1. mem_save("session-state/{project}/checkpoint", ...)
2. If SDD active: mem_save("sdd/{change}/state", ...)
→ Then proceed with /compact
```

**Sí lo tiene para Codex. NO está replicado para Claude Code ni Gemini.**

### [MITIGACIÓN F-05 EXTENDIDA]

1. **Implementar pre-compact hook en Claude Code** via `.claude/settings.json`:
   ```json
   {
     "hooks": {
       "PreToolUse": [{
         "matcher": "Compact",
         "hooks": [{
           "type": "command",
           "command": "architect-ai session-checkpoint --force"
         }]
       }]
     }
   }
   ```
   Donde `architect-ai session-checkpoint` llama `mem_session_summary` si Engram está disponible.

2. **Replicar `engram-compact-prompt.md` para Claude Code y Gemini**:
   ```
   internal/assets/claude/engram-compact-prompt.md   ← nuevo
   internal/assets/gemini/engram-compact-prompt.md   ← nuevo
   ```
   Con el mismo contenido del pre-compact checkpoint.

3. **Fallback de Engram caído mid-session**: cuando una operación Engram falla durante una sesión activa, L0/L1 debe:
   - Escribir checkpoint local en `.atl/emergency-checkpoint.yaml`
   - Emitir alerta al usuario: "[ENGRAM DOWN] Memory persistence lost. Local checkpoint saved to .atl/emergency-checkpoint.yaml"
   - Cambiar artifact_store a `openspec` para continuar

4. **Eliminar el bootstrap paradox**: guardar un fallback de tool-availability en `.atl/session-tools-cache.json` en disco, actualizado por cada sesión exitosa. Si Engram está caído al inicio, leer el cache en disco.

---

## F-06 (EXTENDIDO) — 🟠 ALTO: GATE_ERROR como Amplificador de Circuit Breaker

### Subtópico A: El grafo completo de propagación del GATE_ERROR

```
Orquestrador L1a intenta delegar a sdd-spec sub-agent:
  
  Paso 1: L1a ejecuta Skill Resolver
    → Obtiene foundation_block (Tier 1) ✓
    → Obtiene go-testing compact rules (Tier 2) ✓
    → Construye el prompt del sub-agent
  
  Paso 2: L1a calcula D1, D2, D3, D4 para la tarea
    → D1=2 (systemic), D2=1 (partial specs), D3=0, D4=1
    → Mode = 2 (Tactical)
  
  Paso 3: L1a inyecta el modo en el header del sub-agent:
    "[MODE 2 | D1=2, D2=1, D3=0, D4=1] spec task is systemic complexity"
  
  [FALLO: context-guardian compactó el contexto de L1a entre T22 y T23]
  
  Paso 3 POST-COMPACT: L1a pierde la computación D1-D4
    → El prompt generado tiene literalmente: "[MODE {INJECTED_MODE} | D1={D1}, ...]"
    → El placeholder no fue resuelto porque el template de inyección
      dependía de variables en el contexto que se compactaron
  
  Paso 4: sub-agent recibe prompt con placeholders literales
    → Gate v2: "IF any fields contain unfilled placeholders → GATE_ERROR"
    → Sub-agent emite: [GATE_ERROR: mode not injected]
    → status: blocked, attempt_number: 1
  
  Paso 5: L1a recibe blocked con attempt_number:1
    → Circuit breaker: attempt_count["sdd-spec"] = 1 → "Update approach, retry"
    → L1a reintenta con MISMO template (el bug está en la compactación, no en la spec)
    → GATE_ERROR de nuevo
    → attempt_count["sdd-spec"] = 2 → "Request user context"
    → Usuario no sabe que el problema es infrastructure
    → attempt_count["sdd-spec"] = 3 → CIRCUIT BREAKER TRIPS
    → sdd-spec: abandoned
    → TODO el workflow SDD se detiene
```

### Subtópico B: El GATE_ERROR no distingue entre error de infraestructura y error de dominio

El `result-contract.md` tiene solo 4 status válidos:
```json
"status": "completed|failed|blocked|abandoned"
```

No existe `"infrastructure_blocked"`. El GATE_ERROR produce `"blocked"`, que el circuit breaker cuenta igual que cualquier otra falla. No hay forma de que el orquestrador sepa si el blocked es por:
- Template no inyectado (infrastructure) → siempre retriable con el mismo fix
- Spec contradictoria (domain) → necesita clarificación del usuario
- Prerequisito no completado (DAG) → necesita completar la fase anterior

El circuit breaker aplica la misma escalación para los tres casos.

### Subtópico C: La interacción con el Skill Resolution Feedback

`skill-resolution-feedback.md` define:

```
IF result.skill_resolution.status == "none":
  → Log to Engram: topic_key="skill-resolution-failure/{phase}/{timestamp}"
  → Escalate to human if D5 >= 2
  → Retry with explicit skill list
```

Pero si Engram está caído (F-05), `mem_save("skill-resolution-failure/...")` falla silenciosamente. El log de escalación se pierde. El orquestrador no tiene evidencia de que el skill resolver falló, por lo que el retry se hace sin la corrección del skill.

Este es un **failure cascading triple**: Engram caído → skill log perdido → retry sin corrección → GATE_ERROR de nuevo → circuit breaker cuenta fallo de skill como fallo de dominio.

### [MITIGACIÓN F-06 EXTENDIDA]

1. **Nuevo status en result-contract**: `"infrastructure_blocked"` que el circuit breaker trata diferente:
   ```json
   {
     "status": "infrastructure_blocked",
     "infrastructure_error": {
       "type": "gate_template_unresolved | engram_unavailable | skill_file_missing",
       "retry_without_user": true,
       "max_infra_retries": 2
     }
   }
   ```
   El CB tiene un contador separado: `infrastructure_attempt_counts` que no acumula contra el domain limit de 3.

2. **Gate v2 fallback a auto-score** (ya en v1). Ampliar con logging local:
   ```markdown
   IF placeholders unfilled:
     → Write to .atl/gate-errors.log: timestamp, phase, missing_fields
     → Auto-compute D1-D4 from task description
     → Log: [GATE_FALLBACK] Using auto-scored MODE 2
     → Proceed (don't block on infrastructure error)
   ```

3. **Skill log fallback a filesystem**: si `mem_save("skill-resolution-failure/...")` falla, escribir a `.atl/skill-failures.log`:
   ```go
   if err := memSave(ctx, skillFailureKey, payload); err != nil {
       // Engram unavailable — write to local fallback
       appendToFile(".atl/skill-failures.log", fmt.Sprintf("%s: %s\n", time.Now(), payload))
   }
   ```

---

## F-07 (EXTENDIDO) — 🟠 ALTO: Bucle Infinito sdd-verify — Análisis Completo del Grafo de Estados

### Subtópico A: El Dependency Graph y los estados sin salida

El Dependency Graph del sdd-orchestrator (extraído del código):

```
proposal → specs → design → tasks → apply → verify → archive
                      ↑                          |
                      └──── NEEDS CHANGES ────────┘
                                (Judgment Day)
```

El grafo tiene un ciclo intencional: `verify:NEEDS CHANGES` → `sdd-apply` → `verify` de nuevo. El circuit breaker limita este ciclo a 3 intentos por fase.

**Pero el estado `verify:abandoned` no tiene arco de salida en el grafo**:

```
verify:abandoned ──→ ???
```

El `sdd-archive` tiene como prerequisito:
```markdown
Verify: Confirm verdict is APPROVED or CONDITIONALLY APPROVED
If verdict is NEEDS CHANGES or UNRESOLVED → STOP and return blocked
```

No menciona `verify:abandoned`. El archivo (sdd-archive.md) tampoco define qué hacer con `verify:abandoned`. El grafo tiene un sumidero sin salida.

### Subtópico B: La Severity → Action Matrix y sus huecos

La matrix en `sdd-verify.md` Result Processing:

```
| CRITICAL     | Immediate re-apply     |
| BLOCKING     | Wait for user          |
| WARNING-REAL | Present to user        |
| WARNING-THEO | Note, maintain APPROVED|
| SUGGESTION   | Note, maintain APPROVED|
```

**Hueco 1**: No existe `"INFRASTRUCTURE"` severity. Un test runner unavailable produce hallazgos que sdd-verify no puede clasificar en ninguna de estas categorías. El protocolo dice "If test runner unavailable → flag as RISK in report, do NOT claim APPROVED". ¿En qué severity va ese RISK? No está definido → el agente elige → comportamiento no determinístico.

**Hueco 2**: La deterministic check "Tests exist for each capability" — ¿qué pasa si no hay tests para una funcionalidad que el spec no especificó tests para? ¿Es CRITICAL o WARNING-REAL? No está definido.

**Hueco 3**: WCAG Compliance Check — "Verify aria-labels, contrast ratios, keyboard accessibility." Para un change de backend puro (Odoo model field), estos checks no aplican pero no hay condición `IF frontend change`. El agente intentará aplicar WCAG checks a código Python y puede producir findings espurios.

### Subtópico C: El `--skip-ai` flag como vector de bypass del TDD

El código-reviewer (`sdd-verify.md`) define:
```
SKIP: --skip-ai flag → static checks only, no AI verdict.
```

Y los static checks NO incluyen:
- Semantic audit protocol (sdd-verify semantic-audit es AI)
- Adversarial review (AI)
- Test baseline comparison (requiere ejecutar el test runner, pero el AI verdict es quien interpreta)

Con `--skip-ai`, sdd-verify solo verifica que los archivos existen y el JSON del result contract es válido. Un apply que introdujo código sin tests puede pasar con `--skip-ai`.

**La combinación peligrosa**: `sdd-apply --force-assume` + `sdd-verify --skip-ai` = bypass completo del pipeline de calidad. Ninguno de estos flags tiene logging en Engram ni en sdd-archive.

### Subtópico D: Interaction con el Strict TDD mode — el bucle irrecuperable

En Strict TDD mode (`StrictTDD: true`), el ciclo es más agresivo:

```
sdd-apply (batch N):
  Implementa feature Y
  Escribe test para Y → test FALLA (RED phase del TDD)
  "CRITICAL — test file implements trivial assertions: expect(true).toBe(true)"
  → HARD STOP (strict-tdd-verify.md regla: trivial assertion detection)
  → status: blocked, attempt: 1

sdd-orchestrator:
  Receives blocked with CRITICAL
  → "Immediate re-apply"
  → sdd-apply attempt 2:
    Intenta arreglar el test trivial
    → El test SÍ era el placeholder que el developer iba a reemplazar
    → "expect(true).toBe(true)" era un scaffold, no el test final
    → sdd-apply arregla el scaffold con un test real
    → sdd-verify: "CRITICAL — test doesn't cover the happy path"
    → blocked, attempt 2

  → sdd-apply attempt 3:
    Agrega coverage al happy path
    → sdd-verify: "WARNING-REAL — no negative test case for empty input"
    → CONDITIONALLY APPROVED (no CRITICAL)
    → sdd-archive puede proceder

  ← OK en este caso
```

El problema ocurre cuando el CRITICAL es estructural (no arreglable por sdd-apply):

```
sdd-verify CRITICAL: "Function process_payment undefined in design"
→ El diseño no incluye process_payment porque es legacy code
→ sdd-apply no puede arreglar el diseño desde la fase de apply
→ sdd-apply attempt 2: mismo CRITICAL
→ attempt 3: mismo CRITICAL
→ Circuit breaker trips: verify:abandoned
→ change queda en limbo: apply:completed, verify:abandoned, archive:blocked
```

### [MITIGACIÓN F-07 EXTENDIDA]

1. **Agregar `verify:abandoned` al Dependency Graph con arco de salida**:
   ```
   verify:abandoned → sdd-orchestrator emite DEADLOCK menu:
     "[1] /sdd-hotfix — patch the specific failing assertion (≤3 files)"
     "[2] /sdd-archive --status=abandoned — close as abandoned"
     "[3] /sdd-ff verify --force -- skip and document exception"
     "[4] /sdd-design --amend — go back to design phase"
   ```
   Opción 4 es nueva: re-abrir la fase de design para arreglar la contradicción.

2. **Clasificar CRITICAL por tipo en sdd-verify**:
   ```
   CRITICAL:domain    → re-apply (domain bug, fixable by code)
   CRITICAL:spec      → back to sdd-spec (spec incomplete/wrong)
   CRITICAL:design    → back to sdd-design (design missing function)
   CRITICAL:infra     → infrastructure_blocked (no aplica CB domain count)
   ```

3. **Registrar `--skip-ai` en Engram como exception record**:
   ```
   mem_save(
     topic_key: "sdd/{change}/exceptions",
     content: { skip_ai_used: true, phase: "sdd-verify", timestamp, reason: null }
   )
   ```
   sdd-archive requiere que exceptions estén documentadas con razón justificada.

---

## F-08 (EXTENDIDO) — 🟠 ALTO: Saturación de Contexto por Odoo MCP

### Subtópico A: El Research Order y el camino a la saturación

El Odoo Research Order (odoo-overlay-routing.md §5):

```
1. Engram (project memory + prior Odoo decisions)        ← mem_search
2. rg en local custom addons (project/addons_customs/)   ← bash, contextualizado
3. rg en Odoo Community local (~/gitproj/...)            ← bash, GRANDES directorios
4. Context7 with "odoo" library-id                       ← MCP, bounded
5. OCA GitHub (search)                                   ← web, bounded
6. Odoo Community GitHub (browse)                        ← web, bounded
```

El paso 3 es el vector de saturación real: `rg` en el directorio Odoo Community puede producir miles de líneas de resultados. La instrucción es usar `ctx_execute` para contener el output, pero la instrucción es "Do:" (recomendación) no "MUST" (mandatoria).

El paso MCP (`mcp_odoo_search_records`) también es un vector, pero el Research Order lo posiciona en el paso 4-6 (secundario), no primario. Sin embargo, `odoo-context-gatherer` — el sub-agent delegado — puede llamar `mcp_odoo_search_records` como PRIMER paso si la pregunta es sobre datos vivos (no código).

### Subtópico B: ODOO_YOLO — la configuración que habilita mutaciones bajo contexto degradado

`internal/components/mcp/generator.go` (línea 3767):

```go
"ODOO_URL": opts.OdooURL,
"ODOO_USER": opts.OdooUser,
"ODOO_YOLO": boolStr(opts.OdooYolo),
```

`ODOO_YOLO=true` es pasado como variable de entorno al `mcp-server-odoo`. Esto le indica al servidor MCP que puede ejecutar operaciones de escritura (`write`, `create`, `delete`) sin confirmación adicional del usuario.

**La variable es configurada por el instalador** (`architect-ai init --odoo-yolo`) y persiste en `.env.mcp`. No hay mecanismo en runtime para desactivar YOLO si el D-score del agente sube.

```
Escenario de fallo:
  Sesión Odoo con YOLO=true
  Turno 15: mcp_odoo_search_records(model="account.move", domain=[]) → 5000 facturas
  → context_history sube a 120K chars → D4 = 3 (Guardian Active)
  → Mode 3-CTX: +++Pragmatic, 50% output compression
  → Agent está en modo de máxima compresión cognitiva

  Turno 16: usuario dice "actualiza el impuesto de IVA a 21%"
  → Agent en Mode 3-CTX toma la acción más "pragmática"
  → Con YOLO=true: mcp_odoo_write("account.tax", ids=[...], {"amount": 21})
  → 847 registros de impuestos actualizados sin confirmación
  → PRODUCCIÓN AFECTADA
```

El agent en Mode 3-CTX no tiene ninguna guardia que deshabilite YOLO o requiera confirmación.

### Subtópico C: El CAUTION_POLICY.md — inyección full en cada sub-agent

`odoo-overlay-routing.md §7`:

```markdown
.atl/overlays/odoo-development-skill/rules/CAUTION_POLICY.md → full (critical)
```

**`full` significa el archivo completo, no compact rules**. Para cada sub-agent en un proyecto Odoo, el CAUTION_POLICY.md completo se inyecta en el prompt. Si este archivo tiene 8KB y se invoca en 15 sub-agents durante un pipeline SDD completo:

```
8KB × 15 llamadas = 120KB de CAUTION_POLICY repetido en tokens
```

Esto es aproximadamente **90.000 tokens** (estimado a 750 tokens/KB) solo de CAUTION_POLICY inyectado repetitivamente. A precios de Claude Sonnet, esto representa costos significativos por pipeline.

### Subtópico D: La versión paginada que no existe en el tooling

El Odoo guide menciona:
```python
records = self.search([], order='date DESC', limit=10)
```

Esta es documentación de referencia para CÓDIGO Python. No hay ningún mecanismo en el MCP server que imponga un limit máximo en las llamadas del LLM. Si el agente llama:

```
mcp_odoo_search_records(model="stock.move", domain=[], fields=["all"])
```

El servidor ejecuta la consulta sin límite. La responsabilidad de paginar está en el agente — que puede estar operando bajo Mode 3-CTX con razonamiento degradado.

### [MITIGACIÓN F-08 EXTENDIDA]

1. **YOLO Guard con D-score check** (ya en v1). Implementar en el MCP server:
   ```python
   # mcp-server-odoo/src/server.py
   def handle_write(ctx, model, ids, vals):
       d_score = ctx.get("agent_d3", 0) + ctx.get("agent_d4", 0)
       yolo = os.getenv("ODOO_YOLO", "false").lower() == "true"
       
       if yolo and d_score >= 3:
           return {
               "error": "YOLO_GUARD: D3+D4 >= 3 (context degraded). Explicit confirmation required.",
               "requires_confirmation": True,
               "confirmation_token": generate_token(model, ids, vals)
           }
       # proceed with write
   ```

2. **Hard limit de records en el MCP server**:
   ```python
   MAX_RECORDS_DEFAULT = 50
   MAX_RECORDS_YOLO = 200  # YOLO permite más pero aún limitado
   
   def handle_search_records(ctx, model, domain, fields, limit=None):
       effective_limit = min(limit or MAX_RECORDS_DEFAULT, 
                            MAX_RECORDS_YOLO if yolo else MAX_RECORDS_DEFAULT)
       # enforce at server level, not agent level
   ```

3. **CAUTION_POLICY a Engram como conocimiento semántico**:
   ```markdown
   # odoo-overlay-routing.md §7 PROPUESTA
   .atl/overlays/.../CAUTION_POLICY.md → compact (summary: first 500 chars)
   Full policy available via: mem_search("knowledge/odoo/caution-policy")
   ```
   Esto reduce la inyección de 8KB a ~500 chars por llamada.

---

## F-09 (EXTENDIDO) — 🟡 MEDIO: TUI ProgressFunc No-Op — Análisis Completo

### Subtópico A: El flujo de mensajes BubbleTea y la brecha de `p.Send()`

BubbleTea opera en un event loop: `Init → Update → View`. Los mensajes externos (de goroutines) se envían al programa via `p.Send(msg)`. El TUI de architect-ai tiene esta arquitectura:

```go
// main.go — estructura correcta
p := tea.NewProgram(model.New(...))
// ...
p.Run()
```

Pero el `Model` en `startInstalling` no tiene referencia a `p`:

```go
// model.go
func (m Model) startInstalling(/* sin p *tea.Program */) (tea.Model, tea.Cmd) {
    return m, tea.Batch(tickCmd(), func() tea.Msg {
        onProgress := func(event pipeline.ProgressEvent) {
            // NOTE: ProgressFunc is called synchronously from the pipeline goroutine.
            // We cannot use p.Send() here because we don't have a reference to the
            // tea.Program. Instead, these events are collected in the ExecutionResult
        }
        result := executeFn(selection, resolved, detection, onProgress)
        return PipelineDoneMsg{Result: result}
    })
}
```

El comentario en el código reconoce el problema: "We cannot use p.Send() here because we don't have a reference to the tea.Program."

### Subtópico B: El caso de BubbleTea y el patrón correcto para `p.Send()`

El patrón estándar de BubbleTea para enviar mensajes desde goroutines es:

```go
// Patrón A: pasar p al Model
type Model struct {
    program *tea.Program  // referencia al programa
    // ...
}

// En Init() o Update():
func (m Model) someCmd() tea.Cmd {
    return func() tea.Msg {
        go func() {
            // background work
            m.program.Send(SomeMsg{})  // envío desde goroutine
        }()
        return nil
    }
}
```

```go
// Patrón B: usar tea.ExecProcess o tea.Cmd
func longRunningCmd(progressCh chan ProgressEvent) tea.Cmd {
    return func() tea.Msg {
        for event := range progressCh {
            // Esto NO funciona — Cmd solo puede retornar UN Msg
        }
        return DoneMsg{}
    }
}
```

El Patrón A requiere que el `Model` tenga referencia a `*tea.Program`. El Patrón B (lo que tiene el código actual) solo permite retornar UN mensaje al final — que es exactamente lo que produce la experiencia "todo aparece de golpe al final".

### Subtópico C: El impacto en UX — instalaciones largas sin feedback

La screen `ScreenInstalling` renderiza via `screens.RenderInstalling(m.Progress.ViewModel(), ...)`. El `ProgressViewModel` es actualizado solo cuando llega un `StepProgressMsg` — que nunca llega durante la ejecución (solo el `PipelineDoneMsg` final).

`ProgressFromExecution()` en `model.go` convierte el `ExecutionResult` final en el modelo de progreso. Esto produce una transición instantánea de "todo en progreso" a "todo completado" cuando llega el `PipelineDoneMsg`.

**Para una instalación de 3 minutos**, el usuario ve:
```
⠸ Install dependencies    ← spinner (inmóvil, solo el spinner rota)
⠸ Configure agents        ← spinner
⠸ Inject components       ← spinner

[3 minutos sin cambios]

✓ Install dependencies    ← todo completo de golpe
✓ Configure agents
✓ Inject components
```

El spinner rota (via `TickMsg` cada 100ms), lo cual crea la ilusión de actividad, pero no hay información de progreso real.

### Subtópico D: Error recovery en pipeline sin progress events

Si el pipeline falla en el step 1 de 5, el usuario no lo sabe hasta que `PipelineDoneMsg` llega con `ExecutionResult.Error` en el step 1. No hay forma de intervenir antes de que el pipeline complete su intento.

En instalaciones con errores de red (dependencias que fallan al descargar), el usuario espera el timeout completo de red (potencialmente 60-120s) sin saber qué está pasando.

### [MITIGACIÓN F-09 EXTENDIDA]

```go
// model.go — refactoring completo
type Model struct {
    // ...
    program *tea.Program  // NUEVO: referencia al programa
}

// En el constructor:
func New(version string, p *tea.Program) Model {
    return Model{
        program: p,
        // ...
    }
}

// En startInstalling:
func (m Model) startInstalling() (tea.Model, tea.Cmd) {
    prog := m.program  // capturar referencia antes del closure
    return m, tea.Batch(tickCmd(), func() tea.Msg {
        onProgress := func(event pipeline.ProgressEvent) {
            if prog != nil {
                prog.Send(StepProgressMsg{
                    StepID:    event.StepID,
                    StepName:  event.StepName,
                    Status:    pipeline.StepStatus(event.Status),
                    Err:       event.Err,
                    Timestamp: time.Now(),
                })
            }
        }
        result := executeFn(m.Selection, m.Resolved, m.Detection, onProgress)
        return PipelineDoneMsg{Result: result}
    })
}

// main.go — pasar p al model:
m := model.New(version, nil)  // p=nil inicialmente
p := tea.NewProgram(m)
m.SetProgram(p)  // o re-inicializar después de crear p
```

---

## F-10 (EXTENDIDO) — 🟡 MEDIO: Goroutines Huérfanas en AgentBuilder

### Subtópico A: El ciclo de vida completo de la goroutine de generación

```go
// Ciclo normal:
startGeneration() → goroutine lanzada con ctx
→ engine.Generate(ctx, prompt) → Anthropic API call (HTTP)
→ response retorna → AgentBuilderGeneratedMsg enviado
→ Model recibe msg → m.AgentBuilder.Generating = false

// Ciclo de cancelación:
User presiona Esc → handleKeyPress → m.AgentBuilder.GenerationCancel()
→ ctx cancelado
→ IF engine respeta context: HTTP request abortado → goroutine termina
→ IF engine NO respeta context: HTTP request continúa hasta completion/timeout
   → goroutine vive el timeout completo (5 minutos por defecto)
   → AgentBuilderGeneratedMsg eventualmente llega
   → Model.Update() case AgentBuilderGeneratedMsg:
       if !m.AgentBuilder.Generating { return m, nil }  // ignorado correctamente
```

La goroutine huérfana es ignorada correctamente en el handler. El problema es de recursos, no de correctness.

### Subtópico B: El riesgo de double-generation

```
T=0:  startGeneration() → goroutine G1 lanzada
T=5s: User presiona Esc → cancel() → ctx cancelado
T=6s: User presiona Generate de nuevo → startGeneration() → goroutine G2 lanzada
T=10s: G1 completa (API no respetó cancel) → AgentBuilderGeneratedMsg {Agent: A1}
T=15s: G2 completa → AgentBuilderGeneratedMsg {Agent: A2}

Model recibe {Agent: A1}: m.AgentBuilder.Generating = true → procesado → A1 instalado
Model recibe {Agent: A2}: m.AgentBuilder.Generating = false → ignorado

PERO: hay una ventana entre T=10s y T=15s donde A1 fue instalado y m.AgentBuilder.Generated = A1
Si el usuario navega al preview screen entre T=10 y T=15, ve A1.
Si acepta y A2 llega después (ignorado), el usuario instaló A1 que fue del generate cancelado.
```

### Subtópico C: El timeout de 5 minutos y su racionalidad

El timeout de `context.WithTimeout(context.Background(), 5*time.Minute)` es razonable para una API call, pero hay implicaciones:

1. **Sin Esc + timeout**: si el usuario no presiona Esc y la API no responde, la goroutine bloquea por 5 minutos consumiendo una conexión HTTP.
2. **Con Esc + engine no respeta cancel**: la goroutine sigue 5 minutos después de que el usuario canceló.
3. **Multiple cancels**: en teoría, el usuario puede cancelar y regenear 10 veces → 10 goroutines huérfanas simultáneas, cada una con una conexión HTTP abierta al API.

El `http.DefaultClient` en Go NO respeta el context por defecto en versiones antiguas. Si el Anthropic SDK que usa el engine no propaga el context al `http.Request`, las goroutines no se cancelan.

### [MITIGACIÓN F-10 EXTENDIDA]

1. **Verificar que el HTTP client usa el context**:
   ```go
   // En el engine.Generate():
   req, _ := http.NewRequestWithContext(ctx, "POST", anthropicURL, body)
   // NOT: http.NewRequest(...) — esto ignora el context
   ```

2. **WaitGroup para garantizar solo una goroutine activa**:
   ```go
   type AgentBuilderState struct {
       // ...
       GenerationCancel func()
       GenerationDone   chan struct{}  // NUEVO
   }
   
   func (m Model) startGeneration() (tea.Model, tea.Cmd) {
       // Cancelar generación anterior y esperar que termine
       if m.AgentBuilder.GenerationCancel != nil {
           m.AgentBuilder.GenerationCancel()
           if m.AgentBuilder.GenerationDone != nil {
               select {
               case <-m.AgentBuilder.GenerationDone:
               case <-time.After(2 * time.Second): // no esperar más de 2s
               }
           }
       }
       
       done := make(chan struct{})
       ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
       m.AgentBuilder.GenerationCancel = cancel
       m.AgentBuilder.GenerationDone = done
       
       return m, func() tea.Msg {
           defer close(done)
           defer cancel()
           // ... generate ...
       }
   }
   ```

---

## F-17 — 🟠 ALTO (NUEVO): MergeJSONFile Silencia Corrupción de Settings.json

### [COMPONENTE / VECTOR]
`internal/components/filemerge/merge.go` · `MergeJSONObjects()` · `.gemini/settings.json`, `.claude/settings.json`

### [TRACE DE FALLO — Backtracking]

`MergeJSONObjects` tiene este comportamiento con JSON malformado:

```go
func MergeJSONObjects(baseJSON []byte, overlayJSON []byte) ([]byte, error) {
    base, err := unmarshalJSONObject(baseJSON)
    if err != nil {
        // SILENCIA el error — usa empty map como base
        base = map[string]any{}
    }
    // ...
}
```

**El error de parseo del `baseJSON` es silenciado intencionalmente** (comentario en código: "safe and far preferable to aborting the whole install").

Pero esta "safety" tiene un efecto secundario grave:

```
Escenario: Gemini CLI con settings.json customizado por el usuario:
  ~/.gemini/settings.json:
    {
      "theme": "dark",
      "fontSize": 16,
      "customPrompt": "Always be concise",
      "mcpServers": { ... user MCPs ... }
    }

El usuario edita manualmente el settings.json y comete un typo:
    {
      "theme": "dark",
      "fontSize": 16,
      "customPrompt": "Always be concise"
      "mcpServers": { ... }  // ← falta coma aquí
    }

architect-ai sync ejecuta:
  → osReadFile(settingsPath) → lee el JSON con typo
  → MergeJSONObjects(malformedBase, overlay) 
  → unmarshalJSONObject falla con parse error
  → base = map[string]any{} ← TODO el settings del usuario BORRADO
  → merged = overlay solo (solo los MCPs de architect-ai)
  → WriteFileAtomic escribe el merged al settings.json
  
Usuario ahora tiene:
  ~/.gemini/settings.json:
    {
      "mcpServers": { architect-ai MCPs }
    }
  
  "theme": "dark", "fontSize": 16, "customPrompt": todos PERDIDOS
  Sin mensaje de error ni advertencia
```

### [IMPACTO]

- Pérdida silenciosa de configuración de usuario en cada `architect-ai sync` si el settings.json tiene algún error de formato.
- No hay backup automático antes del merge (el backup del snapshot existe pero no se invoca en todos los paths de merge).
- El usuario no recibe ningún feedback del error.

### [MITIGACIÓN]

1. **Cambiar el comportamiento de silencio a advertencia**:
   ```go
   func MergeJSONObjects(baseJSON []byte, overlayJSON []byte) ([]byte, error) {
       base, err := unmarshalJSONObject(baseJSON)
       if err != nil {
           // Log error and abort merge — don't silently lose user config
           return nil, fmt.Errorf("base JSON malformed: %w — fix manually before running sync", err)
       }
       // ...
   }
   ```

2. **Agregar backup ANTES del merge** en `mergeJSONFile`:
   ```go
   func mergeJSONFile(path string, overlay []byte) (filemerge.WriteResult, error) {
       // Backup antes de cualquier escritura
       if err := backupFile(path); err != nil {
           log.Printf("WARN: could not backup %s: %v", path, err)
       }
       baseJSON, _ := osReadFile(path)
       // ...
   }
   ```

3. **Validar JSON antes de sync** en `architect-ai sync`:
   ```go
   if err := validateJSONFile(settingsPath); err != nil {
       return fmt.Errorf("settings.json invalid JSON at %s: %w\nFix with: architect-ai validate-config", settingsPath, err)
   }
   ```

---

## F-18 — 🟡 MEDIO (NUEVO): Resize No Propaga a Sub-componentes TUI

### [COMPONENTE / VECTOR]
`internal/tui/model.go` · `WindowSizeMsg` handler · Screens con content scrollable

### [TRACE DE FALLO]

El handler de resize:
```go
case tea.WindowSizeMsg:
    m.Width = msg.Width
    m.Height = msg.Height
    return m, nil
```

`m.Width` y `m.Height` son actualizados. Pero estos valores solo son usados en `ScreenAgentBuilderPreview`:

```go
case ScreenAgentBuilderPreview:
    return screens.RenderABPreview(
        m.AgentBuilder.Generated, targets,
        m.AgentBuilder.PreviewScroll, m.Height,  // ← usa m.Height
        m.Cursor, ...
    )
```

**El resto de screens** (ScreenInstalling, ScreenComplete, ScreenBackups, ScreenRestoreConfirm, etc.) reciben parámetros de rendering sin `m.Width` ni `m.Height`. Sus layouts son estáticos.

**Escenario problemático**:
```
Usuario abre TUI en ventana de 80×24
→ ScreenBackups renderiza lista de backups (estática 80 cols)

Usuario redimensiona a 40×24 (ventana pequeña, laptop)
→ WindowSizeMsg: m.Width=40, m.Height=24
→ ScreenBackups sigue renderizando con layout de 80 cols
→ Output truncado, columnas solapadas

Usuario redimensiona a 200×50 (monitor grande)
→ ScreenBackups sigue en layout de 80 cols
→ Mucho espacio en blanco desperdiciado
```

### [IMPACTO]

UX degradada en terminales no-estándar. Puede causar wrapping de texto incorrecto en pantallas estrechas, especialmente en `ScreenDependencyTree` donde la representación árbol puede solaparse.

### [MITIGACIÓN]

Propagar `m.Width` y `m.Height` a todos los renderers:

```go
// Definir una interfaz de rendering con dimensiones
type RenderContext struct {
    Width  int
    Height int
}

// En View():
ctx := screens.RenderContext{Width: m.Width, Height: m.Height}
case ScreenBackups:
    return screens.RenderBackups(ctx, m.Backups, m.Cursor, m.BackupScroll, m.PinErr)
```

Y en cada screen renderer, usar `ctx.Width` para calcular el ancho de columnas y truncamiento de texto.

---

## F-19 — 🟡 MEDIO (NUEVO): Validador de Cycle Detection en State.yaml — Falso Negativo

### [COMPONENTE / VECTOR]
`internal/components/openspec/state.go` · `detectCycle()` · Algoritmo de Kahn

### [TRACE DE FALLO]

`detectCycle()` implementa el algoritmo de Kahn. Pero hay un caso no cubierto:

```go
func detectCycle(phases map[string]*Phase) error {
    inDeg := map[string]int{}
    for name, ph := range phases {
        inDeg[name] = len(ph.DependsOn)
    }
    // ...
}
```

El algoritmo cuenta `len(ph.DependsOn)` como in-degree. Pero no valida que cada elemento de `ph.DependsOn` referencie una fase real. Si `DependsOn` contiene una fase inválida, el algoritmo calcula un in-degree incorrecto.

**Ejemplo**:
```yaml
phases:
  sdd-apply:
    depends_on: ["sdd-nonexistent"]  # fase que no existe
```

`inDeg["sdd-apply"] = 1` (tiene un dep)
La fase "sdd-nonexistent" no está en el mapa → nunca se pone en la queue
`sdd-apply` nunca es procesado → `visited < len(phases)` → `ErrCycle`!

**Falso positivo**: el validator reporta un ciclo cuando en realidad hay una referencia inválida. El error ErrCycle es confuso — el usuario recibe "phases graph has a cycle" cuando el problema es una fase referenciada que no existe.

Pero hay además el **check I9** que debería prevenir esto:
```go
// I9
for _, dep := range ph.DependsOn {
    if !inSet(dep, ValidPhases) {
        return fmt.Errorf("%w: phases.%s.depends_on=%q", ErrEnum, name, dep)
    }
    if _, ok := s.Phases[dep]; !ok {
        return fmt.Errorf("%w: %s -> %s", ErrDanglingDepends, name, dep)
    }
}
```

El I9 ya previene el caso de dangling deps con `ErrDanglingDepends`. Pero si I9 falla antes del cycle check, se retorna ErrDanglingDepends correctamente.

El problema real: **el orden de checks importa**. Si I9 se ejecuta ANTES del cycle check, el dangling dep es detectado con el error correcto. Si el orden cambia en el futuro, puede haber regresión.

### [MITIGACIÓN]

1. Añadir test explícito del orden de checks:
   ```go
   func TestValidateOrderDanglingBeforeCycle(t *testing.T) {
       s := makeState()
       s.Phases["sdd-apply"].DependsOn = []string{"nonexistent-phase"}
       err := Validate(s, "test")
       // Must be ErrDanglingDepends, NOT ErrCycle
       if !errors.Is(err, ErrDanglingDepends) {
           t.Errorf("got %v, want ErrDanglingDepends", err)
       }
   }
   ```

2. Hacer `detectCycle` más robusto con skip de deps inválidas:
   ```go
   func detectCycle(phases map[string]*Phase) error {
       inDeg := map[string]int{}
       for name, ph := range phases {
           validDeps := 0
           for _, dep := range ph.DependsOn {
               if _, ok := phases[dep]; ok {
                   validDeps++  // solo contar deps que existen
               }
           }
           inDeg[name] = validDeps
       }
       // ...
   }
   ```

---

## F-20 — 🟡 MEDIO (NUEVO): Skill Resolver — HAS_ODOO Detection con Falso Positivo/Negativo

### [COMPONENTE / VECTOR]
`internal/assets/_shared/skill-resolver.md` · Detection script · `HAS_ODOO` logic

### [TRACE DE FALLO]

El script de detección de Odoo:

```bash
HAS_ODOO=$([ -f "$(find . -name '__manifest__.py' -maxdepth 5 | head -1)\" ] && echo 1 || echo 0)
```

**Problema 1 — Subshell expansion**:

```bash
[ -f "$(find . -name '__manifest__.py' -maxdepth 5 | head -1)" ]
```

`find` puede retornar múltiples paths. `head -1` toma el primero. El resultado se expande en `$()` y se pasa a `[ -f "..." ]`. Si el path tiene espacios, la expansión falla:

```bash
# Si find retorna: "./my odoo module/__manifest__.py"
[ -f "./my odoo module/__manifest__.py" ]
# → bash lo interpreta como: [ -f "./my" ] odoo ...
# → error de sintaxis silencioso → HAS_ODOO=0 (falso negativo)
```

**Problema 2 — Falso positivo en repos multi-proyecto**:

Un monorepo puede tener:
```
apps/
  web-app/
  odoo-customizations/
    sale_custom/__manifest__.py
  golang-service/
```

`find . -name '__manifest__.py'` encuentra la customización de Odoo. El golang-service también dispara `HAS_ODOO=1`, aunque el task actual es de Go. El skill de Odoo se inyecta en tareas de Go irrelevantes.

**Problema 3 — maxdepth=5 puede ser insuficiente**:

En estructuras de directorio profundas:
```
client/
  projects/
    acme/
      infrastructure/
        odoo/
          addons/
            sale_custom/
              __manifest__.py  ← depth 8
```

El `maxdepth 5` no lo encuentra → HAS_ODOO=0 → las reglas de Odoo no se inyectan → sdd-apply puede romper patrones Odoo por falta de guía.

### [MITIGACIÓN]

```bash
# Detección más robusta con .atl/config.yaml como fuente de verdad
HAS_ODOO=0
if grep -q "^odoo_version:" .atl/config.yaml 2>/dev/null; then
    HAS_ODOO=1
elif find . -name '__manifest__.py' -maxdepth 10 -print -quit 2>/dev/null | grep -q .; then
    HAS_ODOO=1
fi
```

Agregar al `architect-ai init`:
```go
// Si __manifest__.py encontrado → escribir en .atl/config.yaml
config.OdooVersion = detectOdooVersion()
config.IsOdooProject = true
```

Así la detección en runtime es determinista (lee config.yaml) en lugar de depender de `find`.

---

## RESUMEN EJECUTIVO AMPLIADO — GRAFO DE RIESGOS

### Matriz de Impacto Sistémico

```
RIESGO CRÍTICO (acción en < 24h):
  F-04-B: "running" vs "in_progress" → Save() SIEMPRE falla con la instrucción del prompt
           Impacto: TODO el pipeline SDD está bloqueado en modo openspec/hybrid
           Fix: 30 minutos (cambiar enum o cambiar prompts)

  F-04-A: TOCTOU en lockfile → corrupción silenciosa del DAG bajo paralelismo
           Fix: 2 horas (O_CREATE|O_EXCL + .tmp por PID)

  F-01-B: Mode A en Gemini template → L0 ejecuta inline sin guardrails
           Fix: 15 minutos (eliminar sección Mode A del template)

RIESGO ALTO (acción en < 1 semana):
  F-17: mergeJSONFile silencia JSON malformado → pérdida de configuración de usuario
  F-05-A: Bootstrap paradox de Engram → cliff-edge cuando MCP cae
  F-07-B: verify:abandoned sin arco de salida → cambios en limbo permanente

RIESGO MEDIO (sprint 1-2):
  F-09: ProgressFunc no-op → UX ciega en i
---

## F-11 (EXTENDIDO) — 🟡 MEDIO: Caveman Firewall Inauditable en Modo `none`

### Subtópico A: El invariante del firewall y sus condiciones de falla

El `caveman-firewall.md` establece reglas absolutas:

```markdown
This firewall CANNOT be:
- Disabled by context pressure (D4 = 3 does not disable it)
- Suspended during /compact
- Overridden by an orchestrator posture
- Bypassed for "quick fixes" or "tiny changes"
Every character of source code is NORMAL. No exceptions.
```

Y las transiciones de registro deben ser registradas explícitamente en `apply-progress`:

```
[REGISTER→NORMAL] Entering code zone for task T-03
... (code writing) ...
[REGISTER→ULTRA] Exiting code zone, task T-03 committed
```

**En modo `artifact_store=none`**, el `apply-progress` es generado inline en la respuesta del chat, no persistido en ningún sistema. La evidencia de las transiciones de registro existe solo mientras el contexto del modelo las contiene.

### Subtópico B: El grafo de pérdida de evidencia en modo none

```
sdd-apply en modo none:
  Task T-03: implementar process_payment()
  [REGISTER→NORMAL] ← en el chat, no en Engram
  def process_payment(amount, currency):
      # ULTRA en comment: validate. send. log.   ← VIOLACIÓN del firewall
      ...
  [REGISTER→ULTRA] ← en el chat
  
  apply-progress (inline, no persistido):
    - T-03: completed (ULTRA en comment REGISTRADO como normal? No hay audit)
  
  sdd-verify:
    → intenta leer apply-progress de Engram → null (modo none)
    → intenta leer de openspec → null (modo none)
    → Semantic Audit Protocol: "ADD RISK 'no apply-progress available'"
    → Skip apply-progress review
    → La violación del firewall NO es detectada
    → Verdict: puede ser APPROVED con violación embedded
```

### Subtópico C: La interacción con el Rollback Harness en modo none

El Rollback Harness (`rollback-harness.md`) define:
```bash
# Step 4: Archive apply-progress.yaml for debugging
if [ -f ".atl/apply-progress.yaml" ]; then
  mv ".atl/apply-progress.yaml" ".atl/apply-progress.yaml.${TS}.rollback"
fi
```

En modo `none`, el archivo `.atl/apply-progress.yaml` no existe. El rollback no archiva ningún artifact. El post-mortem de un apply fallido en modo none es imposible porque no hay evidencia persistida.

### Subtópico D: El hotfix protocol y el modo none

El `sdd-hotfix-protocol.md` define un ciclo comprimido de 5 pasos que incluye `verify-lite: tests for affected files only; semantic audit SKIPPED`. El hotfix puede ejecutarse en modo none (urgencia).

**Combinación peligrosa**: hotfix + none + --skip-ai = cero trazabilidad, cero verificación, cero audit trail. El código puede ir a producción sin ninguna evidencia de haber pasado por el pipeline.

### [MITIGACIÓN F-11 EXTENDIDA]

1. **Marcar modo `none` como `AUDIT_DEGRADED` en el result contract**:
   ```json
   {
     "status": "completed",
     "audit_mode": "degraded",
     "audit_limitations": [
       "caveman_firewall: not verifiable (no apply-progress artifact)",
       "tdd_compliance: not verifiable (no artifact store)",
       "rollback: not available (no artifact to restore)"
     ]
   }
   ```
   El sdd-archive DEBE incluir estas limitaciones en el Deviation Log.

2. **Forzar modo degraded en sdd-verify cuando artifact_store=none**:
   ```markdown
   IF artifact_store == "none":
     → verdict CANNOT be "APPROVED"
     → minimum verdict: "CONDITIONALLY APPROVED"
     → add to risks: "No artifact store — caveman firewall compliance unverifiable"
   ```

3. **Proteger el hotfix+none+--skip-ai**: el sdd-hotfix debe requerir al menos un artifact store mínimo cuando hay cambios de código:
   ```markdown
   IF sdd-hotfix AND artifact_store == "none" AND code_files_changed > 0:
     → WARN: "Hotfix without artifact store leaves no audit trail for code changes"
     → Require explicit: "I understand the risks, proceed"
   ```

---

## F-12 (EXTENDIDO) — 🟡 MEDIO: Context Bloat Garantizado con Odoo Overlay

### Subtópico A: La cadena de inyección acumulativa

El odoo-overlay-routing.md §7 define 4 archivos inyectados en cada sub-agent:

```
coding-style.md    → compact rules    (~1.5 KB)
security.md        → compact rules    (~2 KB)
CAUTION_POLICY.md  → FULL             (~6-10 KB)
cudio-git.md       → compact rules    (~1 KB)
```

Total por llamada: **~12 KB** de reglas Odoo.

El Skill Resolver también inyecta el foundation block (Tier 1):

```
foundation block   → always first     (~8 KB estimado)
```

Y los patrones Odoo correspondientes a la versión:

```
patterns-19/SKILL.md → compact rules  (~3 KB)
patterns-agnostic/SKILL.md → compact  (~2 KB)
```

Total de inyección base por sub-agent en proyecto Odoo: **~25 KB**.

El límite del Skill Resolver es `SKILLS_TO_INJECT[:4]` (4 bloques máximo). Pero esto no limita el tamaño de cada bloque — solo el número. El CAUTION_POLICY full por sí solo puede exceder el total de 4 bloques compactos.

### Subtópico B: Cálculo de tokens por pipeline SDD completo

```
Pipeline SDD completo en proyecto Odoo:
  sdd-init      → 1 sub-agent × 25 KB = 25 KB
  sdd-explore   → 2 sub-agents × 25 KB = 50 KB (+ odoo-context-gatherer)
  sdd-propose   → 1 sub-agent × 25 KB = 25 KB
  sdd-spec      → 1 sub-agent × 25 KB = 25 KB
  sdd-design    → 1 sub-agent × 25 KB = 25 KB
  sdd-tasks     → 1 sub-agent × 25 KB = 25 KB
  sdd-apply     → 3 batches × 25 KB  = 75 KB
  sdd-verify    → 2 sub-agents × 25 KB = 50 KB (+ odoo-code-reviewer)
  sdd-archive   → 1 sub-agent × 25 KB = 25 KB
  ─────────────────────────────────────────────
  Total overhead de inyección: ~325 KB (~243.000 tokens)
```

A $3/M tokens (Claude Sonnet), esto representa **~$0.73 solo en overhead de inyección** por pipeline. Para un equipo ejecutando 10 pipelines por sprint, son ~$7.30/sprint únicamente en tokens de inyección de reglas repetidas.

### Subtópico C: El efecto en D4 y el modo de razonamiento

Con 25 KB de inyección base + el contexto de la tarea + historial de conversación, el D4 de los sub-agents de Odoo empieza en **valores elevados artificialmente**:

```
Sub-agent sdd-apply en turno 5 del pipeline:
  Inyección base: 25 KB
  Historial de conversación (turnos anteriores): 40 KB
  apply-progress del batch anterior: 15 KB
  ─────────────────────────
  Total contexto inicial: 80 KB → D4 = 2 (50-100KB range)
  → Mode 2-ERR activado automáticamente (D4 >= 2 = "High")
  → El agente opera bajo postura Tactical desde el primer turno
  → Si hay un error en el apply → D3 += 1 → Mode 3-Diagnostic
```

El sistema opera consistentemente en modos de mayor overhead cognitivo por el contexto inflado, no por la complejidad real de la tarea.

### [MITIGACIÓN F-12 EXTENDIDA]

1. **Convertir CAUTION_POLICY a semantic memory** (ya en v1). Implementación concreta:
   ```markdown
   # odoo-overlay-routing.md §7 REVISADO
   coding-style.md    → compact rules    (inject always)
   security.md        → compact rules    (inject always)
   CAUTION_POLICY.md  → summary-only    (inject: first 200 chars + pointer)
   cudio-git.md       → compact rules    (inject always if git ops in task)
   
   Full CAUTION_POLICY: mem_search("knowledge/odoo/caution-policy")
   Load full only when: task involves write ops OR D1 >= 2
   ```

2. **Cap de inyección total en Odoo**: el Skill Resolver debe tener un `MAX_INJECTION_KB` configurable:
   ```python
   MAX_INJECTION_KB = 15  # configurable in .atl/config.yaml
   if total_injection_size > MAX_INJECTION_KB * 1024:
       # Trim lower-priority skills
       log_warn(f"Injection trimmed to {MAX_INJECTION_KB}KB")
   ```

3. **Lazy-load de patrones por versión**: no inyectar `patterns-{version}` en todas las fases, solo en sdd-apply y sdd-verify donde se escribe código real:
   ```
   LAZY_LOAD_PHASES = ["sdd-apply", "sdd-verify"]
   ALWAYS_LOAD_PHASES = ["sdd-explore", "sdd-spec", "sdd-design"]
   
   if current_phase in ALWAYS_LOAD_PHASES:
       inject = [foundation, security_compact, caution_summary]
   elif current_phase in LAZY_LOAD_PHASES:
       inject = [foundation, security_compact, caution_full, patterns_full]
   ```

---

## F-13 (EXTENDIDO) — 🟡 MEDIO: Loop de Compactación sin Cooldown

### Subtópico A: Los 6 trigger conditions y sus interacciones

El context-guardian SKILL.md v3.0 define 6 condiciones de auto-trigger:

```
1. char_count(context_history) >= 100_000
2. Sub-agent skill_resolution: none in last 2 turns
3. D4 >= 2 in current reasoning evaluation
4. 3+ file reads in same context window without compaction
5. User says "compact", "reset context", "what's my state"
6. attempt_count >= 2 for current phase
```

**La condición 1 y la condición 3 pueden coexistir con la condición 4** en un pipeline Odoo normal:

```
Post-compactación (turno N+1):
  char_count = 15K (acabado de compactar)
  D4 = 0 (compactado)
  
  L1a carga el skill-registry (Tier 1 mandatory):
  char_count += 20K → 35K
  
  L1a carga odoo overlay (Odoo project):
  char_count += 12K → 47K
  
  L1a carga el protocolo de fase (Progressive Phase Loading):
  char_count += 15K → 62K
  
  sdd-apply lee 3 archivos del proyecto:
  char_count += 18K → 80K
  → Trigger 4 FIRES: "3+ file reads without compaction"
  → context-guardian invocado de nuevo
  → Pero acabamos de compactar!
```

### Subtópico B: El costo de la compactación sin cooldown

Cada invocación de context-guardian tiene un costo:

1. `mem_save("context-pack/...")` → 1 Engram write call
2. Generar el Context Pack (resumen 2000 tokens) → ~500 tokens de output
3. `/compact` o `/compress` → invocación al modelo
4. `mem_context()` post-compactación → 1 Engram read call

Estimado: **~1000-2000 tokens por ciclo de compactación**. Sin cooldown, en el escenario anterior la compactación se dispara en el turno N+1, luego en N+2 (el mismo ciclo de carga de skills se repite), etc.

### Subtópico C: La condición 6 como amplificador

`attempt_count >= 2` también dispara context-guardian. En un pipeline con fallos repetidos (F-07: sdd-verify loop), el circuit breaker incrementa attempt_count. Esto:

```
attempt_count = 2 → context-guardian trigger 6
→ compactación → context-guardian invocado
→ post-compactación: sdd-verify intenta de nuevo
→ falla → attempt_count = 3 → CIRCUIT BREAKER
→ Pero también dispara context-guardian trigger 6 de nuevo (attempt_count=3 >= 2)
→ compactación durante el trip del circuit breaker
→ El estado del circuit breaker puede perderse en la compactación
→ Circuit breaker reset involuntario
```

### Subtópico D: Los anti-patterns documentados vs la realidad

El SKILL.md documenta como anti-pattern:
```
- Rebuilding pack from scratch every turn when only one fact changed
  (use mem_update on existing pack)
```

Pero el protocolo de compactación no especifica cuándo usar `mem_update` vs `mem_save`. Un agente que sigue el "Compaction Protocol" siempre hace `mem_save` (nuevo pack), no `mem_update`. El anti-pattern está documentado pero el protocolo positivo no incluye la condición para usar `mem_update`.

### [MITIGACIÓN F-13 EXTENDIDA]

1. **Implementar cooldown de compactación en el context-guardian SKILL.md**:
   ```markdown
   ## Cooldown Rule
   After any compaction event, suppress all auto-triggers for the next 3 delegations.
   Track via protected_fact: `last_compaction_turn: {N}`.
   If `current_turn - last_compaction_turn < 3`: skip trigger check.
   Exception: User explicitly requests "compact now" (trigger 5 always fires).
   ```

2. **Umbral dinámico post-compactación**:
   ```markdown
   ## Dynamic Threshold
   Base threshold: 100_000 chars
   After compaction: threshold = base × 1.5 (150_000) for next 5 turns
   After 5 turns: return to base threshold
   This prevents immediate re-trigger during skill/protocol reloading.
   ```

3. **`mem_update` para packs existentes**:
   ```markdown
   ## Pack Update Protocol
   IF protected_fact "context-pack-key" exists AND pack age < 30min:
     → mem_update(existing-pack-key, {only changed sections})
     → Do NOT create new pack
   IF pack age >= 30min OR no existing pack:
     → mem_save new pack (full rebuild)
   ```

---

## F-14 (EXTENDIDO) — 🟡 MEDIO: Exposición de Credenciales Odoo

### Subtópico A: La implementación de `ensureGitignored` y sus límites

```go
func ensureGitignored(path, pattern string) {
    data, _ := os.ReadFile(path)
    if strings.Contains(string(data), pattern) { return }
    // Si ya contiene ".env.mcp" → no agrega nada → asume que está ignorado
    
    f, _ := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    defer f.Close()
    fmt.Fprintf(f, "\n# MCP secrets\n%s\n", pattern)
}
```

El `strings.Contains(string(data), pattern)` verifica si `.env.mcp` aparece en el `.gitignore`. Pero el patrón puede estar presente como **parte de otro patrón** sin cubrir el archivo:

```gitignore
# Caso 1: patrón negativo que EXCLUYE el .env.mcp del gitignore
!.env.mcp    ← strings.Contains encuentra ".env.mcp" → no agrega nada
              ← pero "!" significa "do NOT ignore" → el archivo SÍ se comittea

# Caso 2: patrón comentado
# .env.mcp  ← strings.Contains encuentra ".env.mcp" → no agrega nada
              ← el comentario no aplica gitignore → el archivo SÍ se comittea

# Caso 3: cobertura parcial por extensión que no aplica
.env         ← no contiene ".env.mcp" → el check AGREGA .env.mcp
              ← ahora tenemos .env Y .env.mcp → doble coverage → ok
              
# Caso 4: negación posterior
.env.mcp     ← strings.Contains → no agrega nada → OK
!.env.mcp    ← línea posterior → EXCLUYE de nuevo → el archivo SÍ se comittea
```

Los casos 1, 2 y 4 producen **falso negativo**: `ensureGitignored` cree que el archivo está ignorado cuando en realidad puede ser commiteado.

### Subtópico B: El path de las credenciales en memoria del proceso

Cuando Gemini o Antigravity usan el MCP de Odoo, la contraseña se pasa via variable de entorno:

```go
// generator.go línea 3767
"ODOO_PASSWORD": "${ODOO_PASSWORD}",  // para Antigravity
```

La contraseña está en el environment del proceso shell que lanza el LLM. En Linux, el environment de un proceso es visible en `/proc/{PID}/environ` para el mismo usuario. Si hay otros procesos del mismo usuario, pueden leer la contraseña del ambiente del proceso padre.

En el MCP server, el binding es:

```python
# mcp-server-odoo — hypothetical implementation
password = os.environ.get("ODOO_PASSWORD", "")
```

La contraseña está en el ambiente del MCP server process durante toda la sesión.

### Subtópico C: El sdd-verify code-reviewer y la detección de secrets

El `sdd-verify.md` deterministic check `A1`:
```
BLOCK if diff contains:
- Hardcoded API keys: patterns (api_key|secret|token|password|private_key)\\s*=\\s*["'][^"']{8,}["']
```

Este pattern detecta hardcoded secrets en el diff. Pero:

1. El `.env.mcp` **no está en el diff** si está correctamente gitignored.
2. Si por el fallo F-14-A el `.env.mcp` SÍ está en el diff (gitignore falló), el pattern detecta `ODOO_PASSWORD=real_password` → BLOCK → correcto.

Pero hay un gap: el pattern no detecta:
```
ODOO_PASSWORD = os.getenv("ODOO_PASSWORD")  # referencia a env var, no valor hardcoded
```

Esto es correcto — la referencia a env var es safe. El secret real está en `.env.mcp` (que no debe commitearse).

El gap real: **no hay ninguna verificación de que `.env.mcp` esté correctamente gitignored antes del commit**. El GGA pre-commit hook verifica muchas cosas pero no explícitamente la efectividad del gitignore para `.env.mcp`.

### [MITIGACIÓN F-14 EXTENDIDA]

1. **Reemplazar `strings.Contains` con validación real de gitignore rules**:
   ```go
   func ensureGitignored(path, pattern string) error {
       data, _ := os.ReadFile(path)
       lines := strings.Split(string(data), "\n")
       
       for _, line := range lines {
           trimmed := strings.TrimSpace(line)
           // Skip comments and empty lines
           if trimmed == "" || strings.HasPrefix(trimmed, "#") {
               continue
           }
           // Skip negation patterns
           if strings.HasPrefix(trimmed, "!") {
               continue
           }
           // Check if this line matches our pattern
           if gitignoreMatches(trimmed, pattern) {
               return nil // correctly ignored
           }
       }
       
       // Pattern not found or only in comment/negation — add it
       f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
       if err != nil { return err }
       defer f.Close()
       _, err = fmt.Fprintf(f, "\n# MCP secrets — do not commit\n%s\n", pattern)
       return err
   }
   ```

2. **Agregar verificación de `.env.mcp` al GGA pre-commit**:
   ```bash
   # .atl/scripts/gga-pre-commit.sh — nueva sección
   # Check .env.mcp is not staged
   if git diff --cached --name-only | grep -q "\.env\.mcp$"; then
     echo "ERROR: .env.mcp is staged for commit. This file contains secrets."
     echo "Remove with: git restore --staged .env.mcp"
     exit 1
   fi
   ```

3. **Usar un secreto gestor en lugar de .env.mcp para credenciales de producción**: documentar en `docs/security.md` que para entornos de producción, `ODOO_PASSWORD` debe provenir del gestor de secretos del sistema (keychain en macOS, `secret-tool` en Linux, o el vault del CI/CD), no de `.env.mcp`.

---

## F-15 (EXTENDIDO) — 🟢 BAJO: ScreenUninstallConfirm sin Ruta Forward en linearRoutes

### Subtópico A: La arquitectura del router y el invariante roto

El `linearRoutes` tiene un invariante implícito: toda screen que puede recibir la tecla Enter (confirmación) para avanzar DEBE tener una ruta Forward en el mapa si el avance es vía router.

Inspeccionando el código:
```go
// ScreenUninstallConfirm está AUSENTE en linearRoutes
// Las únicas entradas de Uninstall son:
ScreenUninstallMode:       {Backward: ScreenWelcome},
ScreenUninstall:           {Backward: ScreenUninstallMode},
ScreenUninstallComponents: {Backward: ScreenUninstall},
ScreenUninstallProfiles:   {Backward: ScreenUninstallComponents},
ScreenUninstallResult:     {Backward: ScreenWelcome},
// ScreenUninstallConfirm: MISSING
```

El Enter handler de `ScreenUninstallConfirm` usa `m.setScreen()` directamente:
```go
case ScreenUninstallConfirm:
    if m.Cursor == 0 {
        m.OperationRunning = true
        m.OperationMode = "uninstall"
        return m, tea.Batch(tickCmd(), m.startUninstall())
    }
    // cancel: go back (direct setScreen)
```

Y el completion usa `handlePipelineDone` que llama `setScreen(ScreenUninstallResult)` directamente. El router **nunca** es usado para `ScreenUninstallConfirm → Forward`. Por tanto la ausencia no causa un bug activo.

### Subtópico B: El riesgo de ruta genérica

El código tiene un handler genérico de Enter:
```go
// En handleEnterKey (hipotético):
if next, ok := NextScreen(m.Screen); ok {
    m.setScreen(next)
    return m, nil
}
// Else: screen-specific handling
```

Si en algún refactor se añade una tecla genérica de "confirmar" que llama `NextScreen`, `ScreenUninstallConfirm` retornaría `ScreenUnknown, false` → navega a pantalla 0 (Welcome) en lugar de iniciar el uninstall. El bug es latente, no activo.

### Subtópico C: El test coverage de ScreenUninstallConfirm

Los tests en el batch 7 muestran:
```go
// Líneas 8573-8720: múltiples tests para ScreenUninstallConfirm
m.Screen = ScreenUninstallConfirm
// ... tests de navegación backward, Enter con cursor 0 y 1
```

Los tests cubren el backward correcto. Pero **no hay test** que verifique que `NextScreen(ScreenUninstallConfirm)` retorna `ScreenUnknown, false` explícitamente. Si alguien añade la ruta `ScreenUninstallConfirm: {Forward: ScreenWelcome}` por error, los tests no lo detectarían como bug (irían a Welcome en lugar de iniciar uninstall).

### [MITIGACIÓN F-15 EXTENDIDA]

1. **Añadir comentario explicativo en linearRoutes**:
   ```go
   // ScreenUninstallConfirm intentionally omitted from linearRoutes:
   // Forward navigation is handled by startUninstall() → handlePipelineDone → ScreenUninstallResult
   // Backward navigation is dynamic (mode-dependent) — handled in navigateBackward()
   ```

2. **Agregar test de invariante**:
   ```go
   func TestScreenUninstallConfirmHasNoLinearForward(t *testing.T) {
       _, ok := NextScreen(ScreenUninstallConfirm)
       if ok {
           t.Error("ScreenUninstallConfirm should not have a linear forward route " +
                   "— forward navigation is via startUninstall(), not the router")
       }
   }
   ```

---

## F-16 (EXTENDIDO) — 🟢 BAJO: Colisión de topic_key en ResearchTopicKey

### Subtópico A: El algoritmo completo y sus propiedades de unicidad

```go
func ResearchTopicKey(tool, query string) string {
    cleaned := strings.ToLower(strings.TrimSpace(query))
    cleaned = nonAlnum.ReplaceAllString(cleaned, "-")        // reemplaza no-alfanum con "-"
    cleaned = collapseDash.ReplaceAllString(cleaned, "-")    // colapsa guiones múltiples
    cleaned = strings.Trim(cleaned, "-")                     // quita guiones extremos
    if cleaned == "" { cleaned = "query" }
    if len(cleaned) > 50 {
        cleaned = strings.Trim(cleaned[:50], "-")            // trunca a 50 chars
    }
    return fmt.Sprintf("research/%s/%s-len%d", tool, cleaned, len(query))
}
```

**Propiedades**:
- `len(query)` es la longitud de la query ORIGINAL, antes de slugify.
- El slug está truncado a 50 chars del string procesado (lower + nonAlnum).
- El sufijo `-len{N}` usa la longitud original para diferenciación.

**Test de colisión con la misma longitud original**:

```
Query A: "odoo sale order workflow configuration ent"  (len=42)
  slug: "odoo-sale-order-workflow-configuration-ent"  (< 50: no trunca)
  key:  "research/context7/odoo-sale-order-workflow-configuration-ent-len42"

Query B: "odoo sale order workflow configuration ENT"  (len=42, solo mayúsculas)
  after toLower: "odoo sale order workflow configuration ent"
  slug: "odoo-sale-order-workflow-configuration-ent"
  key:  "research/context7/odoo-sale-order-workflow-configuration-ent-len42"
  ← COLISIÓN EXACTA con Query A
```

**Las queries son semánticamente diferentes** (el usuario puede haber capitalizado por razón). Ambas producen la misma key → upsert silencioso.

### Subtópico B: El caso más frecuente — queries con sufijo distinto pero mismo slug inicial

```
Query A: "how to configure Odoo sale order workflow"   (len=42)
Query B: "How to configure Odoo sale order workflow?"   (len=43)
  
  A: slug = "how-to-configure-odoo-sale-order-workflow"  key ends: -len42
  B: slug = "how-to-configure-odoo-sale-order-workflow"  key ends: -len43
  ← NO colisión (len difiere por el "?")
```

El mecanismo len-suffix funciona correctamente cuando las queries difieren por longitud. El problema es cuando son idénticas en contenido pero difieren solo en capitalización (o trailing spaces trimmeados).

### Subtópico C: El impacto en el Research Order de Odoo

Cuando `odoo-context-gatherer` investiga el mismo módulo en dos sesiones distintas (con queries semánticamente iguales pero con capitalización diferente), la segunda sesión sobreescribe el resultado de la primera en Engram:

```
Sesión 1: mem_save("research/context7/odoo-sale-order-len35", {resultado correcto})
Sesión 2: misma query, difiere en case → misma key → mem_save sobreescribe
  Si el resultado de sesión 2 es peor (menos detallado, stale) → regresión invisible
```

El protocolo de Engram no tiene versioning ni historial de sobreescrituras. La pérdida es silenciosa.

### Subtópico D: El test coverage actual — falta el caso de colisión

Los tests en `TestResearchTopicKey` cubren:
- Texto simple, multi-word, espacios múltiples, query muy larga, query vacía.

**No tienen test para**:
- Dos queries con mismo slug pero diferente capitalización → misma key → colisión esperada o inesperada.
- Unicode en queries (emojis, acentos → colapsan a "-").

### [MITIGACIÓN F-16 EXTENDIDA]

1. **Agregar hash de contenido completo** (ya en v1). Implementación completa:
   ```go
   import "crypto/sha256"
   
   func ResearchTopicKey(tool, query string) string {
       cleaned := strings.ToLower(strings.TrimSpace(query))
       cleaned = nonAlnum.ReplaceAllString(cleaned, "-")
       cleaned = collapseDash.ReplaceAllString(cleaned, "-")
       cleaned = strings.Trim(cleaned, "-")
       if cleaned == "" { cleaned = "query" }
       if len(cleaned) > 50 {
           cleaned = strings.Trim(cleaned[:50], "-")
       }
       // 8-char hash of original query for collision-free disambiguation
       h := sha256.Sum256([]byte(query))
       return fmt.Sprintf("research/%s/%s-%x", tool, cleaned, h[:4])
   }
   ```
   El sufijo `-{8hex chars}` reemplaza `-len{N}`. Ejemplo: `research/context7/odoo-sale-order-workflow-a1b2c3d4`.

2. **Agregar tests de colisión explícitos**:
   ```go
   func TestResearchTopicKey_NoCollision(t *testing.T) {
       queries := []string{
           "odoo sale order workflow",
           "Odoo Sale Order Workflow",
           "ODOO SALE ORDER WORKFLOW",
           "odoo sale order workflow ",  // trailing space
       }
       keys := make(map[string]string)
       for _, q := range queries {
           k := ResearchTopicKey("context7", q)
           if prev, exists := keys[k]; exists {
               t.Errorf("collision: %q and %q produce same key %q", prev, q, k)
           }
           keys[k] = q
       }
   }
   ```

---

## APÉNDICE A: PATCHES CONCRETOS PARA LOS 4 CRÍTICOS

### PATCH A-1: Fix enum `running` → `in_progress` (F-04-B)

**Archivo**: `internal/assets/_shared/phase-dag-enforcement.md`

```diff
-  - Phase start → update status to `running`
+  - Phase start → update status to `in_progress`
   - Phase success → update status to `completed`
   - Phase failure → update status to `failed`
```

**Archivo**: `internal/assets/_shared/rollback-harness.md`

```diff
-content = re.sub(
-    r'(  sdd-apply:.*?status: )\"(running|failed)\"',
+content = re.sub(
+    r'(  sdd-apply:.*?status: )\"(in_progress|failed)\"',
     r'\1\"pending\"',
     content, flags=re.DOTALL
 )
```

**Verificación** — buscar todos los archivos con "running" como valor de status:
```bash
rg '"running"' internal/assets/ | grep -v "PlanStatus\|StepStatus\|enum\|comment\|#"
```

---

### PATCH A-2: Fix TOCTOU en state.go Save() (F-04-A)

**Archivo**: `internal/components/openspec/state.go`

```diff
 func Save(path string, s *State) error {
     s.UpdatedAt = time.Now().UTC()
     parent := filepath.Base(filepath.Dir(path))
     if err := Validate(s, parent); err != nil {
         return err
     }
     out, err := yaml.Marshal(s)
     if err != nil {
         return fmt.Errorf("marshal: %w\", err)
     }
-    tmp := path + ".tmp"
+    // Use PID + nanoseconds to avoid concurrent-write collisions on .tmp
+    tmp := fmt.Sprintf("%s.%d.%d.tmp", path, os.Getpid(), time.Now().UnixNano())
     if err := os.WriteFile(tmp, out, 0o644); err != nil {
         return fmt.Errorf("write tmp: %w\", err)
     }
+    defer func() {
+        // Clean up tmp if rename failed
+        _ = os.Remove(tmp)
+    }()
     if f, err := os.OpenFile(tmp, os.O_RDWR, 0o644); err == nil {
         _ = f.Sync()
         _ = f.Close()
     }
     if err := os.Rename(tmp, path); err != nil {
         return fmt.Errorf("rename: %w\", err)
     }
     return nil
 }
```

---

### PATCH A-3: Eliminar Mode A del template Gemini (F-01)

**Archivo**: `internal/assets/gemini/GEMINI.md`

```diff
-## Mode A (Gemini inline — simple tasks)
-Use bash/read/write tools directly. Do NOT use run_subagent for simple operations.
-Simple tasks: git status, grep, cat single file, echo, pwd
-
-## Mode B (SDD Orchestrator)
+## Mode B — SDD Orchestrator
 run_subagent to sdd-orchestrator for all SDD intents.
 
-## Mode C (General Orchestrator)
+## Mode C — General Orchestrator (ALL non-SDD tasks)
+run_subagent to general-orchestrator for ALL other tasks.
+L0 NEVER uses bash/read/write directly. Zero exceptions.
+Even "git status" goes to general-orchestrator via run_subagent.
+
+## IMMUTABILITY RULE (non-negotiable)
+L0 is a pure router. L0 has no operational context.
+L0 MUST NOT execute any tool directly.
+Violation of this rule collapses the delegation architecture.
```

---

### PATCH A-4: Fix placeholder check en `architect-ai check all` (F-03)

**Archivo**: `internal/verify/checks.go`

```go
// NUEVO CHECK: verificar ausencia de placeholders en CLAUDE.md
{
    ID:          "claude-md-no-placeholders",
    Description: "CLAUDE.md contains no unresolved template placeholders",
    Severity:    SeverityCritical,
    Run: func(ctx context.Context, projectRoot string) error {
        claudePath := filepath.Join(projectRoot, "CLAUDE.md")
        data, err := os.ReadFile(claudePath)
        if err != nil {
            if os.IsNotExist(err) {
                return nil // no CLAUDE.md is a different check
            }
            return err
        }
        s := string(data)
        
        placeholders := []string{
            "{content from",
            "{L0_HASH}",
            "{L1A_HASH}",
            "{L1B_HASH}",
            "{CONTENT_HASH}",
            "{INJECTED_MODE}",
            "{D1}", "{D2}", "{D3}", "{D4}",
        }
        for _, p := range placeholders {
            if strings.Contains(s, p) {
                return fmt.Errorf(
                    "CLAUDE.md contains unresolved placeholder %q — "+
                    "run: architect-ai build\n"+
                    "All agent prompts must be materialized before use.",
                    p,
                )
            }
        }
        // Size sanity: a real CLAUDE.md is > 20KB
        if len(data) < 20_000 {
            return fmt.Errorf(
                "CLAUDE.md is only %d bytes — likely not built. "+
                "Run: architect-ai build",
                len(data),
            )
        }
        return nil
    },
    FixHint: "architect-ai build",
},
```

---

## APÉNDICE B: MÉTRICAS DE CALIDAD DEL SISTEMA

### Test Coverage Gap Analysis

| Componente | Cobertura estimada | Gap crítico |
|---|---|---|
| openspec/state.go | Alta (Validate, Save, detectCycle) | Sin test de concurrencia (`-race`) |
| engramkeys/keys.go | Alta (happy path) | Sin test de colisión por case |
| filemerge/merge.go | Alta (adversarial tests) | Sin test de recovery post-silenced-error |
| mcp/secrets.go | Baja (sin tests visibles) | `ensureGitignored` falsos negativos sin coverage |
| tui/model.go | Media (navegación) | Sin test de goroutine leak en cancelación |
| routing/classifier | Alta (CheckMandatoryTriggers) | Sin test de self-report bypass |

### Superficie de Ataque — Resumen

```
Vectores de ataque identificados:
  1. Prompt injection via Mode A (L0 ejecuta inline) → F-01
  2. Context contamination (L1a→L1b) → F-02
  3. State corruption via concurrent writes → F-04
  4. Memory loss via compaction without checkpoint → F-05
  5. YOLO mode + degraded context → data mutation → F-08
  6. Secret exposure via broken gitignore → F-14

Superficies sin hardening:
  - El guard check es self-reported (sin observador externo)
  - ODOO_YOLO no tiene circuit breaker por D-score
  - La isolation L1a/L1b es solo declarativa en Claude Code/VSCode
```

---

*Extensión completa del análisis adversarial — 1.600+ líneas de reporte forense, 20 hallazgos, 4 patches concretos, 6 tests propuestos, grafo completo de dependencias causales entre fallos.*
*Revisión recomendada: F-04-B (enum) como acción inmediata — desbloquea el SDD pipeline en modo openspec/hybrid.*
