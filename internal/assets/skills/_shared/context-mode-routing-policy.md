<!-- architect-ai:context-mode-routing:start -->
## context-mode Routing Policy (v3)

context-mode is external runtime capability. Use to avoid flooding model context with raw tool output.

### Auto-Trigger Threshold
If any command/tool output estimated > 10KB (e.g. `git log`, large project-wide `ripgrep`, full test suites, `npm install`), redirect to `ctx_execute()` or `ctx_batch_execute()` instead of raw commands.

### Graceful Fallback
If context-mode server/tools unavailable, DO NOT block execution.
- For large outputs, truncate/pipe to `head -50` and add warning: `[TRUNCATED - context-mode unavailable]`.
- For web fetching, fallback to search snippets rather than raw HTML.
- Gracefully log/WARN and proceed.

Do:
- Use native direct reads only for small, exact files.
- Use `ctx_execute` for scripts whose raw I/O may be large.
- Use `ctx_execute_file` for large file processing.
- Use `ctx_batch_execute` for multiple searches/commands.
- Use `ctx_index` + `ctx_search` for large markdown/text corpora.
- Use `ctx_fetch_and_index` for large web pages needing repeated focused retrieval.
- Run `ctx stats` or `ctx doctor` when diagnosing context-mode itself.

Do not:
- Dump entire repos/files/logs into the prompt.
- Use context-mode as durable architecture memory; use Engram/OpenSpec for that.
- Call context-mode tools speculatively.
- Modify dot-directory config during source refactoring.
- NEVER use `ctx_index` as substitute for `mem_save`.
- NEVER use `ctx_search` as substitute for `mem_search` (no cross-session).

Stop when enough evidence exists.
<!-- architect-ai:context-mode-routing:end -->
