## Engram Persistent Memory — Protocol

You have access to Engram, a persistent memory system that survives across sessions and compactions. This protocol is MANDATORY and ALWAYS ACTIVE — not something you activate on demand.

### WHEN TO SAVE (mandatory — not optional)

Call mem_save IMMEDIATELY and WITHOUT BEING ASKED after any of these:
- Bug fix completed
- Architecture or design decision made
- Non-obvious discovery about the codebase
- Configuration change or environment setup
- Pattern established (naming, structure, convention)
- User preference or constraint learned

Format for mem_save:
- **title**: Verb + what — short, searchable (e.g. "Fixed N+1 query in UserList")
- **type**: bugfix | decision | architecture | discovery | pattern | config | preference
- **scope**: project (default) | personal
- **topic_key** (MANDATORY for evolving topics): stable key like `architecture/auth-model`, `decision/logger-refactor`, or `sdd/{change-name}/state`.
- **content**:
  - **What**: One sentence — what was done
  - **Why**: What motivated it (user request, bug, performance, etc.)
  - **Where**: Files or paths affected
  - **Learned**: Gotchas, edge cases, things that surprised you (omit if none)

Topic update rules:
- Different topics must not overwrite each other
- Same evolving topic → ALWAYS reuse same topic_key (upsert). This ensures each topic has exactly ONE high-signal source of truth.
- Unsure about key → call mem_suggest_topic_key first
- Know exact ID → use mem_update

### STRATEGIC RESUMPTION (mandatory)
Before starting ANY new task or phase, you MUST catch up on state:
1. Call mem_context for recent session history.
2. Search for relevant stable topic_keys (e.g., sdd/{change-name}/state) to find the latest "source of truth".
3. Use mem_get_observation for full untruncated content before making new decisions.

### WHEN TO SEARCH MEMORY

On any variation of "remember", "recall", "what did we do", "how did we solve", or references to past work (including Spanish equivalents "recordar" / "qué hicimos"): <!-- trigger-phrase-allowlist -->
1. Call mem_context — checks recent session history (fast, cheap)
2. If not found, call mem_search with relevant keywords
3. If found, use mem_get_observation for full untruncated content

Also search PROACTIVELY when:
- Starting work on something that might have been done before
- User mentions a topic you have no context on
- User's FIRST message references the project, a feature, or a problem — call mem_search with keywords from their message to check for prior work before responding

### SESSION CLOSE PROTOCOL (mandatory)

Before ending a session or saying "done" / "that's it" (including Spanish equivalent "listo"), call `mem_session_summary`: <!-- trigger-phrase-allowlist -->

## Goal
[What we were working on this session]

## Instructions
[User preferences or constraints discovered — skip if none]

## Discoveries
- [Technical findings, gotchas, non-obvious learnings]

## Accomplished
- [Completed items with key details]

## Next Steps
- [What remains to be done — for the next session]

## Relevant Files
- path/to/file — [what it does or what changed]

This is NOT optional. If you skip this, the next session starts blind.

### PASSIVE CAPTURE — automatic learning extraction

When completing a task or subtask, include a "## Key Learnings:" section at the end of your response with numbered items. Engram will automatically extract and save these.

Example:
## Key Learnings:

1. bcrypt cost=12 is the right balance for our server performance
2. JWT refresh tokens need atomic rotation to prevent race conditions

You can also call mem_capture_passive(content) directly with any text containing a learning section.

### AFTER COMPACTION

If you see a compaction message or "FIRST ACTION REQUIRED":
1. IMMEDIATELY call mem_session_summary with the compacted summary content — this persists what was done before compaction
2. Call mem_context to recover additional context from previous sessions
3. Only THEN continue working

Do not skip step 1. Without it, everything done before compaction is lost from memory.
