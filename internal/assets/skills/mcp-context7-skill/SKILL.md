---
name: mcp-context7-skill
description: >
  Live framework and library documentation via Context7 MCP. Resolves the
  exact library target, fetches only the minimum needed docs, and persists
  verified findings to Engram for cross-session recall. Treat as a
  documentation consumer — does NOT install, reconfigure, or duplicate MCP
  wiring.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "2.0"
---

# Context7 Skill v2.0

## Purpose

Use Context7 as a LIVE documentation skill when stale model memory would
produce wrong answers. This skill assumes the agent already has access to the
existing `context7` MCP component and focuses on:

1. Choosing the right library target
2. Fetching only the minimum relevant documentation
3. Turning it into actionable guidance
4. **Persisting verified findings to Engram** (new in v2.0)

## When to Use

- User asks for current library or framework documentation
- User needs API syntax, migration notes, or version-specific behavior
- User wants examples grounded in upstream docs rather than cached model memory
- User asks for a documentation-backed answer before writing or changing code

## When NOT to Use

- The answer is fully contained in the current repository → use `ripgrep`
- The question is general web research → use web search, not Context7
- Context7 is unavailable and local docs already answer the question
- The question is about Odoo modules → use `mcp-notebooklm-orchestrator` instead

## Critical Patterns

- Treat this skill as a CONSUMER of the existing `context7` component. Do not
  install, reconfigure, or duplicate MCP wiring from this skill.
- Resolve the exact library, package, or framework first before pulling docs.
  If the target is ambiguous, ask for the missing identifier instead of
  guessing.
- Prefer version-aware requests. When the user gives a version, include it in
  the lookup. When the version is unknown and materially affects correctness,
  say so.
- Pull the minimum relevant documentation needed to answer the task. Do not
  dump large doc sections when a focused summary plus the decisive API details
  is enough.
- Keep documentation facts separate from your own inference. Label inferred
  migration advice, risk analysis, or implementation guidance clearly.
- Use Context7 for library and framework documentation only. Do not turn it
  into a generic search workflow or a replacement for repository inspection.
- Stay domain-agnostic. This bundled skill must not contain domain-specific,
  repository-specific, or product-specific defaults.

## Inputs

- The package, framework, or library the user is asking about
- Optional version or ecosystem details that narrow the lookup
- The question to answer or the code change being planned
- The live Context7 tool surface exposed by the active host

## Outputs

- One documentation-backed answer scoped to the user request
- One short explanation of any uncertainty caused by missing version or target
- One explicit fallback when Context7 cannot answer or is not available
- **Persistence to Engram** for verified findings (see Persistence Contract)

## Procedure

### Step 1: Confirm Upstream Docs Are Needed

Ask: does answering this require upstream authoritative docs, or can the
repository answer it? If repo → use ripgrep, not Context7.

### Step 2: Identify the Exact Target

Library name + version. If either is missing and version materially affects
correctness, ask the user before proceeding.

### Step 3: Resolve the Target

Use the active Context7 `resolve` tool surface to confirm the correct
identifier before requesting documentation excerpts.

### Step 4: Fetch Minimum Relevant Docs

Request only the sections needed for the current question. Do not pull entire
modules when a focused API reference suffices.

### Step 5: Answer with API Facts

Lead with the relevant API facts, constraints, and version notes. Clearly
separate retrieved facts from your own inference.

### Step 6: Persist Verified Findings (NEW in v2.0)

If the documentation lookup produces a VERIFIED finding that:
- Answers a version-specific behavioral question
- Clarifies an API contract the project depends on
- Resolves a migration compatibility question
- Would need to be re-looked-up in a future session

Then persist to Engram:

```
mem_save(
  title: "context7/{framework}/{version}/{topic}",
  topic_key: "context7/{framework}/{version}/{topic}",
  type: "discovery",
  project: "{project}",
  content: "Framework: {name} v{version}\nQuery: {question}\nFinding: {concise answer}\nSource: {Context7 doc section reference}\nVerified: {date}"
)
```

Example:

```
mem_save(
  title: "context7/react/19.0/use-hook-signature",
  topic_key: "context7/react/19.0/use-hook-signature",
  type: "discovery",
  project: "my-react-app",
  content: "Framework: React v19.0\nQuery: What is the signature of the new 'use' hook?\nFinding: use<T>(promise: Promise<T>): T — unwraps promises and context in render.\nSource: React 19 Hooks Reference / use\nVerified: 2026-04-17"
)
```

**Do NOT persist**:
- Inferred guidance (label it separately and don't save)
- Generic documentation browsing (no specific question answered)
- Findings with low confidence
- Version-agnostic facts (those are in the model's training data already)

### Step 7: Fallback

If Context7 is unavailable or the target cannot be resolved safely:
- Say so explicitly
- Fall back to repository evidence
- Or ask the user for clarification

## Return Envelope

```markdown
**Status**: success | partial | blocked
**Summary**: {One-line description of what was answered}
**Finding**: {Concise API fact or behavior description}
**Version context**: {e.g., "React 19.0, differs from 18.x"}
**Inferred guidance**: {Your inference, clearly labeled if present}
**Source**: {Context7 doc section}
**Engram topic**: context7/{framework}/{version}/{topic} (if persisted)
**Next**: {recommended follow-up or "none"}
**Risks**: {ambiguity or uncertainty flags}
**Skill Resolution**: injected | fallback-registry | fallback-path | none
```

## Guardrails

- Do not invent Context7 tool names or payload fields
- Do not install or patch MCP config from this skill
- Do not assume a domain, framework family, or repository-specific default
- Do not treat stale model memory as equivalent to retrieved documentation
- Do not answer version-sensitive questions without stating the version basis
- Do not persist unverified or low-confidence findings to Engram

## Anti-Patterns

- Querying Context7 for answers already in the repository
- Dumping large doc sections verbatim when a focused summary suffices
- Mixing retrieved facts with inference without clear labels
- Persisting every query to Engram regardless of verification confidence
- Skipping persistence when the finding is clearly valuable (memory becomes stale)

## Resources

- `internal/assets/skills/mcp-notebooklm-orchestrator/SKILL.md` — sibling skill for notebook-based research
- `docs/components.md` — MCP integration overview
- `internal/components/mcp/context7.go` — Go adapter (do not modify from this skill)
