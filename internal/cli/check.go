package cli

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/rd-mg/architect-ai/internal/checker"
	"github.com/rd-mg/architect-ai/internal/config"
	"github.com/rd-mg/architect-ai/internal/paths"
)

var (
	contentRe     = regexp.MustCompile(`\{content from [^\}]+\}`)
	includeRe     = regexp.MustCompile(`\{\{[ \t]*include[ \t]+"([^"]+)"[ \t]*\}\}`)
	hashRe        = regexp.MustCompile(`\{[A-Z0-9_]+_HASH\}`)
	placeholderRe = regexp.MustCompile(`\{[A-Z][a-zA-Z ]+\}`)
)

func RunCheck(args []string, stdout io.Writer) error {
	target := "all"
	devMode := false
	for _, arg := range args {
		if arg == "--dev" {
			devMode = true
		} else if !strings.HasPrefix(arg, "-") {
			target = arg
		}
	}
	ctx := paths.New(".", devMode)

	var checks []checker.Check

	switch target {
	case "all":
		if !ctx.IsDevMode {
			checks = append(checks, checkFoundation(ctx))
			checks = append(checks, checkRegistry(ctx))
		}
		checks = append(checks, checkConfigs())
	case "foundation":
		if !ctx.IsDevMode {
			checks = append(checks, checkFoundation(ctx))
		} else {
			fmt.Fprintln(os.Stderr, "foundation check is not applicable in dev mode")
		}
	case "configs":
		checks = append(checks, checkConfigs())
	case "registry":
		if !ctx.IsDevMode {
			checks = append(checks, checkRegistry(ctx))
		} else {
			fmt.Fprintln(os.Stderr, "registry check is not applicable in dev mode")
		}
	default:
		return fmt.Errorf("Usage: architect-ai check [foundation|configs|registry|all] [--dev]")
	}

	results := checker.RunAll(checks...)
	for _, r := range results {
		fmt.Fprintf(stdout, "CHECK %s: %s", r.Name, strings.ToUpper(string(r.Status)))
		if r.Message != "" {
			fmt.Fprintf(stdout, " (%s)", r.Message)
		}
		fmt.Fprintln(stdout)
	}

	if !checker.AllPassed(results) {
		return fmt.Errorf("checks failed")
	}
	return nil
}

func checkFoundation(ctx paths.Context) checker.Check {
	return checker.Check{
		Name: "foundation",
		Run: func() error {
			_, err := os.Stat(ctx.FoundationPath())
			if os.IsNotExist(err) {
				return fmt.Errorf("not found")
			}
			if err != nil {
				return err
			}
			data, err := os.ReadFile(ctx.FoundationPath())
			if err != nil {
				return err
			}
			if !strings.Contains(string(data), "architect-ai:foundation:start") {
				return fmt.Errorf("file exists but missing foundation markers")
			}
			return nil
		},
	}
}

func checkConfigs() checker.Check {
	return checker.Check{
		Name: "configs",
		Run: func() error {
			var errs []string
			for _, agent := range config.KnownAgents() {
				for _, file := range config.AgentFiles(agent) {
					dst := resolveDestination(agent, file)
					data, err := os.ReadFile(dst)
					if err != nil {
						errs = append(errs, fmt.Sprintf("%s: file not found", dst))
						continue
					}
					content := string(data)
					lines := strings.Split(content, "\n")
					for i, l := range lines {
						if contentRe.MatchString(l) {
							errs = append(errs, fmt.Sprintf("%s:L%d: unresolved {content from ...} placeholder", dst, i+1))
						}
						if includeRe.MatchString(l) {
							errs = append(errs, fmt.Sprintf("%s:L%d: unresolved {{ include }} directive", dst, i+1))
						}
						if hashRe.MatchString(l) {
							errs = append(errs, fmt.Sprintf("%s:L%d: unresolved hash token", dst, i+1))
						}
					}
				}
			}
			if len(errs) > 0 {
				return fmt.Errorf("%s", strings.Join(errs, "; "))
			}
			return nil
		},
	}
}

func checkRegistry(ctx paths.Context) checker.Check {
	return checker.Check{
		Name: "registry",
		Run: func() error {
			if ctx.RegistryPath() == "" {
				return nil
			}
			data, err := os.ReadFile(ctx.RegistryPath())
			if err != nil {
				if os.IsNotExist(err) {
					return nil // It's ok if registry doesn't exist
				}
				return fmt.Errorf("file not found: %w", err)
			}
			content := string(data)
			var unfilled []string
			lines := strings.Split(content, "\n")
			for i, l := range lines {
				if strings.Contains(l, "{") && strings.Contains(l, "}") &&
					!strings.Contains(l, "<!--") && !strings.Contains(l, "-->") {
					if placeholderRe.MatchString(l) {
						unfilled = append(unfilled, fmt.Sprintf("L%d", i+1))
					}
				}
			}
			if len(unfilled) > 0 {
				return fmt.Errorf("unfilled placeholder(s) at lines: %s", strings.Join(unfilled, ", "))
			}
			return nil
		},
	}
}

func resolveDestination(agent, file string) string {
	switch {
	case agent == "antigravity":
		return ".antigravity/" + file
	case file == "architect.md":
		return "CLAUDE.md"
	case file == "general-orchestrator.md":
		return ".github/copilot-instructions.md"
	default:
		return file
	}
}
