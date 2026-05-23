# ripgrep-odoo — Local Codebase Discovery Skill (Odoo Monorepo)

## Purpose

Search code patterns in local Odoo monorepo to obtain real evidence
before designing or implementing. Uses the `rg` binary directly.

**ALWAYS use before:**
- Implementing any functionality interacts with Odoo core
- Assuming API of model, OWL component, or o-spreadsheet action
- Writing imports of Odoo native modules

**Do NOT use when:**
- There are recent Engrams with searched pattern (check first with mem_search)
- search is over custom project code (use ripgrep base)

---

## Monorepo Structure (base paths)

```
~/gitproj/odoo/
├── odoo/            → core CE (models, controllers, views)
├── enterprise/      → EE modules (if available)
├── oca/             → OCA modules (by subfolder per category)
├── o-spreadsheet/   → o-spreadsheet source code
├── owl/             → OWL framework (if separate from core)
└── custom/          → current project modules (use ripgrep base for this)
```

---

## Search Domains and Their Flags

### backend_orm (Python models, ORM, business logic)

```bash
rg "{QUERY}" ~/gitproj/odoo/community/ \
  -t py \
  -g "!__manifest__.py" \
  -g "!__init__.py" \
  -g "!*/tests/*" \
  -g "!*/migrations/*" \
  --max-columns 150 \
  --max-count 3 \
  -C 3
```

Excludes: manifests, inits, tests, migrations. Python business logic only.

### frontend_owl (OWL components, JavaScript, view XML)

```bash
rg "{QUERY}" ~/gitproj/odoo/community/addons/web/static/src/ \
  -t js -t xml \
  -g "static/src/**" \
  -g "!*.min.js" \
  -g "!*.bundle.js" \
  --max-columns 150 \
  --max-count 3 \
  -C 2
```

For specific addon: replace `addons/web/static/src/` with `addons/{addon}/static/src/`

### o_spreadsheet (o-spreadsheet components)

```bash
rg "{QUERY}" ~/gitproj/odoo/o-spreadsheet/src/ \
  -t js -t ts \
  -g "!*.min.js" \
  --max-columns 150 \
  --max-count 3 \
  -C 2
```

### views_xml (Odoo views, actions, menus)

```bash
rg "{QUERY}" ~/gitproj/odoo/community/ \
  -t xml \
  -g "!static/" \
  -g "!*/i18n/*" \
  --max-columns 150 \
  --max-count 3 \
  -C 2
```

### security (ACLs, security rules, groups)

```bash
rg "{QUERY}" ~/gitproj/odoo/community/ \
  -g "security/**" \
  -g "*/ir.model.access.csv" \
  --max-count 5
```

### manifest (module dependencies, version gates)

```bash
rg "{QUERY}" ~/gitproj/odoo/community/ \
  -g "__manifest__.py" \
  --max-count 10
```

---

## Mandatory Protocol: 2 Steps

**NEVER execute content search as first step.**

### Step 1: Identify files (-l, files-only)

```bash
# First: which files contain what I'm looking for?
rg "{QUERY}" ~/gitproj/odoo/community/ \
  -t py \           # (or the domain type)
  -l \              # only file names
  --max-count 5

# Example:
rg "class AccountMove" ~/gitproj/odoo/community/ -t py -l
# Output: 2-3 file paths (~10 tokens)
```

### Step 2: Surgical extraction of relevant file

```bash
# Only if Step 1 returns the correct file:
rg "{SPECIFIC_QUERY}" ~/gitproj/odoo/community/addons/account/models/account_move.py \
  --max-count 2 \
  -C 4

# Example:
rg "def _compute_payment_state" \
  ~/gitproj/odoo/community/addons/account/models/account_move.py \
  --max-count 2 -C 4
# Output: 15-20 lines of code (~60 tokens)
```

**Why 2 steps:** Step 1 costs ~10 tokens. Step 2 costs ~60 tokens. Total: 70 tokens.
Direct unfiltered search: 40,000+ tokens → OOM.

---

## VRAM Protection Flags (ALWAYS active)

flags should NEVER be omitted when searching full monorepo:

| Flag | Value | Why |
|---|---|---|
| `--max-columns` | `150` | Ignores minified file lines (lines of 80,000 chars) |
| `--max-count` | `3` (exploration) / `2` (extraction) | Limits results per file |
| `-C` | `2-4` (depending on domain) | Minimal context to understand the pattern |

If output exceeds 80 lines or search takes more than 5 seconds:
→ STOP. Refine with more specific domain or narrower subfolder.

---

## Search Patterns by Use Case

### "How do I inherit this model in Odoo 18?"

```bash
# Step 1:
rg "class AccountMove" ~/gitproj/odoo/community/ -t py -l

# Step 2 (in the found file):
rg "class AccountMove" ~/gitproj/odoo/community/addons/account/models/account_move.py -C 1
# → see: class AccountMove(models.Model) or class AccountMove(account_move, models.Model)
```

### "How do I import this service in OWL?"

```bash
# Step 1:
rg "useService" ~/gitproj/odoo/community/addons/web/static/src/ -t js -l --max-count 3

# Step 2:
rg "const .* = useService" \
  ~/gitproj/odoo/community/addons/web/static/src/core/utils/hooks.js \
  --max-count 2 -C 2
```

### "How does Odoo add action to o-spreadsheet?"

```bash
# Step 1:
rg "registry.category" ~/gitproj/odoo/community/addons/spreadsheet/static/src/ -t js -l

# Step 2:
rg "registry.category\(\"spreadsheet" \
  ~/gitproj/odoo/community/addons/spreadsheet/static/src/ \
  -t js --max-count 2 -C 3
```

### "What ACL exists for this model?"

```bash
rg "account.bank.statement" \
  ~/gitproj/odoo/community/addons/account/ \
  -g "*/ir.model.access.csv" \
  --max-count 10
```

### "What OCA modules have this pattern?"

```bash
# Step 1: files only
rg "bank_statement" ~/gitproj/odoo/oca/ -t py -l --max-count 10

# Step 2: specific pattern in the relevant module
rg "def _import_bank_statement" \
  ~/gitproj/odoo/oca/account-financial-tools/ \
  -t py --max-count 2 -C 3
```

---

## Persistence to Engram (after_model hook)

When ripgrep-odoo finds pattern resolves design question, result MUST be persisted to Engram to avoid searching same thing in next session. Also, if pattern is high quality and generalizable, it should be proposed for inclusion in corresponding `patterns-{v}` skill.

```
# Format for mem_save after a successful search:
topic_key: knowledge/odoo-v{N}/pattern/{descriptive-slug}
content: {
  "query": "the exact rg command that worked",
  "pattern": "the relevant code snippet (max 20 lines)",
  "source": "relative path from ~/gitproj/odoo/",
  "odoo_version": "18",  # or the confirmed version
  "use_when": "description of when to use this pattern",
  "propose_to_skill": true  # if it should be added to patterns-{v}
}
```

save is executed by the **Orchestrator's after_model hook** (defined in Step 05). If `propose_to_skill` is `true`, Orchestrator will generate a `task` artifact so sub-agent updates corresponding version's `SKILL.md`.

---

## Anti-patterns (what NOT to do)

```bash
#  NEVER — domain-less search on full monorepo
rg "account_move" ~/gitproj/odoo/

#  NEVER — no --max-count in content search
rg "def compute" ~/gitproj/odoo/community/ -t py

#  NEVER — too generic term without specific file
rg "def " ~/gitproj/odoo/community/addons/account/ -t py

#  ALWAYS — 2 steps: files-only first
rg "class BankStatementLine" ~/gitproj/odoo/community/ -t py -l
rg "_compute_amount" {file_from_step_1} --max-count 2 -C 3
```
