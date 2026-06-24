package mcp

import (
	"os/exec"
	"strings"
)

const SeqThinkNPXPackage = "@modelcontextprotocol/server-sequential-thinking"

var defaultSeqThinkServerJSON = []byte(`{
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"]
}`)

var defaultSeqThinkOverlayJSON = []byte(`{
  "mcpServers": {
    "sequential-thinking": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"]
    }
  }
}`)

var openCodeSeqThinkOverlayJSON = []byte(`{
  "mcp": {
    "sequential-thinking": {
      "type": "local",
      "command": ["npx", "-y", "@modelcontextprotocol/server-sequential-thinking"],
      "enabled": true
    }
  }
}`)

func DefaultSeqThinkServerJSON() []byte {
	content := make([]byte, len(defaultSeqThinkServerJSON))
	copy(content, defaultSeqThinkServerJSON)
	return content
}

func DefaultSeqThinkOverlayJSON() []byte {
	content := make([]byte, len(defaultSeqThinkOverlayJSON))
	copy(content, defaultSeqThinkOverlayJSON)
	return content
}

func OpenCodeSeqThinkOverlayJSON() []byte {
	content := make([]byte, len(openCodeSeqThinkOverlayJSON))
	copy(content, openCodeSeqThinkOverlayJSON)
	return content
}

func DetectSequentialThinking() bool {
	if _, err := exec.LookPath("npx"); err != nil {
		return false
	}
	out, _ := exec.Command("npm", "list", "-g", SeqThinkNPXPackage).Output()
	return !strings.Contains(string(out), "empty")
}
