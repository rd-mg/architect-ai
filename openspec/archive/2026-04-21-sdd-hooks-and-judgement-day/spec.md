# Specification: SDD Hooks and Judgement Day

## Capability: before_model Hook (Collision & Context)
The orchestrator MUST execute a `before_model` hook before each sub-agent delegation.
- **Collision Check**: Scan for conflicts with existing Engrams (module-specific and global decisions).
- **State Injection**: Inject module state (phase, failures) into the sub-agent prompt.
- **Error Context**: For `apply` and `verify` phases, inject resolved error history to prevent regression.
- **Posture Shift**: If a collision is detected, forcibly upgrade the sub-agent to `+++Autoreason-lite`.

## Capability: after_model Hook (Auto-Persistence)
The orchestrator MUST execute an `after_model` hook after each sub-agent response.
- **State Persistence**: Update `sdd/{module}/state` with current phase and action preview.
- **Pattern Mining**: If `ripgrep-odoo` was used and found a pattern, save to `knowledge/odoo-v{v}/pattern/{slug}`.
- **Brief Versioning**: If `sdd-propose` generated a brief, save as `sdd/{module}/brief/v{N}`.
- **Research Harvesting**: Persist NotebookLM/Context7 results to the global knowledge base.

## Capability: Judgement Day Gate (Odoo Specific)
The `sdd-verify` phase in the Odoo overlay MUST include a "Judgement Day" audit gate.
- **Criterias**:
  1. **Integrity**: Orphan data and `ondelete` checks.
  2. **Collision**: Direct core model modification vs inheritance.
  3. **Scalability**: N+1 prevention and indexing.
- **Routing**:
  - `PASS` -> Proceed to Archive.
  - `FAIL` -> Re-open `sdd-design` for v{N+1} correction.
- **Protections**: Skip if Mode 3 is active or task is trivial (Complexity <= 1).

## Capability: Multi-Orchestrator Parity
The logic MUST be implemented identically across all 11 orchestrator assets (`antigravity`, `gemini`, `claude`, `opencode`, etc.) to ensure behavioral symmetry.

## Verification Scenarios
1. **Collision Detection**: GIVEN a task modifying `res.partner` AND an Engram saying "use inherit for res.partner", THEN `before_model` must inject a collision warning.
2. **Auto-Persistence**: GIVEN a sub-agent completion, THEN the `after_model` hook MUST update the state Engram.
3. **Judgement Day Fail**: GIVEN a design with an N+1 query, THEN Judgement Day MUST trigger a `FAIL` and route back to design.
