<!-- architect-ai:prompt-caching-anchor:start -->
# SDD Phase Protocol: sdd-init
Project: architect-ai
Adapter: Antigravity
Version: 1.1
<!-- architect-ai:prompt-caching-anchor:end -->

## Dependencies
- **Reads**: project files for detection
- **Writes**: `sdd-init` artifact (project context)

## Cognitive Posture
None — detection and configuration.

## Model
sonnet — needs to read project structure intelligently

## When Triggered
- User invokes `/sdd-init` explicitly
- Orchestrator auto-runs when `mem_search(query: "sdd-init/{project}")` returns nothing before any other SDD command

## Sub-Agent Launch Template

```
## Project Standards (auto-resolved)
{matching compact rules — will likely be empty on first init}

## Available Tools
{verified tool list}

## Phase: sdd-init

Task: Detect project context and persist for the session.

## Detection Procedure

### Step 0: Pre-flight Validation (MANDATORY)
Verify the project's health before initializing SDD:
1. **Build Health**: Check if the project can build/compile.
2. **Test Baseline**: Run the existing test suite and record current failures.
3. **Linter**: Check for critical linting errors.
4. **Environment**: Verify required environment variables and dependencies.
5. **CI status**: (Optional) Check status of latest CI run.

Return a `pre-flight-report` in the artifact content. If critical failures found, set status to `blocked`.

### Steps 1-6
1. Project name: use the directory name or package.json/manifest name
2. Language/framework: detect from build files
3. Test runner: detect from scripts in build files
4. Artifact store mode: default to `engram` if Engram available, else `none`
5. Strict TDD capability: check config and test runners
6. Active overlays: check for `.atl/overlays/` directory

### Step 6b: Overlay Registration

IF Odoo project detected in Steps 1-6 (pyproject.toml contains "odoo" OR
any `__manifest__.py` found in addons/ directory):

1. Determine Odoo version from:
   - `pyproject.toml` `[tool.odoo]` section, OR
   - `__manifest__.py` `"version"` field prefix (e.g., "17.0.x.y.z" → v17), OR
   - ODOO_VERSION environment variable

2. Create overlay manifest:
   ```
   mkdir -p .atl/overlays/odoo-{version}
   write .atl/overlays/odoo-{version}/manifest.json:
   {
     "overlay": "odoo-development-skill",
     "version": "{detected_version}",
     "active": true,
     "addons_path": "{detected_addons_path}",
     "detected_at": "{ISO_8601_timestamp}"
   }
   ```

3. Record IS_ODOO = true in sdd-init artifact (under `active_overlays` key)

4. Notify user:
   "Odoo {version} detected. Overlay activated. Phase protocols will include Odoo-specific rules."

IF NOT Odoo project:
   Skip this step entirely.

### Step 7: Test Baseline Persistence
Record the current test state to establish a baseline for verification:
- `mem_save(topic_key: "sdd/{project}/test-baseline")` with the test output summary.

### Step 8: Surface Mapping
Identify entry points and public APIs affected by this change:
- **UI Entry Points**: URLs, buttons, forms, components.
- **API Entry Points**: Endpoints, webhooks, CLI commands.
- **Data Entry Points**: DB schemas, message queues, file watchers.

Record in the sdd-init artifact.


## Empirical Verification Loop (+++Empirical)
- **MANDATORY**: Before concluding, you MUST perform an empirical verification of your findings/artifacts.
- Examples: run a script, check a file, verify a tool output, or perform a manual check of the logic.
- Record the evidence in the `empirical_proof` field of the return handshake.

## Artifact Store: {mode}

## Persistence (MANDATORY)
mem_save(
  title: "sdd-init/{project}",
  topic_key: "sdd-init/{project}",
  type: "project-context",
  project: "{project}",
  content: "Project: {name}\nLanguage: {lang}\nFramework: {framework}\nTest runner: {cmd}\nArtifact mode: {mode}\nStrict TDD: {true|false}\nActive overlays: {list}\nPre-flight: {report summary}\nInit date: {date}"
)

## Return Envelope & Compliance per sdd-phase-common.md (Sections A-F)
```

## Result Processing

- Cache project context for the session
- Orchestrator uses this for all subsequent phase delegations
- Update state: `uninitialized` → `idle`

## Failure Handling

- If project root cannot be determined → return `blocked`, ask user to run from project root
- If detection is ambiguous → return `partial`, ask user specific questions
- If Engram is unavailable → fall back to `none` mode silently, note in return envelope
