package toolpolicy

import (
	"fmt"
	"os"
	"path/filepath"
)

type ToolPolicy struct {
	PreToolUse []ToolRule `yaml:"pre_tool_use"`
}

type ToolRule struct {
	Matcher   string `yaml:"matcher"`
	Decision  string `yaml:"decision"`
	Condition string `yaml:"condition,omitempty"`
}

func GenerateToolPolicy(projectRoot string) error {
	policy := ToolPolicy{
		PreToolUse: []ToolRule{
			{
				Matcher:  "run_command|Bash|Shell|execute",
				Decision: "ask",
				Condition: "phase == sdd-apply AND mode != tmux",
			},
			{
				Matcher:  "mcp__codegraph__.*",
				Decision: "allow",
			},
			{
				Matcher:  "mcp__engram__mem_save|mcp__engram__mem_update",
				Decision: "allow",
			},
			{
				Matcher:  "mcp__engram__mem_search|mcp__engram__mem_get_observation",
				Decision: "allow",
			},
			{
				Matcher:  "WebFetch|WebSearch|mcp__context7__.*",
				Decision: "ask",
				Condition: "posture == production",
			},
			{
				Matcher:  "mcp__context.mode__ctx_.*",
				Decision: "allow",
			},
		},
	}

	atlDir := filepath.Join(projectRoot, ".atl")
	if err := os.MkdirAll(atlDir, 0o755); err != nil {
		return fmt.Errorf("create .atl directory: %w", err)
	}

	policyPath := filepath.Join(atlDir, "tool_policy.yaml")
	content := generateYAML(policy)
	
	tmpPath := policyPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write tool policy tmp: %w", err)
	}
	
	return os.Rename(tmpPath, policyPath)
}

func generateYAML(policy ToolPolicy) string {
	var result string
	result += "pre_tool_use:\n"
	
	for _, rule := range policy.PreToolUse {
		result += fmt.Sprintf("  - matcher: %q\n", rule.Matcher)
		result += fmt.Sprintf("    decision: %q\n", rule.Decision)
		if rule.Condition != "" {
			result += fmt.Sprintf("    condition: %q\n", rule.Condition)
		}
	}
	
	return result
}
