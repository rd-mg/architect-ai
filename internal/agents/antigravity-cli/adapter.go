package antigravitycli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rd-mg/architect-ai/internal/model"
	"github.com/rd-mg/architect-ai/internal/system"
)

type statResult struct {
	isDir bool
	err   error
}

type Adapter struct {
	lookPath func(string) (string, error)
	statPath func(string) statResult
}

func NewAdapter() *Adapter {
	return &Adapter{
		lookPath: exec.LookPath,
		statPath: defaultStat,
	}
}

// --- Identity ---

func (a *Adapter) Agent() model.AgentID {
	return model.AgentAntigravityCLI
}

func (a *Adapter) Tier() model.SupportTier {
	return model.TierFull
}

// --- Detection ---
// agy is the Antigravity CLI binary.
// Config lives at ~/.gemini/antigravity-cli/plugins/architect-ai/

func (a *Adapter) Detect(_ context.Context, homeDir string) (bool, string, string, bool, error) {
	// Check for agy binary on PATH
	binaryPath, err := a.lookPath("agy")
	installed := err == nil

	// Check plugin dir
	pluginDir := filepath.Join(homeDir, ".gemini", "antigravity-cli", "plugins")
	stat := a.statPath(pluginDir)
	if stat.err != nil {
		if os.IsNotExist(stat.err) {
			return installed, binaryPath, pluginDir, false, nil
		}
		return false, "", "", false, stat.err
	}

	return installed, binaryPath, pluginDir, stat.isDir, nil
}

// --- Installation ---

func (a *Adapter) SupportsAutoInstall() bool {
	// agy is installed via curl script — cannot be automated here.
	return false
}

func (a *Adapter) InstallCommand(_ system.PlatformProfile) ([][]string, error) {
	return nil, AgentNotInstallableError{Agent: model.AgentAntigravityCLI}
}

// --- Config paths ---

func (a *Adapter) pluginRoot(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "antigravity-cli", "plugins", "architect-ai")
}

func (a *Adapter) GlobalConfigDir(homeDir string) string {
	return a.pluginRoot(homeDir)
}

func (a *Adapter) SystemPromptDir(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "antigravity-cli", "skills")
}

// SystemPromptFile — Antigravity CLI uses GEMINI.md in the global gemini dir.
func (a *Adapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "GEMINI.md")
}

func (a *Adapter) SkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "antigravity-cli", "skills")
}

func (a *Adapter) SettingsPath(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "antigravity-cli", "settings.json")
}

// --- Config strategies ---

func (a *Adapter) SystemPromptStrategy() model.SystemPromptStrategy {
	return model.StrategyMarkdownSections
}

// MCPStrategy: Antigravity CLI uses plugin-local mcp_config.json (object format).
func (a *Adapter) MCPStrategy() model.MCPStrategy {
	return model.StrategyMCPConfigFile
}

// --- MCP ---

func (a *Adapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(a.pluginRoot(homeDir), "mcp_config.json")
}

// --- Optional capabilities ---

func (a *Adapter) SupportsOutputStyles() bool  { return false }
func (a *Adapter) OutputStyleDir(_ string) string { return "" }
func (a *Adapter) SupportsSlashCommands() bool { return true }
func (a *Adapter) CommandsDir(_ string) string  { return "" }
func (a *Adapter) SupportsSkills() bool          { return true }
func (a *Adapter) SupportsSystemPrompt() bool    { return true }
func (a *Adapter) SupportsMCP() bool             { return true }

// AgentNotInstallableError is returned when InstallCommand is called on the CLI.
type AgentNotInstallableError struct{ Agent model.AgentID }

func (e AgentNotInstallableError) Error() string {
	return "agent " + string(e.Agent) + " (Antigravity CLI / agy) must be installed manually: curl -fsSL https://antigravity.com/install.sh | bash"
}

func defaultStat(path string) statResult {
	info, err := os.Stat(path)
	if err != nil {
		return statResult{err: err}
	}
	return statResult{isDir: info.IsDir()}
}
