package cli

import (
	"flag"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/rd-mg/architect-ai/internal/platform"
	"github.com/rd-mg/architect-ai/internal/setup"
)

type setupConfig struct {
	PlatformOverride string
	Yes              bool
	DryRun           bool
}

func RunSetup(args []string, stdout io.Writer, stderr io.Writer) error {
	cfg := parseSetupFlags(args)

	fmt.Fprintln(stdout, "=== architect-ai Platform Setup ===")
	fmt.Fprintln(stdout)

	// Step 1: Detect platform
	p, err := platform.Detect(cfg.PlatformOverride)
	if err != nil || p.Name == "unknown" {
		p = platform.PromptSelection()
	}
	fmt.Fprintf(stdout, "Platform: %s\n", p.Name)
	fmt.Fprintf(stdout, "Hook level: %s\n", p.HookLevel)

	// Step 2: Check Node.js
	if err := checkNodeJS(); err != nil {
		fmt.Fprintf(stderr, "ERROR: %v\n", err)
		fmt.Fprintln(stderr, "Install Node.js 18+ from https://nodejs.org/")
		return fmt.Errorf("node dependency missing")
	}
	fmt.Fprintln(stdout, "Node.js: OK")

	// Step 3: Install/verify context-mode
	cmVersion, err := setup.EnsureContextMode(cfg.DryRun)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR installing context-mode: %v\n", err)
		return fmt.Errorf("context-mode install failed: %w", err)
	}
	fmt.Fprintf(stdout, "context-mode: %s\n", cmVersion)

	// Step 4: Write platform config files
	if !cfg.DryRun {
		files, err := setup.ConfigurePlatform(p)
		if err != nil {
			fmt.Fprintf(stderr, "ERROR configuring platform: %v\n", err)
			return fmt.Errorf("platform config failed: %w", err)
		}
		for _, f := range files {
			fmt.Fprintf(stdout, "  Written: %s\n", f)
		}
	} else {
		files := setup.DryRunPlatform(p)
		fmt.Fprintln(stdout, "Would write:")
		for _, f := range files {
			fmt.Fprintf(stdout, "  %s\n", f)
		}
	}

	// Step 5: Run doctor
	if !cfg.DryRun {
		fmt.Fprintln(stdout, "\nRunning context-mode doctor...")
		doctorOut, err := runDoctor()
		if err != nil {
			fmt.Fprintf(stdout, "WARN: doctor failed: %v\n", err)
		} else {
			fmt.Fprintln(stdout, doctorOut)
		}
	}

	// Step 6: Write platform-config.yaml
	if !cfg.DryRun {
		if err := setup.WritePlatformConfig(p, cmVersion); err != nil {
			fmt.Fprintf(stderr, "WARN: cannot write .atl/platform-config.yaml: %v\n", err)
		}
		fmt.Fprintln(stdout, "\nSetup complete. .atl/platform-config.yaml written.")
	}

	// Step 7: Print next steps
	printNextSteps(p, stdout)
	return nil
}

func parseSetupFlags(args []string) setupConfig {
	var cfg setupConfig
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	fs.StringVar(&cfg.PlatformOverride, "platform", "", "Override auto-detection")
	fs.BoolVar(&cfg.Yes, "yes", false, "Non-interactive (no prompts)")
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "Show what would be done, no changes")
	// Ignore errors from Parse as it might just be unknown flags, or log them
	_ = fs.Parse(args)
	return cfg
}

func checkNodeJS() error {
	out, err := exec.Command("node", "--version").Output()
	if err != nil {
		return fmt.Errorf("node not found in PATH")
	}
	version := strings.TrimSpace(string(out))
	major := 0
	fmt.Sscanf(version, "v%d", &major)
	if major < 18 {
		return fmt.Errorf("node %s found but 18+ required", version)
	}
	return nil
}

func runDoctor() (string, error) {
	out, err := exec.Command("context-mode", "doctor").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func printNextSteps(p platform.Platform, stdout io.Writer) {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "=== Next Steps ===")
	fmt.Fprintln(stdout)

	if p.RequiresManualMCPSetup {
		fmt.Fprintf(stdout, "1. Add MCP server manually in %s settings:\n", p.DisplayName)
		fmt.Fprintln(stdout, "   Name: context-mode")
		fmt.Fprintln(stdout, "   Command: context-mode")
	}

	if p.RequiresRestart {
		fmt.Fprintf(stdout, "2. Restart %s to activate context-mode hooks.\n\n", p.DisplayName)
	}

	if p.RoutingFileRequired {
		fmt.Fprintf(stdout, "3. Routing file has been copied. Commit it:\n")
		fmt.Fprintf(stdout, "   git add %s && git commit -m 'chore: add context-mode routing'\n\n", p.RoutingFile)
	}

	fmt.Fprintln(stdout, "4. Build deployed configs:")
	fmt.Fprintln(stdout, "   architect-ai build")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "5. Inject gates:")
	fmt.Fprintln(stdout, "   architect-ai gate inject")
	fmt.Fprintln(stdout, "   architect-ai gate l2-purge")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "6. Inject firewall:")
	fmt.Fprintln(stdout, "   architect-ai firewall inject")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "7. Verify everything:")
	fmt.Fprintln(stdout, "   architect-ai check all")
}
