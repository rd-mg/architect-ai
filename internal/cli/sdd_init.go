package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rd-mg/architect-ai/internal/components/filemerge"
)

var osTimeNow = time.Now

func RunSddInit(args []string, stdout io.Writer) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve current working directory: %w", err)
	}

	// Resolve absolute path upfront to avoid relative path issues
	absProjectRoot, absErr := filepath.Abs(projectRoot)
	if absErr != nil {
		return fmt.Errorf("resolve absolute path for project root: %w", absErr)
	}

	fs := flag.NewFlagSet("sdd-init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	mode := fs.String("mode", "engram", "SDD persistence mode: engram, openspec, or hybrid")
	contextData := fs.String("context", "", "Initial project context designed by LLM (YAML/JSON)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// 1. Ensure the project registry and .atl folder are ready.
	// This centralizes .atl creation, overlay discovery, skill indexing,
	// and core project conventions (AGENTS.md, GEMINI.md).
	if _, err := EnsureProjectRegistryReady(absProjectRoot); err != nil {
		return fmt.Errorf("ensure registry ready: %w", err)
	}

	atlDir := filepath.Join(absProjectRoot, ".atl")

	// 2. Save designed context if provided
	if *contextData != "" {
		contextPath := filepath.Join(atlDir, "project-context.yaml")
		_, err := filemerge.WriteFileAtomic(contextPath, []byte(*contextData), 0o644)
		if err != nil {
			return fmt.Errorf("save project context: %w", err)
		}
		fmt.Printf("Created/Updated: %s\n", contextPath)
	}

	// 3. Bootstrap openspec if requested
	if *mode == "openspec" || *mode == "hybrid" {
		if err := bootstrapOpenSpec(absProjectRoot); err != nil {
			return fmt.Errorf("bootstrap openspec: %w", err)
		}
		fmt.Fprintln(stdout, "Bootstrapped openspec/ directory structure.")
	}

	// 4. Write CLI Bootstrap marker
	bootstrapPath := filepath.Join(atlDir, "state", "bootstrap.json")
	if err := os.MkdirAll(filepath.Dir(bootstrapPath), 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	marker := map[string]any{
		"version":    "1.0",
		"mode":       *mode,
		"bootstrapped_at": osTimeNow().Format(time.RFC3339),
	}
	markerBytes, _ := json.MarshalIndent(marker, "", "  ")
	if _, err := filemerge.WriteFileAtomic(bootstrapPath, markerBytes, 0o644); err != nil {
		return fmt.Errorf("write bootstrap marker: %w", err)
	}

	fmt.Fprintf(stdout, "SDD Bootstrap successful in %s mode.\n", *mode)
	fmt.Fprintln(stdout, "You may now run the 'sdd-init' Phase (AI Analysis) to complete project setup.")
	return nil
}

func bootstrapOpenSpec(projectRoot string) error {
	dirs := []string{
		filepath.Join(projectRoot, "openspec", "specs"),
		filepath.Join(projectRoot, "openspec", "changes", "archive"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	configPath := filepath.Join(projectRoot, "openspec", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := "# openspec/config.yaml\nschema: spec-driven\n"
		_, err = filemerge.WriteFileAtomic(configPath, []byte(defaultConfig), 0o644)
		if err != nil {
			return err
		}
	}

	return nil
}
