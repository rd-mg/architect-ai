package odoo

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// mockEngram implements EngramClient for testing
type mockEngram struct {
	mu      sync.Mutex
	saves   map[string]string
	failKey string // simulate failure for this key
}

func (m *mockEngram) SaveIdempotent(key, content, project string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if key == m.failKey {
		return errors.New("simulated Engram failure")
	}
	if m.saves == nil {
		m.saves = make(map[string]string)
	}
	m.saves[key] = content
	return nil
}

func setupTestOverlay(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create minimal skill structure
	guides := map[string]string{
		filepath.Join(dir, "skills", "odoo-18.0", "odoo-18-model-guide.md"): "# Model Guide v18\nContent here",
		filepath.Join(dir, "skills", "patterns-agnostic", "discovery-index.md"): "# Discovery Index\nOCA patterns",
		filepath.Join(dir, "skills", "patterns-ddd", "aggregate-roots.md"): "# Aggregate Roots\nDDD patterns",
	}
	for path, content := range guides {
		os.MkdirAll(filepath.Dir(path), 0755)
		os.WriteFile(path, []byte(content), 0644)
	}
	return dir
}

func TestIndexAll_BasicIndexing(t *testing.T) {
	overlay := setupTestOverlay(t)
	client := &mockEngram{}

	idx := &OdooIndexer{
		OverlayDir:  overlay,
		OdooVersion: "18",
		Engram:      client,
		Workers:     2,
	}

	results, err := idx.IndexAll("test-project")
	if err != nil {
		t.Fatalf("IndexAll: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected at least one indexed guide")
	}

	// Verify guides were saved
	if len(client.saves) == 0 {
		t.Error("expected guides to be saved to Engram")
	}
}

func TestIndexAll_Idempotent(t *testing.T) {
	overlay := setupTestOverlay(t)
	client := &mockEngram{}
	idx := &OdooIndexer{OverlayDir: overlay, OdooVersion: "18", Engram: client, Workers: 2}

	// Index twice
	results1, _ := idx.IndexAll("test-project")
	results2, _ := idx.IndexAll("test-project")

	// Both should succeed (idempotent = UPDATE not duplicate INSERT)
	failed1 := countFailed(results1)
	failed2 := countFailed(results2)

	if failed1 > 0 {
		t.Errorf("first run: %d failures", failed1)
	}
	if failed2 > 0 {
		t.Errorf("second run: %d failures (should be idempotent)", failed2)
	}
}

func TestIndexAll_PartialFailure(t *testing.T) {
	overlay := setupTestOverlay(t)
	client := &mockEngram{failKey: "knowledge/odoo-agnostic/reference/discovery-index"}
	idx := &OdooIndexer{OverlayDir: overlay, OdooVersion: "18", Engram: client, Workers: 2}

	results, err := idx.IndexAll("test-project")
	if err != nil {
		t.Fatalf("IndexAll should not return error for partial failure: %v", err)
	}

	// Should have some successes and one failure
	failed := countFailed(results)
	succeeded := countSucceeded(results)

	if failed == 0 {
		t.Error("expected at least one failure")
	}
	if succeeded == 0 {
		t.Error("expected at least one success despite partial failure")
	}
}

func TestIndexAll_TopicKeyFormat(t *testing.T) {
	overlay := setupTestOverlay(t)
	client := &mockEngram{}
	idx := &OdooIndexer{OverlayDir: overlay, OdooVersion: "18", Engram: client, Workers: 2}

	idx.IndexAll("test-project")

	for key := range client.saves {
		if !isValidTopicKey(key) {
			t.Errorf("invalid topic key format: %s", key)
		}
	}
}

func countFailed(results []IndexResult) int {
	n := 0
	for _, r := range results {
		if r.Action == "failed" {
			n++
		}
	}
	return n
}

func countSucceeded(results []IndexResult) int {
	n := 0
	for _, r := range results {
		if r.Action == "updated" || r.Action == "created" {
			n++
		}
	}
	return n
}

func isValidTopicKey(key string) bool {
	// Must be: word/word/... pattern, lowercase, no spaces
	for _, char := range key {
		if char != '/' && char != '-' && !(char >= 'a' && char <= 'z') && !(char >= '0' && char <= '9') {
			return false
		}
	}
	return len(key) > 0 && key[0] != '/'
}
