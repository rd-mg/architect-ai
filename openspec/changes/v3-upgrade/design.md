# Design: architect-ai V3 Upgrade

## Architecture Decisions
1. **Consolidation**: judgment-day and autoreason-lite logic moved into adaptive-reasoning SKILL.md.
2. **Gating**: Odoo patterns moved to 'skills/patterns-{version}/' subdirectories to leverage the existing Go skill-bridging logic.
3. **Disclosure**: Orchestrator prompts split into 'sdd-phase-protocols/' to keep context slim.

## Component Map
- `adaptive-reasoning`: New entry point for all advanced reasoning.
- `cognitive-mode`: New skill for posture management.
- `overlay.go`: Patch to skip root 'skills/*.md' copy.
