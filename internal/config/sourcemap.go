// Package config provides lookup maps from source identifiers to asset
// paths and templates used throughout the project.
package config

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rd-mg/architect-ai/internal/assets"
)

// sourceMap holds the parsed content of source-map.json.
type sourceMap struct {
	Version      int                    `json:"version"`
	Agents       map[string]agentEntry `json:"agents"`
	Templates    templateEntry         `json:"templates"`
	Placeholders map[string]string     `json:"placeholders,omitempty"`
	HashSources  map[string]string     `json:"hash_sources,omitempty"`
	IncludesBase string                `json:"includes_base,omitempty"`
}

type agentEntry struct {
	Dir   string   `json:"dir"`
	Files []string `json:"files"`
}

type templateEntry struct {
	Dir string `json:"dir"`
}

var (
	sm   sourceMap
	once sync.Once
)

// load reads and parses source-map.json from embedded assets.
func load() error {
	var loadErr error
	once.Do(func() {
		data, err := assets.Read("source-map.json")
		if err != nil {
			loadErr = fmt.Errorf("read source-map.json: %w", err)
			return
		}
		if err := json.Unmarshal([]byte(data), &sm); err != nil {
			loadErr = fmt.Errorf("parse source-map.json: %w", err)
			return
		}
	})
	return loadErr
}

// AgentDir returns the asset subdirectory for the named agent, or an empty
// string if the agent is unknown.
func AgentDir(name string) string {
	if err := load(); err != nil {
		return ""
	}
	entry, ok := sm.Agents[name]
	if !ok {
		return ""
	}
	return entry.Dir
}

// AgentFiles returns the list of asset file names for the named agent, or nil.
func AgentFiles(name string) []string {
	if err := load(); err != nil {
		return nil
	}
	entry, ok := sm.Agents[name]
	if !ok {
		return nil
	}
	out := make([]string, len(entry.Files))
	copy(out, entry.Files)
	return out
}

// KnownAgents returns the names of all agents registered in the source map.
func KnownAgents() []string {
	if err := load(); err != nil {
		return nil
	}
	names := make([]string, 0, len(sm.Agents))
	for n := range sm.Agents {
		names = append(names, n)
	}
	return names
}

// TemplateDir returns the templates directory name from the source map.
func TemplateDir() string {
	if err := load(); err != nil {
		return ""
	}
	return sm.Templates.Dir
}

// SourceMapVersion returns the version of the loaded source map.
func SourceMapVersion() int {
	if err := load(); err != nil {
		return 0
	}
	return sm.Version
}

// PlaceholderSource returns the source file path for a given placeholder token,
// or empty string if not found.
func PlaceholderSource(token string) string {
	if err := load(); err != nil {
		return ""
	}
	return sm.Placeholders[token]
}

// HashSource returns the source file path for a given hash token (e.g. "L0_HASH"),
// or empty string if not found.
func HashSource(token string) string {
	if err := load(); err != nil {
		return ""
	}
	return sm.HashSources[token]
}

// IncludesBase returns the base directory for {{ include }} resolution.
func IncludesBase() string {
	if err := load(); err != nil {
		return ""
	}
	return sm.IncludesBase
}

// AllPlaceholders returns a copy of the placeholder map.
func AllPlaceholders() map[string]string {
	if err := load(); err != nil {
		return nil
	}
	out := make(map[string]string, len(sm.Placeholders))
	for k, v := range sm.Placeholders {
		out[k] = v
	}
	return out
}

// AllHashSources returns a copy of the hash sources map.
func AllHashSources() map[string]string {
	if err := load(); err != nil {
		return nil
	}
	out := make(map[string]string, len(sm.HashSources))
	for k, v := range sm.HashSources {
		out[k] = v
	}
	return out
}
