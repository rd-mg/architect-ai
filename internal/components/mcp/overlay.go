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
)

type Options struct{}

func OverlayFor(agent model.AgentID, server ServerKind, opts Options) ([]byte, error) {
	switch server {
	case ServerContext7:
		return context7Overlay(agent)
	case ServerNotebookLM:
		return notebookLMOverlay(agent)
	default:
		return nil, fmt.Errorf("unsupported MCP server: %s", server)
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
	case model.AgentClaudeCode:
		return []byte(`{
  "command": "notebooklm-mcp",
  "args": []
}`), nil
	default:
		return nil, fmt.Errorf("unsupported agent for notebooklm: %s", agent)
	}
}
