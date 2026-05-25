# Design: Phase 10 — MCP TUI Configurator

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/10-phase-mcp-tui-configurator.md`
> **Change:** phase-10-mcp-tui-configurator
> **Phase:** sdd-design
> **Generated:** 2026-05-22T19:17:00Z
> **Author:** sdd-design (architect-ai)

## Architecture

The MCP Configurator centralizes divergent MCP config formats via a `GenerateConfig()` strategy pattern in Go. Each platform gets its own generator function that produces the exact JSON schema required.

## FMEA Matrix

| Component | Failure Mode | Effect | P | S | RPN | Mitigation |
|---|---|---|---|---|---|---|
| Schema Generator | Hybrid transport keys emitted | Platform fails to load MCP server | 1 | 4 | 4 | Tests assert absence of wrong keys |
| Secrets Engine | Password written to JSON | Credential leaked to VC | 1 | 5 | 5 | `WriteSecretsEnv()` decouples storage |
| Config Writer | Process crash during write | Corrupted config file | 1 | 4 | 4 | Atomic write: .tmp + os.Rename() |
| Engram Discovery | Binary not found | Local persistence broken | 2 | 3 | 6 | Multi-tier fallback discovery |

## Go Implementation

### `internal/components/mcp/generator.go`

```go
package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
```

### `internal/components/mcp/engram_path.go`

```go
package mcp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func FindEngramBinary() (string, error) {
	if env := os.Getenv("ENGRAM_BIN"); env != "" && isExec(env) { return env, nil }
	if path, err := exec.LookPath("engram"); err == nil { return path, nil }
	home, _ := os.UserHomeDir()
	candidates := []string{
		"/usr/local/bin/engram", "/usr/bin/engram", "/opt/homebrew/bin/engram",
		filepath.Join(home, ".linuxbrew", "bin", "engram"),
		"/home/linuxbrew/.linuxbrew/bin/engram",
	}
	for _, c := range candidates {
		if isExec(c) { return c, nil }
	}
	for _, base := range []string{"/home/linuxbrew/.linuxbrew/Cellar", "/opt/homebrew/Cellar", "/usr/local/Cellar"} {
		pkgDir := filepath.Join(base, "engram")
		entries, err := os.ReadDir(pkgDir)
		if err != nil || len(entries) == 0 { continue }
		bin := filepath.Join(pkgDir, entries[len(entries)-1].Name(), "bin", "engram")
		if isExec(bin) { return bin, nil }
	}
	return "", fmt.Errorf("engram not found; install: brew install engram OR set ENGRAM_BIN")
}

func isExec(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Mode()&0111 != 0
}
```

### `internal/components/mcp/secrets.go`

```go
package mcp

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WriteSecretsEnv(projectDir string, secrets map[string]string) error {
	envPath := filepath.Join(projectDir, ".env.mcp")
	ensureGitignored(filepath.Join(projectDir, ".gitignore"), ".env.mcp")
	existing := readDotEnv(envPath)
	for k, v := range secrets { existing[k] = v }
	var lines []string
	lines = append(lines, "# MCP Secrets — generated by architect-ai", "# DO NOT COMMIT — gitignored", "")
	for k, v := range existing { lines = append(lines, fmt.Sprintf("%s=%s", k, v)) }
	return os.WriteFile(envPath, []byte(strings.Join(lines, "\n")+"\n"), 0600)
}

func readDotEnv(path string) map[string]string {
	result := make(map[string]string)
	f, err := os.Open(path)
	if err != nil { return result }
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") { continue }
		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func ensureGitignored(path, pattern string) {
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), pattern) { return }
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil { return }
	defer f.Close()
	fmt.Fprintf(f, "\n# MCP secrets\n%s\n", pattern)
}

func WriteConfig(targetPath string, config map[string]interface{}) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil { return fmt.Errorf("marshal: %w", err) }
	tmp := targetPath + ".tmp"
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil { return fmt.Errorf("mkdir: %w", err) }
	if err := os.WriteFile(tmp, data, 0644); err != nil { return fmt.Errorf("write tmp: %w", err) }
	return os.Rename(tmp, targetPath)
}
```

### `internal/components/mcp/generator_test.go`

```go
package mcp

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestGenerateVSCode_HasServersKey(t *testing.T) {
	cfg, _ := GenerateConfig("vscode", GenerateOptions{})
	if _, ok := cfg["servers"]; !ok { t.Error("VSCode must use 'servers' key") }
	if _, ok := cfg["mcpServers"]; ok { t.Error("VSCode must NOT use 'mcpServers'") }
}

func TestGenerateAntigravity_Context7PureServerUrl(t *testing.T) {
	cfg, _ := GenerateConfig("antigravity", GenerateOptions{})
	servers := cfg["mcpServers"].(map[string]interface{})
	ctx7 := servers["context7"].(map[string]interface{})
	if _, has := ctx7["command"]; has { t.Error("must NOT have command") }
	if _, has := ctx7["args"]; has { t.Error("must NOT have args") }
	if _, has := ctx7["serverUrl"]; !has { t.Error("must have serverUrl") }
}

func TestGenerateGemini_Context7PureHttpUrl(t *testing.T) {
	cfg, _ := GenerateConfig("gemini", GenerateOptions{})
	servers := cfg["mcpServers"].(map[string]interface{})
	ctx7 := servers["context7"].(map[string]interface{})
	if _, has := ctx7["command"]; has { t.Error("must NOT have command") }
	if _, has := ctx7["httpUrl"]; !has { t.Error("must have httpUrl") }
}

func TestGenerateAntigravity_OdooPasswordNotInline(t *testing.T) {
	cfg, _ := GenerateConfig("antigravity", GenerateOptions{IsOdooProject: true, OdooURL: "http://localhost:8069", OdooDB: "test", OdooUser: "admin"})
	servers := cfg["mcpServers"].(map[string]interface{})
	odoo := servers["odoo"].(map[string]interface{})
	env := odoo["env"].(map[string]string)
	if !strings.HasPrefix(env["ODOO_PASSWORD"], "${") {
		t.Errorf("must be env var ref, got: %s", env["ODOO_PASSWORD"])
	}
}

func TestGenerateOpenCode_GeminiPlugin(t *testing.T) {
	cfg, _ := GenerateConfig("opencode", GenerateOptions{GeminiInstalled: true})
	plugins := cfg["plugin"].([]string)
	found := false
	for _, p := range plugins { if p == "opencode-gemini-auth@latest" { found = true } }
	if !found { t.Error("must include gemini-auth plugin") }
}

func TestGenerateOpenCode_NoGeminiPlugin(t *testing.T) {
	cfg, _ := GenerateConfig("opencode", GenerateOptions{GeminiInstalled: false})
	plugins := cfg["plugin"].([]string)
	for _, p := range plugins {
		if p == "opencode-gemini-auth@latest" { t.Error("must NOT include gemini-auth") }
	}
}

func TestWriteConfig_Atomic(t *testing.T) {
	dir := t.TempDir()
	target := dir + "/subdir/config.json"
	cfg := map[string]interface{}{"test": "value"}
	if err := WriteConfig(target, cfg); err != nil { t.Fatal(err) }
	data, err := os.ReadFile(target)
	if err != nil { t.Fatal(err) }
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil { t.Fatal("invalid JSON") }
	if _, err := os.Stat(target + ".tmp"); !os.IsNotExist(err) { t.Error("tmp should be removed") }
}
```

## Key Decisions

- **Transport Schema Purity**: Strict per-platform functions prevent field leakage.
- **Secrets Architecture**: `.env.mcp` + interpolation removes user-error vectors.
- **Engram Discovery**: Cellar parsing accommodates Homebrew updates without re-init.
- **Atomic Writes**: `.tmp` + rename prevents half-written configs.

## MANDATORY IMPLEMENTATION DIRECTIVE
**For `sdd-apply` (Coder):**
All code definitions, declarations, directives, and logic provided in the Source of Truth MUST be implemented EXACTLY as written on the first attempt. The coder MUST NOT analyze, re-design, or deviate from the source code during the initial implementation. Modification of the initial implementation is ONLY permitted if the tests fail.
