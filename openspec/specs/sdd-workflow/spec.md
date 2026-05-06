# SDD Workflow Specification

## Requirements

### Requirement: Pre-mortem Risk Discovery (Propose Phase)
The `sdd-propose` phase MUST include a mandatory **Pre-mortem** section that identifies potential failure modes of the proposed change.
- The section MUST contain at least 2 unique risks.
- Each risk MUST have a Likelihood (Low/Med/High) and a Mitigation strategy.

#### Scenario: Proposal with Pre-mortem
- GIVEN a user requests a new change
- WHEN the `sdd-propose` agent generates the proposal
- THEN the output MUST contain a `## Pre-mortem` section with structured risks.

### Requirement: Viability Scoring
The `sdd-propose` phase MUST include a **Viability Score** (1-10) representing the agent's confidence in the implementation plan's success based on current codebase state.

#### Scenario: Viability assessment
- GIVEN an exploration of a complex refactor
- WHEN `sdd-propose` is executed
- THEN it MUST include a score (e.g., `Viability: 7/10`) and a brief justification.

### Requirement: Environment Pre-flight Validation (Init Phase)
The `sdd-init` phase MUST perform a **Pre-flight** validation of the environment before proceeding.
- It MUST check for presence of required CLI tools (e.g., `rg`, `git`).
- It MUST check for write permissions in the project root and `openspec/` (if applicable).
- It MUST check connectivity to critical MCP servers (Engram, Context7).

#### Scenario: Missing tool detection
- GIVEN a system where `ripgrep` is not installed
- WHEN `sdd-init` is executed
- THEN it MUST report a `Pre-flight Failure` and list `rg` as missing.

### Requirement: Finding Classification (Verify Phase)
The `sdd-verify` phase MUST classify every finding into one of three triage levels:
- **Blocking**: Must be fixed before merging/archiving (e.g., test failure, spec violation).
- **Warning**: Potential issue or technical debt that should be addressed (e.g., missing docstring, minor complexity).
- **Suggestion**: Non-critical improvement or alternative approach.

#### Scenario: Finding Triage
- GIVEN a verification run that finds a lint error and a broken test
- WHEN `sdd-verify` generates its report
- THEN the broken test MUST be marked as `Blocking` and the lint error MUST be marked as `Warning`.

### Requirement: Mandatory FMEA Table (Spec Phase)
The `sdd-spec` phase MUST include a **Failure Mode and Effects Analysis (FMEA)** table for every new or modified requirement that has behavioral impact.

#### Scenario: FMEA generation for new feature
- GIVEN a new capability `user-auth`
- WHEN `sdd-spec` generates the delta spec
- THEN it MUST include an `## FMEA` section with at least one entry.

### Requirement: Sad-path BDD Scenarios (Spec Phase)
For every requirement, the `sdd-spec` agent MUST include at least one **Sad-path** (error case) scenario using BDD syntax.

#### Scenario: Sad-path enforcement
- GIVEN a requirement for "Login Action"
- WHEN `sdd-spec` is executed
- THEN it MUST include a scenario for "Failed Login".

### Requirement: Task Dependency Graph (Tasks Phase)
The `sdd-tasks` phase MUST generate a **Dependency Graph** using Mermaid syntax at the beginning of the `tasks.md` document.

#### Scenario: Dependency graph generation
- GIVEN a task list with dependencies
- WHEN `sdd-tasks` generates `tasks.md`
- THEN it MUST include a `mermaid` code block representing the flow.

### Requirement: Mandatory Poka-Yoke Checklist (Design Phase)
The `sdd-design` phase MUST include a **Poka-Yoke (Mistake-Proofing)** section to prevent common architectural errors (State Invalidation, Resource Leaks, etc.).

#### Scenario: Design with Poka-Yoke
- GIVEN a design for a new component
- WHEN `sdd-design` generates `design.md`
- THEN it MUST include a `## Poka-Yoke Checklist`.

### Requirement: Adversarial Review Pass (Verify Phase)
The `sdd-verify` phase MUST include a distinct **Adversarial Review** pass to identify "False Positives" and missing test coverage.

#### Scenario: Verification with Adversarial Pass
- GIVEN a verification run
- WHEN `sdd-verify` is executed
- THEN the report MUST contain an `### Adversarial Findings` section.

### Requirement: Deviation Log (Archive Phase)
The `sdd-archive` phase MUST generate a **Deviation Log** documenting any changes made during implementation that differ from the design.

#### Scenario: Archive with Deviation Log
- GIVEN an archived change
- WHEN `sdd-archive` is executed
- THEN it MUST record any file-change or decision deviations in the `archive-report.md`.
