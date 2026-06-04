package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/rd-mg/architect-ai/internal/firewall"
	"github.com/rd-mg/architect-ai/internal/paths"
)

func RunFirewall(args []string, stdout io.Writer, stderr io.Writer) error {
	devMode := false
	var filteredArgs []string
	for _, arg := range args {
		if arg == "--dev" {
			devMode = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	ctx := paths.New(".", devMode)

	if len(filteredArgs) < 1 {
		return fmt.Errorf("usage: architect-ai firewall [check|inject] [--dev]")
	}

	targets := firewall.GetTargets(ctx)

	switch filteredArgs[0] {
	case "check":
		results, ok := firewall.Check(targets)
		for _, r := range results {
			status := "OK"
			if !r.Present {
				status = "MISSING"
			}
			fmt.Fprintf(stdout, "FIREWALL %-60s %s (pattern: %s)\n", r.File, status, r.Pattern)
		}
		if !ok {
			return fmt.Errorf("firewall check failed")
		}
		return nil

	case "inject":
		// Ensure source file exists
		if _, err := os.Stat(firewall.FirewallSource); os.IsNotExist(err) {
			return fmt.Errorf("ERROR: %s not found — create it first", firewall.FirewallSource)
		}

		results := firewall.Inject(targets)
		hasError := false
		for _, r := range results {
			if r.Error == "already present" {
				fmt.Fprintf(stdout, "SKIP  %s: already present\n", r.File)
			} else if !r.OK {
				fmt.Fprintf(stderr, "FAIL  %s: %s\n", r.File, r.Error)
				hasError = true
			} else {
				fmt.Fprintf(stdout, "OK    %s: firewall %s\n", r.File, r.Error)
			}
		}
		if hasError {
			return fmt.Errorf("firewall injection failed")
		}
		return nil

	default:
		return fmt.Errorf("unknown subcommand: %s", filteredArgs[0])
	}
}
