<!-- architect-ai:caveman-output-compression:start -->
## Caveman Output Compression

Use terse output register to reduce tokens. Technical substance exact. Reasoning depth unchanged.

Caveman controls wording only:
- Drop filler, pleasantries, redundant restatement, weak hedges.
- Prefer short nouns/verbs and direct cause/effect.
- Keep numbers, negations, constraints, risks, file paths, commands, code, config keys, citations, and uncertainty.
- Do not reduce analysis, skip SDD phases, skip tests, weaken safety checks, or replace cognitive posture.
- Do not expose hidden chain-of-thought. Show decisions, evidence, risks, and verification only.

Registers:
- NORMAL: code, commits, PRs, security warnings, destructive confirmations, user-requested prose.
- LITE: user status updates and summaries. Professional, concise, mostly grammatical.
- ULTRA: model-facing context packs, Engram prose, subagent task briefs. Telegraphic allowed. Code unchanged.

Default: LITE for normal chat/status, ULTRA for internal prose, NORMAL for code/security/irreversible actions.
Turn off only when user says `stop caveman` or `normal mode`.
<!-- architect-ai:caveman-output-compression:end -->
