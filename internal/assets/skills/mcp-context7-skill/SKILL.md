---
name: mcp-context7-skill
description: >
  Tertiary research source. Fetch framework/library official documentation
  (React, Go stdlib, Django, Odoo, etc.) when NotebookLM has no matching
  notebook AND local repo lacks answer. Defers to NotebookLM first.
  Never used BEFORE NotebookLM and local search.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "2.2"
---

# Context7 v2.2

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as first line of response per gate instructions in prompt.

Context7 = tertiary research source. Fetches official docs for frameworks/libraries via MCP. Use ONLY when:

1. NotebookLM has no matching notebook (verified via `notebooklm_list_notebooks`)
2. Local repo lacks answer (verified with ripgrep)
3. Question is about framework/library public surface

If steps 1 and 2 not tried, you're using this skill too early.

---

## Defers to NotebookLM

V3.1 research routing order:

```
1. NotebookLM   ← PRIMARY
2. Local code   ← SECONDARY
3. Context7     ← TERTIARY (you are here)
4. Internet     ← explicit user request only
```

Before calling Context7:

```
# 1. NotebookLM available?
mem_search(query: "knowledge/", project: "{project}")
  → if yes and answer found, STOP — use NotebookLM
  → if yes but no match, fall through
  → if no (unavailable), note and fall through

# 2. Answer in this repo?
rg "{symbol}" --type {lang}
  → if yes, STOP — use local
  → if no, proceed to Context7
```

---

## When Context7 IS the right tool

- Framework public API (React hooks, Django ORM, Odoo decorators)
- Language stdlib behavior (Go's `context` package, Python's `asyncio`)
- Version-specific framework behavior (React 18 vs 19, Odoo 17 vs 18)

Example: "How does `@api.depends` interact with computed stored fields in Odoo 18?"

Context7 handles this. NotebookLM won't have it unless team curated specific upgrade notebook; local repo doesn't contain framework docs.

---

## When Context7 is NOT the right tool

- This project's code → use ripgrep
- Architecture decisions WE made → use NotebookLM
- Upstream library changes not yet adopted → use internet (with user permission)

---

## Query procedure

### Step 1 — Resolve library

```
context7_resolve(library: "react")
  → returns library_id, version list, latest_version
```

Pick version matching project's actual dependency. Do NOT default to `latest_version`.

### Step 2 — Get docs for topic

```
context7_get_docs(
  library_id: "...",
  version: "18.3.0",       // explicit
  topic: "useTransition Suspense interaction"
)
```

### Step 3 — Persist finding (MANDATORY)

```
mem_save(
  title: "knowledge/{framework}/external/{topic}",
  topic_key: "knowledge/{framework}/external/{topic}",
  type: "external-research",
  project: "{project}",
  content: "Q: {question}\nA: {docs summary}\nFramework: {name}@{version}\nDate: {iso-date}"
)
```

**Persistence mandatory in V3.1** — next session same query hits Engram, not Context7.

---

## Cache check — ALWAYS before calling

```
mem_search(query: "knowledge/{framework}/external/", project: "{project}")
  → if found AND covers question, use it
  → if found but stale, re-query and replace
  → if not found, query Context7
```

Staleness:
- Frozen versions (React 18.3.0): never stale. Published docs not edited post-release.
- Rolling versions (React canary, Go tip): stale after 7 days.

---

## Failure modes

- **Context7 MCP unavailable** → `{source: "context7", status: "unavailable"}`, orchestrator falls through to internet (explicit user request only).
- **Library not in Context7 index** → `{source: "context7", status: "no-match"}`, orchestrator may ask user for internet permission.
- **Version not indexed** → `{source: "context7", status: "version-missing", available_versions: [...]}`, orchestrator asks user which to use.

---

## Return envelope (sub-agent result)

```
{
  "source": "context7",
  "framework": "...",
  "version": "...",
  "topic": "...",
  "answer_summary": "2-3 sentences",
  "engram_key": "knowledge/{fw}/external/{topic}",
  "cached_hit": true|false
}
```

---

## Anti-patterns

**Skipping NotebookLM**
```
# WRONG — jumping straight to context7
context7_resolve(library: "odoo")
```
NotebookLM might have Odoo-upgrade notebook with answer. Check first.

**Not persisting findings**
```
# WRONG — querying Context7 twice same session for same topic
```
Always `mem_save` after first call. All future sessions benefit.

**Using Context7 for our own code**
```
# WRONG
context7_get_docs(library: "our-service", topic: "...")
```
Our code is local. Use ripgrep.

---

## See also

- `_shared/research-routing.md` — 4-step priority
- `mcp-notebooklm-orchestrator/SKILL.md` — primary source you defer to
- `ripgrep/SKILL.md` — secondary (local) source
