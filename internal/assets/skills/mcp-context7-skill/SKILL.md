---
name: mcp-context7-skill
description: >
  Uses the existing Context7 MCP component to resolve libraries and fetch
  current framework or package documentation without relying on stale model
  memory. Trigger: When the user asks for official docs, API syntax, version
  limits, migration guidance, or framework examples grounded in upstream
  documentation.
license: Apache-2.0
metadata:
  author: rd-mg
  version: "1.0"
---

## Purpose
Use Context7 as a live documentation skill, not as an installer. This skill
assumes the agent already has access to the existing `context7` component and
focuses on choosing the right library target, fetching only the relevant
documentation, and turning it into actionable guidance.

## When to Use
- User asks for current library or framework documentation.
- User needs API syntax, migration notes, or version-specific behavior.
- User asks for examples that should be grounded in upstream docs rather than
  cached model memory.
- User wants a documentation-backed answer before writing or changing code.
- Do not use when: the answer is fully contained in the current repository,
  the question is general web research, or Context7 is unavailable and local
  docs already answer the question.

## Critical Patterns
- Treat this skill as a consumer of the existing `context7` component. Do not
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
- The package, framework, or library the user is asking about.
- Optional version or ecosystem details that narrow the lookup.
- The question to answer or the code change being planned.
- The live Context7 tool surface exposed by the active host.

## Outputs
- One documentation-backed answer scoped to the user request.
- One short explanation of any uncertainty caused by missing version or target
  information.
- One explicit fallback when Context7 cannot answer or is not available.

## Steps
1. Confirm whether the request actually needs upstream documentation.
2. Identify the exact library or framework target.
3. Use the active Context7 tool surface to resolve the correct target before
   requesting documentation excerpts.
4. Fetch only the documentation sections needed for the current question.
5. Answer with the relevant API facts, constraints, and version notes.
6. If implementation guidance is needed, derive it from the retrieved docs and
   label any inference separately from quoted behavior.
7. If Context7 is unavailable or the target cannot be resolved safely, say so
   and fall back to repository evidence or ask the user for clarification.

## Guardrails
- Do not invent Context7 tool names or payload fields.
- Do not install or patch MCP config from this skill.
- Do not assume a domain, framework family, or repository-specific default.
- Do not treat stale model memory as equivalent to retrieved documentation.
- Do not answer version-sensitive questions without stating the version basis.

## Resources
- `docs/components.md`
- `docs/intended-usage.md`
- `internal/components/mcp/context7.go`