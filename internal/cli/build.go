package cli

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rd-mg/architect-ai/internal/builder"
	"github.com/rd-mg/architect-ai/internal/config"
	"github.com/rd-mg/architect-ai/internal/paths"
)

func RunBuild(args []string) error {
	devMode := false
	for _, arg := range args {
		if arg == "--dev" {
			devMode = true
			break
		}
	}
	ctx := paths.New(".", devMode)

	// Verify foundation.md exists (prerequisite) - only if in target mode
	if !ctx.IsDevMode {
		if _, err := os.Stat(ctx.FoundationPath()); os.IsNotExist(err) {
			return fmt.Errorf("BUILD_ERROR: .atl/_generated/foundation.md not found. Run: architect-ai foundation")
		}
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
				fmt.Fprintf(os.Stderr, "  FAIL %s: source not found (%v)\n", dstPath, err)
				exitCode = 1
				continue
			}

			result, err := b.Build(string(srcContent), readAsset)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  FAIL %s: %v\n", dstPath, err)
				exitCode = 1
				continue
			}

			// Atomic write
			tmpPath := dstPath + ".tmp"
			dirDest := filepath.Dir(dstPath)
			if dirDest != "." {
				os.MkdirAll(dirDest, 0755)
			}
			if err := os.WriteFile(tmpPath, []byte(result.Content), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "  FAIL %s: write error (%v)\n", dstPath, err)
				exitCode = 1
				continue
			}
			if err := os.Rename(tmpPath, dstPath); err != nil {
				fmt.Fprintf(os.Stderr, "  FAIL %s: rename error (%v)\n", dstPath, err)
				exitCode = 1
				continue
			}

			hash := sha256Of(result.Content)
			fmt.Fprintf(os.Stdout, "  OK   %s (%d bytes, hash:%s)\n", dstPath, len(result.Content), hash[:8])
		}
	}

	if exitCode != 0 {
		return fmt.Errorf("BUILD FAILED — see errors above")
	}
	fmt.Fprintln(os.Stdout, "\nBUILD OK — all deployed configs materialized.")
	return nil
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
