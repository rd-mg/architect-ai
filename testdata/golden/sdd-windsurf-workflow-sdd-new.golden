---
description: Initializes new feature via SDD in Hybrid-First mode for Cascade in Windsurf
---

# /sdd-new

Defines **Cascade** behavior for new features, medium/large scope changes, or high-uncertainty work.

## Purpose

Hybrid-First approach using Windsurf native capabilities:
- **Plan Mode** for planning
- **Memories / MCP (Engram)** for context retrieval
- **Artifacts `.sdd/`** as formal planning contract
- **Code Mode** only after explicit user approval

## When to Use

Activate when: new feature, multi-file/module change, architectural risk, explicit SDD request, or need for formal pre-code contract.
Skip for small/specific/minor maintenance tasks.

---

## Mandatory Operating Rules

### 1. Enter Plan Mode Immediately

1. Analyze request
2. Formulate high-level plan
3. Identify scope, risks, dependencies, likely files

Prohibited: writing production code, entering Code Mode, modifying application logic, partial implementation, assuming approval.

**This workflow is for formal planning, not execution.**

---

### 2. Retrieve Context First

Retrieve context before drafting any SDD artifact.

Priority order:
1. **Engram** MCP: `mem_search` for decisions, `mem_context` for recent context
2. Fallback: `AGENTS.md`
3. Incorporate any existing SDD/architecture context

Search for: architectural decisions, repo conventions, implementation constraints, quality rules, established patterns.

**Do not invent conventions.** If insufficient context, state it explicitly.

---

### 3. Create Formal Contract in `.sdd/`

Create `.sdd/` directory if needed. Generate these two mandatory files:

**`.sdd/proposal.md`** — minimum: title, problem, objective, included/excluded scope, approach, risks, assumptions, pending questions.

**`.sdd/spec.md`** — minimum: functional requirements, non-functional requirements, use cases, acceptance criteria, technical constraints, edge cases.

Artifacts must be: clear, reviewable, executable as implementation contract, consistent with project context.

---

### 4. Present Summary

Present brief summary: feature objective, proposed scope, main risks, confirmation of `.sdd/proposal.md` and `.sdd/spec.md` files. No wall of text.

---

### 5. Absolute Approval Gate

Once documents generated, **stop** and ask:

**Do you approve this implementation plan?**

Rules:
- MUST wait for explicit confirmation
- CANNOT proceed to Code Mode without approval
- CANNOT start implementation "in the meantime"
- CANNOT interpret silence as approval
- CANNOT skip this pause

Valid answers: "yes", "approved", "agreed", "go ahead", or equivalent explicit confirmation.

If changes requested: stay in Plan Mode, adjust `.sdd/` files, present plan again, re-ask approval.

---

## Strict Execution Sequence

1. Detect work warrants `/sdd-new`
2. Enter **Plan Mode**
3. Retrieve context via **Engram** (fallback: `AGENTS.md`)
4. Synthesize constraints, scope, risks
5. Create `.sdd/` if absent
6. Generate `.sdd/proposal.md`
7. Generate `.sdd/spec.md`
8. Present brief summary
9. Ask: **Do you approve this implementation plan?**
10. **Stop and wait**

---

## Explicit Prohibitions

Before user approval: NO production code, NO file edits, NO application tasks, NO Code Mode, NO commits, NO partial implementation, NO auto-continuation to next SDD step.

---

## Exit Criteria

Workflow correct only if: Plan Mode used, context retrieved via Engram or AGENTS.md, `.sdd/proposal.md` generated, `.sdd/spec.md` generated, summary presented, approval asked, stopped for explicit confirmation.
