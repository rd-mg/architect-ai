# Proposal: FASE 7 — GGA v2: Pre-Commit AI Audit con Directivas Agnósticas + cudio-git + Odoo

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/07-phase-gga-v2-guardrails.md`

> Prioridad: 🟡 ALTA
> Fuentes: "AgentGuard: Runtime Verification of AI Agents" · "Tool Annotations as Risk Vocabulary" · "Trustworthy Agentic AI Systems" · "AgentVerify: Compositional Formal Verification"
> Restricción: Go = solo instalador. GGA = bash hook + PS1 shim + AGENTS.md prompt.
> Para: Desarrollador LLM junior — cada regla produce un resultado verificable (BLOCK/WARN/APPROVE).

---

## Cambios v2 respecto al borrador anterior

La auditoría adversaria detectó:
1. **PS1 sin jq** — Windows PowerShell no tiene `jq` instalado por defecto. La v1 fallaba silenciosamente.
2. **Pattern evasion** — `"rm -rf ": "ask"` con space trailing bypassable con doble-espacio o variable expansion.
3. **CI environment detection missing** — en GitHub Actions el hook instala pero el AI endpoint no está disponible. Debería degradar gracefully.
4. **Odoo rules sin version-gating** — las reglas `<list>` vs `<tree>` aplican solo a v18+, no a v14-v16.
5. **`--skip-ai` sin audit trail durable** — solo escribía a `.gga/skip-log.jsonl` local. Si el repo se clona, el log se pierde.

---

