package odoo

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func setupTestDirs(t *testing.T) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "odoo-indexer-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create structure:
	// skills/odoo-18.0/reference/guide1.md
	// skills/patterns-18/pattern1.md
	// skills/patterns-agnostic/agnostic1.md
	// skills/patterns-ddd/ddd1.md
	// skills/migration-17-18/mig1.md

	paths := []string{
		"skills/odoo-18.0/guide1.md",
		"skills/patterns-18/pattern1.md",
		"skills/patterns-agnostic/agnostic1.md",
		"skills/patterns-ddd/ddd1.md",
		"skills/migration-17-18/mig1.md",
	}

	for _, p := range paths {
		fullPath := filepath.Join(tempDir, p)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("content of "+filepath.Base(p)), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestIndexAll_Idempotent(t *testing.T) {
	tempDir, cleanup := setupTestDirs(t)
	defer cleanup()

	mock := &mockEngram{}
	indexer := &OdooIndexer{
		OverlayDir:  tempDir,
		OdooVersion: "18",
		Engram:      mock,
		Workers:     2,
	}

	results, err := indexer.IndexAll("test-proj")
	if err != nil {
		t.Fatalf("IndexAll failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}

	// Check if they are successfully updated
	for _, res := range results {
		if res.Action != "updated" {
			t.Errorf("expected action 'updated' for key %s, got %s", res.TopicKey, res.Action)
		}
		if res.Error != nil {
			t.Errorf("expected no error for key %s, got %v", res.TopicKey, res.Error)
		}
	}

	// Run second time to ensure idempotency and no duplicates
	results2, err := indexer.IndexAll("test-proj")
	if err != nil {
		t.Fatalf("second IndexAll failed: %v", err)
	}

	if len(results2) != 5 {
		t.Errorf("expected 5 results on second run, got %d", len(results2))
	}
}

func TestIndexAll_PartialFailure(t *testing.T) {
	tempDir, cleanup := setupTestDirs(t)
	defer cleanup()

	mock := &mockEngram{
		failKey: "knowledge/odoo-18/patterns/pattern1",
	}
	indexer := &OdooIndexer{
		OverlayDir:  tempDir,
		OdooVersion: "18",
		Engram:      mock,
		Workers:     2,
	}

	results, err := indexer.IndexAll("test-proj")
	if err != nil {
		t.Fatalf("IndexAll failed: %v", err)
	}

	var failedCount, successCount int
	for _, res := range results {
		if res.Action == "failed" {
			failedCount++
			if res.TopicKey != mock.failKey {
				t.Errorf("unexpected failed key: %s", res.TopicKey)
			}
		} else if res.Action == "updated" {
			successCount++
		}
	}

	if failedCount != 1 {
		t.Errorf("expected exactly 1 failure, got %d", failedCount)
	}
	if successCount != 4 {
		t.Errorf("expected exactly 4 successes, got %d", successCount)
	}
}

func TestIndexAll_TopicKeyFormat(t *testing.T) {
	tempDir, cleanup := setupTestDirs(t)
	defer cleanup()

	mock := &mockEngram{}
	indexer := &OdooIndexer{
		OverlayDir:  tempDir,
		OdooVersion: "18",
		Engram:      mock,
		Workers:     2,
	}

	results, err := indexer.IndexAll("test-proj")
	if err != nil {
		t.Fatalf("IndexAll failed: %v", err)
	}

	expectedPrefixes := map[string]string{
		"skills/odoo-18.0/guide1.md":      "knowledge/odoo-18/reference/guide1",
		"skills/patterns-18/pattern1.md":    "knowledge/odoo-18/patterns/pattern1",
		"skills/patterns-agnostic/agnostic1.md": "knowledge/odoo-agnostic/reference/agnostic1",
		"skills/patterns-ddd/ddd1.md":       "knowledge/odoo-agnostic/ddd/ddd1",
		"skills/migration-17-18/mig1.md":   "knowledge/odoo-migration/17-18/mig1",
	}

	for _, res := range results {
		found := false
		for _, expected := range expectedPrefixes {
			if res.TopicKey == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("topic key %s does not match any expected format", res.TopicKey)
		}
		if strings.Contains(res.TopicKey, ".md") {
			t.Errorf("topic key %s should not contain file extension", res.TopicKey)
		}
	}
}
