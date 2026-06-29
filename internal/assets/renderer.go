package assets

import "regexp"

// includePattern matches {{ include "path" }} directives in asset files.
// The resolved content of the referenced file replaces the directive.
var includePattern = regexp.MustCompile(`\{\{\s*include\s+"([^"]+)"\s*\}\}`)

// RenderIncludes replaces {{ include "path" }} directives with the
// content of the referenced embedded asset file.
// Panics if any referenced asset cannot be found (programming error).
func RenderIncludes(content []byte) []byte {
	return includePattern.ReplaceAllFunc(content, func(match []byte) []byte {
		submatches := includePattern.FindSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		assetPath := string(submatches[1])
		included := MustRead(assetPath)
		return []byte(included)
	})
}
