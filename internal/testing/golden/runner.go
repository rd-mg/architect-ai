package golden

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// GoldenDir is the directory containing golden test files
const GoldenDir = ".atl/tests/golden"

// AssertGoldenJSON compares generated JSON against a golden file
// Set UPDATE_GOLDEN=1 to regenerate golden files
func AssertGoldenJSON(t *testing.T, name string, actual interface{}) {
	t.Helper()
	actualJSON, err := json.MarshalIndent(actual, "", "  ")
	if err != nil {
		t.Fatalf("marshal actual: %v", err)
	}

	goldenPath := filepath.Join(GoldenDir, name+".golden")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		os.MkdirAll(filepath.Dir(goldenPath), 0755)
		os.WriteFile(goldenPath, actualJSON, 0644)
		t.Logf("Updated golden: %s", goldenPath)
		return
	}

	expectedJSON, err := os.ReadFile(goldenPath)
	if os.IsNotExist(err) {
		t.Fatalf("golden file missing: %s\nRun with UPDATE_GOLDEN=1 to create", goldenPath)
	}
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	if !bytes.Equal(normalizeJSON(actualJSON), normalizeJSON(expectedJSON)) {
		t.Errorf("golden mismatch for %s\nDiff:\n%s", name,
			diffStrings(string(expectedJSON), string(actualJSON)))
	}
}

// AssertGoldenMD compares generated Markdown section against a golden file
func AssertGoldenMD(t *testing.T, name string, actual string) {
	t.Helper()
	goldenPath := filepath.Join(GoldenDir, name+".golden")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		os.MkdirAll(filepath.Dir(goldenPath), 0755)
		os.WriteFile(goldenPath, []byte(actual), 0644)
		t.Logf("Updated golden: %s", goldenPath)
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if os.IsNotExist(err) {
		t.Fatalf("golden file missing: %s\nRun with UPDATE_GOLDEN=1 to create", goldenPath)
	}
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}

	if strings.TrimSpace(string(expected)) != strings.TrimSpace(actual) {
		t.Errorf("golden mismatch for %s\nExpected:\n%s\nActual:\n%s",
			name, string(expected), actual)
	}
}

func normalizeJSON(data []byte) []byte {
	var v interface{}
	json.Unmarshal(data, &v)
	normalized, _ := json.MarshalIndent(v, "", "  ")
	return normalized
}

func diffStrings(expected, actual string) string {
	expLines := strings.Split(expected, "\n")
	actLines := strings.Split(actual, "\n")
	var diff []string
	maxLen := len(expLines)
	if len(actLines) > maxLen {
		maxLen = len(actLines)
	}
	for i := 0; i < maxLen && i < 20; i++ {
		var e, a string
		if i < len(expLines) { e = expLines[i] }
		if i < len(actLines) { a = actLines[i] }
		if e != a {
			diff = append(diff, fmt.Sprintf("line %d:\n  expected: %q\n  actual:   %q", i+1, e, a))
		}
	}
	return strings.Join(diff, "\n")
}
