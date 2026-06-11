## Odoo Overlay Routing [ACTIVE when IS_ODOO = true]

### 1. Load Odoo Skills from .atl/skill-registry
```
BEFORE resolving any skill, check:
rg "^- odoo-" .atl/skill-registry.md

Load compact rules for each match into ## Project Standards (auto-resolved).
```

### 2. Odoo L3 Sub-Agents (callable by any L2 — SDD or non-SDD)

Location: `.atl/overlays/odoo-development-skill/agents/`

| Agent | Trigger condition | Primary use |
|---|---|---|
| `odoo-expert` | Any Odoo model/view/OWL task | Implementation |
| `odoo-code-reviewer` | Post-apply review | Style, security, OCA |
| `odoo-context-gatherer` | Module exploration | Structure, deps, version |
| `odoo-database-query` | Schema / data inspection | DB audit, field analysis |
| `odoo-upgrade-analyzer` | Migration readiness | Version upgrade impact |
| `odoo-skill-finder` | Before new feature design | OCA/core module search |
| `odoo-plan` | Pre-SDD planning | Odoo-specific scoping |
| `odoo-spreadsheet-dashboard-architect` | .osheet dashboard | JSON generation |
| `odoo-ui-automation` | UI tests | Playwright/Selenium Odoo |
| `odoo-addons-maintainer` | Module maintenance | Manifest, upgrades |

### 3. SDD Phase → Supplement Auto-Injection

SDD orchestrator MUST inject supplement alongside base phase skill on Odoo projects:

| Phase | Base Skill | + Odoo Supplement |
|---|---|---|
| sdd-init | sdd-init | `sdd-supplements/init-odoo.md` |
| sdd-explore | sdd-explore | `sdd-supplements/explore-odoo.md` |
| sdd-propose | sdd-propose | `sdd-supplements/propose-odoo.md` |
| sdd-spec | sdd-spec | `sdd-supplements/spec-odoo.md` |
| sdd-design | sdd-design | `sdd-supplements/design-odoo.md` |
| sdd-tasks | sdd-tasks | `sdd-supplements/tasks-odoo.md` |
| sdd-apply | sdd-apply | `sdd-supplements/apply-odoo.md` |
| sdd-verify | sdd-verify | `sdd-supplements/verify-odoo.md` |
| sdd-archive | sdd-archive | `sdd-supplements/archive-odoo.md` |

### 4. Version-Gated Pattern Bundle

```
v19 → patterns-19/ + patterns-agnostic/
v18 → patterns-18/ + patterns-agnostic/
v17 → patterns-17/ + patterns-agnostic/
v16 → patterns-16/ + patterns-agnostic/
v14/v15 → patterns-14-15/ + patterns-agnostic/
Multi-version → all detected + migration-{from}-{to}/ + patterns-agnostic/
```

### 5. Odoo Research Order (overrides general routing)
```
1. Engram (project memory + prior Odoo decisions)
2. rg in local custom addons (project/addons_customs/)
3. rg in Odoo Community local (~/gitproj/odoo/community/{version}/addons/)
4. Context7 with "odoo" library-id
5. OCA GitHub (https://github.com/OCA?q={keyword})
6. Odoo Community GitHub (https://github.com/odoo/odoo/tree/{version}.0/addons)
```

### 6. Cross-Agent Odoo Calling Patterns
```
sdd-explore     → odoo-context-gatherer (module structure)
sdd-explore     → odoo-database-query (schema inspection)
sdd-design      → odoo-skill-finder (avoid reinventing wheel)
sdd-verify      → odoo-code-reviewer (Odoo-specific review)
researcher      → odoo-database-query (schema investigation)
solver          → odoo-expert (complex ORM/Python debugging)
general-orch    → odoo-plan (Odoo feature planning)
```

### 7. Odoo Rules Injection (alongside base guardrails)

Rules are injected in two tiers to minimize token overhead. The orchestrator
reads `current_phase` from sdd-state.yaml and selects the appropriate tier.

#### Tier A — Read-only phases (sdd-explore, sdd-propose, sdd-spec, sdd-design, sdd-tasks, sdd-archive)

```
.atl/overlays/odoo-development-skill/rules/coding-style.md → compact rules  (~1.5KB)
.atl/overlays/odoo-development-skill/rules/security.md     → compact rules  (~2KB)
.atl/overlays/odoo-development-skill/rules/CAUTION_POLICY.md → SUMMARY ONLY (~500 chars)
  Full CAUTION_POLICY available via: mem_search("knowledge/odoo/caution-policy")
  Load full version only when: task involves write ops OR D1 >= 2 OR security risk detected
```

**Tier A total injection: ~4KB per sub-agent** (vs ~13KB before — 69% reduction)

#### Tier B — Write phases (sdd-apply, sdd-verify)

```
.atl/overlays/odoo-development-skill/rules/coding-style.md     → compact rules  (~1.5KB)
.atl/overlays/odoo-development-skill/rules/security.md          → compact rules  (~2KB)
.atl/overlays/odoo-development-skill/rules/CAUTION_POLICY.md    → FULL           (~8KB)
.atl/overlays/odoo-development-skill/rules/cudio-git.md         → compact rules  (~1KB)
```

**Tier B total injection: ~14KB per sub-agent** (same as before — no quality regression on write phases)

#### Tier Selection Logic (inject at orchestrator launch prompt construction)

```
IF current_phase IN [sdd-explore, sdd-propose, sdd-spec, sdd-design, sdd-tasks, sdd-archive]:
  → USE Tier A injection
ELIF current_phase IN [sdd-apply, sdd-verify]:
  → USE Tier B injection
ELSE (general-orchestrator non-SDD tasks with IS_ODOO=true):
  → USE Tier A injection (default to minimal)
```

### 8. Odoo MCP Pagination Guard (MANDATORY for ALL mcp_odoo_search_records calls)

Every call to `mcp_odoo_search_records` MUST include an explicit `limit` parameter.

**Maximum limit per call: 50 records.**

PROHIBITED patterns (will exceed context budget and trigger D4 saturation):
```
mcp_odoo_search_records(model="sale.order", domain=[])
mcp_odoo_search_records(model="account.move", domain=[], limit=None)
mcp_odoo_search_records(model="stock.move", fields=["all"], domain=[])
```

REQUIRED pattern:
```
mcp_odoo_search_records(model="sale.order", domain=[("state","=","sale")], limit=50)
```

For volume/count analysis: use `mcp_odoo_aggregate_records` (returns metadata only, no payload).
For specific record exploration: filter `domain` to target ≤ 50 records before calling.

If more than 50 records are needed: paginate with `offset` in sequential calls, processing
one page at a time and summarizing to masked_evidence before fetching the next page.

### 9. YOLO Mode Guard (MANDATORY when ODOO_YOLO=true)

When `ODOO_YOLO=true` is set in the MCP environment:

**CHECK before ANY write operation (create, write, delete, unlink):**

```
IF D3 >= 2 OR D4 >= 2:
  BLOCK the write operation.
  EMIT: "[YOLO_GUARD] Context degraded (D3={D3}, D4={D4}). Write operation suspended."
  EMIT: "Resolve context pressure before executing mutations. Run context-guardian or /compact."
  DO NOT proceed with the write — return status: "blocked", blocked_reason: "yolo_guard_d_score"

IF D3 < 2 AND D4 < 2:
  Proceed with YOLO write (normal behavior).
```

**CHECK for large-volume writes:**

```
IF records_to_modify > 100 AND ODOO_YOLO=true:
  STOP and emit to user:
  "[YOLO_GUARD] About to modify {N} records in {model}.
   This operation cannot be undone automatically.
   Confirm? Type: confirm-yolo-{model}-{N}"
  Wait for exact confirmation string before proceeding.
```
