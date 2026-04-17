---
description: Initializes a new feature or medium/large task using SDD in Hybrid-First mode for Cascade in Windsurf
---

# /sdd-new

This workflow defines the mandatory behavior of **Cascade** when starting a new feature, medium/large scope change, or work with enough uncertainty to require formal planning.

## Purpose

Use Windsurf's native capabilities in a **Hybrid-First** way:

- **Plan Mode** for planning
- **Memories / MCP (Engram)** to retrieve previous context
- **Artifacts `.sdd/`** only as a formal planning contract
- **Code Mode** only after explicit user approval

## When to use this workflow

Activate this workflow when any of these conditions occur:

- The user starts a **new feature**
- The task affects **multiple files or modules**
- The change has **architectural risk** or uncertainty
- The user explicitly asks to work with **SDD**
- The implementation requires a formal contract before writing code

If the task is small, specific, or clearly minor maintenance, this workflow is NOT the correct path.

---

## Mandatory Operating Rules

### 1. Switch immediately to Plan Mode

When starting this workflow, **you MUST enter Plan Mode immediately**.

Mandatory actions:

1. Analyze the user's request
2. Formulate a high-level plan
3. Identify scope, risks, dependencies, and likely files

Prohibited actions at this stage:

- DO NOT write production code
- DO NOT enter Code Mode
- DO NOT modify application logic
- DO NOT execute partial implementation "to get ahead"
- DO NOT assume implicit approval

**This workflow is for formal planning, not execution.**

---

### 2. Retrieve context before proposing anything

Before drafting any SDD artifact, **you MUST retrieve architectural context and project constraints**.

Order of preference:

1. Use **Engram** via canonical MCP tools: `mem_search` to find previous decisions and `mem_context` to retrieve recent project context
2. If Engram is not available or does not return enough context, read `AGENTS.md`
3. If additional project context related to SDD or architecture exists, incorporate it as well

You must search, at a minimum, for:

- Previous architectural decisions
- Repository conventions
- Implementation constraints
- Quality or review rules
- Established patterns for similar changes

If you don't find enough context, you must state it explicitly in the plan. **Do not invent conventions.**

---

### 3. Create the initial formal contract in `.sdd/`

You must create the `.sdd/` directory if it doesn't exist.

Then you must generate exactly these two initial files:

- `.sdd/proposal.md`
- `.sdd/spec.md`

In this phase, those two files are **mandatory**.

#### Minimum content of `.sdd/proposal.md`

It must capture, at a minimum:

- Change title
- Problem to solve
- Objective
- Included scope
- Excluded scope
- Proposed approach
- Main risks
- Open assumptions
- Pending questions or decisions

#### Minimum content of `.sdd/spec.md`

It must capture, at a minimum:

- Functional requirements
- Non-functional requirements if applicable
- Use cases
- Acceptance criteria
- Relevant technical constraints
- Known edge cases or important assumptions

Artifacts must be:

- Clear
- Reviewable
- Executable as an implementation contract
- Consistent with retrieved project context

---

### 4. Present planning summary to the user

After creating `.sdd/proposal.md` and `.sdd/spec.md`, you must present a brief and clear summary in chat.

That summary must include:

- Feature objective
- Proposed scope
- Main risks or doubts
- Confirmation that the files were created:
  - `.sdd/proposal.md`
  - `.sdd/spec.md`

Do not show an unnecessary wall of text. Summarize the essentials for review.

---

### 5. Absolute Approval Gate

Once the documents are generated, you must **ABSOLUTELY stop**.

You must ask **exactly**:

**Do you approve this implementation plan?**

Then:

- You **must wait for explicit confirmation**
- You CANNOT proceed to Code Mode without approval
- You CANNOT start implementation "in the meantime"
- You CANNOT interpret silence as approval
- You CANNOT replace this pause with an informal summary

Valid answers to continue:

- "yes"
- "approved"
- "agreed"
- "go ahead"
- any equivalent explicit confirmation

If the user requests changes:

- You must stay in Plan Mode
- You must adjust `.sdd/proposal.md` and/or `.sdd/spec.md`
- You must present the plan again
- You must ask again: **Do you approve this implementation plan?**

---

## Strict Execution Sequence

Follow this sequence without skips:

1. Detect that the work warrants `/sdd-new`
2. Enter **Plan Mode**
3. Retrieve context with **Engram** or, failing that, read `AGENTS.md`
4. Synthesize constraints, scope, and risks
5. Create `.sdd/` if it doesn't exist
6. Generate `.sdd/proposal.md`
7. Generate `.sdd/spec.md`
8. Present a brief summary to the user
9. Ask exactly: **Do you approve this implementation plan?**
10. **Stop and wait for a response**

---

## Explicit Prohibitions

While this workflow has not been approved by the user:

- DO NOT write production code
- DO NOT edit implementation files
- DO NOT execute application tasks
- DO NOT switch to Code Mode
- DO NOT create commits
- DO NOT run a partial implementation
- DO NOT continue automatically to the next SDD step

---

## Exit Criteria for this workflow

This workflow is considered correctly executed only if:

- Cascade used **Plan Mode**
- Context was retrieved with **Engram** or `AGENTS.md`
- Generated `.sdd/proposal.md`
- Generated `.sdd/spec.md`
- Presented a summary to the user
- Asked exactly: **Do you approve this implementation plan?**
- Stopped to wait for explicit approval

If any of those points do not occur, the workflow is poorly executed.
