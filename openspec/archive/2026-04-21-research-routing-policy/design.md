# Design: 07-research-routing-policy

## Component: Orchestrator Layer 5 Injection
Update the `Sub-Agent Launch Template` in all 11 orchestrators to include the finalized Research-Routing Policy.

```markdown
## RESEARCH-ROUTING POLICY (Layer 5)
1. **Engram**: mem_search first. If hit, skip others.
2. **ripgrep-odoo**: Use ~/gitproj/odoo/ for core evidence. Mandatory 2-step protocol.
3. **Context7**: Project structure understanding. No Odoo core packing.
4. **NotebookLM**: Domain knowledge (API changes). Disabled in Mode 3.
5. **Web search**: Last resort. Disabled in Mode 3.
```

## Component: Odoo Path Hardening
Update `internal/assets/overlays/odoo-development-skill/skills/ripgrep-odoo/SKILL.md` to set the base path to `~/gitproj/odoo/`.

## Component: Mode Decision Matrix
Inject the following matrix into the orchestrator logic:
| Mode | Allowed Sources |
|---|---|
| Mode 1 | All |
| Mode 2 | All (Context7/NotebookLM limited) |
| Mode 3-ERR | Engram, ripgrep-odoo |
| Mode 3-CTX | Engram (save only) |

## Implementation Strategy
- Batch update all 11 orchestrator assets.
- Update the ripgrep-odoo skill asset.
