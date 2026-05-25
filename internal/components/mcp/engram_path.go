package mcp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func FindEngramBinary() (string, error) {
	if env := os.Getenv("ENGRAM_BIN"); env != "" && isExec(env) { return env, nil }
	if path, err := exec.LookPath("engram"); err == nil { return path, nil }
	home, _ := os.UserHomeDir()
	candidates := []string{
		"/usr/local/bin/engram", "/usr/bin/engram", "/opt/homebrew/bin/engram",
		filepath.Join(home, ".linuxbrew", "bin", "engram"),
		"/home/linuxbrew/.linuxbrew/bin/engram",
	}
	for _, c := range candidates {
		if isExec(c) { return c, nil }
	}
	for _, base := range []string{"/home/linuxbrew/.linuxbrew/Cellar", "/opt/homebrew/Cellar", "/usr/local/Cellar"} {
		pkgDir := filepath.Join(base, "engram")
		entries, err := os.ReadDir(pkgDir)
		if err != nil || len(entries) == 0 { continue }
		bin := filepath.Join(pkgDir, entries[len(entries)-1].Name(), "bin", "engram")
		if isExec(bin) { return bin, nil }
	}
	return "", fmt.Errorf("engram not found; install: brew install engram OR set ENGRAM_BIN")
}

func isExec(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Mode()&0111 != 0
}
