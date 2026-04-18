# Proposal: Engram Research Persistence

## Intent

Implement a prompt-side research caching and persistence mechanism using Engram. This optimizes token usage and speeds up SDD phases by avoiding redundant calls to expensive or slow research tools (NotebookLM, Context7) when the information has already been retrieved recently.

## Scope

### In Scope
- Implement "Cache-Lookup-Before-Call" logic in 9 orchestrator assets.
- Define a deterministic `topic_key` convention for research findings: `research/{tool}/{slug}-len{N}`.
- Add `research_cache_hits` and `research_cache_misses` to sub-agent result envelopes.
- Implement a 7-day (168h) staleness policy for research observations.
- Update `engram-instructions.md` to include research-class topic handling.

### Out of Scope
- Global cross-project research sharing (stay per-project for v1).
- Automatic cache invalidation via hooks (manual staleness check only).
- Persisting transient tool outputs (grep, ripgrep).

## Capabilities

### New Capabilities
- research-persistence: The system MUST check Engram for relevant research findings before calling external tools.

### Modified Capabilities
- sdd-orchestration: Extended with research routing and cache-aware delegation.
- engram-protocol: Added research-class topic handling and staleness rules.

## Approach

1. **Deterministic Keys**: Compute `topic_key` by taking a 50-char prefix of the query, lowercasing, and appending the original query length (`len{N}`) to minimize collisions.
2. **Orchestrator Logic**:
   - Before launching a sub-agent with research tasks, search Engram for existing findings.
   - Inject any found (and fresh) content into the sub-agent prompt as "Previously Found Knowledge".
   - Sub-agent returns whether it hit the cache or performed fresh research.
3. **Persistence**:
   - Save successful, non-empty documentation snippets to Engram with `type: research` and the computed `topic_key`.
   - Truncate findings at 800 bytes or last whitespace before 800 to maintain compact memory.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/assets/claude/sdd-phase-protocols/*` | Modified | Orchestrator protocols updated for cache lookup. |
| `skills/_shared/engram-convention.md` | Modified | Added research topic key and TTL rules. |
| `internal/assets/claude/sdd-orchestrator.md` | Modified | Global SDD orchestrator updated with research metrics. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| ID Collisions | Low | Append query length to prefix (`-len{N}`). |
| Stale Documentation | Med | Enforce 168h TTL; skip cache if older. |
| Cache Bloat | Low | Limit findings to 800 bytes; only save "verified" snippets. |

## Rollback Plan

Revert orchestrator assets and Engram instructions to the previous version. The change is purely additive to the prompt/memory logic and does not alter the underlying binary or data formats.

## Dependencies

- TOPIC-02 (Adaptive reasoning mandatory gate) — for metrics reporting.
- Engram MCP availability.

## Success Criteria

- [ ] Orchestrators correctly perform `mem_search` before external research.
- [ ] Research findings are saved to Engram with the correct `research/*` topic key.
- [ ] Sub-agent results accurately report cache hits/misses.
- [ ] Documentation > 168h old is ignored by the cache logic.
