package platform

import (
	"os"
)

// HookLevel represents the level of hook support a platform has.
type HookLevel string

const (
	HookLevelFull    HookLevel = "full"
	HookLevelPartial HookLevel = "partial"
	HookLevelNone    HookLevel = "none"
)

// Platform represents a supported AI coding platform.
type Platform struct {
	Name                   string
	DisplayName            string
	HookLevel              HookLevel
	RequiresManualMCPSetup bool   // JetBrains — must use UI
	RequiresRestart        bool
	RoutingFileRequired    bool
	RoutingFile            string
	ConfigFiles            []ConfigFile
}

// ConfigFile represents a configuration file to write for a platform.
type ConfigFile struct {
	Path      string
	Content   string
	IsGlobal  bool // writes to ~/ vs project root
	Merge     bool // JSON merge vs overwrite
}

// platforms is the canonical map of all supported platforms.
var platforms = map[string]Platform{
	"opencode": {
		Name:      "opencode",
		DisplayName: "OpenCode",
		HookLevel: HookLevelFull,
		ConfigFiles: []ConfigFile{
			{Path: "opencode.json", Content: opencodeJSON, Merge: true},
		},
	},
	"kilo-code": {
		Name:        "kilo-code",
		DisplayName: "KiloCode",
		HookLevel:   HookLevelFull,
		ConfigFiles: []ConfigFile{
			{Path: "kilo.json", Content: kiloJSON, Merge: true},
		},
	},
	"antigravity": {
		Name:                "antigravity",
		DisplayName:         "Antigravity",
		HookLevel:           HookLevelNone,
		RoutingFileRequired: true,
		RoutingFile:         "GEMINI.md",
		ConfigFiles: []ConfigFile{
			{Path: "~/.gemini/antigravity/mcp_config.json", Content: antigravityMCP, Merge: true, IsGlobal: true},
		},
	},
	"vscode-copilot": {
		Name:              "vscode-copilot",
		DisplayName:       "VS Code Copilot",
		HookLevel:         HookLevelFull,
		RequiresRestart:   true,
		RoutingFileRequired: false,
		ConfigFiles: []ConfigFile{
			{Path: ".vscode/mcp.json", Content: vscodeMCP, Merge: true},
			{Path: ".github/hooks/context-mode.json", Content: vscodeHooks},
		},
	},
	"jetbrains-copilot": {
		Name:                   "jetbrains-copilot",
		DisplayName:            "JetBrains Copilot",
		HookLevel:              HookLevelFull,
		RequiresManualMCPSetup: true,
		RequiresRestart:        true,
		ConfigFiles: []ConfigFile{
			{Path: ".github/hooks/context-mode.json", Content: jetbrainsHooks},
		},
	},
	"cursor": {
		Name:                "cursor",
		DisplayName:         "Cursor",
		HookLevel:           HookLevelPartial,
		RoutingFileRequired: true,
		RoutingFile:         ".cursor/rules/context-mode.mdc",
		ConfigFiles: []ConfigFile{
			{Path: ".cursor/mcp.json", Content: cursorMCP, Merge: true},
			{Path: ".cursor/hooks.json", Content: cursorHooks},
			{Path: ".cursor/rules/context-mode.mdc", Content: cursorRules},
		},
	},
	"kiro": {
		Name:                "kiro",
		DisplayName:         "Kiro",
		HookLevel:           HookLevelPartial,
		RoutingFileRequired: true,
		RoutingFile:         "KIRO.md",
		ConfigFiles: []ConfigFile{
			{Path: ".kiro/settings/mcp.json", Content: kiroMCP, Merge: true},
			{Path: ".kiro/hooks/context-mode.json", Content: kiroHooks},
		},
	},
	"zed": {
		Name:                "zed",
		DisplayName:         "Zed",
		HookLevel:           HookLevelNone,
		RoutingFileRequired: true,
		RoutingFile:         "AGENTS.md",
		ConfigFiles: []ConfigFile{
			{Path: "~/.config/zed/settings.json", Content: zedSettings, Merge: true, IsGlobal: true},
		},
	},
	"codex-cli": {
		Name:                "codex-cli",
		DisplayName:         "Codex CLI",
		HookLevel:           HookLevelPartial,
		RoutingFileRequired: true,
		RoutingFile:         "AGENTS.md",
		ConfigFiles: []ConfigFile{
			{Path: "~/.codex/config.toml", Content: codexTOML, Merge: true, IsGlobal: true},
			{Path: "~/.codex/hooks.json", Content: codexHooks, IsGlobal: true},
		},
	},
	"pi-agent": {
		Name:      "pi-agent",
		DisplayName: "Pi Coding Agent",
		HookLevel: HookLevelFull,
		ConfigFiles: []ConfigFile{
			{Path: "~/.pi/agent/mcp.json", Content: piMCP, Merge: true, IsGlobal: true},
		},
	},
	"claude-code": {
		Name:        "claude-code",
		DisplayName: "Claude Code",
		HookLevel:   HookLevelFull,
		// Claude Code uses plugin marketplace — no file config needed
	},
}

// Detect identifies the current platform from environment and file system.
func Detect(override string) (Platform, error) {
	if override != "" {
		if p, ok := platforms[override]; ok {
			return p, nil
		}
	}

	// Environment variable detection
	if os.Getenv("OPENCODE_SESSION") != "" {
		return platforms["opencode"], nil
	}
	if os.Getenv("CLAUDE_CODE_SESSION") != "" || os.Getenv("CLAUDE_CODE") != "" {
		return platforms["claude-code"], nil
	}
	if os.Getenv("GEMINI_CLI_SESSION") != "" || os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") != "" {
		return platforms["gemini-cli"], nil
	}

	// File-based detection
	switch {
	case fileExists("opencode.json"):
		return platforms["opencode"], nil
	case fileExists("kilo.json"):
		return platforms["kilo-code"], nil
	case fileExists(".cursor"):
		return platforms["cursor"], nil
	case fileExists(".kiro"):
		return platforms["kiro"], nil
	case fileExists(".antigravity"):
		return platforms["antigravity"], nil
	case fileExists(".vscode"):
		return platforms["vscode-copilot"], nil
	}

	return Platform{Name: "unknown"}, nil
}

// AllPlatforms returns the list of all configured platforms.
func AllPlatforms() []Platform {
	result := make([]Platform, 0, len(platforms))
	for _, p := range platforms {
		result = append(result, p)
	}
	return result
}

// PromptSelection prompts the user to select a platform from a list.
func PromptSelection() Platform {
	// Simple selection via prompt — returns unknown, caller handles
	return Platform{Name: "unknown"}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// String constants for platform-specific config files

const opencodeJSON = `{
  "mcp": {
    "context-mode": {
      "type": "local",
      "command": ["context-mode"]
    }
  },
  "plugin": ["context-mode"]
}`

const kiloJSON = `{
  "mcp": {
    "context-mode": {
      "type": "local",
      "command": ["context-mode"]
    }
  },
  "plugin": ["context-mode"]
}`

const geminiCLISettings = `{
  "mcpServers": {
    "context-mode": {
      "command": "context-mode"
    }
  },
  "hooks": {
    "BeforeTool": [{ "type": "command", "command": "context-mode hook gemini-cli beforetool" }],
    "AfterTool": [{ "type": "command", "command": "context-mode hook gemini-cli aftertool" }],
    "PreCompress": [{ "type": "command", "command": "context-mode hook gemini-cli precompress" }],
    "SessionStart": [{ "type": "command", "command": "context-mode hook gemini-cli sessionstart" }]
  }
}`

const antigravityMCP = `{
  "mcpServers": {
    "context-mode": {
      "command": "context-mode"
    }
  }
}`

const vscodeMCP = `{
  "servers": {
    "context-mode": {
      "command": "context-mode"
    }
  }
}`

const vscodeHooks = `{
  "hooks": {
    "PreToolUse": [{ "type": "command", "command": "context-mode hook vscode-copilot pretooluse" }],
    "PostToolUse": [{ "type": "command", "command": "context-mode hook vscode-copilot posttooluse" }],
    "SessionStart": [{ "type": "command", "command": "context-mode hook vscode-copilot sessionstart" }]
  }
}`

const jetbrainsHooks = `{
  "hooks": {
    "PreToolUse": [{ "type": "command", "command": "context-mode hook jetbrains-copilot pretooluse" }],
    "PostToolUse": [{ "type": "command", "command": "context-mode hook jetbrains-copilot posttooluse" }],
    "SessionStart": [{ "type": "command", "command": "context-mode hook jetbrains-copilot sessionstart" }]
  }
}`

const cursorMCP = `{
  "mcpServers": {
    "context-mode": {
      "command": "context-mode"
    }
  }
}`

const cursorHooks = `{
  "version": 1,
  "hooks": {
    "preToolUse": [{ "command": "context-mode hook cursor pretooluse", "matcher": "Shell|Read|Grep|WebFetch|Task|MCP:ctx_execute|MCP:ctx_execute_file|MCP:ctx_batch_execute" }],
    "postToolUse": [{ "command": "context-mode hook cursor posttooluse" }],
    "stop": [{ "command": "context-mode hook cursor stop" }]
  }
}`

const cursorRules = `# context-mode Routing Rules

Always use ctx_execute instead of raw bash for any command producing > 1KB output.
Always use ctx_batch_execute for 3+ parallel searches/commands.
Always use ctx_fetch_and_index + ctx_search for documentation lookups.

See: https://github.com/mksglu/context-mode
`

const kiroMCP = `{
  "mcpServers": {
    "context-mode": {
      "command": "context-mode"
    }
  }
}`

const kiroHooks = `{
  "hooks": {
    "preToolUse": [{ "command": "context-mode hook kiro pretooluse" }],
    "postToolUse": [{ "command": "context-mode hook kiro posttooluse" }]
  }
}`

const zedSettings = `{
  "context_servers": {
    "context-mode": {
      "command": { "path": "context-mode" }
    }
  }
}`

const codexTOML = `[mcp_servers.context-mode]
command = "context-mode"
`

const codexHooks = `{
  "hooks": {
    "PreToolUse": [{ "type": "command", "command": "context-mode hook codex-cli pretooluse" }],
    "PostToolUse": [{ "type": "command", "command": "context-mode hook codex-cli posttooluse" }],
    "SessionStart": [{ "type": "command", "command": "context-mode hook codex-cli sessionstart" }]
  }
}`

const piMCP = `{
  "mcpServers": {
    "context-mode": {
      "command": "context-mode"
    }
  }
}`
