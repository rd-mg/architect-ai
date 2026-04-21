# Research Routing Policy (Shared Fragment)

**Scope**: Any skill or phase protocol that decides between external research tools MUST follow this priority order.

---

## The priority

```
1. NotebookLM           ← project-curated knowledge (first choice)
2. Local code + docs    ← repo itself (ripgrep, find, cat, extract-text)
3. Context7             ← framework/library official docs
4. Internet             ← LAST, only on EXPLICIT user request
```

No other order is acceptable. Deviation requires explicit user approval.

---

## Decision tree

```
Does the question involve project-specific knowledge
(our architecture, our decisions, our conventions)?
  │
  YES → Step 1: NotebookLM
  │       mem_search / notebooklm_query for matching notebook
  │       If hit → use it, STOP
  │       If miss → fall through to Step 2
  │
  NO  → The question is framework/tool-specific (Odoo, React, Go stdlib, etc.)?
         │
         YES → Step 3: Context7 directly (skip 1, 2)
         │
         NO  → Step 2: Local code + docs (ripgrep)
```

---

## Step 1 — NotebookLM (FIRST CHOICE)

Use when:
- The question is about project-specific context (our repo's architecture, our past decisions, our conventions)
- The user references "our docs", "the team's guide", "the notebook"
- The answer might be in a curated notebook — if unsure, TRY NotebookLM FIRST

Probe:
```
mem_search(query: "{user-question topic}", project: "{project}")
  → look for knowledge/{domain}/external/{topic} topic-keys
```

If a matching notebook exists, read it and answer. If not, fall through.

**NotebookLM is query-only** (enforced in V3). Never attempt to create notebooks or write artifacts.

---

## Step 2 — Local code + docs

Use when:
- The question is about THIS repo's actual code (function signatures, types, call sites)
- The answer is definitely in the tree (e.g., "how does our retry logic work?")

Tools in preference order:
1. `ripgrep` — pattern search (ALWAYS preferred over `grep`)
2. `find -name` — filename search
3. `cat` / `extract-text` — read specific files
4. Language toolchain (`gopls`, `tsc --listFiles`, `python -c 'import X; print(X.__file__)'`) for semantic questions

Budget: 2 minutes and ≤10 tool calls. If you can't find it in the local tree in 10 calls, it's probably not there — fall through.

---

## Step 3 — Context7

Use when:
- The question is about a framework, library, or language feature (NOT our code)
- Steps 1 + 2 failed, OR the question is obviously external (e.g., "how does React's `useTransition` behave under Suspense?")

Probe:
```
context7_resolve(library: "react")
context7_get_docs(library_id: "...", topic: "useTransition Suspense")
```

Persist findings in Engram under `context7/{framework}/{version}/{topic}` so the next session skips the API call.

---

## Step 4 — Internet (LAST RESORT)

**Only use when the user explicitly asks.** Trigger phrases:
- "search the web"
- "look online"
- "check the internet"
- "google it"
- "what does the web say about..."
- "busca en internet", "busca online"

Absent an explicit trigger, do NOT call `web_search` / `web_fetch`. Return what you have from steps 1-3 and tell the user you did not search the web.

When you DO search:
- Prefer original sources (official docs, vendor blog, RFC, spec) over aggregators
- Cite URLs in your response
- Persist findings in a NotebookLM-compatible note if the project has a research notebook

---

## Explicit override

The user can ask to skip steps:
- "use Context7 directly" → skip 1, 2
- "search the internet" → skip 1, 2, 3

Honor the override, but note in your response what was skipped:
> Using Context7 directly as requested. Skipped NotebookLM + local search.

---

## Rationale

NotebookLM-first protects project-specific knowledge from being overwritten by generic external answers. Local-second keeps you fast and grounded. Context7 is expensive (network, tokens, rate limits) and should be a deliberate choice. Internet is the slowest, noisiest, and most prone to hallucinated facts — reserve for when the human explicitly wants it.

---

## See also

- `mcp-notebooklm-orchestrator/SKILL.md` — how to query NotebookLM
- `mcp-context7-skill/SKILL.md` — how to query Context7; includes "defers to NotebookLM" section
- `ripgrep/SKILL.md` — local search
