---
name: mcp-context7-skill
description: >
  Tertiary research source. Used for framework and library official
  documentation (React, Go stdlib, Django, Odoo, etc.) when NotebookLM
  has no matching notebook AND the local repo does not contain the
  answer. Defers to NotebookLM first. Never used BEFORE NotebookLM
  and local search.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "2.1"
---

# Context7 v2.1

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.


Context7 is the tertiary research source. It fetches official documentation for frameworks and libraries via MCP. Use it ONLY when:

1. NotebookLM has no matching notebook (verified with `notebooklm_list_notebooks`)
2. The local repo does not contain the answer (verified with ripgrep)
3. The question is about a framework/library's public surface

If steps 1 and 2 haven't been tried, you're using this skill too early.

---

## Defers to NotebookLM

In V3.1, the research routing order is:

```
1. NotebookLM   ← PRIMARY
2. Local code   ← SECONDARY
3. Context7     ← TERTIARY (you are here)
4. Internet     ← explicit user request only
```

Before calling Context7:

```
# 1. Is NotebookLM available?
mem_search(query: "notebooklm/", project: "{project}")
  → if yes and answer found, STOP — use NotebookLM result
  → if yes but no match, fall through
  → if no (unavailable), note in response and fall through

# 2. Is the answer in this repo?
rg "{symbol}" --type {lang}
  → if yes, STOP — use local
  → if no, proceed to Context7
```

---

## When Context7 IS the right tool

- Questions about a framework's public API (React hooks, Django ORM, Odoo decorators)
- Language stdlib behavior (Go's `context` package, Python's `asyncio`)
- Version-specific framework behavior (React 18 vs 19, Odoo 17 vs 18)

Example:
> How does `@api.depends` interact with computed stored fields in Odoo 18?

This is Context7's job. NotebookLM won't have it unless the team specifically curated an Odoo 18 upgrade notebook; local repo doesn't contain framework docs.

---

## When Context7 is NOT the right tool

- Questions about THIS project's code — use ripgrep
- Questions about architecture decisions WE made — use NotebookLM
- Questions about upstream changes in a library we haven't adopted yet — use the internet (with user permission)

---

## Query procedure

### Step 1 — Resolve the library

```
context7_resolve(library: "react")
  → returns library_id, version list, latest_version
```

Pick the version matching the project's actual dependency. Do NOT default to `latest_version` — the user may be on an older version.

### Step 2 — Get docs for a topic

```
context7_get_docs(
  library_id: "...",
  version: "18.3.0",       // explicit
  topic: "useTransition Suspense interaction"
)
```

### Step 3 — Persist the finding

```
mem_save(
  title: "context7/{framework}/{version}/{topic}",
  topic_key: "context7/{framework}/{version}/{topic}",
  type: "external-research",
  project: "{project}",
  content: "Q: {question}\nA: {docs summary}\nFramework: {name}@{version}\nDate: {iso-date}"
)
```

**This persistence is mandatory in V3.1** — new requirement. Next session asking the same question hits Engram, not Context7.

---

## Cache check — ALWAYS before calling

```
mem_search(query: "context7/{framework}/{version}/", project: "{project}")
  → if found AND content covers the question, use it
  → if found but stale, re-query and replace
  → if not found, query Context7
```

Staleness:
- Frozen versions (React 18.3.0): never stale. Frameworks don't edit published docs post-release.
- Rolling versions (React canary, Go tip): stale after 7 days.

---

## Failure modes

- **Context7 MCP not available** → return `{source: "context7", status: "unavailable"}`, orchestrator falls through to internet (only on explicit user request).
- **Library not in Context7 index** → return `{source: "context7", status: "no-match"}`, orchestrator may ask user permission to search the internet.
- **Version not indexed** → return `{source: "context7", status: "version-missing", available_versions: [...]}`, orchestrator asks user which to use.

---

## Return envelope (in sub-agent result)

```
{
  "source": "context7",
  "framework": "...",
  "version": "...",
  "topic": "...",
  "answer_summary": "2-3 sentences",
  "engram_key": "context7/{fw}/{ver}/{topic}",
  "cached_hit": true|false
}
```

---

## Anti-patterns

**❌ Skipping NotebookLM**
```
# WRONG — jumping straight to context7
context7_resolve(library: "odoo")
```
NotebookLM might have an Odoo-upgrade notebook with exactly the answer. Check first.

**❌ Not persisting findings**
```
# WRONG — querying Context7 twice in the same session for the same topic
```
Always `mem_save` after the first call. Every future session benefits.

**❌ Using Context7 for our own code**
```
# WRONG
context7_get_docs(library: "our-service", topic: "...")
```
Our code is local. Use ripgrep.

---

## See also

- `_shared/research-routing.md` — the 4-step priority
- `mcp-notebooklm-orchestrator/SKILL.md` — the primary source you defer to
- `ripgrep/SKILL.md` — the secondary (local) source
