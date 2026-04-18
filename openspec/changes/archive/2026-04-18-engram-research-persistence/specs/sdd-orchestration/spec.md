---
openspec_delta:
  base_sha: "8b4b70f36b7cecaf08aeb192a1a1220c68b5e6de4b5fd0fb52d6f3a469132a78"
  base_path: "openspec/specs/sdd-orchestration/spec.md"
  base_captured_at: "2026-04-18T05:33:04.878Z"
  generator: sdd-spec
  generator_version: 1
---

# Delta for SDD Orchestration Contract

## ADDED Requirements

### Requirement: Research Caching (Engram)

The orchestrator MUST verify the existence and freshness of research findings in Engram before delegating tasks to research-heavy sub-agents.

#### Scenario: Cache Hit
- GIVEN a research question with an existing finding in Engram (age < 168h)
- WHEN the orchestrator plans research
- THEN it MUST retrieve the finding and inject it into the sub-agent prompt as "Previously Found Knowledge".

#### Scenario: Cache Miss (Stale)
- GIVEN a research finding in Engram older than 168h
- WHEN the orchestrator plans research
- THEN it MUST ignore the stale finding and proceed with fresh research.

### Requirement: Research Class Topic Keys

The system MUST compute deterministic topic keys for research findings to enable efficient lookups.

#### Scenario: Topic Key Computation
- GIVEN a query "How does Odoo 19 handle SQL constraints?"
- WHEN computing the topic key for NotebookLM
- THEN the result MUST follow the pattern `research/notebooklm/how-does-odoo-19-handle-sql-constraints-len43`.

## MODIFIED Requirements

### Requirement: Result Contract Extension

The orchestrator result contract MUST include `chosen_mode`, `mode_rationale`, `research_cache_hits`, and `research_cache_misses` fields.

(Previously: Result contract only included status, summary, artifacts, chosen_mode, and mode_rationale.)

#### Scenario: Metrics Reporting
- GIVEN a sub-agent execution that used 2 cached findings and performed 1 fresh search
- WHEN the orchestrator processes the result
- THEN the result envelope MUST contain `research_cache_hits: 2` and `research_cache_misses: 1`.
