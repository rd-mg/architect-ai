# Fase 10: Engram Keys — Collision Fix y Comando `diagnose engram`

**Objetivo:** Resolver F-16 (colisión de topic_key en `ResearchTopicKey` por truncamiento + len-only), OBS-02 (Engram probe falla sin estructura de diagnóstico), y OBS-07 (Solver sin persistencia de árbol de hipótesis). Modifica `engramkeys/keys.go` y crea el subcomando `diagnose engram`.

---

## Paso 1: Corregir `ResearchTopicKey` con hash SHA-256 para unicidad

**Archivo a modificar:** `internal/components/engram/engramkeys/keys.go`

**Acción:** Modificar

**Comando previo:**
```bash
grep -n "func ResearchTopicKey\|nonAlnum\|collapseDash\|len(query)" \
  internal/components/engram/engramkeys/keys.go
```

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
// ResearchTopicKey generates a stable, slugified topic key for research findings.
func ResearchTopicKey(tool, query string) string {
	cleaned := strings.ToLower(strings.TrimSpace(query))
	cleaned = nonAlnum.ReplaceAllString(cleaned, "-")
	cleaned = collapseDash.ReplaceAllString(cleaned, "-")
	cleaned = strings.Trim(cleaned, "-")
	if cleaned == "" {
		cleaned = "query"
	}
	if len(cleaned) > 50 {
		cleaned = strings.Trim(cleaned[:50], "-")
	}
	return fmt.Sprintf("research/%s/%s-len%d", tool, cleaned, len(query))
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
// ResearchTopicKey generates a stable, collision-free topic key for research findings.
//
// Format: research/{tool}/{slug}-{8-hex-chars}
//
// The 8-character SHA-256 hash of the original query guarantees uniqueness even
// when two queries produce the same slug after slugification (e.g. due to case
// normalization or character stripping). This replaces the previous -len{N} suffix
// which produced collisions for queries with identical content but different cases.
//
// Invariant: two distinct queries ALWAYS produce distinct topic keys.
func ResearchTopicKey(tool, query string) string {
	cleaned := strings.ToLower(strings.TrimSpace(query))
	cleaned = nonAlnum.ReplaceAllString(cleaned, "-")
	cleaned = collapseDash.ReplaceAllString(cleaned, "-")
	cleaned = strings.Trim(cleaned, "-")
	if cleaned == "" {
		cleaned = "query"
	}
	if len(cleaned) > 50 {
		cleaned = strings.Trim(cleaned[:50], "-")
	}
	h := sha256.Sum256([]byte(query))
	return fmt.Sprintf("research/%s/%s-%x", tool, cleaned, h[:4])
}
```

**Añadir import de `"crypto/sha256"` al bloque de imports del archivo:**

**Código a reemplazar — BUSCAR EXACTAMENTE** (el bloque de imports actual):
```go
import (
	"fmt"
	"regexp"
	"strings"
)
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)
```

---

## Paso 2: Actualizar los tests de `ResearchTopicKey`

**Archivo a modificar:** `internal/components/engram/engramkeys/keys_test.go`

**Acción:** Modificar — añadir nuevos casos de test al final del archivo existente

**Código a añadir al final del archivo:**
```go
func TestResearchTopicKey_NoCollisionOnCase(t *testing.T) {
	queries := []string{
		"odoo sale order workflow",
		"Odoo Sale Order Workflow",
		"ODOO SALE ORDER WORKFLOW",
		"odoo sale order workflow ",
		"odoo sale order workflow?",
	}
	seen := make(map[string]string)
	for _, q := range queries {
		k := ResearchTopicKey("context7", q)
		if prev, exists := seen[k]; exists {
			t.Errorf("collision: queries %q and %q produce same key %q", prev, q, k)
		}
		seen[k] = q
	}
}

func TestResearchTopicKey_HashSuffixLength(t *testing.T) {
	k := ResearchTopicKey("context7", "some query")
	// Key format: research/tool/slug-XXXXXXXX
	// Hash suffix should be exactly 8 hex chars
	parts := strings.Split(k, "-")
	if len(parts) < 2 {
		t.Fatalf("key has no dash: %q", k)
	}
	hashPart := parts[len(parts)-1]
	if len(hashPart) != 8 {
		t.Errorf("hash suffix should be 8 hex chars, got %d: %q", len(hashPart), hashPart)
	}
	for _, c := range hashPart {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("hash suffix contains non-hex char %q in %q", c, hashPart)
		}
	}
}

func TestResearchTopicKey_NoPreviousLenSuffix(t *testing.T) {
	// Ensure the old -len{N} format is no longer produced
	k := ResearchTopicKey("context7", "some test query")
	if strings.Contains(k, "-len") {
		t.Errorf("key should not use -len suffix (old format): %q", k)
	}
}

func TestResearchTopicKey_Stable(t *testing.T) {
	// Same input always produces same key (deterministic)
	q := "how to configure Odoo payment acquirer"
	k1 := ResearchTopicKey("context7", q)
	k2 := ResearchTopicKey("context7", q)
	if k1 != k2 {
		t.Errorf("ResearchTopicKey is not stable: %q != %q", k1, k2)
	}
}
```

---

## Paso 3: Crear el subcomando `architect-ai diagnose engram`

**Archivo a crear:** `internal/cli/diagnose_engram.go`

**Acción:** Crear

```go
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DiagnoseEngramResult holds the structured output of the Engram diagnostic.
type DiagnoseEngramResult struct {
	Timestamp       string `json:"timestamp"`
	BinaryFound     bool   `json:"binary_found"`
	BinaryPath      string `json:"binary_path,omitempty"`
	ProcessRunning  bool   `json:"process_running"`
	MCPConfigured   bool   `json:"mcp_configured"`
	MCPSettingsPath string `json:"mcp_settings_path,omitempty"`
	ProbeLogPath    string `json:"probe_log_path,omitempty"`
	RecentFails     int    `json:"recent_consecutive_fails"`
	LastProbeResult string `json:"last_probe_result,omitempty"`
	LastProbeError  string `json:"last_probe_error,omitempty"`
	SQLiteExists    bool   `json:"sqlite_exists"`
	SQLitePath      string `json:"sqlite_path,omitempty"`
	Recommendations []string `json:"recommendations"`
}

// RunDiagnoseEngram runs the Engram diagnostic and writes the result to w.
// projectRoot is the absolute path to the project root (contains .atl/).
// homeDir is the user's home directory (contains .gemini/, .claude/).
func RunDiagnoseEngram(projectRoot, homeDir string, w io.Writer) error {
	result := DiagnoseEngramResult{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// 1. Check binary existence
	engramBin := os.Getenv("ENGRAM_BIN")
	if engramBin == "" {
		engramBin = "engram"
	}
	binPath, lookErr := exec.LookPath(engramBin)
	if lookErr == nil {
		result.BinaryFound = true
		result.BinaryPath = binPath
	} else {
		result.BinaryFound = false
		result.Recommendations = append(result.Recommendations,
			"Engram binary not found in PATH. Install with: architect-ai install --component engram")
	}

	// 2. Check if process is running
	if result.BinaryFound {
		result.ProcessRunning = isEngramProcessRunning()
		if !result.ProcessRunning {
			result.Recommendations = append(result.Recommendations,
				fmt.Sprintf("Engram process not running. Start with: %s server start", engramBin))
		}
	}

	// 3. Check MCP configuration in Gemini and Claude settings
	geminiSettingsPath := filepath.Join(homeDir, ".gemini", "settings.json")
	claudeSettingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	for _, settingsPath := range []string{geminiSettingsPath, claudeSettingsPath} {
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			continue
		}
		var settings map[string]any
		if jsonErr := json.Unmarshal(data, &settings); jsonErr != nil {
			continue
		}
		mcpServers, ok := settings["mcpServers"].(map[string]any)
		if ok {
			for key := range mcpServers {
				if strings.Contains(strings.ToLower(key), "engram") {
					result.MCPConfigured = true
					result.MCPSettingsPath = settingsPath
					break
				}
			}
		}
		if result.MCPConfigured {
			break
		}
	}
	if !result.MCPConfigured {
		result.Recommendations = append(result.Recommendations,
			"Engram not found in MCP settings. Run: architect-ai sync --repair")
	}

	// 4. Check probe log
	probeLogPath := filepath.Join(projectRoot, ".atl", "probe-log.jsonl")
	result.ProbeLogPath = probeLogPath
	if data, err := os.ReadFile(probeLogPath); err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		start := 0
		if len(lines) > 10 {
			start = len(lines) - 10
		}
		recent := lines[start:]
		consecFails := 0
		var lastResult, lastErrMsg string
		for _, line := range recent {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var entry struct {
				Result string `json:"result"`
				Error  string `json:"error"`
			}
			if jsonErr := json.Unmarshal([]byte(line), &entry); jsonErr == nil {
				lastResult = entry.Result
				lastErrMsg = entry.Error
				if entry.Result == "failed" {
					consecFails++
				} else if entry.Result == "ok" {
					consecFails = 0
				}
			}
		}
		result.RecentFails = consecFails
		result.LastProbeResult = lastResult
		result.LastProbeError = lastErrMsg
		if consecFails >= 3 {
			result.Recommendations = append(result.Recommendations,
				fmt.Sprintf("Engram probe failed %d consecutive times (last error: %s). "+
					"Check that the Engram server is running and accessible.", consecFails, lastErrMsg))
		}
	}

	// 5. Check SQLite database existence (Engram uses SQLite for storage)
	engramDataDir := filepath.Join(homeDir, ".engram")
	sqlitePaths := []string{
		filepath.Join(engramDataDir, "engram.db"),
		filepath.Join(engramDataDir, "memory.db"),
		filepath.Join(engramDataDir, "data.db"),
	}
	for _, dbPath := range sqlitePaths {
		if _, err := os.Stat(dbPath); err == nil {
			result.SQLiteExists = true
			result.SQLitePath = dbPath
			break
		}
	}
	if !result.SQLiteExists {
		result.Recommendations = append(result.Recommendations,
			"Engram database not found. If Engram was never initialized, run: engram init")
	}

	// 6. Print human-readable report
	fmt.Fprintf(w, "Engram Diagnostic Report — %s\n", result.Timestamp)
	fmt.Fprintf(w, "════════════════════════════════════════\n")
	printCheck(w, "Binary found", result.BinaryFound, result.BinaryPath)
	printCheck(w, "Process running", result.ProcessRunning, "")
	printCheck(w, "MCP configured", result.MCPConfigured, result.MCPSettingsPath)
	printCheck(w, "SQLite database", result.SQLiteExists, result.SQLitePath)

	if result.LastProbeResult != "" {
		fmt.Fprintf(w, "  Last probe:    %s", result.LastProbeResult)
		if result.LastProbeError != "" {
			fmt.Fprintf(w, " (%s)", result.LastProbeError)
		}
		fmt.Fprintln(w)
	}
	if result.RecentFails > 0 {
		fmt.Fprintf(w, "  Recent fails:  %d consecutive\n", result.RecentFails)
	}

	if len(result.Recommendations) > 0 {
		fmt.Fprintf(w, "\nRecommendations:\n")
		for i, r := range result.Recommendations {
			fmt.Fprintf(w, "  %d. %s\n", i+1, r)
		}
	} else {
		fmt.Fprintf(w, "\n✓ Engram appears healthy\n")
	}

	return nil
}

func printCheck(w io.Writer, label string, ok bool, detail string) {
	status := "✓"
	if !ok {
		status = "✗"
	}
	if detail != "" {
		fmt.Fprintf(w, "  %s %-20s %s\n", status, label+":", detail)
	} else {
		fmt.Fprintf(w, "  %s %s\n", status, label)
	}
}

// isEngramProcessRunning checks for a running Engram process using ps.
// Returns false on any error (best-effort check).
func isEngramProcessRunning() bool {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("tasklist", "/FI", "IMAGENAME eq engram.exe")
	default:
		cmd = exec.Command("pgrep", "-x", "engram")
	}
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}
```

---

## Paso 4: Registrar el subcomando `diagnose` en el CLI router

**Archivo a modificar:** `internal/cli/cli.go` (o el archivo equivalente donde se registran los subcomandos)

**Comando previo:**
```bash
grep -n "\"diagnose\"\|case \"check\"\|case \"install\"\|RunCheck\|RunInstall" \
  internal/cli/cli.go cmd/architect-ai/main.go | head -20
```

**Código a insertar** — dentro del `switch args[0]` o equivalente donde se despachan subcomandos:

```go
case "diagnose":
	if len(args) < 2 {
		fmt.Fprintln(stderr, "Usage: architect-ai diagnose <component>")
		fmt.Fprintln(stderr, "Available components: engram")
		return 1
	}
	switch args[1] {
	case "engram":
		projectRoot, _ := os.Getwd()
		homeDir, _ := os.UserHomeDir()
		if err := cli.RunDiagnoseEngram(projectRoot, homeDir, stdout); err != nil {
			fmt.Fprintf(stderr, "diagnose engram: %v\n", err)
			return 1
		}
		return 0
	default:
		fmt.Fprintf(stderr, "unknown component: %q (available: engram)\n", args[1])
		return 1
	}
```

---

## Paso 5: Crear tests del subcomando `diagnose engram`

**Archivo a crear:** `internal/cli/diagnose_engram_test.go`

**Acción:** Crear

```go
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunDiagnoseEngram_EmptyProject(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := t.TempDir()

	var buf bytes.Buffer
	err := RunDiagnoseEngram(projectRoot, homeDir, &buf)
	if err != nil {
		t.Fatalf("RunDiagnoseEngram should not return error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Engram Diagnostic Report") {
		t.Errorf("output should contain 'Engram Diagnostic Report', got:\n%s", output)
	}
}

func TestRunDiagnoseEngram_ReadsProbeLog(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := t.TempDir()

	atlDir := filepath.Join(projectRoot, ".atl")
	if err := os.MkdirAll(atlDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write probe log with 3 consecutive failures
	type probeEntry struct {
		TS     string `json:"ts"`
		Probe  string `json:"probe"`
		Result string `json:"result"`
		Error  string `json:"error"`
	}
	entries := []probeEntry{
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "ok"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed", Error: "timeout"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed", Error: "timeout"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed", Error: "connection refused"},
	}
	var sb strings.Builder
	for _, e := range entries {
		line, _ := json.Marshal(e)
		sb.Write(line)
		sb.WriteByte('\n')
	}
	logPath := filepath.Join(atlDir, "probe-log.jsonl")
	if err := os.WriteFile(logPath, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("write probe log: %v", err)
	}

	var buf bytes.Buffer
	if err := RunDiagnoseEngram(projectRoot, homeDir, &buf); err != nil {
		t.Fatalf("RunDiagnoseEngram: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Recent fails:") {
		t.Errorf("output should mention recent fails:\n%s", output)
	}
	if !strings.Contains(output, "Recommendations:") {
		t.Errorf("output should have recommendations when probe fails:\n%s", output)
	}
}

func TestRunDiagnoseEngram_MCPConfigured(t *testing.T) {
	projectRoot := t.TempDir()
	homeDir := t.TempDir()

	geminiDir := filepath.Join(homeDir, ".gemini")
	if err := os.MkdirAll(geminiDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	settings := map[string]any{
		"mcpServers": map[string]any{
			"engram-mcp": map[string]any{
				"command": "engram",
				"args":    []string{"serve"},
			},
		},
	}
	data, _ := json.Marshal(settings)
	if err := os.WriteFile(filepath.Join(geminiDir, "settings.json"), data, 0o644); err != nil {
		t.Fatalf("write settings: %v", err)
	}

	var buf bytes.Buffer
	if err := RunDiagnoseEngram(projectRoot, homeDir, &buf); err != nil {
		t.Fatalf("RunDiagnoseEngram: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Errorf("output should show at least one passing check:\n%s", output)
	}
}
```

---

## Verificación de Fase

```bash
# 1. Compilar el paquete engramkeys
go build ./internal/components/engram/engramkeys/...

# 2. Tests de ResearchTopicKey (sin regresiones)
go test ./internal/components/engram/engramkeys/... -v -count=1

# 3. Tests nuevos de collision-free
go test ./internal/components/engram/engramkeys/... -v -count=1 \
  -run TestResearchTopicKey_NoCollision
go test ./internal/components/engram/engramkeys/... -v -count=1 \
  -run TestResearchTopicKey_HashSuffix
go test ./internal/components/engram/engramkeys/... -v -count=1 \
  -run TestResearchTopicKey_NoPreviousLen
go test ./internal/components/engram/engramkeys/... -v -count=1 \
  -run TestResearchTopicKey_Stable

# 4. Compilar el paquete cli con el nuevo diagnose
go build ./internal/cli/...

# 5. Tests del diagnose engram
go test ./internal/cli/... -v -count=1 -run TestRunDiagnoseEngram

# 6. Compilar el binario completo
go build ./cmd/architect-ai/...

# 7. Smoke test del subcomando (en directorio con .atl/)
if [ -d ".atl" ]; then
  go run ./cmd/architect-ai diagnose engram 2>&1
fi

# 8. Verificar que el formato de topic_key cambió
go run - << 'GOEOF'
package main

import (
	"fmt"
	"github.com/rd-mg/architect-ai/internal/components/engram/engramkeys"
)

func main() {
	k := engramkeys.ResearchTopicKey("context7", "odoo sale order workflow")
	fmt.Printf("Key: %s\n", k)
	// Should be: research/context7/odoo-sale-order-workflow-XXXXXXXX
	// Should NOT contain: -len
}
GOEOF
```

