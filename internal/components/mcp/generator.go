package mcp

import (
	"fmt"
	"os/exec"
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
		"general": map[string]interface{}{"defaultApprovalMode": "auto_edit"},
		"ide":     map[string]interface{}{"enabled": true},
		"mcpServers": map[string]interface{}{
			"context7":    map[string]interface{}{"httpUrl": "https://mcp.context7.com/mcp", "timeout": 30000, "trust": false},
			"context-mode": map[string]interface{}{"command": "npx", "args": []string{"-y", "@mksglu/context-mode"}, "timeout": 15000},
			"engram":      map[string]interface{}{"command": engramBin, "args": []string{"mcp", "--tools=agent"}},
			"sequential-thinking": map[string]interface{}{
				"command": "npx", "args": []string{"-y", "@modelcontextprotocol/server-sequential-thinking"},
				"timeout": 30000, "trust": true,
			},
		},
		"model":    map[string]interface{}{"name": ""},
		"security": map[string]interface{}{"auth": map[string]interface{}{"selectedType": "oauth-personal"}},
		"ui":       map[string]interface{}{"hideFooter": true, "showCitations": true, "showMemoryUsage": true, "showModelInfoInChat": true},
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
		},
	}
}

func generateClaude(engramBin string, opts GenerateOptions) map[string]interface{} {
	return map[string]interface{}{
		"mcp_servers": map[string]interface{}{
			"engram":             map[string]interface{}{"command": engramBin, "args": []string{"mcp", "--tools=agent"}, "type": "stdio"},
			"context7":           map[string]interface{}{"command": "npx", "args": []string{"-y", "@upstash/context7-mcp@latest"}, "type": "stdio"},
			"sequential_thinking": map[string]interface{}{"command": "npx", "args": []string{"-y", "@modelcontextprotocol/server-sequential-thinking"}, "type": "stdio"},
			"context_mode":       map[string]interface{}{"command": "npx", "args": []string{"-y", "@mksglu/context-mode"}, "type": "stdio"},
		},
	}
}

func boolStr(b bool) string { if b { return "true" }; return "false" }
func isGeminiInstalled() bool { _, err := exec.LookPath("gemini"); return err == nil }
