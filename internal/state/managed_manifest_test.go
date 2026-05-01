package state

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestManagedManifest_SaveLoad(t *testing.T) {
	dir, err := ioutil.TempDir("", "manifest_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "manifest.json")
	m := NewManagedManifest("test-agent", "/test/root")

	content := "test content"
	hash := sha256.Sum256([]byte(content))
	hashStr := hex.EncodeToString(hash[:])

	m.AddEntry(ManagedEntry{
		Component:      "core",
		Path:           "/test/file.txt",
		Kind:           KindFile,
		SHA256AtWrite:  hashStr,
		RemoveStrategy: DeleteIfUnchanged,
	})

	if err := m.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	m2, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	if m2.Agent != "test-agent" {
		t.Errorf("Agent mismatch: got %v, want %v", m2.Agent, "test-agent")
	}

	if len(m2.Entries) != 1 {
		t.Fatalf("Entries length mismatch: got %v, want 1", len(m2.Entries))
	}

	entry := m2.Entries[0]
	if entry.Path != "/test/file.txt" || entry.Kind != KindFile || entry.SHA256AtWrite != hashStr || entry.RemoveStrategy != DeleteIfUnchanged {
		t.Errorf("Entry mismatch: %+v", entry)
	}
}

func TestManagedManifest_AddEntryUpdatesExisting(t *testing.T) {
	m := NewManagedManifest("test-agent", "/test/root")

	RecordManagedFile(m, "core", "/test/file.txt", "old content", DeleteIfUnchanged)
	if len(m.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(m.Entries))
	}

	RecordManagedFile(m, "core", "/test/file.txt", "new content", DeleteIfUnchanged)
	if len(m.Entries) != 1 {
		t.Fatalf("Expected 1 entry after update, got %d", len(m.Entries))
	}

	hash := sha256.Sum256([]byte("new content"))
	expectedHash := hex.EncodeToString(hash[:])

	if m.Entries[0].SHA256AtWrite != expectedHash {
		t.Errorf("Hash not updated: got %v, want %v", m.Entries[0].SHA256AtWrite, expectedHash)
	}
}

func TestManagedManifest_RecordHelpers(t *testing.T) {
	m := NewManagedManifest("test", "")

	RecordManagedJSONPath(m, "mcp", "/settings.json", "mcpServers.test")
	if len(m.Entries) != 1 {
		t.Fatalf("Expected 1 entry")
	}
	if m.Entries[0].Kind != KindJSONPath || m.Entries[0].JSONPath != "mcpServers.test" {
		t.Errorf("Invalid json path entry: %+v", m.Entries[0])
	}

	RecordManagedSection(m, "persona", "/README.md", "architect-ai:persona")
	if len(m.Entries) != 2 {
		t.Fatalf("Expected 2 entries")
	}
	if m.Entries[1].Kind != KindMarkdownSection || m.Entries[1].Marker != "architect-ai:persona" {
		t.Errorf("Invalid markdown section entry: %+v", m.Entries[1])
	}
}
