package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rd-mg/architect-ai/internal/components/openspec"
)

// RunSDDStatus implements `architect-ai sdd-status [change-name]`.
func RunSDDStatus(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("sdd-status", flag.ContinueOnError)
	fs.SetOutput(stderr)
	projectRoot := fs.String("project", ".", "Project root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	changeName := ""
	if fs.NArg() >= 1 {
		changeName = fs.Arg(0)
	}

	absRoot, err := filepath.Abs(*projectRoot)
	if err != nil {
		return fmt.Errorf("resolve project root: %w", err)
	}

	if changeName == "" {
		// List all active changes.
		changesDir := filepath.Join(absRoot, "openspec", "changes")
		entries, err := os.ReadDir(changesDir)
		if err != nil {
			return fmt.Errorf("read changes directory: %w", err)
		}

		fmt.Fprintln(stdout, "Active Changes:")
		found := false
		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == "archive" {
				continue
			}
			found = true
			fmt.Fprintf(stdout, "  - %s\n", entry.Name())
		}
		if !found {
			fmt.Fprintln(stdout, "  (none)")
		}
		fmt.Fprintln(stdout, "\nUse 'architect-ai sdd-status [change-name]' for details.")
		return nil
	}

	statePath := filepath.Join(absRoot, "openspec", "changes", changeName, "state.yaml")
	s, err := openspec.Load(statePath)
	if err != nil {
		return fmt.Errorf("state.yaml invalid for change %q: %w\n\nFollow the runbook for recovery: docs/openspec-state-recovery.md", changeName, err)
	}

	renderStateTable(stdout, s)
	return nil
}

func renderStateTable(w io.Writer, s *openspec.State) {
	fmt.Fprintf(w, "Change: %s\n", s.ChangeName)
	fmt.Fprintf(w, "Store:  %s\n", s.ArtifactStore)
	fmt.Fprintf(w, "Created: %s\n", s.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "Updated: %s\n", s.UpdatedAt.Format(time.RFC3339))
	fmt.Fprintln(w, "\nPhases:")
	fmt.Fprintln(w, "  PHASE            STATUS        STARTED               COMPLETED             ARTIFACT(S)")
	fmt.Fprintln(w, "  -----------------------------------------------------------------------------------------")

	// Sort phases for consistent output.
	phases := make([]string, 0, len(s.Phases))
	for p := range s.Phases {
		phases = append(phases, p)
	}
	// Use canonical order if possible, else alphabetical.
	sort.Slice(phases, func(i, j int) bool {
		idxI := phaseIdx(phases[i])
		idxJ := phaseIdx(phases[j])
		if idxI != -1 && idxJ != -1 {
			return idxI < idxJ
		}
		return phases[i] < phases[j]
	})

	for _, name := range phases {
		ph := s.Phases[name]
		start := "-"
		if ph.StartedAt != nil {
			start = ph.StartedAt.Format("01-02 15:04")
		}
		end := "-"
		if ph.CompletedAt != nil {
			end = ph.CompletedAt.Format("01-02 15:04")
		}
		artifacts := "-"
		if ph.Artifact != "" {
			artifacts = ph.Artifact
		} else if len(ph.Artifacts) > 0 {
			artifacts = fmt.Sprintf("%d files", len(ph.Artifacts))
		}

		fmt.Fprintf(w, "  %-16s %-12s %-20s %-20s %s\n",
			name, strings.ToUpper(ph.Status), start, end, artifacts)
	}

	if s.Metering != nil {
		fmt.Fprintln(w, "\nMetering:")
		fmt.Fprintf(w, "  Tokens: %d\n", s.Metering.TotalTokens)
		fmt.Fprintf(w, "  Sessions: %d\n", s.Metering.Sessions)
		fmt.Fprintf(w, "  Cost (est): $%.2f\n", s.Metering.EstimatedCostUSD)
	}
}

func phaseIdx(name string) int {
	for i, v := range openspec.ValidPhases {
		if v == name {
			return i
		}
	}
	return -1
}
