package mcp_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/rd-mg/architect-ai/internal/agents/antigravity"
	antigravitycli "github.com/rd-mg/architect-ai/internal/agents/antigravity-cli"
	"github.com/rd-mg/architect-ai/internal/agents/claude"
	"github.com/rd-mg/architect-ai/internal/agents/gemini"
	"github.com/rd-mg/architect-ai/internal/agents/opencode"
	"github.com/rd-mg/architect-ai/internal/agents/vscode"
	"github.com/rd-mg/architect-ai/internal/components/mcp"
)

// requiredMCPs lists the server names that EVERY agent must have after install.
var requiredMCPs = []string{
	"engram",
	"context7",
	"sequential-thinking",
	"context-mode",
	"codegraph",
	"notebooklm-mcp",
}

// allGeneratedPlatforms lists all agents that GenerateConfig supports.
// Agents using TOML (codex), mcp.json (cursor), or other config mechanisms
// are configured via inject.go's strategy dispatch, not GenerateConfig.
var allGeneratedPlatforms = []struct {
	name    string
	adapter interface{ SupportsMCP() bool }
}{
	{"claude", claude.NewAdapter()},
	{"opencode", opencode.NewAdapter()},
	{"gemini", gemini.NewAdapter()},
	{"vscode", vscode.NewAdapter()},
	{"antigravity", antigravity.NewAdapter()},
	{"antigravity-cli", antigravitycli.NewAdapter()},
}

func TestAllAgentsHaveRequiredMCPs(t *testing.T) {
	for _, tc := range allGeneratedPlatforms {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.adapter.SupportsMCP() {
				t.Skipf("agent %q does not support MCP", tc.name)
			}

			cfg, err := mcp.GenerateConfig(tc.name, mcp.GenerateOptions{})
			if err != nil {
				t.Fatalf("GenerateConfig(%s): %v", tc.name, err)
			}

			cfgJSON, _ := json.Marshal(cfg)
			cfgStr := string(cfgJSON)

			for _, required := range requiredMCPs {
				if !containsJSONString(cfgStr, required) {
					t.Errorf("agent %q missing required MCP server %q in generated config", tc.name, required)
				}
			}
		})
	}
}

func TestGeminiMCPOnlyDoesNotContainSettingsKeys(t *testing.T) {
	// Verify that generateGeminiMCPOnly returns ONLY mcpServers, not general/ui/model/security keys.
	// This is a safety check: the MCP-only path must NOT overwrite user settings on merge.
	t.Skip("generateGeminiMCPOnly is internal — tested via injectMergeIntoSettings integration")
}

func containsJSONString(s, substr string) bool {
	return strings.Contains(s, `"`+substr+`"`)
}
