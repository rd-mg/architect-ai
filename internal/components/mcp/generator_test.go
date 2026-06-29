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
	if isGeminiInstalled() {
		t.Skip("gemini binary is on PATH — cannot test absent case in this environment")
	}
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

func TestGenerateClaudeKeyNames(t *testing.T) {
	cfg, err := GenerateConfig("claude", GenerateOptions{})
	if err != nil {
		t.Fatalf("GenerateConfig(claude) error = %v", err)
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal error = %v", err)
	}
	text := string(data)

	for _, key := range []string{"sequential-thinking", "context-mode", "notebooklm-mcp"} {
		if !strings.Contains(text, `"`+key+`"`) {
			t.Errorf("claude config missing key %q", key)
		}
	}
	for _, key := range []string{"sequential_thinking", "context_mode"} {
		if strings.Contains(text, `"`+key+`"`) {
			t.Errorf("claude config should NOT contain key %q", key)
		}
	}
}

func TestAllGeneratorsHaveNotebookLM(t *testing.T) {
	platforms := []string{"claude", "vscode", "antigravity", "gemini", "opencode"}
	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			cfg, err := GenerateConfig(platform, GenerateOptions{})
			if err != nil {
				t.Fatalf("GenerateConfig(%s) error = %v", platform, err)
			}
			data, err := json.Marshal(cfg)
			if err != nil {
				t.Fatalf("Marshal(%s) error = %v", platform, err)
			}
			if !strings.Contains(string(data), "notebooklm-mcp") {
				t.Errorf("%s config missing notebooklm-mcp", platform)
			}
		})
	}
}

func TestGenerateGemini_OnlyMcpServers(t *testing.T) {
	cfg, err := GenerateConfig("gemini", GenerateOptions{})
	if err != nil {
		t.Fatalf("GenerateConfig(gemini) error = %v", err)
	}
	for _, key := range []string{"general", "ide", "model", "security", "ui"} {
		if _, ok := cfg[key]; ok {
			t.Errorf("gemini config should NOT contain top-level key %q", key)
		}
	}
	if _, ok := cfg["mcpServers"]; !ok {
		t.Error("gemini config must contain 'mcpServers' key")
	}
}
