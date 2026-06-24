package mcp

import (
	"fmt"

	"github.com/rd-mg/architect-ai/internal/model"
)

type ServerKind string

const (
	ServerContext7   ServerKind = "context7"
	ServerEngram     ServerKind = "engram"
	ServerNotebookLM ServerKind = "notebooklm-mcp"
	ServerSequentialThinking ServerKind = "sequential-thinking"
	ServerCodeGraph  ServerKind = "codegraph"
)

type Options struct{}

func OverlayFor(agent model.AgentID, server ServerKind, opts Options) ([]byte, error) {
	switch server {
	case ServerContext7:
		return context7Overlay(agent)
	case ServerNotebookLM:
		return notebookLMOverlay(agent)
	case ServerSequentialThinking:
		return sequentialThinkingOverlay(agent)
	case ServerCodeGraph:
		return codegraphOverlay(agent)
	default:
		return nil, fmt.Errorf("unsupported MCP server: %s", server)
	}
}

func sequentialThinkingOverlay(agent model.AgentID) ([]byte, error) {
	switch agent {
	case model.AgentGeminiCLI:
		return []byte(`{
  "mcpServers": {
    "sequential-thinking": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-sequential-thinking"
      ],
      "timeout": 30000,
      "trust": true
    }
  }
}`), nil
	case model.AgentOpenCode, model.AgentKilocode:
		return []byte(`{
  "mcp": {
    "sequential-thinking": {
      "type": "local",
      "command": [
        "npx",
        "-y",
        "@modelcontextprotocol/server-sequential-thinking"
      ],
      "enabled": true
    }
  }
}`), nil
	case model.AgentVSCodeCopilot, model.AgentCursor:
		return []byte(`{
  "servers": {
    "sequential-thinking": {
      "type": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-sequential-thinking"
      ]
    }
  }
}`), nil
	case model.AgentAntigravity, model.AgentWindsurf, model.AgentQwenCode, model.AgentKiroIDE:
		return []byte(`{
  "mcpServers": {
    "sequential-thinking": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-sequential-thinking"
      ]
    }
  }
}`), nil
	case model.AgentClaudeCode:
		return []byte(`{
  "command": "npx",
  "args": [
    "-y",
    "@modelcontextprotocol/server-sequential-thinking"
  ]
}`), nil
	default:
		return nil, fmt.Errorf("unsupported agent for sequential-thinking: %s", agent)
	}
}

func context7Overlay(agent model.AgentID) ([]byte, error) {
	switch agent {
	case model.AgentGeminiCLI:
		return []byte(`{
  "mcpServers": {
    "context7": {
      "httpUrl": "https://mcp.context7.com/mcp",
      "timeout": 30000,
      "trust": false
    }
  }
}`), nil
	case model.AgentOpenCode, model.AgentKilocode:
		return []byte(`{
  "mcp": {
    "context7": {
      "type": "remote",
      "url": "https://mcp.context7.com/mcp",
      "enabled": true
    }
  }
}`), nil
	case model.AgentVSCodeCopilot, model.AgentCursor:
		return []byte(`{
  "servers": {
    "context7": {
      "type": "http",
      "url": "https://mcp.context7.com/mcp"
    }
  }
}`), nil
	case model.AgentAntigravity:
		return []byte(`{
  "mcpServers": {
    "context7": {
      "serverUrl": "https://mcp.context7.com/mcp"
    }
  }
}`), nil
	case model.AgentWindsurf:
		return []byte(`{
  "mcpServers": {
    "context7": {
      "serverUrl": "https://mcp.context7.com/mcp"
    }
  }
}`), nil
	case model.AgentQwenCode:
		return []byte(`{
  "mcpServers": {
    "context7": {
      "serverUrl": "https://mcp.context7.com/mcp"
    }
  }
}`), nil
	case model.AgentKiroIDE:
		return []byte(`{
  "mcpServers": {
    "context7": {
      "serverUrl": "https://mcp.context7.com/mcp"
    }
  }
}`), nil
	case model.AgentClaudeCode:
		return []byte(`{
  "command": "npx",
  "args": [
    "-y",
    "@upstash/context7-mcp"
  ]
}`), nil
	default:
		return nil, fmt.Errorf("unsupported agent for context7: %s", agent)
	}
}

func notebookLMOverlay(agent model.AgentID) ([]byte, error) {
	switch agent {
	case model.AgentGeminiCLI:
		return []byte(`{
  "mcpServers": {
    "notebooklm-mcp": {
      "command": "notebooklm-mcp",
      "args": [],
      "timeout": 30000,
      "trust": false
    }
  }
}`), nil
	case model.AgentOpenCode, model.AgentKilocode:
		return []byte(`{
  "mcp": {
    "notebooklm-mcp": {
      "type": "local",
      "command": ["notebooklm-mcp"],
      "enabled": false
    }
  }
}`), nil
	case model.AgentVSCodeCopilot, model.AgentCursor:
		return []byte(`{
  "servers": {
    "notebooklm-mcp": {
      "type": "stdio",
      "command": "notebooklm-mcp",
      "args": []
    }
  }
}`), nil
	case model.AgentAntigravity:
		return []byte(`{
  "mcpServers": {
    "notebooklm-mcp": {
      "command": "notebooklm-mcp",
      "args": []
    }
  }
}`), nil
	case model.AgentWindsurf:
		return []byte(`{
  "mcpServers": {
    "notebooklm-mcp": {
      "command": "notebooklm-mcp",
      "args": []
    }
  }
}`), nil
	case model.AgentQwenCode:
		return []byte(`{
  "mcpServers": {
    "notebooklm-mcp": {
      "command": "notebooklm-mcp",
      "args": []
    }
  }
}`), nil
	case model.AgentKiroIDE:
		return []byte(`{
  "mcpServers": {
    "notebooklm-mcp": {
      "command": "notebooklm-mcp",
      "args": []
    }
  }
}`), nil
	case model.AgentClaudeCode:
		return []byte(`{
  "command": "notebooklm-mcp",
  "args": []
}`), nil
	default:
		return nil, fmt.Errorf("unsupported agent for notebooklm: %s", agent)
	}
}

func codegraphOverlay(agent model.AgentID) ([]byte, error) {
	switch agent {
	case model.AgentGeminiCLI:
		return []byte(`{
  "mcpServers": {
    "codegraph": {
      "command": "npx",
      "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"],
      "timeout": 30000,
      "trust": true
    }
  }
}`), nil
	case model.AgentOpenCode, model.AgentKilocode:
		return []byte(`{
  "mcp": {
    "codegraph": {
      "type": "local",
      "command": ["npx", "-y", "@colbymchenry/codegraph", "serve", "--mcp"],
      "enabled": true
    }
  }
}`), nil
	case model.AgentVSCodeCopilot, model.AgentCursor:
		return []byte(`{
  "servers": {
    "codegraph": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"]
    }
  }
}`), nil
	case model.AgentAntigravity, model.AgentWindsurf, model.AgentQwenCode, model.AgentKiroIDE:
		return []byte(`{
  "mcpServers": {
    "codegraph": {
      "command": "npx",
      "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"]
    }
  }
}`), nil
	case model.AgentClaudeCode:
		return []byte(`{
  "command": "npx",
  "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"]
}`), nil
	default:
		return nil, fmt.Errorf("unsupported agent for codegraph: %s", agent)
	}
}
