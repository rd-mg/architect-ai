package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/rd-mg/architect-ai/internal/builder"
	"github.com/rd-mg/architect-ai/internal/config"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	// Verify foundation.md exists (prerequisite)
	if _, err := os.Stat(".atl/_generated/foundation.md"); os.IsNotExist(err) {
		fmt.Fprintln(stderr, "BUILD_ERROR: .atl/_generated/foundation.md not found.")
		fmt.Fprintln(stderr, "  Run: architect-ai foundation")
		return 1
	}

	// Build each deployed config
	resolver := builder.NewResolver()
	b := builder.New(resolver)

	exitCode := 0
	for _, agent := range config.KnownAgents() {
		dir := config.AgentDir(agent)
		for _, file := range config.AgentFiles(agent) {
			srcPath := fmt.Sprintf("internal/assets/%s/%s", dir, file)
			dstPath := resolveDestination(agent, file)

			srcContent, err := os.ReadFile(srcPath)
			if err != nil {
				fmt.Fprintf(stderr, "  FAIL %s: source not found (%v)\n", dstPath, err)
				exitCode = 1
				continue
			}

			result, err := b.Build(string(srcContent), readAsset)
			if err != nil {
				fmt.Fprintf(stderr, "  FAIL %s: %v\n", dstPath, err)
				exitCode = 1
				continue
			}

			// Atomic write
			tmpPath := dstPath + ".tmp"
			if err := os.WriteFile(tmpPath, []byte(result.Content), 0644); err != nil {
				fmt.Fprintf(stderr, "  FAIL %s: write error (%v)\n", dstPath, err)
				exitCode = 1
				continue
			}
			if err := os.Rename(tmpPath, dstPath); err != nil {
				fmt.Fprintf(stderr, "  FAIL %s: rename error (%v)\n", dstPath, err)
				exitCode = 1
				continue
			}

			hash := sha256Of(result.Content)
			fmt.Fprintf(stdout, "  OK   %s (%d bytes, hash:%s)\n", dstPath, len(result.Content), hash[:8])
		}
	}

	if exitCode != 0 {
		fmt.Fprintln(stderr, "\nBUILD FAILED — see errors above.")
		return 1
	}
	fmt.Fprintln(stdout, "\nBUILD OK — all deployed configs materialized.")
	return 0
}

// resolveDestination maps agent config files to their deployment paths.
func resolveDestination(agent, file string) string {
	switch {
	case agent == "antigravity":
		return ".antigravity/" + file
	case file == "architect.md":
		return "CLAUDE.md"
	case file == "general-orchestrator.md":
		return ".github/copilot-instructions.md"
	default:
		return file
	}
}

// readAsset implements builder.ReadFunc, reading from internal/assets/.
func readAsset(path string) (string, error) {
	fullPath := "internal/assets/" + path
	data, err := os.ReadFile(fullPath)
	if err != nil {
		// Try relative to project root
		data, err = os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("asset not found: %s", path)
		}
	}
	return string(data), nil
}

func sha256Of(content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", sum)
}
