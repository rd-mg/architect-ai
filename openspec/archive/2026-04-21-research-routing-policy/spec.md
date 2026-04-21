# Specification: 07-research-routing-policy

## Capability: 5-Step Research Routing
The orchestrator MUST mandate that all sub-agents follow a hierarchical research protocol:
1. **Engram**: Always first. `mem_search` specific topic keys.
2. **ripgrep-odoo**: For Odoo core implementation evidence.
3. **Context7**: For current project structure (not Odoo core).
4. **NotebookLM**: For updated framework/API knowledge (migration guides).
5. **Web search**: Last resort with site filters (`site:github.com/odoo`).

## Capability: Mode-Based Source Restrictions
Routing availability MUST be gated by the active Adaptive Reasoning Mode:
- **Mode 1/2**: All sources available (Context7/NotebookLM limited in Mode 2).
- **Mode 3-ERR**: Engram and ripgrep-odoo only.
- **Mode 3-CTX**: Engram `mem_save` only.
- **Web search**: Prohibited in all Mode 3 variations.

## Capability: Odoo Path Standard
- The `ripgrep-odoo` skill MUST point all local Odoo searches to the canonical directory: `~/gitproj/odoo/`.
- The orchestrator MUST inject this path when delegating Odoo-specific research tasks.

## Capability: 2-Step Extraction Protocol
The orchestrator MUST enforce the 2-step protocol for all local code searches:
- Step A: Find files with matches (`-l`).
- Step B: Targeted extraction from identified files.

## Verification
- GIVEN a sub-agent delegation, THEN the prompt MUST contain the Layer 5 Research-Routing Policy.
- GIVEN Mode 3-CTX, THEN the prompt MUST prohibit external research.
- GIVEN an Odoo research task, THEN the search path MUST be `~/gitproj/odoo/`.
