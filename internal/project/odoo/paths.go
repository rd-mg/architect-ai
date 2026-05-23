package odoo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type PathConfig struct {
	WorkspaceRoot  string
	CommunityPath  string
	EnterprisePath string
	OCAPath        string
	OdooVersion    string
}

func ResolvePaths(workspaceRoot string) *PathConfig {
	cfg := &PathConfig{WorkspaceRoot: workspaceRoot}
	cfg.loadFromAtlConfig(filepath.Join(workspaceRoot, ".atl", "config.yaml"))
	if cfg.CommunityPath == "" {
		cfg.CommunityPath = os.Getenv("ODOO_COMMUNITY_PATH")
	}
	if cfg.EnterprisePath == "" {
		cfg.EnterprisePath = os.Getenv("ODOO_ENTERPRISE_PATH")
	}
	if cfg.CommunityPath == "" {
		cfg.CommunityPath = discoverOdooPath()
	}
	return cfg
}

func discoverOdooPath() string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, "gitproj", "odoo", "community"),
		filepath.Join(home, "odoo", "community"),
		filepath.Join(home, "src", "odoo"),
		"/opt/odoo",
		"/usr/local/lib/odoo",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	if out, err := exec.Command("python3", "-c",
		"import odoo, os; print(os.path.dirname(odoo.__file__))").Output(); err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}

func (c *PathConfig) loadFromAtlConfig(configPath string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "odoo_community_path:") {
			c.CommunityPath = strings.Trim(strings.TrimSpace(
				strings.TrimPrefix(line, "odoo_community_path:")), `"'`)
		}
		if strings.HasPrefix(line, "odoo_version:") {
			c.OdooVersion = strings.Trim(strings.TrimSpace(
				strings.TrimPrefix(line, "odoo_version:")), `"'`)
		}
	}
}

func (c *PathConfig) ToEnvMap() map[string]string {
	m := map[string]string{
		"WORKSPACE_ROOT":       c.WorkspaceRoot,
		"ODOO_COMMUNITY_PATH":  c.CommunityPath,
		"ODOO_ENTERPRISE_PATH": c.EnterprisePath,
		"ODOO_OCA_PATH":        c.OCAPath,
		"ODOO_VERSION":         c.OdooVersion,
	}
	for k, v := range m {
		if v == "" {
			delete(m, k)
		}
	}
	return m
}
