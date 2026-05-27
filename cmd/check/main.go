package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/rd-mg/architect-ai/internal/checker"
	"github.com/rd-mg/architect-ai/internal/config"
)

var (
	contentRe = regexp.MustCompile(`\{content from [^\}]+\}`)
	includeRe = regexp.MustCompile(`\{\{[ \t]*include[ \t]+"([^"]+)"[ \t]*\}\}`)
	hashRe    = regexp.MustCompile(`\{[A-Z0-9_]+_HASH\}`)
)

func main() {
	target := "all"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}

	var checks []checker.Check

	switch target {
	case "all":
		checks = append(checks, checkFoundation())
		checks = append(checks, checkConfigs())
		checks = append(checks, checkRegistry())
	case "foundation":
		checks = append(checks, checkFoundation())
	case "configs":
		checks = append(checks, checkConfigs())
	case "registry":
		checks = append(checks, checkRegistry())
	default:
		fmt.Fprintf(os.Stderr, "Usage: architect-ai check [foundation|configs|registry|all]\n")
		os.Exit(2)
	}

	results := checker.RunAll(checks...)
	for _, r := range results {
		fmt.Printf("CHECK %s: %s", r.Name, strings.ToUpper(string(r.Status)))
		if r.Message != "" {
			fmt.Printf(" (%s)", r.Message)
		}
		fmt.Println()
	}

	if !checker.AllPassed(results) {
		os.Exit(1)
	}
}

func checkFoundation() checker.Check {
	return checker.Check{
		Name: "foundation",
		Run: func() error {
			fi, err := os.Stat(".atl/_generated/foundation.md")
			if os.IsNotExist(err) {
				return fmt.Errorf("not found")
			}
			if err != nil {
				return err
			}
			data, err := os.ReadFile(".atl/_generated/foundation.md")
			if err != nil {
				return err
			}
			if !strings.Contains(string(data), "architect-ai:foundation:start") {
				return fmt.Errorf("file exists but missing foundation markers")
			}
			return fmt.Errorf("age=%s, size=%d", fi.ModTime().Format("15:04"), fi.Size())
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
					line := 1
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
						_ = line
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

func checkRegistry() checker.Check {
	return checker.Check{
		Name: "registry",
		Run: func() error {
			data, err := os.ReadFile(".atl/skill-registry.md")
			if err != nil {
				return fmt.Errorf("file not found")
			}
			content := string(data)
			var unfilled []string
			lines := strings.Split(content, "\n")
			for i, l := range lines {
				if strings.Contains(l, "{") && strings.Contains(l, "}") &&
					!strings.Contains(l, "<!--") && !strings.Contains(l, "-->") {
					// Check for {Capitalized text} patterns (unfilled placeholders)
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

var placeholderRe = regexp.MustCompile(`\{[A-Z][a-zA-Z ]+\}`)

// resolveDestination mirrors cmd/build logic.
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
