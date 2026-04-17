package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func RunCleanup(args []string, stdout io.Writer) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home directory: %w", err)
	}

	// Directorios a limpiar
	cleanupTargets := []string{
		filepath.Join(homeDir, ".architect-ai", "tmp"),
		filepath.Join(homeDir, ".architect-ai", "logs"),
		filepath.Join(".atl", "tmp"), // Local project temp
	}

	_, _ = fmt.Fprintln(stdout, "Starting deep clean up...")

	for _, target := range cleanupTargets {
		if _, err := os.Stat(target); os.IsNotExist(err) {
			continue
		}

		_, _ = fmt.Fprintf(stdout, "  [clean] Removing %s...\n", target)
		if err := os.RemoveAll(target); err != nil {
			_, _ = fmt.Fprintf(stdout, "  [error] Failed to remove %s: %v\n", target, err)
		} else {
			_, _ = fmt.Fprintf(stdout, "  [done] %s cleared.\n", target)
		}
	}

	_, _ = fmt.Fprintln(stdout, "\nClean up complete. System is lean.")
	return nil
}
