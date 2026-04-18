# PRD: OpenCode SDD Profiles

> **Create interchangeable model profiles for OpenCode — switch between orchestrator configurations with a single Tab.**

**Version**: 0.1.0-draft
**Author**: Gentleman Programming
**Date**: 2026-04-03
**Status**: Draft

---

## 1. Problem Statement

Today, OpenCode allows only ONE `sdd-orchestrator` with ONE set of models assigned to SDD sub-agents. This forces the user to choose between:

- **Maximum Quality** (Opus everywhere → expensive and slow)
- **Balance** (Opus orchestrator + Sonnet sub-agents → current default)
- **Economy** (Sonnet/Haiku everywhere → fast and cheap but less powerful)

The problem: **you cannot switch between these configurations without manually editing `opencode.json`** every time you want to change modes. In practice, a developer needs different profiles for different moments:

- **"I'm doing something heavy"** → Opus orchestrator, Sonnet sub-agents
- **"Simple task, don't want to burn tokens"** → Haiku everywhere
- **"Want to test a new Google model"** → Gemini orchestrator, mixed sub-agents
- **"Just reviewing a PR"** → Lightweight profile

Currently, this is a manual headache. This feature solves it.

---

## 2. Vision

**The user creates N model profiles from the TUI. Each profile generates its own `sdd-orchestrator-{name}` with its own sub-agents in `opencode.json`. In OpenCode, the user hits Tab and sees all available orchestrators — switching between profiles like shifting gears.**

```
┌─────────────────────────────────────────────────────────────┐
│  opencode.json                                               │
│                                                              │
│  ┌──────────────────────┐   ┌──────────────────────────────┐ │
│  │  sdd-orchestrator    │   │  sdd-orchestrator-cheap      │ │
│  │  (opus + sonnet)     │   │  (haiku everywhere)          │ │
│  │                      │   │                              │ │
│  │  sdd-init     sonnet │   │  sdd-init-cheap     haiku   │ │
│  │  sdd-explore  sonnet │   │  sdd-explore-cheap  haiku   │ │
│  │  sdd-apply    sonnet │   │  sdd-apply-cheap    haiku   │ │
│  │  ...                 │   │  ...                        │ │
│  └──────────────────────┘   └──────────────────────────────┘ │
│                                                              │
│  Tab in OpenCode → choose which orchestrator to use         │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Target Users

| User | Pain Point | How Profiles Help |
|------|-----------|-------------------|
| **Power user with multiple providers** | Wants to test Anthropic vs Google vs OpenAI for SDD without touching config | Create one profile per provider, switch with Tab |
| **Cost-conscious developer** | Wants a "cheap" mode for simple tasks | "cheap" profile with Haiku/Flash, "premium" profile with Opus |
| **Team lead** | Wants to standardize profiles for the team | Profiles live in `opencode.json`, syncable |
| **Experimenter** | Wants to test new models without breaking default config | Experimental profile, default remains intact |

---

## 4. Scope

### In Scope (v1)
- Profile creation from the TUI (new screen)
- List existing profiles
- **Edit existing profiles from the TUI** (select profile → modify models → sync)
- **Delete profiles from the TUI** (select profile → confirm → remove orchestrator + sub-agents from JSON → sync)
- Generation of N orchestrators + N×9 sub-agents in `opencode.json`
- Update existing profiles during Sync / Update+Sync
- Shared prompts: one file per phase, reused by all profiles
- CLI flag to create profiles (`--profile`)

### Out of Scope (permanently)
- **Profiles for Claude Code** — DOES NOT APPLY. Claude Code uses a completely different mechanism (CLAUDE.md + Task tool). The profiles feature is exclusive to OpenCode because it depends on the `opencode.json` agent/sub-agent system and Tab selection. This is NOT "future" — it is an architectural decision.

### Out of Scope (v1, future consideration)
- Export/import profiles between machines

---

## 5. Detailed Requirements

### 5.1 TUI: Profile Creation Screen

**R-PROF-01**: The Welcome screen MUST include a new option **"OpenCode SDD Profiles"** below "Configure Models".

**R-PROF-02**: If profiles already exist, the option MUST show the count: `"OpenCode SDD Profiles (2)"`.

**R-PROF-03**: The profiles screen MUST show existing profiles with available actions:

```
┌─────────────────────────────────────────────────────────┐
│  OpenCode SDD Profiles                                   │
│                                                          │
│  Existing profiles:                                      │
│    ✦ default ─── anthropic/claude-opus-4                 │
│    • cheap ───── anthropic/claude-haiku-3.5              │
│    • gemini ──── google/gemini-2.5-pro                   │
│                                                          │
│  ▸ Create new profile                                    │
│    Back                                                  │
│                                                          │
│  j/k: navigate • enter: edit • n: new • d: delete       │
│  esc: back                                               │
└─────────────────────────────────────────────────────────┘
```

**R-PROF-04**: When selecting "Create new profile" (or pressing `n`), the user MUST:
1. **Enter a profile name** (free text, validated to slug: lowercase, no spaces, alphanumeric + hyphens)
2. **Select orchestrator model** (reusing existing ModelPicker — provider → model)
3. **Select models for sub-agents** (reusing existing ModelPicker with 9+1 rows: Set all + 9 phases)
4. **Confirm** → profile is generated and sync is executed

**R-PROF-05**: The name "default" IS RESERVED for the base orchestrator (`sdd-orchestrator`). The user CANNOT create a profile named "default".

**R-PROF-06**: If the user enters an existing name, they MUST be asked if they want to overwrite.

### 5.1b TUI: Profile Editing

**R-PROF-07**: Pressing `enter` on an existing profile enters edit mode. The flow is IDENTICAL to creation but:
- The name CANNOT be changed (shown as fixed header)
- Orchestrator model is pre-selected with current value
- Sub-agent models are pre-selected with current values
- On confirmation, existing profile is overwritten and sync is executed

**R-PROF-07b**: The `default` profile CAN also be edited — it is the base `sdd-orchestrator`. Editing default is equivalent to what "Configure Models → OpenCode" does today but integrated into the profiles flow.

### 5.1c TUI: Profile Deletion

**R-PROF-08**: Pressing `d` on an existing profile MUST show a confirmation screen:

```
┌─────────────────────────────────────────────────────────┐
│  Delete Profile                                          │
│                                                          │
│  Are you sure you want to delete profile "cheap"?        │
│                                                          │
│  This will remove from opencode.json:                    │
│    • sdd-orchestrator-cheap                              │
│    • sdd-init-cheap                                      │
│    • sdd-explore-cheap                                   │
│    • ... (10 agents total)                               │
│                                                          │
│  ▸ Delete                                                │
│    Cancel                                                │
│                                                          │
│  enter: select • esc: cancel                             │
└─────────────────────────────────────────────────────────┘
```

**R-PROF-08b**: On confirmation:
1. ALL agent keys for the profile are removed from `opencode.json` (`sdd-orchestrator-{name}` + 10 sub-agents `sdd-{phase}-{name}`)
2. Atomic write of updated JSON
3. Result is shown (success/error)
4. Return to profile list

**R-PROF-08c**: The `default` profile CANNOT be deleted. Pressing `d` on default does nothing (keybinding ignored). The default is the base orchestrator that must always exist.

**R-PROF-08d**: Deleting a profile DOES NOT delete shared prompt files (`~/.config/opencode/prompts/sdd/*.md`) — these are shared by all profiles and remain as long as at least one profile exists.

### 5.2 Naming Convention

**R-PROF-10**: The DEFAULT profile (no suffix) generates agents with current names:
- `sdd-orchestrator`
- `sdd-init`, `sdd-explore`, `sdd-propose`, `sdd-spec`, `sdd-design`, `sdd-tasks`, `sdd-apply`, `sdd-verify`, `sdd-archive`

**R-PROF-11**: A profile named `cheap` generates agents with suffixes:
- `sdd-orchestrator-cheap`
- `sdd-init-cheap`, `sdd-explore-cheap`, ..., `sdd-archive-cheap`

**R-PROF-12**: The `sdd-orchestrator-{name}` MUST have `"mode": "primary"` to appear as selectable with Tab in OpenCode. Sub-agents `sdd-{phase}-{name}` MUST have `"mode": "subagent"` and `"hidden": true`.

**R-PROF-13**: Orchestrator permissions for a profile MUST be scoped to its own sub-agents:
```json
{
  "permission": {
    "task": {
      "*": "deny",
      "sdd-*-cheap": "allow"
    }
  }
}
```

### 5.3 Shared Prompt Architecture

**R-PROF-20**: Prompts for each SDD phase MUST live in separate files under `~/.config/opencode/prompts/sdd/`:
```
~/.config/opencode/prompts/sdd/
├── orchestrator.md
├── sdd-init.md
├── sdd-explore.md
├── sdd-propose.md
├── sdd-spec.md
├── sdd-design.md
├── sdd-tasks.md
├── sdd-apply.md
├── sdd-verify.md
├── sdd-archive.md
└── sdd-onboard.md
```

**R-PROF-21**: The `prompt` for each agent in opencode.json MUST reference the shared file using OpenCode `{file:path}` syntax:
```json
{
  "sdd-apply": {
    "mode": "subagent",
    "hidden": true,
    "model": "anthropic/claude-sonnet-4-20250514",
    "prompt": "{file:~/.config/opencode/prompts/sdd/sdd-apply.md}"
  },
  "sdd-apply-cheap": {
    "mode": "subagent",
    "hidden": true,
    "model": "anthropic/claude-haiku-3.5-20241022",
    "prompt": "{file:~/.config/opencode/prompts/sdd/sdd-apply.md}"
  }
}
```

**R-PROF-22**: Content of these prompt files MUST be EXACTLY the same as what is currently inlined in the JSON overlay. The refactor is an extraction without behavioral change.

**R-PROF-23**: The orchestrator prompt (`orchestrator.md`) MUST include an `<!-- architect-ai:sdd-model-assignments -->` block that is dynamically injected with that specific profile's model table.

**R-PROF-24**: For a profile's orchestrator, the prompt MUST reference sub-agents WITH SUFFIX. 

**Architectural Decision**: The orchestrator prompt is NOT shared between profiles — each profile generates its own version with:
- That profile's model assignments table
- Correct `sdd-{phase}-{suffix}` references

Sub-agent prompts ARE shared because they are identical across profiles (only the model changes, not the prompt).

### 5.4 Sync & Update Behavior

**R-PROF-30**: During `Sync` or `Update+Sync`, the system MUST:
1. Detect ALL existing profiles in `opencode.json` (pattern: `sdd-orchestrator-*`)
2. Update shared prompt files in `~/.config/opencode/prompts/sdd/`
3. Regenerate orchestrator prompts for each profile (to inject updated model assignments)
4. NOT modify model assignments of profiles — only prompts

**R-PROF-31**: If a profile has a sub-agent referencing a model that no longer exists in OpenCode's cache, Sync MUST:
- **Warning** to user (not error)
- Preserve existing assignment (user might have configured it manually)

**R-PROF-32**: Shared prompt files MUST be covered by pre-sync backup, same as `opencode.json`.

**R-PROF-33**: Sync MUST be idempotent: if prompts are already updated, `filesChanged` MUST NOT increment.

### 5.5 Profile Detection & State

**R-PROF-40**: Profiles MUST be detected by reading existing `opencode.json`, NOT from a separate state file. `opencode.json` IS the source of truth.

**R-PROF-41**: A profile is detected by the presence of an agent key matching `sdd-orchestrator-{name}` with `"mode": "primary"`.

**R-PROF-42**: Upon detecting existing profiles, the system MUST infer:
- **Name**: suffix after `sdd-orchestrator-`
- **Orchestrator model**: `"model"` field of the orchestrator
- **Sub-agent models**: `"model"` fields of `sdd-{phase}-{name}`

**R-PROF-43**: The default profile (`sdd-orchestrator` without suffix) ALWAYS exists when SDD is configured. Additional profiles are optional.

### 5.6 CLI Support

**R-PROF-50**: `sync` command MUST accept a `--profile <name>:<orchestrator-model>` flag that creates/updates a profile during sync:
```bash
architect-ai sync --profile cheap:anthropic/claude-haiku-3.5-20241022
```

**R-PROF-51**: Multiple `--profile` flags MUST be supported:
```bash
architect-ai sync \
  --profile cheap:anthropic/claude-haiku-3.5-20241022 \
  --profile premium:anthropic/claude-opus-4-20250514
```

**R-PROF-52**: Flag format is `name:provider/model`. To assign individual models to sub-agents via CLI, extended syntax is used:
```bash
architect-ai sync --profile cheap:anthropic/claude-haiku-3.5-20241022 \
  --profile-phase cheap:sdd-apply:anthropic/claude-sonnet-4-20250514
```

---

## 6. Technical Design

### 6.1 Data Model

```go
// Profile represents a named SDD orchestrator configuration with model assignments.
type Profile struct {
    Name                string                       // e.g. "cheap", "premium"
    OrchestratorModel   model.ModelAssignment         // orchestrator model
    PhaseAssignments    map[string]model.ModelAssignment // per-phase models (optional overrides)
}
```

### 6.2 OpenCode JSON Structure (per profile)

For a profile named "cheap" with Haiku:

```json
{
  "agent": {
    "sdd-orchestrator-cheap": {
      "mode": "primary",
      "description": "SDD Orchestrator (cheap profile) — haiku everywhere",
      "model": "anthropic/claude-haiku-3.5-20241022",
      "prompt": "... orchestrator prompt with cheap-specific model table and sub-agent references ...",
      "permission": {
        "task": {
          "*": "deny",
          "sdd-*-cheap": "allow"
        }
      },
      "tools": {
        "read": true,
        "write": true,
        "edit": true,
        "bash": true,
        "delegate": true,
        "delegation_read": true,
        "delegation_list": true
      }
    },
    "sdd-init-cheap": {
      "mode": "subagent",
      "hidden": true,
      "model": "anthropic/claude-haiku-3.5-20241022",
      "description": "Bootstrap SDD context (cheap profile)",
      "prompt": "{file:~/.config/opencode/prompts/sdd/sdd-init.md}"
    },
    "sdd-explore-cheap": {
      "mode": "subagent",
      "hidden": true,
      "model": "anthropic/claude-haiku-3.5-20241022",
      "description": "Investigate codebase (cheap profile)",
      "prompt": "{file:~/.config/opencode/prompts/sdd/sdd-explore.md}"
    }
    // ... remaining 7 sub-agents with -cheap suffix
  }
}
```

### 6.3 Prompt File Architecture

```
~/.config/opencode/
├── opencode.json          (agents with model + prompt refs)
├── prompts/
│   └── sdd/
│       ├── sdd-init.md        (shared by all profiles)
│       ├── sdd-explore.md     (shared by all profiles)
│       ├── sdd-propose.md     (shared)
│       ├── sdd-spec.md        (shared)
│       ├── sdd-design.md      (shared)
│       ├── sdd-tasks.md       (shared)
│       ├── sdd-apply.md       (shared)
│       ├── sdd-verify.md      (shared)
│       ├── sdd-archive.md     (shared)
│       └── sdd-onboard.md     (shared)
├── skills/                (existing SDD skills)
├── commands/              (existing slash commands)
├── plugins/               (existing plugins)
└── plugins/               (existing plugins)
```

**Key insight**: Orchestrator prompts ARE NOT shared as external files because each profile needs its own model assignments table and sub-agent references with suffixes. They are inlined in each orchestrator's JSON during generation.

Sub-agent prompts ARE shared as `{file:...}` files because they are identical between profiles — only the `"model"` field changes.

### 6.4 Affected Files (Implementation Map)

| Area | File | Changes |
|------|------|---------|
| **Domain model** | `internal/model/types.go` | Add `Profile` struct |
| **Domain model** | `internal/model/selection.go` | Add `Profiles []Profile` to `Selection` and `SyncOverrides` |
| **TUI: screens** | `internal/tui/screens/profiles.go` | NEW — profile list screen (list + edit + delete actions) |
| **TUI: screens** | `internal/tui/screens/profile_create.go` | NEW — profile creation/edit flow (name → models → confirm) |
| **TUI: screens** | `internal/tui/screens/profile_delete.go` | NEW — profile delete confirmation screen |
| **TUI: model** | `internal/tui/model.go` | Add `ScreenProfiles`, `ScreenProfileCreate`, `ScreenProfileEdit`, `ScreenProfileDelete`, `ScreenProfileResult` |
| **TUI: router** | `internal/tui/router.go` | Add routes for all profile screens |
| **TUI: welcome** | `internal/tui/screens/welcome.go` | Add "OpenCode SDD Profiles" option |
| **SDD inject** | `internal/components/sdd/inject.go` | Extract prompts to files, generate profile agents |
| **SDD inject** | `internal/components/sdd/profiles.go` | NEW — profile CRUD: generate, detect, delete agents from JSON |
| **SDD inject** | `internal/components/sdd/prompts.go` | NEW — shared prompt file management |
| **SDD inject** | `internal/components/sdd/read_assignments.go` | Add profile detection from opencode.json |
| **Sync** | `internal/cli/sync.go` | Update sync to handle profiles, add `--profile` flag |
| **Assets** | `internal/assets/opencode/sdd-overlay-multi.json` | Refactor to use `{file:...}` references |
| **OpenCode models** | `internal/opencode/models.go` | No changes (reuse existing) |

### 6.5 Sync Flow (Updated)

```
Sync Start
  │
  ├─ 1. Read opencode.json → detect existing profiles
  │     (pattern: sdd-orchestrator-*)
  │
  ├─ 2. Write/update shared prompt files
  │     ~/.config/opencode/prompts/sdd/*.md
  │     (from embedded assets, same as today's inline prompts)
  │
  ├─ 3. Update DEFAULT orchestrator + sub-agents
  │     (sdd-orchestrator, sdd-init, ..., sdd-archive)
  │     - Update prompts (inline for orchestrator, {file:} for sub-agents)
  │     - Preserve model assignments
  │
  ├─ 4. For EACH detected profile:
  │     ├─ Update sub-agent prompts (they use {file:}, auto-updated in step 2)
  │     ├─ Regenerate orchestrator prompt (inline, with profile's model table)
  │     └─ Preserve model assignments
  │
  └─ 5. Verify: all profile orchestrators + sub-agents present
```

### 6.6 Migration Path

**Backward compatibility**: Users without profiles see no changes. The refactor of prompts to files is transparent:

1. **First sync after update**: 
   - Creates `~/.config/opencode/prompts/sdd/` directory
   - Writes prompt files
   - Migrates sub-agents in overlay from inline prompt to `{file:...}` reference
   - Result: identical behavior, only prompt location changes

2. **Users with existing multi-mode**:
   - Model assignments are preserved
   - Sub-agents are automatically migrated to `{file:...}`
   - Zero disruption

---

## 7. UX Flow

### 7.1 Welcome Screen (Updated)

```
┌─────────────────────────────────────────────────────────┐
│                                                          │
│  ★  Gentleman AI Ecosystem — v0.x.x                     │
│     Supercharge your AI agents.                          │
│                                                          │
│  ▸ Install Ecosystem                                     │
│    Update                                                │
│    Sync                                                  │
│    Update + Sync                                         │
│    Configure Models                                      │
│    OpenCode SDD Profiles (2)                     ← NEW   │
│    Manage Backups                                        │
│    Quit                                                  │
│                                                          │
│  j/k: navigate • enter: select • q: quit                │
└─────────────────────────────────────────────────────────┘
```

### 7.2 Profile List Screen

```
┌─────────────────────────────────────────────────────────┐
│  OpenCode SDD Profiles                                   │
│                                                          │
│  Your SDD model profiles for OpenCode. Each profile      │
│  creates its own orchestrator (visible with Tab).        │
│                                                          │
│  Existing profiles:                                      │
│    ✦ default ─── anthropic/claude-opus-4                 │
│  ▸   cheap ───── anthropic/claude-haiku-3.5              │
│      gemini ──── google/gemini-2.5-pro                   │
│                                                          │
│    Create new profile                                    │
│    Back                                                  │
│                                                          │
│  j/k: navigate • enter: edit • n: new • d: delete       │
│  esc: back                                               │
└─────────────────────────────────────────────────────────┘
```

Profiles are navigable items. The cursor can be on a profile OR on "Create new profile" / "Back":
- **enter on a profile** → edit mode (modify models, then sync)
- **d on a profile** → delete confirmation (except default)
- **enter on "Create new profile"** → creation flow
- **n anywhere** → shortcut for "Create new profile"

### 7.3 Profile Edit Flow

Identical to creation but with pre-populated values:

```
┌─────────────────────────────────────────────────────────┐
│  Edit Profile "cheap"                                    │
│                                                          │
│  Current orchestrator: anthropic/claude-haiku-3.5        │
│                                                          │
│  ▸ Change orchestrator model                             │
│    Change sub-agent models                               │
│    Save & Sync                                           │
│    Cancel                                                │
│                                                          │
│  j/k: navigate • enter: select • esc: cancel            │
└─────────────────────────────────────────────────────────┘
```

### 7.4 Profile Delete Flow

```
┌─────────────────────────────────────────────────────────┐
│  Delete Profile                                          │
│                                                          │
│  Are you sure you want to delete profile "cheap"?        │
│                                                          │
│  This will remove from opencode.json:                    │
│    • sdd-orchestrator-cheap                              │
│    • sdd-init-cheap ... sdd-archive-cheap                │
│    • (11 agents total)                                   │
│                                                          │
│  ▸ Delete & Sync                                         │
│    Cancel                                                │
│                                                          │
│  enter: select • esc: cancel                             │
└─────────────────────────────────────────────────────────┘
```

### 7.5 Profile Creation Flow

```
Step 1: Name
┌─────────────────────────────────────────────────────────┐
│  Create SDD Profile                                      │
│                                                          │
│  Profile name: cheap_                                    │
│                                                          │
│  (lowercase, hyphens allowed, no spaces)                 │
│  Reserved: "default"                                     │
│                                                          │
│  enter: confirm • esc: cancel                            │
└─────────────────────────────────────────────────────────┘

Step 2: Orchestrator Model
┌─────────────────────────────────────────────────────────┐
│  Profile "cheap" — Select Orchestrator Model             │
│                                                          │
│  ▸ anthropic                                             │
│    google                                                │
│    openai                                                │
│    Back                                                  │
│                                                          │
│  (reuses existing ModelPicker)                           │
└─────────────────────────────────────────────────────────┘

Step 3: Sub-agent Models
┌─────────────────────────────────────────────────────────┐
│  Profile "cheap" — Assign Sub-agent Models               │
│                                                          │
│  ▸ Set all phases ──── (none)                            │
│    sdd-init ────────── (none)                            │
│    sdd-explore ─────── (none)                            │
│    sdd-propose ─────── (none)                            │
│    sdd-spec ─────────── (none)                           │
│    sdd-design ──────── (none)                            │
│    sdd-tasks ────────── (none)                           │
│    sdd-apply ────────── (none)                           │
│    sdd-verify ──────── (none)                            │
│    sdd-archive ─────── (none)                            │
│    Continue                                              │
│    Back                                                  │
│                                                          │
│  (reuses existing ModelPicker with provider/model drill) │
└─────────────────────────────────────────────────────────┘

Step 4: Confirm + Sync
┌─────────────────────────────────────────────────────────┐
│  Profile "cheap" — Ready to Create                       │
│                                                          │
│  Orchestrator: anthropic/claude-haiku-3.5-20241022      │
│  Sub-agents:   anthropic/claude-haiku-3.5-20241022 (all)│
│                                                          │
│  This will:                                              │
│  • Add sdd-orchestrator-cheap to opencode.json           │
│  • Add 10 sub-agents (sdd-init-cheap ... sdd-archive-cheap) │
│  • Run sync to apply changes                             │
│                                                          │
│  ▸ Create & Sync                                         │
│    Cancel                                                │
│                                                          │
│  enter: select • esc: cancel                             │
└─────────────────────────────────────────────────────────┘
```

---

## 8. Edge Cases & Decisions

### 8.1 OpenCode Model Cache Not Available

If `~/.cache/opencode/models.json` does not exist (OpenCode has never run), profile creation MUST:
- Show explanatory message: "Run OpenCode at least once to populate the model cache"
- Offer only "Back"
- NOT block the rest of the TUI

### 8.2 Profile Name Validation

| Input | Valid? | Reason |
|-------|--------|--------|
| `cheap` | ✓ | Simple slug |
| `premium-v2` | ✓ | Hyphens allowed |
| `my profile` | ✗ | Spaces not allowed |
| `default` | ✗ | Reserved |
| `LOUD` | → `loud` | Auto-lowercased |
| `sdd-orchestrator` | ✗ | Would create `sdd-orchestrator-sdd-orchestrator` — confusing |
| `a` | ✓ | Minimum 1 char |
| (empty) | ✗ | Must have a name |

### 8.3 Model Inheritance for Sub-agents

When a sub-agent doesn't have an explicit model assignment:
1. Use orchestrator model from the same profile
2. If orchestrator model is not set, use root `"model"` from `opencode.json`
3. If nothing is set, OpenCode uses its default

### 8.4 Deleting a Profile

Deletion is fully supported from the TUI (press `d` on a profile → confirm → agents removed from JSON → sync). Operation:
1. Reads `opencode.json`
2. Removes ALL keys matching `sdd-orchestrator-{name}` and `sdd-{phase}-{name}` (11 keys total)
3. Writes updated JSON atomically
4. Runs sync to ensure consistency
5. `default` profile CANNOT be deleted — keybinding ignored

### 8.5 Orchestrator Prompt — Sub-agent References

Default profile orchestrator prompt references sub-agents as `sdd-apply`. A "cheap" profile needs its orchestrator to reference `sdd-apply-cheap`. 

**Solution**: When generating a profile's orchestrator prompt, string replacement of pattern `sdd-{phase}` → `sdd-{phase}-{suffix}` is done ONLY within sections referencing sub-agents (Model Assignments table, delegation rules). This happens at generation time, not in the shared file.

---

## 9. Success Metrics

| Metric | Target |
|--------|--------|
| Profile creation time (TUI) | < 60 seconds |
| Sync time with 3 profiles | < 5 seconds additional |
| Zero regression on users without profiles | 100% backward compatible |
| Profile count supported | Tested up to 10 |
| Files changed per sync (no actual changes) | 0 (idempotent) |

---

## 10. Implementation Phases

### Phase 1: Shared Prompt Refactor (Foundation)
- Extract sub-agent prompts to `~/.config/opencode/prompts/sdd/*.md`
- Update `sdd-overlay-multi.json` to use `{file:...}` references
- Update `inject.go` to write prompt files
- Update sync to maintain prompt files
- **Zero behavioral change** — same prompts, different location

### Phase 2: Profile Data Model & Generation
- Add `Profile` type to domain model
- Implement profile agent generation (orchestrator + sub-agents with suffix)
- Profile detection from existing `opencode.json`
- Update `injectModelAssignments` to handle multiple profiles

### Phase 3: TUI Screens — Create & List
- Profile list screen (shows existing profiles with actions)
- Profile creation flow (name → orchestrator model → sub-agent models → confirm)
- Wire into Welcome screen
- Integrate with sync flow (auto-sync after profile creation)

### Phase 4: TUI Screens — Edit & Delete
- Profile edit flow (select profile → modify models → save & sync)
- Profile delete confirmation screen + JSON cleanup
- `d` keybinding on profile list for delete
- `enter` keybinding on profile for edit
- Default profile protection (no delete, yes edit)

### Phase 5: Sync Integration
- Update sync to detect and maintain all profiles
- Add `--profile` CLI flag
- Update backup targets to include prompt files
- Update post-sync verification for profiles

### Phase 6: Polish & Testing
- E2E tests for profile creation, edit, delete + sync
- Edge case handling (missing cache, invalid names, etc.)
- Documentation update

---

## 11. Open Questions

1. **Is each profile's orchestrator prompt inlined in the JSON or saved as a file?**
   → Decision: INLINED in JSON. Orchestrator prompt is profile-specific (model table + sub-agent references), cannot be shared as a file. Sub-agent prompts ARE shared as files.

2. **What happens to `sdd-onboard` in profiles?**
   → Decision: `sdd-onboard-{name}` is generated as a sub-agent of the profile, just like the other 9 sub-agents.

3. **Do SDD slash commands (`/sdd-new`, `/sdd-ff`, etc.) work with custom profiles?**
   → Yes. Commands are bound to the orchestrator. When the user selects `sdd-orchestrator-cheap` with Tab, commands run against that orchestrator which delegates to `sdd-*-cheap` sub-agents.

4. **How does OpenCode handle `{file:...}` in prompts? Does it support `~` expansion?**
   → Validate with OpenCode docs. If not supported, use expanded absolute path during generation.

5. **Does the `gentleman` agent (persona) also need per-profile variants?**
   → No. The `gentleman` agent is the general persona, not part of SDD. It only mirrors the default orchestrator's model.
