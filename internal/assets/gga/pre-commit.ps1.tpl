# GGA v2.0 — Windows PowerShell pre-commit hook
# Uses only built-in PowerShell — no-jq required

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

# Build payload using ConvertTo-Json (built-in — no-jq needed)
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
