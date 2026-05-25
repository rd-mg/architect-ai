package orchestrator

import (
	"log"

	"github.com/rd-mg/architect-ai/internal/reasoning/gate"
)

// ModeStore defines the interface for persisting scored cognitive mode dimensions.
// The Go runtime implementation uses Engram (mem_save); this interface allows
// testing with a mock and graceful degradation when Engram is unavailable.
type ModeStore interface {
	// SaveScoredMode persists the scored cognitive dimensions for a phase.
	// Returns an error if persistence fails; callers MUST NOT block dispatch on failure.
	SaveScoredMode(changeID, phase string, dims Dims48, mode int, postures []string) error
}

// noopModeStore is a no-op implementation that logs and discards.
type noopModeStore struct{}

func (n *noopModeStore) SaveScoredMode(changeID, phase string, dims Dims48, mode int, postures []string) error {
	log.Printf("[CognitiveMode] (no-op) would persist: change=%s phase=%s dims=%v mode=%d postures=%v",
		changeID, phase, dims, mode, postures)
	return nil
}

// NewNoopModeStore returns a no-op ModeStore that logs and discards.
// Used when Engram is unavailable or during development.
func NewNoopModeStore() ModeStore {
	return &noopModeStore{}
}

// PersistCognitiveDims persists scored cognitive dimensions to the provided ModeStore.
// It computes postures from dims via gate.Score and delegates to the store.
//
// Topic key format: reasoning/{changeID}/phase/{phase}/dims
//
// This function MUST NOT block dispatch on persistence failure — errors are logged and
// the function returns nil.
func PersistCognitiveDims(store ModeStore, changeID, phase string, dims Dims48, expectedMode, declaredMode int) error {
	if store == nil {
		log.Println("[CognitiveMode] no ModeStore provided, skipping persistence")
		return nil
	}

	// Compute postures from dims for completeness
	_, postures := gate.Score(dims)

	err := store.SaveScoredMode(changeID, phase, dims, expectedMode, postures)
	if err != nil {
		log.Printf("[CognitiveMode] failed to persist dimensions (non-blocking): %v", err)
		return nil // non-blocking — log and continue
	}
	_ = declaredMode // available for future enhancement
	return nil
}
