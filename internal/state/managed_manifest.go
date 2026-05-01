package state

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rd-mg/architect-ai/internal/model"
)

type RemoveStrategy string

const (
	DeleteIfUnchanged   RemoveStrategy = "delete_if_unchanged"
	DeleteAlwaysOwned   RemoveStrategy = "delete_always_owned"
	RemoveJSONPath      RemoveStrategy = "remove_json_path"
	RemoveMarkedSection RemoveStrategy = "remove_marked_section"
	ManualReview        RemoveStrategy = "manual_review"
)

type Kind string

const (
	KindFile            Kind = "file"
	KindJSONPath        Kind = "json_path"
	KindMarkdownSection Kind = "markdown_section"
	KindDirectory       Kind = "directory"
)

type ManagedEntry struct {
	Component      model.ComponentID `json:"component"`
	Path           string            `json:"path"`
	Kind           Kind              `json:"kind"`
	JSONPath       string            `json:"json_path,omitempty"`
	Marker         string            `json:"marker,omitempty"`
	SHA256AtWrite  string            `json:"sha256_at_write,omitempty"`
	RemoveStrategy RemoveStrategy    `json:"remove_strategy"`
}

type ManagedManifest struct {
	SchemaVersion int               `json:"schema_version"`
	Agent         model.AgentID     `json:"agent"`
	ProjectRoot   string            `json:"project_root"`
	CreatedAt     time.Time         `json:"created_at"`
	Entries       []ManagedEntry    `json:"entries"`
	
	mu sync.RWMutex
}

func NewManagedManifest(agent model.AgentID, projectRoot string) *ManagedManifest {
	return &ManagedManifest{
		SchemaVersion: 1,
		Agent:         agent,
		ProjectRoot:   projectRoot,
		CreatedAt:     time.Now().UTC(),
		Entries:       make([]ManagedEntry, 0),
	}
}

func LoadManifest(path string) (*ManagedManifest, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest ManagedManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

func (m *ManagedManifest) Save(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	tempFile := path + ".tmp"
	if err := ioutil.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp manifest: %w", err)
	}

	if err := os.Rename(tempFile, path); err != nil {
		os.Remove(tempFile) // Best effort cleanup
		return fmt.Errorf("failed to commit manifest: %w", err)
	}

	return nil
}

func (m *ManagedManifest) AddEntry(entry ManagedEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update if exists, otherwise append
	for i, e := range m.Entries {
		if e.Path == entry.Path && e.Kind == entry.Kind && e.Component == entry.Component {
			// For JSONPath and MarkdownSection, also check specific fields
			if e.Kind == KindJSONPath && e.JSONPath != entry.JSONPath {
				continue
			}
			if e.Kind == KindMarkdownSection && e.Marker != entry.Marker {
				continue
			}
			m.Entries[i] = entry
			return
		}
	}
	m.Entries = append(m.Entries, entry)
}

// Helpers

func RecordManagedFile(m *ManagedManifest, component model.ComponentID, path, content string, strategy RemoveStrategy) {
	hash := sha256.Sum256([]byte(content))
	m.AddEntry(ManagedEntry{
		Component:      component,
		Path:           path,
		Kind:           KindFile,
		SHA256AtWrite:  hex.EncodeToString(hash[:]),
		RemoveStrategy: strategy,
	})
}

func RecordManagedJSONPath(m *ManagedManifest, component model.ComponentID, path, jsonPath string) {
	m.AddEntry(ManagedEntry{
		Component:      component,
		Path:           path,
		Kind:           KindJSONPath,
		JSONPath:       jsonPath,
		RemoveStrategy: RemoveJSONPath,
	})
}

func RecordManagedSection(m *ManagedManifest, component model.ComponentID, path, marker string) {
	m.AddEntry(ManagedEntry{
		Component:      component,
		Path:           path,
		Kind:           KindMarkdownSection,
		Marker:         marker,
		RemoveStrategy: RemoveMarkedSection,
	})
}
func LoadOrNewManifest(homeDir, projectRoot string, agent model.AgentID) (*ManagedManifest, string) {
	path := AgentManifestPath(homeDir, projectRoot, agent)
	m, err := LoadManifest(path)
	if err != nil {
		return NewManagedManifest(agent, projectRoot), path
	}
	return m, path
}
