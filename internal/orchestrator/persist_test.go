package orchestrator

import (
	"testing"
)

// mockModeStore implements ModeStore for testing.
type mockModeStore struct {
	savedChangeID string
	savedPhase    string
	savedDims     Dims48
	savedMode     int
	savedPostures []string
	err           error // if non-nil, SaveScoredMode returns this error
}

func (m *mockModeStore) SaveScoredMode(changeID, phase string, dims Dims48, mode int, postures []string) error {
	m.savedChangeID = changeID
	m.savedPhase = phase
	m.savedDims = dims
	m.savedMode = mode
	m.savedPostures = postures
	return m.err
}

func TestPersistCognitiveDims_Success(t *testing.T) {
	mock := &mockModeStore{}
	dims := Dims48{1, 0, 0, 0}

	err := PersistCognitiveDims(mock, "test-change", "sdd-apply", dims, 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.savedChangeID != "test-change" {
		t.Errorf("expected changeID=test-change, got %s", mock.savedChangeID)
	}
	if mock.savedPhase != "sdd-apply" {
		t.Errorf("expected phase=sdd-apply, got %s", mock.savedPhase)
	}
	if mock.savedDims != dims {
		t.Errorf("expected dims=%v, got %v", dims, mock.savedDims)
	}
	if mock.savedMode != 1 {
		t.Errorf("expected mode=1, got %d", mock.savedMode)
	}
}

func TestPersistCognitiveDims_NilStore(t *testing.T) {
	err := PersistCognitiveDims(nil, "test-change", "sdd-apply", Dims48{}, 1, 1)
	if err != nil {
		t.Errorf("expected nil error for nil store, got %v", err)
	}
}

func TestPersistCognitiveDims_StoreFails(t *testing.T) {
	mock := &mockModeStore{err: ErrStoreUnavailable}
	dims := Dims48{2, 0, 2, 0}

	err := PersistCognitiveDims(mock, "test-change", "sdd-verify", dims, 3, 3)
	// Must not propagate error — non-blocking contract
	if err != nil {
		t.Errorf("expected nil error even on store failure (non-blocking), got %v", err)
	}
	if mock.savedChangeID != "test-change" {
		t.Errorf("expected save to be attempted")
	}
}

func TestPersistCognitiveDims_ZeroDims(t *testing.T) {
	mock := &mockModeStore{}
	dims := Dims48{0, 0, 0, 0}

	err := PersistCognitiveDims(mock, "test-change", "sdd-apply", dims, 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.savedDims != dims {
		t.Errorf("expected zero dims=%v, got %v", dims, mock.savedDims)
	}
	if mock.savedMode != 1 {
		t.Errorf("expected mode=1 for zero dims, got %d", mock.savedMode)
	}
	if len(mock.savedPostures) == 0 || mock.savedPostures[0] != "+++Pragmatic" {
		t.Errorf("expected Pragmatic posture for zero dims, got %v", mock.savedPostures)
	}
}

func TestPersistCognitiveDims_HighPressureDims(t *testing.T) {
	mock := &mockModeStore{}
	dims := Dims48{3, 0, 2, 0}

	err := PersistCognitiveDims(mock, "pressure-test", "sdd-apply", dims, 3, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.savedDims != dims {
		t.Errorf("expected dims=%v, got %v", dims, mock.savedDims)
	}
	if mock.savedMode != 3 {
		t.Errorf("expected mode=3, got %d", mock.savedMode)
	}
}

// ErrStoreUnavailable is a sentinel error for testing store failures.
var ErrStoreUnavailable = &storeError{"engram unavailable"}

type storeError struct{ msg string }

func (e *storeError) Error() string { return e.msg }
