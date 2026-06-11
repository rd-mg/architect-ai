# Fase 0: Orden de Ejecución y Mapa de Dependencias

**Objetivo:** Proporcionar el orden exacto de aplicación de las 13 fases de hardening, el mapa de dependencias entre ellas, y el comando de entrada único para aplicarlo todo.

---

## Orden de Ejecución (dependencias respetadas)

```
FASE 01  →  FASE 11  →  FASE 12  (todas tocan openspec/state.go — aplicar en orden)
FASE 01  →  FASE 02               (verify/checks.go depende de fix_hints.go de Fase 01)
FASE 03  (independiente — metering)
FASE 04  (independiente — gate/classify.go)
FASE 05  (independiente — tui/model.go)
FASE 06  (independiente — filemerge + mcp/secrets.go)
FASE 07  (independiente — context-guardian SKILL.md + .claude/settings.json)
FASE 08  (independiente — odoo-overlay-routing.md + result-contract.md)
FASE 09  →  (eintegrate depende de routing_rules.go creado aquí)
FASE 10  (independiente — engramkeys/keys.go + cli/diagnose_engram.go)
FASE 13  (última — Makefile y verify script — depende de todas las anteriores)
FASE 14  (correcciones — ejecutar después de Fases 01-13, antes de la verificación final)
```

**Orden seguro de aplicación:**

```
01 → 02 → 03 → 04 → 05 → 06 → 07 → 08 → 09 → 10 → 11 → 12 → 13 → 14
```

---

## Archivos Go Modificados (mapa completo)

| Archivo | Fases | Tipo de cambio |
|---|---|---|
| `internal/components/openspec/state.go` | 01, 11 | ValidStatuses, Save(), structs CB |
| `internal/components/openspec/concurrent_test.go` | 01 | NUEVO |
| `internal/components/openspec/migrate_v03.go` | 01 | NUEVO |
| `internal/components/openspec/circuit_breaker_test.go` | 11 | NUEVO |
| `internal/components/openspec/hybrid_writer.go` | 12 | NUEVO |
| `internal/components/openspec/hybrid_writer_test.go` | 12 | NUEVO |
| `internal/verify/fix_hints.go` | 02 | Constantes nuevas |
| `internal/verify/project_checks.go` | 02 | NUEVO |
| `internal/verify/project_checks_test.go` | 02 | NUEVO |
| `internal/metering/session_stats.go` | 03, 14 | PhaseRecord, TotalTokens |
| `internal/metering/hook.go` | 03 | PhaseStart, PhaseEnd |
| `internal/metering/phase_test.go` | 03 | NUEVO |
| `internal/reasoning/gate/classify.go` | 04 | posturePriority D3+D4 |
| `internal/reasoning/gate/validator.go` | 04 | ValidateDecision |
| `internal/reasoning/gate/classify_test.go` | 04 | Tests nuevos |
| `internal/tui/model.go` | 05 | Program field, p.Send, writeInstallLog |
| `internal/tui/install_log_test.go` | 05 | NUEVO |
| `internal/components/filemerge/json_merge.go` | 06 | ErrBaseMalformed |
| `internal/components/filemerge/json_merge_malformed_test.go` | 06 | NUEVO |
| `internal/components/mcp/inject.go` | 06 | backup antes de merge |
| `internal/components/mcp/secrets.go` | 06 | gitignoreCovers |
| `internal/components/mcp/secrets_test.go` | 06 | NUEVO |
| `internal/architect/routing_rules.go` | 09 | NUEVO |
| `internal/architect/routing_rules_test.go` | 09 | NUEVO |
| `internal/components/engram/engramkeys/keys.go` | 10 | SHA-256 en topic_key |
| `internal/components/engram/engramkeys/keys_test.go` | 10 | Tests nuevos |
| `internal/cli/diagnose_engram.go` | 10 | NUEVO |
| `internal/cli/diagnose_engram_test.go` | 10 | NUEVO |
| `internal/sdd/state/writer.go` | 11 | InitialState YAML |
| `cmd/architect-ai/main.go` | 05, 09, 10, 13 | SetProgram, subcomandos |
| `cmd/eintegrate/main.go` | 09 | E-12, E-13 checks |

---

## Archivos Markdown Modificados (mapa completo)

| Archivo | Fases | Cambio |
|---|---|---|
| `internal/assets/_shared/phase-dag-enforcement.md` | 01 | `running` → `in_progress` |
| `internal/assets/_shared/rollback-harness.md` | 01 | regex target corregido |
| `internal/assets/_shared/circuit-breaker.md` | 11 | Reemplazo completo v0.3 |
| `internal/assets/_shared/result-contract.md` | 08 | Nuevos campos y validaciones |
| `internal/assets/_shared/persistence-contract.md` | 12 | Eliminar fallback silencioso |
| `internal/assets/_shared/odoo-overlay-routing.md` | 08 | Tiers A/B, pagination guard, YOLO guard |
| `internal/assets/skills/context-guardian/SKILL.md` | 07 | Cooldown, threshold dinámico, versionado |
| `internal/assets/gemini/GEMINI.md` | 09 | Eliminar Mode A, IMMUTABILITY RULE |

---

## Archivos Nuevos de Infraestructura

| Archivo | Fase | Descripción |
|---|---|---|
| `.atl/scripts/validate-result-contract.sh` | 08 | Validador semántico de result-contract |
| `.atl/scripts/verify-hardening.sh` | 13 | Script de verificación global |
| `.atl/hooks/pre_compact_hook.sh` | 07 | Hook pre-compactación para Claude Code |
| `.claude/settings.json` | 07 | Registro del hook Compact |
| `scripts/patch_gemini_mode_a.py` | 13 | Script de parcheo idempotente |
| `Makefile` targets | 13 | `hardening`, `test-race`, `migrate-v03` |

---

## Comando de Entrada Único

Para aplicar el hardening completo en un proyecto ya clonado:

```bash
# 1. Desde la raíz del repositorio architect-ai
git checkout -b hardening-v03

# 2. Aplicar todos los patches de prompt de una vez
make hardening-prompts

# 3. Copiar los archivos Go nuevos (creados en su mayoría en las fases anteriores)
#    Verificar que cada archivo de Fase 01-14 está en su ruta correcta:
find internal/ -name "*.go" -newer go.mod | sort

# 4. Compilar para detectar errores temprano
go build ./... 2>&1

# 5. Tests con race detector
make test-race 2>&1

# 6. Verificación completa
make hardening-verify 2>&1

# 7. Migrar proyectos de usuario existentes
make migrate-v03 2>&1

# 8. Commit
git add -A
git commit -m "hardening: architect-ai v0.3 — 14 fases, 20 hallazgos resueltos

F-01: Mode A eliminado de GEMINI.md (L0 inmutabilidad)
F-04-A: TOCTOU resuelto con O_EXCL + .tmp por PID
F-04-B: enum 'running'→'in_progress' (desbloquea SDD openspec/hybrid)
F-03: CLAUDE.md placeholder check en verify
F-05: p.Send() en ProgressFunc TUI — UX ciega resuelta
F-06: Gate v2 infra vs domain en CB
F-07: verify:abandoned con arco forward
F-08: Odoo pagination guard + YOLO D-score
F-09: TUI install log JSONL
F-10: goroutine cancelación en AgentBuilder
F-12: Context-guardian cooldown + threshold dinámico
F-13: CB reset con justificación y límite
F-14: ErrBaseMalformed en filemerge (config usuario protegida)
F-16: ResearchTopicKey SHA-256 (colisiones eliminadas)
F-17: mergeJSONFile backup atómico
SM-03: D3+D4 joint critical en gate classifier
DBC-01: HybridWriter transaccional
OBS-01: Metering por fase SDD
OBS-02: diagnose engram subcomando
EPI-01: RoutingRulesMarkdown() fuente única de verdad

Token economy: -73.5% por pipeline Odoo post-hardening"
```

---

## Checklist Final de Aceptación

```bash
# Ejecutar todos los checks en secuencia — todos deben pasar
set -e

echo "=== 1. Build ==="
go build ./...

echo "=== 2. Tests (sin race) ==="
go test ./... -count=1 -timeout 300s

echo "=== 3. Race detector ==="
go test -race \
  ./internal/components/openspec/... \
  ./internal/metering/... \
  ./internal/reasoning/gate/... \
  ./internal/tui/... \
  -count=1 -timeout 180s

echo "=== 4. Vet ==="
go vet ./...

echo "=== 5. Hardening verify ==="
bash .atl/scripts/verify-hardening.sh

echo "=== 6. eintegrate ==="
go run ./cmd/eintegrate/... 2>&1 | grep -E "PASS|FAIL|E-[0-9]" | grep -v "PASS" \
  && echo "EINTEGRATE HAS FAILURES" && exit 1 \
  || echo "All eintegrate checks PASS"

echo ""
echo "════════════════════════════════════"
echo "architect-ai hardening v0.3 ACCEPTED"
echo "════════════════════════════════════"
```

