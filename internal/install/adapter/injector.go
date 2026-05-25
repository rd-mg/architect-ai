// internal/install/adapter/injector.go
package adapter

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Platform config for each supported IDE/CLI
type Platform struct {
	ID                    string
	EntryFile             string
	SupportsRealSubagents bool
	SupportsParallel      bool
	SupportsNativeMCP     bool
	CompressCommand       string
	HasDegradedMode       bool
}

var Supported = map[string]Platform{
	"opencode":    {ID: "opencode",    EntryFile: "opencode.json",                          SupportsRealSubagents: true,  SupportsParallel: true,  SupportsNativeMCP: true,  CompressCommand: "/compact"},
	"claude":      {ID: "claude",      EntryFile: "CLAUDE.md",                              SupportsRealSubagents: true,  SupportsParallel: true,  SupportsNativeMCP: true,  CompressCommand: "/compact"},
	"cursor":      {ID: "cursor",      EntryFile: ".github/copilot-instructions.md",        SupportsRealSubagents: false, SupportsParallel: false, SupportsNativeMCP: false, HasDegradedMode: true},
	"antigravity": {ID: "antigravity", EntryFile: ".antigravity/agent.md",                  SupportsRealSubagents: false, SupportsParallel: false, SupportsNativeMCP: false, HasDegradedMode: true},
	"gemini":      {ID: "gemini",      EntryFile: "GEMINI.md",                              SupportsRealSubagents: true,  SupportsParallel: true,  SupportsNativeMCP: true,  CompressCommand: "/compress"},
}

// contentHash computes a short SHA256 hash of content for marker idempotency
func contentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", sum[:4]) // 8 hex chars is enough for change detection
}

// sectionPattern matches <!-- architect-ai:{name}:start hash:{hash} --> ... <!-- architect-ai:{name}:end -->
var sectionStartRe = regexp.MustCompile(`<!--\s*architect-ai:([^:]+):start(?:\s+hash:[a-f0-9]+)?\s*-->`)
var sectionEndRe = regexp.MustCompile(`<!--\s*architect-ai:([^:]+):end\s*-->`)

// InjectSection updates a named section in a Markdown config file.
// Uses content hash to skip injection if content is unchanged (idempotency fix).
func InjectSection(filePath, sectionName, content string) (bool, error) {
	startMarker := fmt.Sprintf("<!-- architect-ai:%s:start hash:%s -->", sectionName, contentHash(content))
	endMarker := fmt.Sprintf("<!-- architect-ai:%s:end -->", sectionName)
	newSection := startMarker + "\n" + content + "\n" + endMarker

	existing := ""
	if data, err := os.ReadFile(filePath); err == nil {
		existing = string(data)
	}

	// Check if content is already up-to-date (idempotency)
	if strings.Contains(existing, startMarker) {
		return false, nil // already up-to-date, skip injection
	}

	// Find and replace existing section (any hash variant)
	startPattern := regexp.MustCompile(`<!--\s*architect-ai:` + regexp.QuoteMeta(sectionName) + `:start[^>]*-->`)
	endPattern := regexp.MustCompile(`<!--\s*architect-ai:` + regexp.QuoteMeta(sectionName) + `:end\s*-->`)

	if startPattern.MatchString(existing) {
		startIdx := startPattern.FindStringIndex(existing)
		endLoc := endPattern.FindStringIndex(existing)
		if endLoc != nil && endLoc[1] > startIdx[0] {
			existing = existing[:startIdx[0]] + newSection + existing[endLoc[1]:]
		}
	} else {
		existing = existing + "\n\n" + newSection + "\n"
	}

	tmp := filePath + ".tmp"
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return false, fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(tmp, []byte(existing), 0644); err != nil {
		return false, fmt.Errorf("write tmp: %w", err)
	}
	return true, os.Rename(tmp, filePath)
}

// ValidateInstallation checks all required agent files exist for a platform
func ValidateInstallation(platformID, projectDir string) []string {
	p, ok := Supported[platformID]
	if !ok {
		return []string{fmt.Sprintf("unknown platform: %s", platformID)}
	}

	var issues []string
	entryPath := filepath.Join(projectDir, p.EntryFile)
	if _, err := os.Stat(entryPath); os.IsNotExist(err) {
		issues = append(issues, fmt.Sprintf("[MISSING] entry file: %s", p.EntryFile))
	}

	// Check .atl structure
	required := []string{
		".atl/sdd-state.yaml",
		".atl/skill-manifest.yaml",
		".atl/_generated/foundation.md",
		".atl/agents/architect.md",
		".atl/agents/sdd-orchestrator.md",
		".atl/agents/general-orchestrator.md",
	}
	for _, r := range required {
		if _, err := os.Stat(filepath.Join(projectDir, r)); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("[MISSING] %s", r))
		}
	}

	if !p.SupportsRealSubagents {
		issues = append(issues, fmt.Sprintf("[WARN] %s: single-thread only — sub-agents are simulated", platformID))
	}
	if !p.SupportsNativeMCP {
		issues = append(issues, fmt.Sprintf("[WARN] %s: no native MCP — degraded mode active", platformID))
	}
	if p.CompressCommand == "" {
		issues = append(issues, fmt.Sprintf("[WARN] %s: no compress command — manual summary fallback only", platformID))
	}
	return issues
}
