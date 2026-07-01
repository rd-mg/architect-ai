package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rd-mg/architect-ai/internal/backup"
	"github.com/rd-mg/architect-ai/internal/model"
)

func TestSync_DistributesGeneralOrchestrator(t *testing.T) {
	temp, _ := filepath.EvalSymlinks(t.TempDir())
	homeDir := temp
	workspaceDir := filepath.Join(temp, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	
	restoreHome := osUserHomeDir
	restoreBackupHome := backup.UserHomeDirFn
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		osUserHomeDir = restoreHome
		backup.UserHomeDirFn = restoreBackupHome
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})
	
	// mock both home dir fns to the temp home
	osUserHomeDir = func() (string, error) { return homeDir, nil }
	backup.UserHomeDirFn = func() (string, error) { return homeDir, nil }
	runCommand = func(string, ...string) error { return nil }
	cmdLookPath = func(name string) (string, error) { return "/usr/local/bin/" + name, nil }

	adapters := []string{
		".claude",
		filepath.Join(".config", "opencode"),
		filepath.Join(".config", "opencode-insiders"),
		".cursor",
		".cursor-tutor",
		filepath.Join(".config", "windsurf"),
		".windsurf",
		".vscode",
		".gemini",
		filepath.Join(".config", "github-copilot"),
		filepath.Join(".gemini", "antigravity"),
	}

	for _, a := range adapters {
		if err := os.MkdirAll(filepath.Join(homeDir, a), 0755); err != nil {
			t.Fatalf("Failed to create mock config dir for %s: %v", a, err)
		}
	}

	// Use agents with well-supported sync paths to avoid verification failures
	// from incomplete mock setup for agents with special directory requirements.
	sel := BuildSyncSelection(SyncFlags{
		Agents: []string{"claude-code", "opencode", "cursor", "windsurf", "vscode-copilot", "codex", "kiro-ide", "antigravity", "qwen-code"},
	}, []model.AgentID{
		model.AgentClaudeCode, model.AgentOpenCode,
		model.AgentCursor, model.AgentWindsurf,
		model.AgentVSCodeCopilot, model.AgentCodex, model.AgentKiroIDE,
		model.AgentAntigravity, model.AgentQwenCode,
	})
	
	// Windsurf workflows require a project root (e.g. .git) to be detected
	if err := os.WriteFile(filepath.Join(workspaceDir, ".git"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create .git marker: %v", err)
	}
	
	originalWD, _ := os.Getwd()
	os.Chdir(workspaceDir)
	defer os.Chdir(originalWD)

	// Pass absolute homeDir and ensure UserHomeDirFn returns absolute path
	_, err := RunSyncWithSelection(homeDir, sel)
	if err != nil {
		t.Fatalf("RunSyncWithSelection failed: %v", err)
	}

	var foundFiles []string
	filepath.Walk(homeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		
		content, _ := os.ReadFile(path)
		if strings.Contains(string(content), "general-orchestrator") || strings.Contains(string(content), "Generalist") {
			foundFiles = append(foundFiles, path)
		}
		return nil
	})

	// We expect general-orchestrator in multiple agents
	if len(foundFiles) < 5 {
		t.Errorf("general-orchestrator was not distributed correctly. found: %v", len(foundFiles))
	}
}
