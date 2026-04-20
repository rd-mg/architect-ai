# Cognitive Mode Specification

## Purpose
Define the behavioral postures for Architect-AI agents, ranging from investigative to purely tactical execution.

## Requirements

### Requirement: Posture #7 — Caveman (Minimalist Execution)
The system MUST activate Posture #7 (Caveman) when context pressure is critical or when performing high-volume, execution-only tasks. Communication MUST be restricted to a four-field status block.

#### Scenario: High Context Pressure Activation
- GIVEN a session with Context Pressure (D4) = 3 (Contexto > 80%)
- WHEN a sub-agent is launched
- THEN the `+++Caveman` posture MUST be injected
- AND the sub-agent MUST only respond with STATUS/ACTION/RESULT/NEXT fields.

#### Scenario: Execution-only Task
- GIVEN a sub-agent assigned to a mechanical task (e.g., creating 5 files)
- WHEN it completes the task
- THEN it MUST use the Caveman format to report success, avoiding narrative explanations.

### Requirement: Posture #8 — Autoreason-lite (Autonomous Micro-Conflict Resolution)
The system MUST provide a posture for resolving low-complexity conflicts (e.g., Engram collisions or minor execution errors) without user intervention or full diagnostic mode.

#### Scenario: Engram Collision Resolution
- GIVEN a sub-agent detects a conflict between its task and a "Decision" Engram
- WHEN Posture #8 is active
- THEN it MUST formulate 3 hypotheses (H1-H3)
- AND it MUST execute the most probable fix (H1) automatically.

#### Scenario: Initial Execution Error (Error Pressure = 1)
- GIVEN a sub-agent fails a command for the first time
- WHEN Autoreason-lite is active
- THEN it MUST attempt a quick fix (e.g., path adjustment or import correction)
- AND it MUST NOT escalate to the user unless the fix fails.

## Invariants
- Caveman posture MUST NOT restrict internal reasoning depth, only external output.
- Autoreason-lite MUST escalate to Mode 3 (Diagnostic) if the initial fix fails.
