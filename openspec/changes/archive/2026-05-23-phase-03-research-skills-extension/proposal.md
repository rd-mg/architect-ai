# Proposal: FASE 3 — Research Universal + bash-expert/fish + Skill Registry v3 Tiers

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/03-phase-research-skills-extension.md`

> Prioridad: 🔴 CRÍTICA
> Fuentes: ByteRover LCM (arxiv 2604.01599) · Context Engineering Azure SRE · Agent Skills vs MCP · "Context Kubernetes" (declarative orchestration)
> Restricción: Go = solo instalador. Enforcement = MD + YAML + JSON + hooks.
> Para: Desarrollador LLM junior — cada paso es atómico y verificable.

---

## Cambios v2 respecto al borrador anterior

La auditoría adversaria detectó:
1. **mcp-notebooklm-orchestrator en bridge:always** = overhead en agentes que nunca la usan
2. **Skill Registry sin tiers** = todos los skills cargan igual, sin progresividad
3. **bash-expert sin Fish** = muchos desarrolladores usan Fish shell como default
4. **Researcher duplicado** en cada orchestrator = divergencia de routing entre fases
5. **Fallback inline faltante** si researcher falla o no está disponible

Esta versión corrige los cinco problemas.

---

