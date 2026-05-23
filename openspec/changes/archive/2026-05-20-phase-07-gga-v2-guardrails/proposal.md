# Proposal: Phase 7 - GGA v2 Agnostic Directives + cudio-git + Odoo Rules

## Intent
Upgrade GGA (Gentleman Guardian Angel) to v2 with language-agnostic rules, cudio-git commit format validation, Odoo-specific security/architecture rules, and emergency `--skip-ai` mode with an Engram durable audit trail.

## Cambios v2 respecto al borrador anterior
La auditoría adversaria detectó:
1. **PS1 sin jq** — Windows PowerShell no tiene `jq` instalado por defecto. La v1 fallaba silenciosamente.
2. **Pattern evasion** — `"rm -rf ": "ask"` con space trailing bypassable con doble-espacio o variable expansion.
3. **CI environment detection missing** — en GitHub Actions el hook instala pero el AI endpoint no está disponible. Debería degradar gracefully.
4. **Odoo rules sin version-gating** — las reglas `<list>` vs `<tree>` aplican solo a v18+, no a v14-v16.
5. **`--skip-ai` sin audit trail durable** — solo escribía a `.gga/skip-log.jsonl` local. Si el repo se clona, el log se pierde.

## Scope
### In Scope
- Section A: Agnostic rules (secrets, architecture, code quality, testing, error handling, dependencies) for ALL projects. Robust regex for pattern evasion (`\s*`).
- Section B: Commit format rules — cudio-git `[TAG][ID] module: desc` and conventional commits fallback.
- Section C: Odoo-specific rules (SQL injection, sudo abuse) with strict Version-Gating auto-activated when IS_ODOO=true.
- Section D: `--skip-ai` mode with non-skippable secret detection and local + Engram durable audit trail (`.gga/skip-log.jsonl`).
- Section E: CI mode auto-detection for graceful degradation (static checks only).
- Section F: Non-blocking mode when AI provider is down — warn-only.
- Go implementation: `internal/gga/installer.go` with platform-aware hook generation.
- Multi-OS hooks: bash (Linux/macOS) + native PowerShell without `jq` dependency (Windows).

### Out of Scope
- Custom user-defined rules beyond A-F sections.

## Impact
- ALL projects get security scanning (secrets, credentials) — non-skippable and bypass-resistant.
- Odoo projects get specialized SQL injection and architecture checks gated by detected version.
- Windows users no longer require `jq` to run the hook.
- CI environments automatically degrade to static checks without hanging.
- Emergency production fixes are never blocked by AI downtime.
- Audit trail ensures skip events are traceable across sessions via Engram.
