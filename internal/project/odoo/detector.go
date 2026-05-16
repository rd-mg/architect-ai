package odoo

import (
    "bufio"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

// Info holds detected Odoo project metadata
type Info struct {
    IsOdoo          bool
    Version         string
    ManifestPath    string
    OverlayInstalled bool
    AvailableAgents []string
    AvailableSkills []string
}

var versionRe = regexp.MustCompile(`["']version["']\s*:\s*["'](\d+)\.`)

// Detect inspects projectDir for Odoo signals
func Detect(projectDir string) (*Info, error) {
    info := &Info{}

    // Signal 1: __manifest__.py
    if mp := findFile(projectDir, "__manifest__.py"); mp != "" {
        info.IsOdoo = true
        info.ManifestPath = mp
        info.Version = extractVersion(mp)
    }

    // Signal 2: odoo in requirements / pyproject
    if !info.IsOdoo {
        for _, f := range []string{"requirements.txt", "pyproject.toml"} {
            if containsLine(filepath.Join(projectDir, f), "odoo") {
                info.IsOdoo = true
                break
            }
        }
    }

    if !info.IsOdoo {
        return info, nil
    }

    overlayDir := filepath.Join(projectDir, ".atl", "overlays", "odoo-development-skill")
    if _, err := os.Stat(overlayDir); err == nil {
        info.OverlayInstalled = true
        info.AvailableAgents = listMDs(filepath.Join(overlayDir, "agents"))
    }

    return info, nil
}

func findFile(root, name string) string {
    var found string
    filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error { //nolint:errcheck
        if err != nil {
            return nil
        }
        skip := map[string]bool{".git": true, "node_modules": true, "__pycache__": true}
        if d.IsDir() && skip[d.Name()] {
            return filepath.SkipDir
        }
        if !d.IsDir() && d.Name() == name {
            found = p
            return filepath.SkipAll
        }
        return nil
    })
    return found
}

func extractVersion(path string) string {
    f, _ := os.Open(path)
    if f == nil {
        return "unknown"
    }
    defer f.Close()
    s := bufio.NewScanner(f)
    for s.Scan() {
        if m := versionRe.FindStringSubmatch(s.Text()); len(m) > 1 {
            return m[1]
        }
    }
    return "unknown"
}

func containsLine(path, prefix string) bool {
    f, _ := os.Open(path)
    if f == nil {
        return false
    }
    defer f.Close()
    s := bufio.NewScanner(f)
    for s.Scan() {
        if strings.HasPrefix(strings.ToLower(strings.TrimSpace(s.Text())), prefix) {
            return true
        }
    }
    return false
}

func listMDs(dir string) []string {
    entries, _ := os.ReadDir(dir)
    var names []string
    for _, e := range entries {
        if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
            names = append(names, strings.TrimSuffix(e.Name(), ".md"))
        }
    }
    return names
}
