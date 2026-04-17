---
name: mcp-notebooklm-orchestrator
description: >
  Query-focused NotebookLM MCP skill. Ask questions, search content, compare
  notebooks, and update notebook configurations (tags, descriptions, search
  instructions). NEVER creates new notebooks, sources, or artifacts (audio,
  video, reports, quizzes, etc.). Persists verified findings to Engram for
  cross-session recall.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "2.0"
---

# NotebookLM Orchestrator v2.0 (Query-Focused)

## Purpose

Query and configure NotebookLM notebooks. Use NotebookLM's long-context
understanding for investigation, comparison, and synthesis from already-loaded
sources.

This skill is a CONSUMER of an existing NotebookLM MCP surface. It does not
install, configure, or provision MCP servers. If the host lacks NotebookLM
tools, this skill is unavailable — fall back to other research methods.

## Scope Boundary

| Operation | Allowed | Notes |
|-----------|:-:|-------|
| Ask questions to a notebook | ✅ | Primary use case |
| Search across notebooks | ✅ | For discovery and routing |
| Read notebook descriptions | ✅ | For selecting the right notebook |
| List available notebooks | ✅ | Discovery |
| Compare content across notebooks | ✅ | Synthesis from existing sources |
| Update notebook TAGS | ✅ | For organization |
| Update notebook DESCRIPTIONS | ✅ | For improved discovery |
| Update notebook INSTRUCTIONS | ✅ | E.g., "prioritize code over docs" |
| Update search/response CONFIGURATIONS | ✅ | E.g., result ordering preferences |
| CREATE new notebook | ⛔ NO | Human-only action |
| ADD sources to notebook | ⛔ NO | Human-only action |
| DELETE notebook or sources | ⛔ NO | Human-only action |
| GENERATE artifacts (audio overview, video, report, quiz, flashcards, mind map, slides, infographic, data table) | ⛔ NO | Human-only action |

**Rationale**: Artifact generation and notebook creation are human-authorized
workflows. This skill restricts itself to query and configuration to avoid
accidentally consuming user quota or generating undesired artifacts.

## When to Use

- User asks a question about content already loaded in a NotebookLM notebook
- User needs to compare findings across multiple notebooks
- User wants to improve how a notebook responds (via configuration updates)
- User asks "what does notebook X say about Y?"

## When NOT to Use

- User needs documentation lookup for a well-known framework → use `mcp-context7-skill`
- User needs repository-specific answers → use `ripgrep` or direct file reads
- User wants to create a new notebook → tell the user to create it manually, then this skill can query it
- User wants an artifact (audio, report, etc.) → tell the user the skill is query-only and suggest they generate the artifact manually

## Odoo Notebook Best Practice

When querying Odoo-related notebooks, **always inject this instruction context
before the question**:

```
Base answers on source code first, then on technical documentation, then on
functional documentation. Match the module's Odoo version from __manifest__.py.
```

Empirically, this ordering produces significantly more precise responses for
Odoo-specific questions because the code reflects actual behavior while docs
may lag.

The instruction can be set permanently on the notebook via the "Update
notebook instructions" operation, so subsequent queries inherit it.

## Procedure

### Step 1: Tool Discovery

Inspect the active NotebookLM MCP tool surface. Do NOT assume tool names —
they vary across host implementations.

Typical tool names (examples, verify with the actual surface):
- `notebook_list` / `list_notebooks`
- `notebook_query` / `ask_question` / `query_notebook`
- `notebook_describe` / `get_notebook_info`
- `notebook_update` / `update_notebook_metadata`

### Step 2: Notebook Resolution

Resolve the target notebook in this order:

1. **User-provided identifier** — if the user gives a notebook ID or URL, use it
2. **Project-specific default** — if local context or memory mentions a default notebook for this project, use it
3. **Notebook discovery** — list available notebooks and match by description or tag

**Do NOT** hardcode a notebook ID, URL, or repository-specific default into
this skill. Domain-agnostic by design.

### Step 3: Context Recovery

Apply progressive disclosure. Recover only the minimum local or memory
context needed to formulate the question. Avoid dumping entire file contents
when a focused summary suffices.

### Step 4: Query Formulation

For each question:
- Be specific about what version / what module / what aspect
- Include a hint about the preferred source ordering if known (see Odoo Best Practice above)
- Ask one question per query call — compound questions produce muddy answers

### Step 5: Answer Processing

- If the answer is decisive and verifiable → PERSIST to Engram (see below)
- If the answer is speculative or confidence-low → mark as unverified, do not persist
- If the answer contradicts local code or docs → report the contradiction, do not persist

### Step 6: Configuration Updates (optional)

If the user asks to update notebook configuration:
- Tags: add/remove per the user's request
- Description: update to improve future discovery
- Instructions: update to shape response behavior (e.g., Odoo code-first instruction)

NEVER update configurations without explicit user confirmation, UNLESS the
update is adding the standard Odoo code-first instruction to an Odoo-tagged
notebook (safe default).

## Persistence Contract

Verified findings MUST be saved to Engram for cross-session recall:

```
mem_save(
  title: "notebooklm/{notebook-slug}/{topic-slug}",
  topic_key: "notebooklm/{notebook-slug}/{topic-slug}",
  type: "discovery",
  project: "{project}",
  content: "Notebook: {name}\nQuery: {question}\nAnswer: {concise finding}\nSources cited: {source names}\nVerified: {date}"
)
```

Example:

```
mem_save(
  title: "notebooklm/odoo-v18/account-move-reconciliation",
  topic_key: "notebooklm/odoo-v18/account-move-reconciliation",
  type: "discovery",
  project: "acme-accounting",
  content: "Notebook: Odoo 18 Accounting\nQuery: How does account.move.line.reconcile work in v18?\nAnswer: Uses batch grouping with partial reconciliation. Key method is account.partial.reconcile._create_reconciliation.\nSources cited: odoo/addons/account/models/account_move_line.py\nVerified: 2026-04-17"
)
```

**Do NOT persist**:
- Speculative or low-confidence answers
- Answers that contradict authoritative local sources
- Full raw answers (extract the concise finding)

## Return Envelope

```markdown
**Status**: success | partial | blocked
**Summary**: Queried notebook "{name}". {One-line answer summary}.
**Answer**: {Concise finding}
**Sources cited**: {list from notebook response}
**Verified**: {yes/no — based on your confidence in the answer}
**Engram topic**: notebooklm/{notebook-slug}/{topic-slug} (if persisted)
**Next**: {recommended follow-up or "none"}
**Risks**: {any contradictions with local code or low-confidence flags}
**Skill Resolution**: injected | fallback-registry | fallback-path | none
```

## Rules

- NEVER create a notebook
- NEVER add or delete sources
- NEVER generate artifacts (audio, video, report, etc.)
- NEVER assume tool names — discover them each session
- NEVER hardcode notebook IDs or URLs
- ALWAYS prefer synchronous query over asynchronous workflow
- ALWAYS persist verified findings to Engram
- ALWAYS inject the Odoo code-first instruction when querying Odoo notebooks
- ALWAYS distinguish retrieved facts from your own inference in the answer

## Anti-Patterns

- Generating artifacts "just in case" the user wants them
- Creating new notebooks without user consent
- Querying without first discovering available notebooks
- Persisting unverified or contradictory answers
- Answering version-sensitive questions without stating the version basis
- Mixing NotebookLM answers with your inference without labeling which is which

## Resources

- `internal/assets/skills/mcp-context7-skill/SKILL.md` — sibling skill for framework docs
- `docs/components.md` — MCP integration overview
