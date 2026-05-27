package cli

import (
	"os"
	"path/filepath"
	"testing"

	componentuninstall "github.com/rd-mg/architect-ai/internal/components/uninstall"
	"github.com/rd-mg/architect-ai/internal/model"
)

func TestRunPurgeWithSelection(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(workspaceDir, 0755)

	// Stub out AppVersion for testing
	AppVersion = "test-version"

	scope := componentuninstall.PurgeScope{
		ManagedConfig: false,
		EngramProject: true,
		WorkspaceATL:  true,
	}
	
	// Creating .atl directory to test removal
	atlPath := filepath.Join(workspaceDir, ".atl")
	os.MkdirAll(atlPath, 0755)

	res, err := RunPurgeWithSelection(homeDir, workspaceDir, []model.AgentID{}, scope)
	if err != nil {
		t.Fatalf("RunPurgeWithSelection failed: %v", err)
	}

	if !res.EngramRemoved {
		t.Errorf("Expected EngramRemoved to be true")
	}
	if !res.ATLRemoved {
		t.Errorf("Expected ATLRemoved to be true")
	}
	
	if _, err := os.Stat(atlPath); !os.IsNotExist(err) {
		t.Errorf("Expected .atl directory to be removed")
	}
}
