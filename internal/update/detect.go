package update

import (
	"context"
	"regexp"
	"strings"

	"github.com/rd-mg/architect-ai/internal/process"
)

// versionRegexp extracts a semver-like version from command output.
// Same pattern as internal/system/deps.go for consistency.
var versionRegexp = regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?)`)

// devVersionRegexp matches common unversioned source-build output like
// "engram dev" or "version: dev".
var devVersionRegexp = regexp.MustCompile(`(?i)(?:^|\s)dev(?:$|\s)`)

// detectInstalledVersion determines the installed version of a tool.
// For tools with nil DetectCmd (architect-ai), returns currentBuildVersion.
// For other tools, checks LookPath then runs the detect command.
func detectInstalledVersion(ctx context.Context, tool ToolInfo, currentBuildVersion string) string {
	if tool.DetectCmd == nil {
		return currentBuildVersion
	}

	if len(tool.DetectCmd) == 0 {
		return ""
	}

	// Run version command using bounded process runner.
	res, _ := process.Run(ctx, tool.DetectCmd[0], tool.DetectCmd[1:], process.OptionsFor(process.FastCheck))
	if res.Error != nil {
		return "" // command failed or timed out — binary exists but version unknown
	}

	return parseVersionFromOutput(strings.TrimSpace(string(res.Stdout)))
}

// parseVersionFromOutput extracts the first semver-like pattern from raw output.
func parseVersionFromOutput(output string) string {
	if output == "" {
		return ""
	}

	if devVersionRegexp.MatchString(output) {
		return "dev"
	}

	match := versionRegexp.FindStringSubmatch(output)
	if len(match) >= 2 {
		return match[1]
	}

	return ""
}
