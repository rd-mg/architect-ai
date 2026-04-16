package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rd-mg/architect-ai/internal/components/filemerge"
)

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
	fs.SetOutput(ioDiscard{})
	mode := fs.String("mode", "engram", "SDD persistence mode: engram, openspec, or hybrid")
	contextData := fs.String("context", "", "Initial project context designed by LLM (YAML/JSON)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// 1. Ensure the project registry and .atl folder are ready.
	// This centralizes .atl creation, overlay discovery, skill indexing,
	// and core project conventions (AGENTS.md, GEMINI.md).
	if err := EnsureSDDReady(absProjectRoot); err != nil {
		return fmt.Errorf("ensure sdd ready: %w", err)
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
		fmt.Println("Bootstrapped openspec/ directory structure.")
	}

	fmt.Printf("SDD initialized successfully in %s mode.\n", *mode)
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
