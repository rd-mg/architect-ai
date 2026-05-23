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
```
.atl/overlays/odoo-development-skill/rules/coding-style.md     → compact rules
.atl/overlays/odoo-development-skill/rules/security.md          → compact rules
.atl/overlays/odoo-development-skill/rules/CAUTION_POLICY.md    → full (critical)
.atl/overlays/odoo-development-skill/rules/cudio-git.md         → compact rules
```
