# Proposal: FASE 2 — SDD v3: Phase DAG Enforced + Result Contract + Circuit Breaker + Apply Continuity

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/02-phase-sdd.md`

> Prioridad: 🔴 CRÍTICA  
> Fuentes: "Towards Structured, State-Aware, Execution-Grounded Reasoning" · "Spec-Driven Development: From Code to Contract" · "Constitutional SDD" · TESTS AS INSTRUCTIONS  
> Restricción: Go = solo instalador. Enforcement = MD + YAML + shell hooks + JSON  
> Para: Desarrollador LLM junior — cada instrucción es atómica y verificable

---

## Por qué esta versión (v3) es diferente

**v1**: Fases documentadas en MD, sin enforcement mecánico.  
**v2**: git branch isolation + semantic audit + Odoo detection añadidos.  
**v3 (esta)**: Phase DAG enforced via YAML state · Result Contract con JSON schema · Circuit Breaker (Ralph Loop) · Apply Continuity robusta · artifact-store mode selection.

### Problema fundamental detectado en auditoría

El Phase DAG estaba solo en prompts MD — bypassable bajo context pressure. El agente podía ejecutar `sdd-apply` sin que `sdd-design` existiera. Esto es un **Failure Mode RPN=15** (crítico).

**Solución**: `.atl/sdd-state.yaml` es la fuente de verdad del estado SDD. Cada agente **DEBE** leerlo al inicio para verificar sus prerequisitos. Si los prerequisitos no están marcados como `completed`, el agente **DEBE** rechazar la ejecución y reportar STATUS: BLOCKED.

---

