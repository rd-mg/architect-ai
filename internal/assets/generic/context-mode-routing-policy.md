<!-- architect-ai:context-mode-routing:start -->
## context-mode Routing Policy

External runtime capability. Use to avoid flooding model context with raw output.

Do:
- Use native reads only for small, exact files.
- Use `ctx_execute` for scripts whose raw I/O may be large.
- Use `ctx_execute_file` for large file processing.
- Use `ctx_batch_execute` for multiple searches/commands.
- Use `ctx_index` + `ctx_search` for large markdown/text corpora.
- Use `ctx_fetch_and_index` for large web pages needing repeated retrieval.
- Run `ctx stats` or `ctx doctor` when diagnosing context-mode itself.

Do not:
- Dump entire repos/files/logs into prompt.
- Use context-mode as durable architecture memory; use Engram/OpenSpec.
- Call context-mode tools speculatively.
- Modify dot-directory config during source refactoring.

Stop when enough evidence exists.

## ctx_batch_execute Pattern

For multiple related searches in one phase step:
  ctx_batch_execute(["cmd1", "cmd2", "cmd3"])
  → Single MCP call, all outputs compressed, returned as one block

Better than: sequential rg calls that each flood the context window.
Threshold: use batch_execute when running ≥3 shell commands in the same step.
Exception: Commands with dependencies (output of cmd1 feeds cmd2) → stay sequential.

## ctx_fetch_and_index Pattern

For large web pages or external docs needed more than once:
  ctx_fetch_and_index(url)  → fetch, chunk, index into context-mode store
  ctx_search(query)         → retrieve relevant chunks by keyword
  NOT: Raw file/web content pasted into prompt
## RESEARCH-ROUTING POLICY (enforce before any external lookup)

Use sources in strict priority order. Escalate ONLY when current source yields no result.

**STEP 1 — Engram (always first)**
`mem_search(query: "{specific_topic_key}", project: "{project}")`
→ Hit: USE IT. Skip steps 2–6.
→ Miss: proceed to step 2.

**STEP 2 — Local ripgrep (Project Evidence)**
Use when: understanding project's own structure or specific file contents.
`ctx_batch_execute(["rg '{keyword}' -l", "rg '^func|^type' {file}"])`
→ Hit: use it. For semantic relationships → also run Step 2b.
→ Miss: proceed to step 2b.

**STEP 2b — CodeGraph (Semantic Relationships)**
Use when: need call chains, callers, impact radius, or cross-file relationships.
Only available when `codegraph_context` is in verified tool list.
```
codegraph_context(query: "{topic}", maxNodes: 25, format: "markdown")
codegraph_callers(nodeId: "{node}")    // who calls this?
codegraph_impact(nodeId: "{node}")     // what breaks?
```
→ Result supplements or replaces multi-file ripgrep chaining.
→ Miss or unavailable: proceed to step 3.

**STEP 3 — Context7 (Framework/Library Docs)**
Use when: documentation for third-party library or API needed.
```
mcp__context7__resolve-library-id(libraryName: "{detected_library}")
mcp__context7__get-library-docs(context7CompatibleLibraryID: "{id}", topic: "{specific_aspect}", tokens: 5000)
```
ALWAYS specify `topic`. Never fetch full docs without topic filter.
→ Hit: use it. → Miss: proceed to step 4.

**STEP 4 — NotebookLM (Optional Synthesis)**
Use when: version-specific changes, migration guides, domain synthesis.
ONLY available in Mode 1 or Mode 2. NOT in Mode 3.
→ Result persists to Engram via after_model hook.

**STEP 5 — Web search (last resort)**
Use when: steps 1–4 all yield no result.
Include `site:` filter when possible.
NOT available in Mode 3.

<!-- architect-ai:context-mode-routing:end -->
