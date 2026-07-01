<!-- architect-ai:codegraph-tools-reference:v2 -->
## CodeGraph MCP Tools Reference

Requires: `codegraph` server active in MCP config.
Priority: Use BEFORE ripgrep for all semantic, relationship, and impact queries.

### Tool Signatures

```
codegraph_explore(query: string)
  → Answer almost any question in one call — "how does X work", a flow ("how does X reach Y"), or surveying an area — returning the relevant symbols' verbatim source grouped by file, plus the call paths between them and a blast-radius summary. Surfaces dynamic-dispatch hops (callbacks, React re-render, interface→impl) grep can't follow. Name a file or symbol in the query to read its current line-numbered source, the same shape the Read tool gives you.
```

### CLI Subcommands and Equivalents

When not running as an MCP server or when utilizing CodeGraph via the command line interface, the following commands are available:

- `codegraph init [path]`             # Initialize a project + build its graph (one step)
- `codegraph uninit [path]`           # Remove CodeGraph from a project (--force to skip prompt)
- `codegraph index [path]`            # Full index (--force to re-index, --quiet for less output)
- `codegraph sync [path]`             # Incremental update
- `codegraph status [path]`           # Show statistics
- `codegraph unlock [path]`           # Remove a stale lock file that's blocking indexing
- `codegraph query <search>`          # Search symbols (--kind, --limit, --json)
- `codegraph explore <query>`         # Relevant symbols' source + call paths in one shot (same output as the codegraph_explore MCP tool)
- `codegraph node <symbol|file>`      # One symbol's source + callers, or read a file with line numbers (same output as codegraph_node)
- `codegraph files [path]`            # Show file structure (--format, --filter, --max-depth, --json)
- `codegraph callers <symbol>`        # Find what calls a function/method (--limit, --json)
- `codegraph callees <symbol>`        # Find what a function/method calls (--limit, --json)
- `codegraph impact <symbol>`         # Analyze what code is affected by changing a symbol (--depth, --json)
- `codegraph affected [files...]`     # Find test files affected by changes
- `codegraph daemon`                  # Manage background daemons — pick one to stop (alias: daemons)

### MCP Server Tool Configuration

When running as an MCP server, CodeGraph exposes a single tool — `codegraph_explore` — by default.
The other tools (`codegraph_node`, `codegraph_search`, `codegraph_callers`, `codegraph_callees`, `codegraph_impact`, `codegraph_files`, `codegraph_status`) are unlisted by default but remain fully functional. Their output is automatically integrated into `codegraph_explore` (e.g. its blast-radius section, the relationship map, and callee lists).

To re-enable specific sub-tools on the MCP surface, use the `CODEGRAPH_MCP_TOOLS` environment variable:
```bash
CODEGRAPH_MCP_TOOLS=explore,node,search,callers
```

### Decision Matrix: CodeGraph vs ripgrep

| Query type | Use CodeGraph | Use ripgrep |
|-----------|--------------|-------------|
| Tracing/exploration & call paths | ✅ codegraph_explore / CLI explore | ❌ manual search & tracing |
| Find all callers of function X | ✅ codegraph_explore / CLI callers | ❌ slow, error-prone |
| Find impact radius of change X | ✅ codegraph_explore / CLI impact | ❌ impossible |
| Read file or symbol body | ✅ codegraph_explore / CLI node | ❌ line-number grep |
| Find string literal "config_key" | ❌ | ✅ rg -l "config_key" |
| Find all YAML env var references | ❌ | ✅ rg "MY_VAR" --type yaml |

### Initialization
First use in a session: verify codegraph index is current.
Run in project root:
```bash
codegraph init
```
If codegraph command not found: skip to ripgrep fallback silently.

### Fallback (if codegraph unavailable)
If CodeGraph is not installed or indexed, use standard tools:
- Tracing/exploration → sequential_thinking + rg "^func|^type|^class" sweep
- Callers/impact tracing → manual multi-file grep + dependency tracing
