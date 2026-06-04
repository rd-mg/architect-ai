package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func RunInstallHooks(args []string, stdout io.Writer) error {
	gitRootCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	gitRootBytes, err := gitRootCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to find git repository root (are you in a git repo?): %w", err)
	}

	gitRoot := string(gitRootBytes)
	gitRoot = gitRoot[:len(gitRoot)-1] // remove trailing newline

	hooksDir := filepath.Join(gitRoot, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create .git/hooks directory: %w", err)
	}

	preCommitPath := filepath.Join(hooksDir, "pre-commit")

	wrapperScript := `#!/bin/sh
exec "$(git rev-parse --show-toplevel)/architect-ai" check all
`
	if err := os.WriteFile(preCommitPath, []byte(wrapperScript), 0755); err != nil {
		return fmt.Errorf("failed to write pre-commit hook: %w", err)
	}

	fmt.Fprintf(stdout, "Successfully installed pre-commit hook at %s\n", preCommitPath)
	return nil
}
