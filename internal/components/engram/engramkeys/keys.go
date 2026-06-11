package engramkeys

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
var collapseDash = regexp.MustCompile(`-+`)

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
	hash := sha256.Sum256([]byte(query))
	return fmt.Sprintf("research/%s/%s-%x", tool, cleaned, hash[:4])
}
