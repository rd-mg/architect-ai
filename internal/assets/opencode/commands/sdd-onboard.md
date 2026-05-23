---
description: Guided SDD walkthrough — onboard user through full SDD cycle using their real codebase
agent: sdd-orchestrator
subtask: true
---

SDD sub-agent. Read the skill file at ~/.config/opencode/skills/sdd-onboard/SKILL.md FIRST, then follow its instructions exactly.

CONTEXT:
- Working directory: !`echo -n "$(pwd)"`
- Current project: !`echo -n "$(basename $(pwd))"`
- Artifact store mode: engram

TASK:
Guide user through complete SDD cycle using their actual codebase. Real change with real artifacts, not a toy example. Goal: teach by doing — walk through exploration, proposal, spec, design, tasks, apply, verify, archive.

ENGram PERSISTENCE (artifact store mode: engram):
Save onboarding progress as you go:
  mem_save(title: "sdd-onboard/{project}", topic_key: "sdd-onboard/{project}", type: "architecture", project: "{project}", content: "{onboarding state}")
topic_key enables upserts — re-running updates, not duplicates.

Return structured result with: status, executive_summary, artifacts, and next_recommended.
