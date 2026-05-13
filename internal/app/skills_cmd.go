package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rd-mg/architect-ai/internal/skills"
)

// runSkillsPatternsCmd handles `architect-ai skills patterns [--skill <name>] [--clear]`.
func runSkillsPatternsCmd(ctx context.Context, args []string, stdout io.Writer) error {
	skillFilter := ""
	clearMode := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--skill":
			if i+1 < len(args) {
				skillFilter = args[i+1]
				i++
			}
		case "--clear":
			clearMode = true
		}
	}

	if clearMode {
		return clearPatterns(ctx, skillFilter, stdout)
	}
	return listPatterns(ctx, skillFilter, stdout)
}

func listPatterns(ctx context.Context, skillFilter string, stdout io.Writer) error {
	query := "knowledge/_global/skill/"
	if skillFilter != "" {
		query = "knowledge/_global/skill/" + skillFilter + "/learned-patterns"
	}

	cmd := exec.CommandContext(ctx, "engram", "search", query)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("engram search failed (is engram installed?): %w", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "" || strings.Contains(result, "no results") {
		_, _ = fmt.Fprintln(stdout, "No learned patterns found for this project.")
		_, _ = fmt.Fprintln(stdout, "Patterns accumulate after each successful sdd-archive cycle.")
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "Learned patterns (query: %s)\n\n", query)
	_, _ = fmt.Fprintln(stdout, result)
	return nil
}

func clearPatterns(ctx context.Context, skillFilter string, stdout io.Writer) error {
	if skillFilter == "" {
		return fmt.Errorf("--clear requires --skill <name>; use --clear-all to remove all patterns")
	}

	topicKey := "knowledge/_global/skill/" + skillFilter + "/learned-patterns"
	_, _ = fmt.Fprintf(stdout, "Clearing patterns for skill: %s\n", skillFilter)
	_, _ = fmt.Fprintf(stdout, "Topic key: %s\n\n", topicKey)
	_, _ = fmt.Fprintln(stdout, "To delete this observation, run:")
	_, _ = fmt.Fprintf(stdout, "  engram tui  (search for %q, then delete)\n", topicKey)
	_, _ = fmt.Fprintln(stdout, "\nAutomatic deletion is not supported — engram delete requires the observation ID.")
	_, _ = fmt.Fprintln(stdout, "Run `engram search \""+topicKey+"\"` to find the ID, then `engram delete <id>`.")
	return nil
}

func runSkillsCmd(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(stdout, "Usage: architect-ai skills <subcommand>")
		_, _ = fmt.Fprintln(stdout, "  patterns              Show learned patterns from Engram")
		_, _ = fmt.Fprintln(stdout, "  add <owner>/<repo>    Install a community skill")
		_, _ = fmt.Fprintln(stdout, "  remove <skill-id>     Remove a community skill")
		_, _ = fmt.Fprintln(stdout, "  list                  List installed community skills")
		_, _ = fmt.Fprintln(stdout, "  update                Update all community skills")
		return nil
	}
	switch args[0] {
	case "patterns":
		return runSkillsPatternsCmd(ctx, args[1:], stdout)
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("usage: architect-ai skills add <owner>/<repo>[/<path>]")
		}
		return runSkillsAdd(ctx, args[1], stdout)
	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("usage: architect-ai skills remove <skill-id>")
		}
		return runSkillsRemove(ctx, args[1], stdout)
	case "list":
		return runSkillsList(ctx, stdout)
	case "update":
		return runSkillsUpdate(ctx, stdout)
	default:
		return fmt.Errorf("unknown skills subcommand %q", args[0])
	}
}

func runSkillsAdd(ctx context.Context, ref string, stdout io.Writer) error {
	owner, repo, path, err := parseRef(ref)
	if err != nil {
		return err
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if path != "" {
		return installOne(ctx, owner, repo, path, homeDir, stdout)
	}

	// No path specified — list all skills in the repo
	_, _ = fmt.Fprintf(stdout, "Discovering skills in %s/%s...\n", owner, repo)
	paths, err := skills.ListSkillPathsInRepo(ctx, owner, repo)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return fmt.Errorf("no SKILL.md files found in %s/%s", owner, repo)
	}
	_, _ = fmt.Fprintf(stdout, "Found %d skill(s) — installing all.\n", len(paths))
	for _, p := range paths {
		if err := installOne(ctx, owner, repo, p, homeDir, stdout); err != nil {
			_, _ = fmt.Fprintf(stdout, "  [warn] %s: %v\n", p, err)
		}
	}
	return nil
}

func installOne(ctx context.Context, owner, repo, path, homeDir string, stdout io.Writer) error {
	id := deriveID(repo, path)
	result, err := skills.FetchSkillMD(ctx, owner, repo, path)
	if err != nil {
		return fmt.Errorf("%s: %w", id, err)
	}
	// Write to agent skill directories using the existing skill install helper.
	// We need to implement skills.WriteToAgentDirs and skills.RemoveFromAgentDirs.
	if err := skills.WriteToAgentDirs(homeDir, id, result.Content); err != nil {
		return err
	}
	m, _ := skills.LoadManifest(homeDir)
	m.Add(skills.CommunitySkillEntry{
		ID:          id,
		Source:      owner + "/" + repo,
		Path:        path,
		SHA:         result.SHA,
		InstalledAt: time.Now().UTC(),
	})
	_ = skills.SaveManifest(homeDir, m)

	// Also update skills-lock.json in the current project directory if present.
	if wd, err := os.Getwd(); err == nil {
		if lockfile, lErr := skills.LoadLockfile(wd); lErr == nil {
			lockfile.Add(skills.SkillManifestEntry{
				ID:          id,
				Source:      "community",
				Path:        path,
				SHA:         result.SHA,
				Kind:        "Community",
				InstalledAt: time.Now().UTC(),
			})
			_ = skills.SaveLockfile(wd, lockfile)
		}
	}

	_, _ = fmt.Fprintf(stdout, "  [ok] %s installed\n", id)
	return nil
}

func runSkillsRemove(ctx context.Context, id string, stdout io.Writer) error {
	homeDir, _ := os.UserHomeDir()
	m, err := skills.LoadManifest(homeDir)
	if err != nil {
		return err
	}
	if m.FindByID(id) == nil {
		return fmt.Errorf("skill %q not in community manifest", id)
	}
	if err := skills.RemoveFromAgentDirs(homeDir, id); err != nil {
		return err
	}
	m.Remove(id)
	_ = skills.SaveManifest(homeDir, m)

	// Also remove from skills-lock.json in the current project directory if present.
	if wd, err := os.Getwd(); err == nil {
		if lockfile, err := skills.LoadLockfile(wd); err == nil {
			if lockfile.Remove(id) {
				_ = skills.SaveLockfile(wd, lockfile)
			}
		}
	}

	_, _ = fmt.Fprintf(stdout, "[ok] %s removed\n", id)
	return nil
}

func runSkillsList(_ context.Context, stdout io.Writer) error {
	homeDir, _ := os.UserHomeDir()
	m, err := skills.LoadManifest(homeDir)
	if err != nil {
		return err
	}
	if len(m.Skills) == 0 {
		_, _ = fmt.Fprintln(stdout, "No community skills installed.")
		_, _ = fmt.Fprintln(stdout, "Install with: architect-ai skills add <owner>/<repo>")
		return nil
	}
	_, _ = fmt.Fprintf(stdout, "Community skills (%d installed):\n", len(m.Skills))
	for _, s := range m.Skills {
		_, _ = fmt.Fprintf(stdout, "  %-30s  %s  (%s)\n",
			s.ID, s.InstalledAt.Format("2006-01-02"), s.Source)
	}
	return nil
}

func runSkillsUpdate(ctx context.Context, stdout io.Writer) error {
	homeDir, _ := os.UserHomeDir()
	m, err := skills.LoadManifest(homeDir)
	if err != nil {
		return err
	}
	if len(m.Skills) == 0 {
		_, _ = fmt.Fprintln(stdout, "No community skills to update.")
		return nil
	}
	updated := 0
	for i, e := range m.Skills {
		parts := strings.SplitN(e.Source, "/", 2)
		if len(parts) != 2 {
			continue
		}
		result, err := skills.FetchSkillMD(ctx, parts[0], parts[1], e.Path)
		if err != nil {
			_, _ = fmt.Fprintf(stdout, "  [skip] %s: %v\n", e.ID, err)
			continue
		}
		if result.SHA != "" && result.SHA == e.SHA {
			_, _ = fmt.Fprintf(stdout, "  [up-to-date] %s\n", e.ID)
			continue
		}
		if err := skills.WriteToAgentDirs(homeDir, e.ID, result.Content); err != nil {
			_, _ = fmt.Fprintf(stdout, "  [error] %s: %v\n", e.ID, err)
			continue
		}
		m.Skills[i].SHA = result.SHA
		updated++
		_, _ = fmt.Fprintf(stdout, "  [ok] %s updated\n", e.ID)
	}
	if updated > 0 {
		_ = skills.SaveManifest(homeDir, m)
		// Also update skills-lock.json SHA entries.
		if wd, err := os.Getwd(); err == nil {
			if lockfile, err := skills.LoadLockfile(wd); err == nil {
				for _, e := range m.Skills {
					if entry := lockfile.FindByID(e.ID); entry != nil {
						entry.SHA = e.SHA
						entry.InstalledAt = time.Now().UTC()
					}
				}
				_ = skills.SaveLockfile(wd, lockfile)
			}
		}
	}
	_, _ = fmt.Fprintf(stdout, "%d skill(s) updated.\n", updated)
	return nil
}

func parseRef(ref string) (owner, repo, path string, err error) {
	parts := strings.SplitN(ref, "/", 3)
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("invalid ref %q — expected <owner>/<repo>[/<path>]", ref)
	}
	if len(parts) == 3 {
		return parts[0], parts[1], parts[2], nil
	}
	return parts[0], parts[1], "", nil
}

func deriveID(repo, path string) string {
	if path == "" {
		return strings.ToLower(repo)
	}
	return strings.ToLower(filepath.Base(path))
}
