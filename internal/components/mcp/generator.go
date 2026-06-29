package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/rd-mg/architect-ai/internal/system"
)

type GenerateOptions struct {
	IsOdooProject   bool
	OdooURL         string
	OdooDB          string
	OdooUser        string
	OdooYolo        bool
	PostgresURL     string
	GeminiInstalled bool
}

func GenerateConfig(platform string, opts GenerateOptions) (map[string]interface{}, error) {
	engramBin, err := FindEngramBinary()
	if err != nil {
		engramBin = "engram"
	}

	// Emit warning if Odoo MCP is requested but uv/uvx is not available.
	if opts.IsOdooProject {
		uvAvailable, uvVersion := system.CheckUV(context.Background())
		if !uvAvailable {
			fmt.Fprintln(os.Stderr,
				"WARNING: Odoo MCP requires 'uvx' (uv package manager). "+
					"Install: curl -LsSf https://astral.sh/uv/install.sh | sh")
		} else {
			_ = uvVersion // available for debugging
		}
	}

	switch platform {
	case "vscode":
		return generateVSCode(engramBin, opts), nil
	case "antigravity":
		return generateAntigravity(engramBin, opts), nil
	case "gemini":
		return generateGemini(engramBin, opts), nil
	case "opencode":
		return generateOpenCode(engramBin, opts), nil
	case "claude":
		return generateClaude(engramBin, opts), nil
	case "antigravity-cli":
		return generateAntigravityCLI(engramBin, opts), nil
	default:
		return nil, fmt.Errorf("unknown platform: %s", platform)
	}
}

func generateVSCode(engramBin string, opts GenerateOptions) map[string]interface{} {
	servers := map[string]interface{}{
		"context-mode": map[string]interface{}{
			"type": "stdio", "command": "npx",
			"args": []string{"-y", "@mksglu/context-mode"},
		},
		"context7": map[string]interface{}{"type": "http", "url": "https://mcp.context7.com/mcp"},
		"engram": map[string]interface{}{
			"type": "stdio", "command": engramBin,
			"args": []string{"mcp", "--tools=agent"},
		},
		"sequential-thinking": map[string]interface{}{
			"type": "stdio", "command": "npx",
			"args": []string{"-y", "@modelcontextprotocol/server-sequential-thinking"},
		},
		"codegraph": map[string]interface{}{
			"type": "stdio", "command": "npx",
			"args": []string{"-y", "@colbymchenry/codegraph", "serve", "--mcp"},
		},
		"notebooklm-mcp": map[string]interface{}{
			"type": "stdio", "command": "notebooklm-mcp", "args": []string{},
		},
	}
	result := map[string]interface{}{"servers": servers}
	if opts.IsOdooProject {
		servers["odoo"] = map[string]interface{}{
			"type": "stdio", "command": "uvx", "args": []string{"mcp-server-odoo"},
			"env": map[string]string{
				"ODOO_DB": opts.OdooDB, "ODOO_PASSWORD": "${input:odoo-password}",
				"ODOO_URL": opts.OdooURL, "ODOO_USER": opts.OdooUser, "ODOO_YOLO": boolStr(opts.OdooYolo),
			},
		}
		if opts.PostgresURL != "" {
			servers["postgres"] = map[string]interface{}{
				"type": "stdio", "command": "npx",
				"args": []string{"-y", "@modelcontextprotocol/server-postgres", opts.PostgresURL},
			}
		}
		result["inputs"] = []map[string]interface{}{
			{"type": "promptString", "id": "odoo-password", "description": "Odoo MCP password", "password": true},
		}
	}
	return result
}

func generateAntigravity(engramBin string, opts GenerateOptions) map[string]interface{} {
	servers := map[string]interface{}{
		"context7":    map[string]interface{}{"serverUrl": "https://mcp.context7.com/mcp"},
		"context-mode": map[string]interface{}{"command": "npx", "args": []string{"-y", "@mksglu/context-mode"}},
		"engram":      map[string]interface{}{"command": engramBin, "args": []string{"mcp", "--tools=agent"}},
		"sequential-thinking": map[string]interface{}{
			"command": "npx", "args": []string{"-y", "@modelcontextprotocol/server-sequential-thinking"},
			"timeout": 30000, "trust": true,
		},
		"codegraph": map[string]interface{}{
			"command": "npx", "args": []string{"-y", "@colbymchenry/codegraph", "serve", "--mcp"},
			"timeout": 30000, "trust": true,
		},
		"notebooklm-mcp": map[string]interface{}{
			"command": "notebooklm-mcp", "args": []string{},
		},
	}
	if opts.IsOdooProject {
		servers["odoo"] = map[string]interface{}{
			"command": "uvx", "args": []string{"mcp-server-odoo"},
			"env": map[string]string{
				"ODOO_DB": opts.OdooDB, "ODOO_PASSWORD": "${ODOO_PASSWORD}",
				"ODOO_URL": opts.OdooURL, "ODOO_USER": opts.OdooUser, "ODOO_YOLO": boolStr(opts.OdooYolo),
			},
		}
	}
	return map[string]interface{}{"mcpServers": servers}
}

func generateGemini(engramBin string, opts GenerateOptions) map[string]interface{} {
	return map[string]interface{}{
		"mcpServers": generateGeminiMCPServers(engramBin, opts),
	}
}

// generateGeminiMCPServers returns the shared MCP servers map for Gemini CLI.
// Extracted so it can be reused by generateGemini and generateGeminiMCPOnly.
func generateGeminiMCPServers(engramBin string, opts GenerateOptions) map[string]interface{} {
	return map[string]interface{}{
		"context7":    map[string]interface{}{"httpUrl": "https://mcp.context7.com/mcp", "timeout": 30000, "trust": false},
		"context-mode": map[string]interface{}{"command": "npx", "args": []string{"-y", "@mksglu/context-mode"}, "timeout": 15000},
		"engram":      map[string]interface{}{"command": engramBin, "args": []string{"mcp", "--tools=agent"}},
		"sequential-thinking": map[string]interface{}{
			"command": "npx", "args": []string{"-y", "@modelcontextprotocol/server-sequential-thinking"},
			"timeout": 30000, "trust": true,
		},
		"codegraph": map[string]interface{}{
			"command": "npx", "args": []string{"-y", "@colbymchenry/codegraph", "serve", "--mcp"},
			"timeout": 30000, "trust": true,
		},
		"notebooklm-mcp": map[string]interface{}{
			"command": "notebooklm-mcp", "args": []string{}, "timeout": 30000, "trust": false,
		},
	}
}

// generateGeminiMCPOnly returns ONLY the mcpServers section for MERGE operations.
// Used by injectMergeIntoSettings when settings.json already exists,
// to avoid overwriting user settings (general, ui, model, etc.).
func generateGeminiMCPOnly(engramBin string, opts GenerateOptions) map[string]interface{} {
	return map[string]interface{}{
		"mcpServers": generateGeminiMCPServers(engramBin, opts),
	}
}

func generateOpenCode(engramBin string, opts GenerateOptions) map[string]interface{} {
	tr := true
	plugins := []string{".atl/plugins/background-agents.ts"}
	if opts.GeminiInstalled || isGeminiInstalled() {
		plugins = append(plugins, "opencode-gemini-auth@latest")
	}
	return map[string]interface{}{
		"$schema": "https://opencode.ai/config.json",
		"plugin":  plugins,
		"mcp": map[string]interface{}{
			"context7":    map[string]interface{}{"enabled": &tr, "type": "remote", "url": "https://mcp.context7.com/mcp"},
			"context-mode": map[string]interface{}{"type": "local", "command": []string{"npx", "-y", "@mksglu/context-mode"}, "enabled": &tr},
			"engram":      map[string]interface{}{"type": "local", "command": []string{engramBin, "mcp", "--tools=agent"}},
			"sequential-thinking": map[string]interface{}{"type": "local", "command": []string{"npx", "-y", "@modelcontextprotocol/server-sequential-thinking"}, "enabled": &tr},
			"codegraph": map[string]interface{}{"type": "local", "command": []string{"npx", "-y", "@colbymchenry/codegraph", "serve", "--mcp"}, "enabled": &tr},
			"notebooklm-mcp": map[string]interface{}{"type": "local", "command": []string{"notebooklm-mcp"}, "enabled": &tr},
		},
	}
}

func generateClaude(engramBin string, opts GenerateOptions) map[string]interface{} {
	return map[string]interface{}{
		"mcp_servers": map[string]interface{}{
			"engram":              map[string]interface{}{"command": engramBin, "args": []string{"mcp", "--tools=agent"}, "type": "stdio"},
			"context7":            map[string]interface{}{"command": "npx", "args": []string{"-y", "@upstash/context7-mcp"}, "type": "stdio"},
			"sequential-thinking": map[string]interface{}{"command": "npx", "args": []string{"-y", "@modelcontextprotocol/server-sequential-thinking"}, "type": "stdio"},
			"context-mode":        map[string]interface{}{"command": "npx", "args": []string{"-y", "@mksglu/context-mode"}, "type": "stdio"},
			"codegraph":           map[string]interface{}{"command": "npx", "args": []string{"-y", "@colbymchenry/codegraph", "serve", "--mcp"}, "type": "stdio"},
			"notebooklm-mcp":      map[string]interface{}{"command": "notebooklm-mcp", "args": []string{}, "type": "stdio"},
		},
	}
}

func generateAntigravityCLI(engramBin string, opts GenerateOptions) map[string]interface{} {
	servers := map[string]interface{}{
		"engram":              map[string]interface{}{"command": engramBin, "args": []string{"mcp", "--tools=agent"}},
		"context7":            map[string]interface{}{"serverUrl": "https://mcp.context7.com/mcp"},
		"sequential-thinking": map[string]interface{}{"command": "npx", "args": []string{"-y", "@modelcontextprotocol/server-sequential-thinking"}},
		"context-mode":        map[string]interface{}{"command": "context-mode", "args": []string{"--mcp"}},
		"codegraph":           map[string]interface{}{"command": "npx", "args": []string{"-y", "@colbymchenry/codegraph", "serve", "--mcp"}},
		"notebooklm-mcp":      map[string]interface{}{"command": "notebooklm-mcp", "args": []string{}},
	}
	return map[string]interface{}{"mcpServers": servers}
}

func boolStr(b bool) string { if b { return "true" }; return "false" }
func isGeminiInstalled() bool { _, err := exec.LookPath("gemini"); return err == nil }
