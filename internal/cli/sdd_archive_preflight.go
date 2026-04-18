package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rd-mg/architect-ai/internal/components/openspec"
)

// RunSDDArchivePreflight runs a dry-run of the archive's conflict check.
// Returns nil on success (no conflicts), or an error if conflicts exist or
// a technical error occurs.
func RunSDDArchivePreflight(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("sdd-archive-preflight", flag.ContinueOnError)
	fs.SetOutput(stderr)
	projectRoot := fs.String("project", ".", "Project root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: architect-ai sdd-archive-preflight <change-name>")
	}
	changeName := fs.Arg(0)

	absRoot, err := filepath.Abs(*projectRoot)
	if err != nil {
		return fmt.Errorf("resolve project root: %w", err)
	}

	deltaSpecsRoot := filepath.Join(absRoot, "openspec", "changes", changeName, "specs")
	entries, err := os.ReadDir(deltaSpecsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(stdout, "No delta specs to check.")
			return nil
		}
		return fmt.Errorf("read delta specs directory: %w", err)
	}

	var conflicts []openspec.ConflictReport
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		domain := e.Name()
		// Check for spec.md inside the domain folder.
		specPath := filepath.Join(deltaSpecsRoot, domain, "spec.md")
		if _, err := os.Stat(specPath); os.IsNotExist(err) {
			continue
		}

		report, err := openspec.CheckConflict(absRoot, changeName, domain)
		if err != nil && report == nil {
			return fmt.Errorf("error checking conflict for domain %q: %w", domain, err)
		}
		if report != nil {
			conflicts = append(conflicts, *report)
		}
	}

	if len(conflicts) == 0 {
		fmt.Fprintln(stdout, "OK: no conflicts.")
		return nil
	}

	if err := openspec.WriteConflictReport(absRoot, changeName, conflicts); err != nil {
		return fmt.Errorf("write conflict report: %w", err)
	}

	// Update state.yaml sdd-archive phase to failed.
	statePath := filepath.Join(absRoot, "openspec", "changes", changeName, "state.yaml")
	if s, err := openspec.Load(statePath); err == nil {
		if ph, ok := s.Phases["sdd-archive"]; ok {
			ph.Status = "failed"
			ph.Error = fmt.Sprintf("%d domain(s) in conflict; see merge-conflict.md", len(conflicts))
		}
		_ = openspec.Save(statePath, s)
	}

	fmt.Fprintf(stdout, "CONFLICTS: %d domain(s) have changed on main since this change was specced.\n", len(conflicts))
	fmt.Fprintf(stdout, "Report written to: openspec/changes/%s/merge-conflict.md\n", changeName)
	fmt.Fprintln(stdout, "\nPlease follow the recovery runbook: docs/openspec-merge-conflict.md")
	
	return fmt.Errorf("merge conflict detected in %d domain(s)", len(conflicts))
}
