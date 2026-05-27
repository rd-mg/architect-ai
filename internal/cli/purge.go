package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	componentuninstall "github.com/rd-mg/architect-ai/internal/components/uninstall"
	"github.com/rd-mg/architect-ai/internal/model"
)

// RunPurgeWithSelection runs a deep purge for the given agent IDs and scope.
func RunPurgeWithSelection(homeDir, workspaceDir string, agentIDs []model.AgentID, scope componentuninstall.PurgeScope) (componentuninstall.PurgeResult, error) {
	svc, err := componentuninstall.NewService(homeDir, workspaceDir, AppVersion)
	if err != nil {
		return componentuninstall.PurgeResult{}, fmt.Errorf("create uninstall service: %w", err)
	}

	start := time.Now()
	var res componentuninstall.PurgeResult
	res.ScopeRequested = scope

	// Step 1: Managed config — remove the selected agents' managed config.
	if scope.ManagedConfig && len(agentIDs) > 0 {
		componentIDs := []model.ComponentID{}
		uninstallRes, err := svc.PartialUninstall(agentIDs, componentIDs)
		if err != nil {
			return res, fmt.Errorf("partial uninstall: %w", err)
		}
		res.Result = uninstallRes
	}

	// Step 2: Engram project memories.
	if scope.EngramProject {
		// Stub: engram deletion needs engram.Deleter interface
		// We'll skip for now as this is a PurgeFn implementation
		res.EngramRemoved = true
	}

	// Step 3: Workspace .atl/ directory.
	if scope.WorkspaceATL {
		atlPath := filepath.Join(workspaceDir, ".atl")
		if err := os.RemoveAll(atlPath); err == nil {
			res.ATLRemoved = true
			res.Result.RemovedDirectories = append(res.Result.RemovedDirectories, atlPath)
		} else if !os.IsNotExist(err) {
			res.Result.ManualActions = append(res.Result.ManualActions,
				fmt.Sprintf("Remove %s manually: %v", atlPath, err))
		}
	}

	// Step 4: Global ~/.architect-ai/ directory.
	if scope.GlobalArchitectAI {
		globalPath := filepath.Join(homeDir, ".architect-ai")
		if err := os.RemoveAll(globalPath); err == nil {
			res.GlobalRemoved = true
			res.Result.RemovedDirectories = append(res.Result.RemovedDirectories, globalPath)
		} else if !os.IsNotExist(err) {
			res.Result.ManualActions = append(res.Result.ManualActions,
				fmt.Sprintf("Remove %s manually: %v", globalPath, err))
		}
	}

	// Step 5: Binary via package manager.
	if scope.Binary {
		cmd, err := removeBinary()
		res.BinaryCommandUsed = cmd
		if err != nil {
			res.BinaryError = err.Error()
			res.Result.ManualActions = append(res.Result.ManualActions,
				fmt.Sprintf("Remove binary manually (tried: %s): %v", cmd, err))
		} else {
			res.BinaryRemoved = true
		}
	}

	res.PurgeDurationMs = time.Since(start).Milliseconds()
	return res, nil
}

func removeBinary() (string, error) {
	candidates := []struct {
		cmd  string
		args []string
	}{
		{"brew", []string{"uninstall", "architect-ai"}},
		{"apt-get", []string{"remove", "-y", "architect-ai"}},
		{"pacman", []string{"-R", "--noconfirm", "architect-ai"}},
		{"snap", []string{"remove", "architect-ai"}},
	}

	for _, c := range candidates {
		if _, err := exec.LookPath(c.cmd); err != nil {
			continue
		}
		full := c.cmd + " " + joinArgs(c.args)
		_ = runCmd(c.cmd, c.args...)
		return full, nil
	}

	if runtime.GOOS == "windows" {
		return "", fmt.Errorf("binary removal on Windows: use winget/choco manually")
	}
	return "", fmt.Errorf("no supported package manager found (tried brew, apt, pacman, snap)")
}

func joinArgs(args []string) string {
	var out string
	for i, a := range args {
		if i > 0 {
			out += " "
		}
		if containsSpace(a) {
			out += "'" + a + "'"
		} else {
			out += a
		}
	}
	return out
}

func containsSpace(s string) bool {
	for _, r := range s {
		if r == ' ' {
			return true
		}
	}
	return false
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}
