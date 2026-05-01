package engramkeys

import (
	"fmt"
	"regexp"
	"strings"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
var collapseDash = regexp.MustCompile(`-+`)

// ResearchTopicKey generates a stable, slugified topic key for research findings.
// Format: research/{tool}/{slug}-len{N}
func ResearchTopicKey(tool, query string) string {
	cleaned := strings.ToLower(strings.TrimSpace(query))
	cleaned = nonAlnum.ReplaceAllString(cleaned, "-")
	cleaned = collapseDash.ReplaceAllString(cleaned, "-")
	cleaned = strings.Trim(cleaned, "-")
	if cleaned == "" {
		cleaned = "query"
	}
	if len(cleaned) > 50 {
		cleaned = strings.Trim(cleaned[:50], "-")
	}
	return fmt.Sprintf("research/%s/%s-len%d", tool, cleaned, len(query))
}
