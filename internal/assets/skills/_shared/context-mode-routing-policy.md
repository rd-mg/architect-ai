<!-- architect-ai:context-mode-routing:start -->
## context-mode Routing Policy

context-mode is an external runtime capability. Use it to avoid flooding the model context with raw tool output.

Do:
- Use native direct reads only for small, exact files.
- Use `ctx_execute` for scripts whose raw inputs/outputs may be large.
- Use `ctx_execute_file` for large file processing.
- Use `ctx_batch_execute` for multiple searches/commands.
- Use `ctx_index` + `ctx_search` for large markdown/text corpora.
- Use `ctx_fetch_and_index` for large web pages that need repeated focused retrieval.
- Run `ctx stats` or `ctx doctor` when diagnosing context-mode itself.

Do not:
- Dump entire repos/files/logs into the prompt.
- Use context-mode as durable architecture memory; use Engram/OpenSpec for that.
- Call context-mode tools speculatively.
- Modify dot-directory config during source refactoring.

Stop when enough evidence exists.
<!-- architect-ai:context-mode-routing:end -->
