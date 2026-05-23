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
<!-- architect-ai:context-mode-routing:end -->
