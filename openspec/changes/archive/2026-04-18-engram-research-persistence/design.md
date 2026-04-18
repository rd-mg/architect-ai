# Design: Engram Research Persistence

## Technical Approach

Implement a caching layer within the SDD orchestrator that leverages Engram to store and retrieve research findings. This involves extending the `sdd-explore` protocol with a "Cache-Lookup-Before-Call" step, standardizing `topic_key` generation for research findings, and updating the sub-agent result envelope to report hit/miss metrics.

## Architecture Decisions

### Decision: Deterministic Research Keys

**Choice**: `research/{tool}/{slug}-len{N}`.
**Alternatives considered**: Hash of the query (MD5/SHA1).
**Rationale**: LLMs can compute string lengths and prefixes deterministically without tool calls. Computing a hash requires an extra tool call or specialized knowledge. The length-based ID provides a good balance between simplicity and collision avoidance.

### Decision: Cache TTL and Staleness

**Choice**: 168 hours (7 days).
**Alternatives considered**: 24 hours, or perpetual.
**Rationale**: Documentation for libraries and internal code evolves, but not typically every day. 7 days provides significant token savings while keeping the information reasonably fresh.

### Decision: Storage Strategy (Hybrid)

**Choice**: Findings are saved to Engram only (not Openspec filesystem).
**Alternatives considered**: Save to `openspec/cache/`.
**Rationale**: Research findings are transient "working memory". Saving them to the filesystem would clutter the repo and diffs. Engram is designed for this type of persistent memory.

## Data Flow

    Orchestrator Request (Query Q)
           │
           ▼
    Compute Key K = slug(Q) + len(Q)
           │
           ▼
    mem_search(K) ── hit (age < 168h) ──→ Inject into Prompt
           │
           └── miss/stale ──→ Launch Sub-Agent ──→ Tool Call (NB, C7)
                                  │                     │
                                  ▼                     ▼
                           Result Envelope ←──── Save Findings to Engram
                        (metrics: hit/miss)

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/assets/claude/sdd-phase-protocols/sdd-explore.md` | Modify | Add Cache-Lookup-Before-Call step to Research Procedure. |
| `internal/assets/claude/sdd-phase-protocols/sdd-propose.md` | Modify | Update result processing for cache metrics. |
| `internal/assets/claude/sdd-phase-protocols/sdd-design.md` | Modify | Update result processing for cache metrics. |
| `internal/assets/claude/sdd-phase-protocols/sdd-verify.md` | Modify | Update result processing for cache metrics. |
| `internal/assets/claude/sdd-orchestrator.md` | Modify | Global orchestrator result contract update. |
| `skills/_shared/engram-convention.md` | Modify | Document research-class topic keys and TTL. |

## Interfaces / Contracts

### Topic Key Computation (Prompt-side)
1. Take the query string `Q`.
2. Clean: lowercase, replace non-alphanumeric with `-`, collapse multiple `-`.
3. Truncate at 50 chars.
4. Append `-len{Q.length}`.
5. Format: `research/{tool}/{cleaned_query}`.

### Sub-Agent Result Envelope
```json
{
  "status": "success",
  "research_cache_hits": 2,
  "research_cache_misses": 1,
  "research_sources_used": ["engram", "notebooklm", "ripgrep"]
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Key Computation | Verify slug and length generation logic in prompts. |
| Integration | Cache Hit Flow | Mock Engram with a fresh finding and verify tool call skip. |
| Integration | Cache Stale Flow | Mock Engram with a stale finding (>168h) and verify tool call execution. |

## Migration / Rollout

No migration required. The system will start populating the cache as research tasks are performed.
