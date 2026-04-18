package vscode

import (
	"testing"
)

func TestSessionHookEnabled(t *testing.T) {
	if SessionHookEnabled() {
		t.Error("VSCode should report SessionHookEnabled() == false")
	}
}

func TestRecordResponse_Safety(t *testing.T) {
	// Should not panic
	RecordResponse(nil)
	RecordResponse([]byte("garbage"))
}
