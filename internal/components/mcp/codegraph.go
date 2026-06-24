package mcp

import (
	"os/exec"
	"strings"
)

const CodeGraphNPXPackage = "@colbymchenry/codegraph"

var defaultCodeGraphServerJSON = []byte(`{
  "command": "npx",
  "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"]
}`)

var defaultCodeGraphOverlayJSON = []byte(`{
  "mcpServers": {
    "codegraph": {
      "command": "npx",
      "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"]
    }
  }
}`)

var openCodeCodeGraphOverlayJSON = []byte(`{
  "mcp": {
    "codegraph": {
      "type": "local",
      "command": ["npx", "-y", "@colbymchenry/codegraph", "serve", "--mcp"],
      "enabled": true
    }
  }
}`)

var vsCodeCodeGraphOverlayJSON = []byte(`{
  "servers": {
    "codegraph": {
      "type": "http",
      "url": "http://localhost:7340/mcp"
    }
  }
}`)

func DefaultCodeGraphServerJSON() []byte {
	content := make([]byte, len(defaultCodeGraphServerJSON))
	copy(content, defaultCodeGraphServerJSON)
	return content
}

func DefaultCodeGraphOverlayJSON() []byte {
	content := make([]byte, len(defaultCodeGraphOverlayJSON))
	copy(content, defaultCodeGraphOverlayJSON)
	return content
}

func OpenCodeCodeGraphOverlayJSON() []byte {
	content := make([]byte, len(openCodeCodeGraphOverlayJSON))
	copy(content, openCodeCodeGraphOverlayJSON)
	return content
}

func VSCodeCodeGraphOverlayJSON() []byte {
	content := make([]byte, len(vsCodeCodeGraphOverlayJSON))
	copy(content, vsCodeCodeGraphOverlayJSON)
	return content
}

func CodeGraphMCPCommand() (string, []string) {
	if path, err := exec.LookPath("codegraph"); err == nil {
		return path, []string{"serve", "--mcp"}
	}
	return "npx", []string{"-y", CodeGraphNPXPackage, "serve", "--mcp"}
}

func CodeGraphIsAvailable() bool {
	_, errBin := exec.LookPath("codegraph")
	_, errNPX := exec.LookPath("npx")
	return errBin == nil || errNPX == nil
}

func DetectCodeGraph() bool {
	out, _ := exec.Command("npx", "--yes", CodeGraphNPXPackage, "--version").Output()
	return len(strings.TrimSpace(string(out))) > 0
}
