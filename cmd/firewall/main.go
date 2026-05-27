package main

import (
	"fmt"
	"os"

	"github.com/rd-mg/architect-ai/internal/firewall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: architect-ai firewall [check|inject]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "check":
		results, ok := firewall.Check(firewall.Targets)
		for _, r := range results {
			status := "OK"
			if !r.Present {
				status = "MISSING"
			}
			fmt.Printf("FIREWALL %-60s %s (pattern: %s)\n", r.File, status, r.Pattern)
		}
		if !ok {
			os.Exit(1)
		}

	case "inject":
		// Ensure source file exists
		if _, err := os.Stat(firewall.FirewallSource); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "ERROR: %s not found — create it first\n", firewall.FirewallSource)
			os.Exit(1)
		}

		results := firewall.Inject(firewall.Targets)
		for _, r := range results {
			if r.Error == "already present" {
				fmt.Printf("SKIP  %s: already present\n", r.File)
			} else if !r.OK {
				fmt.Printf("FAIL  %s: %s\n", r.File, r.Error)
			} else {
				fmt.Printf("OK    %s: firewall %s\n", r.File, r.Error)
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", os.Args[1])
		os.Exit(1)
	}
}
