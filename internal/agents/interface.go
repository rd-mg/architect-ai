package agents

import (
	"context"

	"github.com/rd-mg/architect-ai/internal/model"
	"github.com/rd-mg/architect-ai/internal/system"
)

// Capability tags for adapter feature checks.
type Capability string

const (
	CapabilityAutoInstall Capability = "auto-install"
)

// Adapter is the core abstraction for AI agent integration.
//
// Adding a new agent requires compliance with ADAPTER-CONTRACT.md:
//
// 1. A new adapter implementation.
// 2. Factory/Registry registration.
// 3. Catalog registration.
// 4. Component-specific integrations for optional capabilities (see below).
type Adapter interface {
	// Identity
	Agent() model.AgentID
	Tier() model.SupportTier

	// Detection
	Detect(ctx context.Context, homeDir string) (installed bool, binaryPath string, configPath string, configFound bool, err error)

	// Installation
	SupportsAutoInstall() bool
	InstallCommand(profile system.PlatformProfile) ([][]string, error)

	// Config paths — components use these instead of hardcoding paths per agent.
	GlobalConfigDir(homeDir string) string
	SystemPromptDir(homeDir string) string
	SystemPromptFile(homeDir string) string
	SkillsDir(homeDir string) string
	SettingsPath(homeDir string) string

	// Config strategies — HOW to inject content, not WHERE (that's paths above).
	SystemPromptStrategy() model.SystemPromptStrategy
	MCPStrategy() model.MCPStrategy

	// MCP path resolution
	MCPConfigPath(homeDir string, serverName string) string

	// Basic capabilities — common across all tiers.
	SupportsSkills() bool
	SupportsSystemPrompt() bool
	SupportsMCP() bool

	// Optional capabilities — agents declare what they support via these methods
	// or via separate interface type assertions (SubAgentCapable, WorkflowCapable).
	SupportsOutputStyles() bool
	OutputStyleDir(homeDir string) string

	SupportsSlashCommands() bool
	CommandsDir(homeDir string) string
}

// SubAgentCapable identifies adapters that support native sub-agent files
// (e.g., Cursor .cursorrules, Kiro .kiro, Gemini agents/).
type SubAgentCapable interface {
	SupportsSubAgents() bool
	SubAgentsDir(homeDir string) string
	EmbeddedSubAgentsDir() string
}

// WorkflowCapable identifies adapters that support native workflow files
// (e.g., Kiro workflows/).
type WorkflowCapable interface {
	SupportsWorkflows() bool
	WorkflowsDir(workspaceDir string) string
	EmbeddedWorkflowsDir() string
}
