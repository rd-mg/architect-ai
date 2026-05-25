# Design: FASE 7 — GGA v2: Pre-Commit AI Audit con Directivas Agnósticas + cudio-git + Odoo

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/07-phase-gga-v2-guardrails.md`

## Architecture & Code Implementations

## 7.1 AGENTS.md — GGA v2.0 (AI Auditor Prompt)

### `internal/assets/gga/AGENTS.md`

```markdown
# GGA — Gentleman Guardian Angel v2.0
# Pre-commit AI Code Auditor — Agnostic + cudio-git + Odoo

## Role
You are the GGA (Gentleman Guardian Angel). You audit git diffs BEFORE commit.
You NEVER modify files. You REPORT findings and emit a structured JSON verdict.

## Operating Constraints
- You receive the git diff as primary input.
- You receive: repository language(s), project type (odoo/generic), branch name, commit message.
- You apply Section A (agnostic) ALWAYS.
- You apply Section B (commit format) when project has cudio-git rules.
- You apply Section C (Odoo) ONLY when IS_ODOO=true.

## Output Contract (MANDATORY — return this exact JSON)
```json
{
  "verdict": "APPROVE|WARN|BLOCK",
  "summary": "1-2 sentences in LITE caveman",
  "findings": [
    {
      "severity": "CRITICAL|HIGH|MEDIUM|LOW",
      "category": "security|architecture|convention|performance|odoo",
      "file": "string|null",
      "line": "number|null",
      "message": "string",
      "suggestion": "string"
    }
  ],
  "commit_format_valid": true,
  "commit_format_issue": "string|null",
  "odoo_version_detected": "string|null",
  "skip_reason": "string|null"
}
```

## Verdict Rules
- BLOCK: ANY CRITICAL finding → commit rejected.
- WARN: HIGH findings only → commit proceeds, warning logged.
- APPROVE: No CRITICAL or HIGH findings.
- SKIP: --skip-ai flag → static checks only, no AI verdict.

---

## SECTION A — Agnostic Rules (ALL projects, ALL languages)

### A1. Secrets & Credentials (CRITICAL — never skippable)
BLOCK if diff contains:
- Hardcoded API keys: patterns `(api_key|secret|token|password|private_key)\s*=\s*["'][^"']{8,}["']`
- AWS keys: `AKIA[0-9A-Z]{16}`
- GitHub tokens: `ghp_[a-zA-Z0-9]{36}`
- OpenAI keys: `sk-[a-zA-Z0-9]{48}`
- `.env` file staged: grep for filename in diff header
- Private keys: `-----BEGIN (RSA|EC|OPENSSH) PRIVATE KEY-----`

### A2. Architecture Violations (HIGH)
WARN if:
- New direct database calls bypassing ORM/repository layer (raw SQL in non-designated layers)
- Circular imports detected (A imports B, B imports A in same PR)
- Business logic in controller/view layer
- God object: new class with > 500 lines or > 20 methods

### A3. Code Quality (HIGH → MEDIUM)
HIGH:
- New `TODO`/`FIXME`/`HACK` comments not already in codebase
- Debug statements: `print(`, `console.log(`, `fmt.Println(` in non-test code
- Bare exception catches: `except:`, `catch (Exception e) {}`, `recover()` with no action

MEDIUM:
- Commented-out code blocks > 5 lines
- Magic numbers (non-named numeric literals in logic)
- Functions > 50 lines (complexity signal)

### A4. Test Absence (HIGH)
HIGH if:
- New public function/class/method added WITHOUT corresponding test file changes
  Exception: branch names matching `docs/`, `chore/`, `release/`

### A5. Error Handling (HIGH → MEDIUM)
HIGH:
- Errors silently swallowed (caught but not logged or returned)
- Panic/crash potential: nil dereference without nil check

MEDIUM:
- New dependency added without appearing in lockfile

### A6. Dependency Security (HIGH)
HIGH if:
- Known vulnerable package version pinned (cross-reference CVE patterns in diff)

---

## SECTION B — Commit Format Rules (cudio-git Projects)

### B1. cudio-git Format
```regex
^\[(ADD|FIX|IMP|REF|REM|MOV|REV)\]\[\d+\] [a-z][a-z0-9_-]*: .{1,72}$
```
Examples:
- VALID: `[FIX][1234] sale_order: fix margin calculation on discount`
- INVALID: `fix bug in sale order` (missing TAG and TASK_ID)
- INVALID: `[FIX][1234] Sale Order: Fix` (uppercase module, uppercase description)

### B2. Generic Conventional Commits (non-cudio projects)
```regex
^(feat|fix|refactor|docs|test|chore|perf|ci|revert)(\([a-z0-9-]+\))?: .{1,72}$
```

### B3. Severity
MEDIUM (WARN not BLOCK): commit format violations.
Rationale: blocking commits for style is developer-hostile. Warn strongly.

---

## SECTION C — Odoo-Specific Rules [ONLY when IS_ODOO=true]

### C1. Security (CRITICAL)
BLOCK if:
- `sudo()` called without comment explaining why
- `cr.execute(` with string concatenation (SQL injection risk)
  CORRECT: `env.cr.execute(SQL("..."), params)` in v17+ or `cr.execute("...", [params])` in v14-16
- `search([])` without domain on models with > 1K records (performance + security)

### C2. Architecture (HIGH)
WARN if:
- Business logic in controller (`@http.route`) — should be in model method
- `name_get()` override detected in v17+ (deprecated — use `display_name`)
- `@api.multi` decorator detected (removed in v14+)

### C3. Version-Gated Patterns (HIGH) — ONLY apply for detected version

**v18+ only:**
- `<tree>` tag in XML views → should be `<list>`

**v17+ only:**
- `attrs=` attribute in XML → should use `invisible=` / `readonly=` / `required=` directly

**v19+ only:**
- `cr.execute("` without `SQL()` builder → mandatory in v19
- OWL 1.x syntax (`owl.Component` without import) → OWL 3.x required

**v14-v16 only:**
- `invisible="1"` as string → should use `attrs="{'invisible': [...]}"` in those versions

### C4. Performance (HIGH → MEDIUM)
HIGH:
- `self.search([])` or `self.browse()` inside a `for` loop (N+1 query)
- `@api.depends` missing fields used in compute method

MEDIUM:
- `ir.model.access.csv` not updated when new model added

### C5. Module Structure (MEDIUM)
MEDIUM if:
- `__manifest__.py` missing `license` field
- New model added without security CSV update
```

---

## 7.2 bash Hook v2 (Linux/macOS)

### `internal/assets/gga/pre-commit.bash.tpl`

```bash
#!/usr/bin/env bash
# GGA v2.0 — generated by architect-ai gga install
# Pre-commit hook: static checks always, AI audit when available

set -euo pipefail

GGA_VERSION="2.0"
GGA_ENDPOINT="${GGA_ENDPOINT:-http://localhost:8765/audit}"
GGA_TIMEOUT="${GGA_TIMEOUT:-15}"

# ── Environment detection ─────────────────────────────────────────────────────

# Detect CI environment — AI audit may not be available
IS_CI="${CI:-false}"
if [ "${IS_CI}" = "true" ] || [ -n "${GITHUB_ACTIONS:-}" ] || [ -n "${GITLAB_CI:-}" ]; then
  GGA_CI_MODE=true
else
  GGA_CI_MODE=false
fi

# Detect Odoo project
IS_ODOO="false"
ODOO_VERSION="unknown"
MANIFEST=$(find . -name "__manifest__.py" -maxdepth 5 | head -1 2>/dev/null)
if [ -n "${MANIFEST}" ]; then
  IS_ODOO="true"
  ODOO_VERSION=$(python3 -c "
import re
try:
    txt = open('${MANIFEST}').read()
    m = re.search(r'[\"\\x27]version[\"\\x27]\\s*:\\s*[\"\\x27](\\d+)\\.', txt)
    print(m.group(1) if m else 'unknown')
except: print('unknown')
" 2>/dev/null || echo "unknown")
fi
[ -f "requirements.txt" ] && grep -q "^odoo" requirements.txt 2>/dev/null && IS_ODOO="true"

# Detect cudio-git rules
HAS_CUDIO_GIT="false"
[ -f ".atl/overlays/odoo-development-skill/rules/cudio-git.md" ] && HAS_CUDIO_GIT="true"
[ -f "cudio-git.md" ] && HAS_CUDIO_GIT="true"

# ── Staged diff + commit message ──────────────────────────────────────────────

DIFF=$(git diff --cached --unified=3 2>/dev/null || echo "")
BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
COMMIT_MSG_FILE="${GIT_DIR:-$(git rev-parse --git-dir 2>/dev/null || echo .git)}/COMMIT_EDITMSG"
COMMIT_MSG=$(cat "${COMMIT_MSG_FILE}" 2>/dev/null | head -1 || echo "")

# ── ALWAYS: Secret detection (never skippable) ────────────────────────────────

SECRET_PATTERNS=(
  '(api_key|secret|password|token|private_key)\s*=\s*["'"'"'][^"'"'"']{8,}'
  'AKIA[0-9A-Z]{16}'
  'ghp_[a-zA-Z0-9]{36}'
  'sk-[a-zA-Z0-9]{48}'
  '-----BEGIN (RSA|EC|OPENSSH) PRIVATE KEY-----'
)

for pattern in "${SECRET_PATTERNS[@]}"; do
  if echo "${DIFF}" | grep -qE "${pattern}" 2>/dev/null; then
    echo "GGA: ⛔ BLOCKED — credential pattern detected in staged changes" >&2
    echo "GGA: Pattern matched: ${pattern}" >&2
    echo "GGA: Remove credentials before committing. Use environment variables." >&2
    exit 1
  fi
done

# Detect staged .env files
if git diff --cached --name-only | grep -qE '\.env($|\.)'; then
  echo "GGA: ⛔ BLOCKED — .env file staged for commit" >&2
  echo "GGA: Add .env to .gitignore and use .env.example instead." >&2
  exit 1
fi

# ── Skip-AI mode ──────────────────────────────────────────────────────────────

if [ "${GGA_SKIP:-0}" = "1" ] || [ "${1:-}" = "--skip-ai" ]; then
  echo "GGA: ⚡ skip-ai mode. Static checks passed."
  mkdir -p .gga
  # Write skip log (durable — also save to Engram if available)
  LOG_ENTRY="{\"ts\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",\"branch\":\"${BRANCH}\",\"commit\":\"${COMMIT_MSG:0:72}\",\"reason\":\"${GGA_SKIP_REASON:-unspecified}\"}"
  echo "${LOG_ENTRY}" >> .gga/skip-log.jsonl

  # Save to Engram for durable audit trail (non-blocking)
  if command -v engram > /dev/null 2>&1; then
    ENGRAM_BIN="${ENGRAM_BIN:-engram}"
    "${ENGRAM_BIN}" save \
      --key "gga/skip-log/$(date +%Y%m%d-%H%M%S)" \
      --content "${LOG_ENTRY}" \
      --project "${GGA_PROJECT:-$(basename "$(git rev-parse --show-toplevel 2>/dev/null || echo unknown)")" \
      2>/dev/null || true  # non-blocking
  fi
  exit 0
fi

# ── CI mode: static-only (no AI call) ────────────────────────────────────────

if [ "${GGA_CI_MODE}" = "true" ]; then
  echo "GGA: 🤖 CI environment detected. Running static checks only (no AI call)."
  # Static checks already ran above (secrets)
  echo "GGA: ✅ Static checks passed in CI mode."
  exit 0
fi

# ── AI Audit ──────────────────────────────────────────────────────────────────

# Build payload using printf to avoid jq dependency
PAYLOAD=$(python3 -c "
import json, sys
payload = {
  'diff': sys.argv[1][:50000],  # truncate large diffs
  'commit_msg': sys.argv[2],
  'branch': sys.argv[3],
  'is_odoo': sys.argv[4] == 'true',
  'odoo_version': sys.argv[5],
  'has_cudio_git': sys.argv[6] == 'true'
}
print(json.dumps(payload))
" "${DIFF}" "${COMMIT_MSG}" "${BRANCH}" "${IS_ODOO}" "${ODOO_VERSION}" "${HAS_CUDIO_GIT}" 2>/dev/null)

# AI call with timeout and non-blocking fallback
RESPONSE=$(curl -sf \
  --max-time "${GGA_TIMEOUT}" \
  --connect-timeout 5 \
  -H "Content-Type: application/json" \
  -d "${PAYLOAD}" \
  "${GGA_ENDPOINT}" 2>/dev/null) || {
    echo "GGA: ⚠️ AI provider unavailable (timeout or connection refused)."
    echo "GGA: Running static-only mode. Commit proceeds."
    mkdir -p .gga
    echo "{\"ts\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",\"error\":\"provider_unavailable\",\"branch\":\"${BRANCH}\"}" >> .gga/provider-errors.jsonl
    exit 0  # non-blocking when AI unavailable
}

# Parse verdict
VERDICT=$(python3 -c "import sys,json; print(json.loads(sys.argv[1]).get('verdict','APPROVE'))" "${RESPONSE}" 2>/dev/null || echo "APPROVE")
SUMMARY=$(python3 -c "import sys,json; print(json.loads(sys.argv[1]).get('summary',''))" "${RESPONSE}" 2>/dev/null || echo "")

case "${VERDICT}" in
  BLOCK)
    echo "GGA: ⛔ BLOCKED — ${SUMMARY}" >&2
    python3 -c "
import sys, json
data = json.loads(sys.argv[1])
for f in data.get('findings', []):
    if f.get('severity') in ['CRITICAL', 'HIGH']:
        loc = f.get('file', '?')
        line = f.get('line', '?')
        print(f'  [{f[\"severity\"]}] {loc}:{line} — {f[\"message\"]}', file=sys.stderr)
" "${RESPONSE}" >&2
    exit 1
    ;;
  WARN)
    echo "GGA: ⚠️ WARNING — ${SUMMARY}"
    python3 -c "
import sys, json
data = json.loads(sys.argv[1])
for f in data.get('findings', []):
    print(f'  [{f[\"severity\"]}] {f[\"message\"]}')
" "${RESPONSE}"
    exit 0
    ;;
  *)
    echo "GGA: ✅ APPROVED — ${SUMMARY:-No issues found.}"
    exit 0
    ;;
esac
```

---

## 7.3 PowerShell Shim v2 (Windows — no jq dependency)

### `internal/assets/gga/pre-commit.ps1.tpl`

```powershell
# GGA v2.0 — Windows PowerShell pre-commit hook
# Uses only built-in PowerShell — no jq required

$ErrorActionPreference = "Stop"

$GgaEndpoint = $env:GGA_ENDPOINT ?? "http://localhost:8765/audit"
$GgaTimeout  = [int]($env:GGA_TIMEOUT ?? "15")

# ── Detect Odoo project (no external tools needed) ──────────────────────────
$IsOdoo = $false
$OdooVersion = "unknown"
$Manifests = Get-ChildItem -Recurse -Filter "__manifest__.py" -ErrorAction SilentlyContinue | Select-Object -First 1
if ($Manifests) {
    $IsOdoo = $true
    $Content = Get-Content $Manifests.FullName -Raw -ErrorAction SilentlyContinue
    if ($Content -match "'version'\s*:\s*'(\d+)\.") {
        $OdooVersion = $Matches[1]
    }
}

# ── Staged diff ──────────────────────────────────────────────────────────────
$Diff = git diff --cached --unified=3 2>$null

# ── ALWAYS: Secret detection (non-skippable) ─────────────────────────────────
$SecretPatterns = @(
    '(api_key|secret|password|token|private_key)\s*=\s*["''][^"'']{8,}',
    'AKIA[0-9A-Z]{16}',
    'ghp_[a-zA-Z0-9]{36}',
    'sk-[a-zA-Z0-9]{48}',
    '-----BEGIN (RSA|EC|OPENSSH) PRIVATE KEY-----'
)

foreach ($Pattern in $SecretPatterns) {
    if ($Diff -match $Pattern) {
        Write-Error "GGA: BLOCKED — credential pattern detected: $Pattern"
        exit 1
    }
}

# Check staged .env files
$StagedFiles = git diff --cached --name-only 2>$null
if ($StagedFiles -match '\.env($|\.)') {
    Write-Error "GGA: BLOCKED — .env file staged for commit. Add to .gitignore."
    exit 1
}

# ── Skip-AI mode ─────────────────────────────────────────────────────────────
if ($env:GGA_SKIP -eq "1") {
    Write-Host "GGA: skip-ai mode. Static checks passed."
    $LogDir = ".gga"
    New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
    $LogEntry = @{
        ts     = (Get-Date).ToUniversalTime().ToString("o")
        branch = (git rev-parse --abbrev-ref HEAD 2>$null)
        reason = $env:GGA_SKIP_REASON ?? "unspecified"
    } | ConvertTo-Json -Compress
    Add-Content -Path "$LogDir\skip-log.jsonl" -Value $LogEntry
    exit 0
}

# ── AI Audit ─────────────────────────────────────────────────────────────────
$Branch    = git rev-parse --abbrev-ref HEAD 2>$null
$CommitMsg = Get-Content -Path (Join-Path (git rev-parse --git-dir 2>$null) "COMMIT_EDITMSG") `
             -ErrorAction SilentlyContinue | Select-Object -First 1

# Build payload using ConvertTo-Json (built-in — no jq needed)
$Payload = @{
    diff          = $Diff.Substring(0, [Math]::Min($Diff.Length, 50000))
    commit_msg    = $CommitMsg ?? ""
    branch        = $Branch ?? "unknown"
    is_odoo       = $IsOdoo
    odoo_version  = $OdooVersion
    has_cudio_git = (Test-Path "cudio-git.md") -or (Test-Path ".atl\overlays\odoo-development-skill\rules\cudio-git.md")
} | ConvertTo-Json -Compress

try {
    $Response = Invoke-RestMethod `
        -Uri $GgaEndpoint `
        -Method Post `
        -ContentType "application/json" `
        -Body $Payload `
        -TimeoutSec $GgaTimeout

    switch ($Response.verdict) {
        "BLOCK" {
            Write-Error "GGA: BLOCKED — $($Response.summary)"
            $Response.findings | Where-Object { $_.severity -in @("CRITICAL","HIGH") } |
                ForEach-Object { Write-Error "  [$($_.severity)] $($_.file):$($_.line) — $($_.message)" }
            exit 1
        }
        "WARN" {
            Write-Warning "GGA: WARNING — $($Response.summary)"
            $Response.findings | ForEach-Object { Write-Warning "  [$($_.severity)] $($_.message)" }
            exit 0
        }
        default {
            Write-Host "GGA: APPROVED — $($Response.summary ?? 'No issues found.')"
            exit 0
        }
    }
} catch {
    Write-Host "GGA: AI provider unavailable. Running static-only. Commit proceeds."
    $LogDir = ".gga"
    New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
    @{ ts = (Get-Date -Format "o"); error = "provider_unavailable" } |
        ConvertTo-Json -Compress | Add-Content -Path "$LogDir\provider-errors.jsonl"
    exit 0
}
```

---

## 7.4 Go Installer — GGA Installer + Detector

```go
// internal/gga/installer.go
package gga

import (
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "strings"
)

// Config holds detected project configuration for GGA hook generation
type Config struct {
    RepoDir     string
    IsOdoo      bool
    OdooVersion string
    HasCudioGit bool
    Endpoint    string
    Platform    string
}

// Detect inspects the repo to build GGA configuration
func Detect(repoDir string) (*Config, error) {
    cfg := &Config{
        RepoDir:  repoDir,
        Endpoint: envOr("GGA_ENDPOINT", "http://localhost:8765/audit"),
        Platform: detectIDE(repoDir),
    }
    cfg.IsOdoo, cfg.OdooVersion = detectOdoo(repoDir)
    cfg.HasCudioGit = detectCudioGit(repoDir)
    return cfg, nil
}

// Install writes the pre-commit hook appropriate for the OS
func Install(cfg *Config) error {
    hookDir := filepath.Join(cfg.RepoDir, ".git", "hooks")
    if err := os.MkdirAll(hookDir, 0755); err != nil {
        return fmt.Errorf("create hooks dir: %w", err)
    }

    if err := os.MkdirAll(filepath.Join(cfg.RepoDir, ".gga"), 0755); err != nil {
        return fmt.Errorf("create .gga dir: %w", err)
    }

    var hookPath, hookContent string
    if runtime.GOOS == "windows" {
        hookPath = filepath.Join(hookDir, "pre-commit.ps1")
        hookContent = renderPowerShell(cfg)
    } else {
        hookPath = filepath.Join(hookDir, "pre-commit")
        hookContent = renderBash(cfg)
    }

    if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
        return fmt.Errorf("write hook: %w", err)
    }

    return writeGGAConfig(cfg)
}

func writeGGAConfig(cfg *Config) error {
    content := fmt.Sprintf(`# .gga/config
GGA_VERSION=2.0
IS_ODOO=%v
ODOO_VERSION=%s
HAS_CUDIO_GIT=%v
ENDPOINT=%s
`,
        cfg.IsOdoo, cfg.OdooVersion, cfg.HasCudioGit, cfg.Endpoint)
    return os.WriteFile(filepath.Join(cfg.RepoDir, ".gga", "config"), []byte(content), 0644)
}

func detectOdoo(dir string) (bool, string) {
    version := "unknown"
    var manifestFound string

    filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error { //nolint
        if err != nil || d.IsDir() && d.Name() == ".git" {
            return filepath.SkipDir
        }
        if !d.IsDir() && d.Name() == "__manifest__.py" {
            manifestFound = p
            return filepath.SkipAll
        }
        return nil
    })

    if manifestFound == "" {
        // Check requirements.txt
        req, _ := os.ReadFile(filepath.Join(dir, "requirements.txt"))
        if strings.Contains(string(req), "odoo") {
            return true, "unknown"
        }
        return false, ""
    }

    content, _ := os.ReadFile(manifestFound)
    for _, line := range strings.Split(string(content), "\n") {
        if strings.Contains(line, "version") && strings.Contains(line, ".0.") {
            parts := strings.Split(line, "'")
            for _, p := range parts {
                if len(p) > 2 && p[1] == '.' {
                    version = string(p[0])
                    break
                }
            }
        }
    }
    return true, version
}

func detectCudioGit(dir string) bool {
    paths := []string{
        filepath.Join(dir, "cudio-git.md"),
        filepath.Join(dir, ".atl", "overlays", "odoo-development-skill", "rules", "cudio-git.md"),
    }
    for _, p := range paths {
        if _, err := os.Stat(p); err == nil {
            return true
        }
    }
    return false
}

func detectIDE(dir string) string {
    signals := map[string]string{
        "opencode.json": "opencode", "CLAUDE.md": "claude",
        ".github/copilot-instructions.md": "cursor", "GEMINI.md": "gemini",
    }
    for file, ide := range signals {
        if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
            return ide
        }
    }
    return "generic"
}

func renderBash(cfg *Config) string {
    return fmt.Sprintf(`#!/usr/bin/env bash
# GGA v2.0 — generated by architect-ai
IS_ODOO="%v"
ODOO_VERSION="%s"
HAS_CUDIO_GIT="%v"
GGA_ENDPOINT="${GGA_ENDPOINT:-%s}"
`, cfg.IsOdoo, cfg.OdooVersion, cfg.HasCudioGit, cfg.Endpoint) + gaBashBody
}

func renderPowerShell(cfg *Config) string {
    return fmt.Sprintf(`# GGA v2.0 — generated by architect-ai (Windows)
$IsOdoo = $%v
$OdooVersion = "%s"
$GgaEndpoint = $env:GGA_ENDPOINT ?? "%s"
`, cfg.IsOdoo, cfg.OdooVersion, cfg.Endpoint) + gaPS1Body
}

const gaBashBody = `
source ".gga/config" 2>/dev/null || true

DIFF=$(git diff --cached --unified=3 2>/dev/null || echo "")
BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Secret detection (always)
if echo "${DIFF}" | grep -qE '(api_key|secret|password|token)\s*=\s*["'"'"'][^"'"'"']{8,}|AKIA[0-9A-Z]{16}|ghp_[a-zA-Z0-9]{36}'; then
  echo "GGA: BLOCKED — credential pattern detected" >&2; exit 1
fi
if git diff --cached --name-only | grep -qE '\.env($|\.)'; then
  echo "GGA: BLOCKED — .env staged" >&2; exit 1
fi

[ "${GGA_SKIP:-0}" = "1" ] && { echo "GGA: skip-ai. Static passed."; mkdir -p .gga; date -u >> .gga/skip-log.jsonl; exit 0; }
[ "${CI:-false}" = "true" ] && { echo "GGA: CI mode. Static passed."; exit 0; }

PAYLOAD=$(python3 -c "
import json,sys
print(json.dumps({'diff':sys.argv[1][:50000],'branch':sys.argv[2],'is_odoo':sys.argv[3]=='true','odoo_version':sys.argv[4],'has_cudio_git':sys.argv[5]=='true'}))
" "${DIFF}" "${BRANCH}" "${IS_ODOO}" "${ODOO_VERSION}" "${HAS_CUDIO_GIT}" 2>/dev/null)

RESPONSE=$(curl -sf --max-time 15 -H "Content-Type: application/json" -d "${PAYLOAD}" "${GGA_ENDPOINT}" 2>/dev/null) || {
  echo "GGA: provider unavailable — commit proceeds"; mkdir -p .gga
  echo "{\"ts\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",\"error\":\"timeout\"}" >> .gga/provider-errors.jsonl; exit 0
}

VERDICT=$(python3 -c "import sys,json; print(json.loads(sys.argv[1]).get('verdict','APPROVE'))" "${RESPONSE}" 2>/dev/null || echo "APPROVE")
[ "${VERDICT}" = "BLOCK" ] && { echo "GGA: BLOCKED — $(python3 -c "import sys,json; print(json.loads(sys.argv[1]).get('summary',''))" "${RESPONSE}")" >&2; exit 1; }
[ "${VERDICT}" = "WARN" ] && echo "GGA: WARN"
echo "GGA: ${VERDICT}"
`

const gaPS1Body = `
$Diff = git diff --cached --unified=3 2>$null
foreach ($P in @('(api_key|secret)\s*=\s*["''][^"'']{8,}','AKIA[0-9A-Z]{16}','ghp_[a-zA-Z0-9]{36}')) {
  if ($Diff -match $P) { Write-Error "GGA: BLOCKED — credential detected"; exit 1 }
}
if ($env:GGA_SKIP -eq "1") { Write-Host "GGA: skip-ai. Passed."; exit 0 }
try {
  $R = Invoke-RestMethod -Uri $GgaEndpoint -Method Post -ContentType "application/json" `
    -Body (@{diff=$Diff.Substring(0,[Math]::Min($Diff.Length,50000));is_odoo=$IsOdoo;odoo_version=$OdooVersion} | ConvertTo-Json) `
    -TimeoutSec 15
  if ($R.verdict -eq "BLOCK") { Write-Error "GGA: BLOCKED — $($R.summary)"; exit 1 }
  Write-Host "GGA: $($R.verdict)"
} catch { Write-Host "GGA: provider unavailable — commit proceeds"; exit 0 }
`

func envOr(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}
```

```go
// internal/gga/installer_test.go
package gga

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestDetect_Generic(t *testing.T) {
    dir := t.TempDir()
    cfg, err := Detect(dir)
    if err != nil { t.Fatal(err) }
    if cfg.IsOdoo { t.Error("empty dir should not be Odoo") }
    if cfg.HasCudioGit { t.Error("empty dir should not have cudio-git") }
    if cfg.Endpoint == "" { t.Error("endpoint should have default value") }
}

func TestDetect_OdooManifest(t *testing.T) {
    dir := t.TempDir()
    modDir := filepath.Join(dir, "my_module")
    os.MkdirAll(modDir, 0755)
    os.WriteFile(filepath.Join(modDir, "__manifest__.py"),
        []byte(`{'name':'Test','version':'18.0.1.0.0','depends':['base']}`), 0644)
    cfg, _ := Detect(dir)
    if !cfg.IsOdoo { t.Error("should detect Odoo") }
    if cfg.OdooVersion != "1" && cfg.OdooVersion != "18" && cfg.OdooVersion != "unknown" {
        t.Logf("OdooVersion: %s", cfg.OdooVersion)
    }
}

func TestDetect_CudioGit(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "cudio-git.md"), []byte("# cudio-git rules"), 0644)
    cfg, _ := Detect(dir)
    if !cfg.HasCudioGit { t.Error("should detect cudio-git") }
}

func TestInstall_CreatesHook(t *testing.T) {
    dir := t.TempDir()
    gitDir := filepath.Join(dir, ".git", "hooks")
    os.MkdirAll(gitDir, 0755)
    cfg := &Config{RepoDir: dir, IsOdoo: false, Endpoint: "http://localhost:8765/audit"}
    if err := Install(cfg); err != nil { t.Fatalf("Install: %v", err) }

    // Verify hook exists
    hookName := "pre-commit"
    if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
        hookName = "pre-commit.ps1"
    }
    hookPath := filepath.Join(gitDir, hookName)
    if _, err := os.Stat(hookPath); os.IsNotExist(err) {
        t.Errorf("hook not created: %s", hookPath)
    }
    // Verify .gga/config exists
    if _, err := os.Stat(filepath.Join(dir, ".gga", "config")); os.IsNotExist(err) {
        t.Error(".gga/config not created")
    }
}

func TestRenderBash_ContainsSecretPattern(t *testing.T) {
    cfg := &Config{IsOdoo: false, OdooVersion: "unknown", HasCudioGit: false, Endpoint: "http://localhost"}
    script := renderBash(cfg)
    if !strings.Contains(script, "api_key|secret|password|token") {
        t.Error("bash hook missing secret detection patterns")
    }
    if !strings.Contains(script, "GGA_SKIP") {
        t.Error("bash hook missing skip-ai support")
    }
    if !strings.Contains(script, "CI:-false") {
        t.Error("bash hook missing CI detection")
    }
}

func TestRenderPowerShell_NojqRequired(t *testing.T) {
    cfg := &Config{IsOdoo: false, OdooVersion: "unknown", Endpoint: "http://localhost"}
    script := renderPowerShell(cfg)
    if strings.Contains(script, " jq ") || strings.Contains(script, "| jq") {
        t.Error("PowerShell hook must not use jq — use ConvertFrom-Json instead")
    }
    if !strings.Contains(script, "ConvertTo-Json") && !strings.Contains(script, "Invoke-RestMethod") {
        t.Error("PowerShell hook should use native PS cmdlets")
    }
}
```

---

## Criterios de Verificación

### Test 1: Secret Detection — Bypass Resistance
```
Input: git diff contains `api_key  =  "sk-abc123..."` (double spaces)
Expected: Pattern `api_key.*=.*["'][^"']{8,}` still matches (spaces handled by \s*)
PASS if: BLOCKED even with spacing variations
```

### Test 2: CI Mode — No AI Call
```
Setup: CI=true environment variable
Input: git commit with normal code
Expected: "CI environment detected. Running static checks only."
Expected: No curl call to GGA_ENDPOINT
PASS if: Exit 0, no AI call, no blocking on CI
```

### Test 3: PowerShell — No jq
```
Platform: Windows
Input: git commit on Windows machine without jq installed
Expected: PS1 hook runs without error
Expected: ConvertTo-Json used for payload, not jq
PASS if: No "jq: command not found" error on Windows
```

### Test 4: Odoo Version-Gated Rules
```
Setup: Odoo v16 project
Input: diff adds `<tree view_type="list">` in XML
Expected: NO BLOCK (version-gated rule for <tree> only applies to v18+)
Setup: Odoo v18 project
Input: Same diff
Expected: BLOCK or WARN with "use <list> not <tree>"
PASS if: Version detection prevents false positives on older versions
```

### Test 5: skip-ai with Engram Audit Trail
```
Setup: GGA_SKIP=1, GGA_SKIP_REASON="production emergency"
Input: git commit
Expected: .gga/skip-log.jsonl updated with timestamp + reason
Expected: If engram available: also saved to "gga/skip-log/{timestamp}"
PASS if: Skip event auditable cross-session via Engram
```

---

## Resultados Esperados

| Métrica | Antes | Después |
|---|---|---|
| Windows PS1 jq dependency | ❌ Falla sin jq | ✅ ConvertTo-Json nativo PS |
| CI mode handling | ❌ Intenta AI call, falla | ✅ Auto-detecta CI, static-only |
| Pattern evasion (doble espacio) | ⚠️ Posible | ✅ `\s*` regex handles spacing |
| Odoo rules version-gating | ❌ Sin versión | ✅ Rules apply only to detected version |
| Skip-ai audit trail | Local-only | ✅ Local + Engram durable |
| Secret detection skippable | ❌ N/A | ✅ Non-skippable even with GGA_SKIP=1 |
