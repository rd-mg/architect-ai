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

func RunDiagnoseEngram(projectRoot, homeDir string, w io.Writer) error {
	result := DiagnoseEngramResult{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

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

	if result.BinaryFound {
		result.ProcessRunning = isEngramProcessRunning()
		if !result.ProcessRunning {
			result.Recommendations = append(result.Recommendations,
				fmt.Sprintf("Engram process not running. Start with: %s server start", engramBin))
		}
	}

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
