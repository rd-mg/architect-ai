# Tasks: Phase 7 - GGA v2 Agnostic Directives + cudio-git + Odoo Rules

## Status: COMPLETED

## Implementation Tasks

### 7.1 GGA v2 AGENTS.md
- [x] `internal/assets/gga/AGENTS.md` — v2.0 full rewrite
  - [x] Section A: Agnostic Rules (Secrets, Architecture, Quality, Testing, Error Handling, Dependencies)
  - [x] Section B: Commit Format Rules (cudio-git projects)
  - [x] Section C: Odoo-Specific Rules (Security, Architecture, Version-Gated Patterns, Performance, Module Structure)
  - [x] Output JSON schema: verdict, summary, findings, commit_format_valid, commit_format_issue, odoo_version_detected, skip_reason

### 7.2 bash Hook v2 (Linux/macOS)
- [x] `internal/assets/gga/pre-commit.bash.tpl`
  - [x] Secret detection (always, never skippable)
  - [x] Skip mode logging locally and to Engram
  - [x] CI auto-detection (`CI=true`) for static-only mode
  - [x] Non-blocking behavior when AI provider is down (exit 0, log to `.gga/provider-errors.jsonl`)

### 7.3 PowerShell Shim v2 (Windows)
- [x] `internal/assets/gga/pre-commit.ps1.tpl`
  - [x] Migrate completely off `jq` to native PowerShell `ConvertTo-Json`
  - [x] Replicate all bash hook capabilities (secret detection, skip mode, non-blocking)

### 7.4 Go Installer — GGA Installer + Detector
- [x] `internal/gga/installer.go`
  - [x] Implement `Detect()` for Odoo, version, cudio-git
  - [x] Implement `Install()` to write OS-specific hook and `.gga/config`
- [x] `internal/gga/installer_test.go`
  - [x] Add tests for `Detect` on empty dir, Odoo manifest, cudio-git files
  - [x] Add tests for template rendering confirming absence of `jq` and presence of secret patterns
