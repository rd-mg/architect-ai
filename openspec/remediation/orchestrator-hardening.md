# Orchestrator Hardening

## Goal
Restrict tool access for `sdd-orchestrator` and `general-orchestrator` to delegation and memory tools only.

## Changes
- Removed filesystem, shell, and editing tools: `bash`, `read`, `edit`, `write`.
- Added Engram memory tools: `mem_search`, `mem_get_observation`, `mem_save`, `mem_context`.

## Configuration Changes
- `internal/assets/opencode/opencode.json` updated to restrict tool sets for `sdd-orchestrator` and `general-orchestrator`.
