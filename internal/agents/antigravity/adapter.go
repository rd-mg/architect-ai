package antigravity

import (
	"context"
	"os"
	"path/filepath"

	"github.com/rd-mg/architect-ai/internal/model"
	"github.com/rd-mg/architect-ai/internal/system"
)

type statResult struct {
	isDir bool
	err   error
}

type Adapter struct {
	statPath      func(string) statResult
	workspaceRoot string
}

func NewAdapter() *Adapter {
	return &Adapter{
		statPath: defaultStat,
	}
}

// --- WorkspaceAware ---

func (a *Adapter) SetWorkspaceRoot(root string) {
	a.workspaceRoot = root
}

// --- Identity ---

func (a *Adapter) Agent() model.AgentID {
	return model.AgentAntigravity
}

func (a *Adapter) Tier() model.SupportTier {
	return model.TierFull
}

// --- Detection ---

func (a *Adapter) Detect(_ context.Context, homeDir string) (bool, string, string, bool, error) {
	configPath := filepath.Join(homeDir, ".gemini", "antigravity")

	stat := a.statPath(configPath)
	if stat.err != nil {
		if os.IsNotExist(stat.err) {
			return false, "", configPath, false, nil
		}
		return false, "", "", false, stat.err
	}

	// Antigravity is a desktop IDE — no binary on PATH to detect.
	// If config dir exists, it's installed.
	return stat.isDir, "", configPath, stat.isDir, nil
}

// --- Installation ---

func (a *Adapter) SupportsAutoInstall() bool {
	return false // Desktop IDE — cannot install via CLI.
}

func (a *Adapter) InstallCommand(_ system.PlatformProfile) ([][]string, error) {
	return nil, AgentNotInstallableError{Agent: model.AgentAntigravity}
}

// --- Config paths ---

func (a *Adapter) GlobalConfigDir(homeDir string) string {
	if a.workspaceRoot != "" {
		return filepath.Join(a.workspaceRoot, ".agent")
	}
	return filepath.Join(homeDir, ".gemini", "antigravity")
}

func (a *Adapter) SystemPromptDir(homeDir string) string {
	if a.workspaceRoot != "" {
		return filepath.Join(a.workspaceRoot, ".agent")
	}
	return filepath.Join(homeDir, ".gemini")
}

func (a *Adapter) SystemPromptFile(homeDir string) string {
	if a.workspaceRoot != "" {
		return filepath.Join(a.workspaceRoot, ".agent", "GEMINI.md")
	}
	return filepath.Join(homeDir, ".gemini", "GEMINI.md")
}

func (a *Adapter) SkillsDir(homeDir string) string {
	if a.workspaceRoot != "" {
		return filepath.Join(a.workspaceRoot, ".agent", "skills")
	}
	return filepath.Join(homeDir, ".gemini", "antigravity", "skills")
}

func (a *Adapter) SettingsPath(homeDir string) string {
	if a.workspaceRoot != "" {
		return filepath.Join(a.workspaceRoot, ".agent", "settings.json")
	}
	return filepath.Join(homeDir, ".gemini", "antigravity", "settings.json")
}

// --- Config strategies ---

func (a *Adapter) SystemPromptStrategy() model.SystemPromptStrategy {
	return model.StrategyMarkdownSections
}

func (a *Adapter) MCPStrategy() model.MCPStrategy {
	return model.StrategyMCPConfigFile
}

// --- MCP ---

func (a *Adapter) MCPConfigPath(homeDir string, _ string) string {
	if a.workspaceRoot != "" {
		return filepath.Join(a.workspaceRoot, ".agent", "mcp_config.json")
	}
	return filepath.Join(homeDir, ".gemini", "antigravity", "mcp_config.json")
}

// --- Optional capabilities ---

func (a *Adapter) SupportsOutputStyles() bool {
	return false
}

func (a *Adapter) OutputStyleDir(_ string) string {
	return ""
}

func (a *Adapter) SupportsSlashCommands() bool {
	return true
}

func (a *Adapter) CommandsDir(homeDir string) string {
	if a.workspaceRoot != "" {
		return filepath.Join(a.workspaceRoot, ".agent", "workflows")
	}
	return ""
}

func (a *Adapter) SupportsSkills() bool {
	return true
}

func (a *Adapter) SupportsSystemPrompt() bool {
	return true
}

func (a *Adapter) SupportsMCP() bool {
	return true
}

// --- WorkflowCapable ---

func (a *Adapter) SupportsWorkflows() bool {
	return true
}

func (a *Adapter) WorkflowsDir(_ string) string {
	if a.workspaceRoot != "" {
		return filepath.Join(a.workspaceRoot, ".agent", "workflows")
	}
	return ""
}

func (a *Adapter) EmbeddedWorkflowsDir() string {
	return "antigravity/workflows"
}

// AgentNotInstallableError is returned when InstallCommand is called on a desktop-only agent.
type AgentNotInstallableError struct {
	Agent model.AgentID
}

func (e AgentNotInstallableError) Error() string {
	return "agent " + string(e.Agent) + " is a desktop IDE and cannot be installed via CLI"
}

func defaultStat(path string) statResult {
	info, err := os.Stat(path)
	if err != nil {
		return statResult{err: err}
	}

	return statResult{isDir: info.IsDir()}
}
