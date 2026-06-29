<!-- architect-ai:codegraph-tools-reference:v1 -->
## CodeGraph MCP Tools Reference

Requires: `codegraph` server active in MCP config.
Priority: Use BEFORE ripgrep for all relationship/impact queries.

### Tool Signatures

```
codegraph_context(query: string, maxNodes?: int = 25, format?: "markdown" | "json")
  → Semantic context pack: nodes related to query (functions, types, files, call chains)
  → Use as PRIMARY exploration tool (replaces multi-file ripgrep)

codegraph_trace(entry: string)
  → Full call chain from entry point to all leaf calls
  → Use for: "what does this function ultimately call?"

codegraph_callers(nodeId: string)
  → All nodes that call this node (inbound references)
  → Equivalent to: LspFindReferences, "who calls this?"

codegraph_callees(nodeId: string)
  → All nodes this node calls (outbound references)
  → Equivalent to: "what does this call?"

codegraph_impact(nodeId: string, depth?: int = 3)
  → Blast radius: what breaks if nodeId changes
  → Use in sdd-explore (impact surface) and sdd-verify (change validation)

codegraph_search(query: string, kind?: string, limit?: int = 10)
  → Symbol search: find function/type/class by name
  → Use instead of: rg "^func {name}"
```

### Decision Matrix: CodeGraph vs ripgrep

| Query type | Use CodeGraph | Use ripgrep |
|-----------|--------------|-------------|
| Find all callers of function X | ✅ codegraph_callers | ❌ slow, error-prone |
| Find impact radius of change X | ✅ codegraph_impact | ❌ impossible |
| Find string literal "config_key" | ❌ | ✅ rg -l "config_key" |
| Find all YAML env var references | ❌ | ✅ rg "MY_VAR" --type yaml |
| Find related functions/types | ✅ codegraph_context | ❌ pattern-only |
| Find file containing "func AuthHandler" | ✅ codegraph_search | ✅ either |

### Initialization
First use in a session: verify codegraph index is current.
Run in project root:
```bash
codegraph init --quiet  # idempotent — only re-indexes changed files
```
If codegraph command not found: skip to ripgrep fallback silently.

### Fallback (if codegraph unavailable)
All codegraph tool calls → replace with:
- codegraph_callers → rg "{function_name}" -l
- codegraph_context → sequential_thinking + rg "^func|^type|^class" sweep
- codegraph_impact → manual multi-file grep + dependency tracing
