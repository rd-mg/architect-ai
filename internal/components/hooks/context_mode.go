package hooks

import (
	"os"
	"path/filepath"

	"github.com/rd-mg/architect-ai/internal/agents"
	"github.com/rd-mg/architect-ai/internal/components/filemerge"
	"github.com/rd-mg/architect-ai/internal/model"
)

// InjectContextModeHook writes a pre-tool-use hook file that routes
// shell/read/grep tool calls through context-mode.
func InjectContextModeHook(homeDir string, adapter agents.Adapter) error {
	hookDir, hookFile, hookContent := hookConfigFor(adapter.Agent(), homeDir)
	if hookDir == "" {
		return nil // agent does not support hook files
	}

	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		return err
	}

	_, err := filemerge.WriteFileAtomic(filepath.Join(hookDir, hookFile), hookContent, 0o644)
	return err
}

func hookConfigFor(agent model.AgentID, homeDir string) (dir, file string, content []byte) {
	switch agent {
	case model.AgentClaudeCode:
		return filepath.Join(homeDir, ".claude", "hooks"),
			"context-mode.json",
			claudeContextModeHookJSON

	case model.AgentCursor:
		return filepath.Join(homeDir, ".cursor", "hooks"),
			"context-mode.json",
			cursorContextModeHookJSON


	case model.AgentAntigravityCLI:
		// CLI uses plugin-level hooks.json (written by installer.go)
		return "", "", nil

	default:
		return "", "", nil
	}
}

var claudeContextModeHookJSON = []byte(`{
  "preToolUse": [
    {
      "type": "command",
      "command": "context-mode hook claude pretooluse",
      "matcher": "Bash|Shell|Read|Grep|WebFetch|Task",
      "timeout": 5
    }
  ]
}
`)

var cursorContextModeHookJSON = []byte(`{
  "preToolUse": [
    {
      "type": "command",
      "command": "context-mode hook cursor pretooluse",
      "matcher": "shell|read|grep|fetch",
      "timeout": 5
    }
  ]
}
`)

