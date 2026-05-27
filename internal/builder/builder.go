package builder

import "fmt"

// Builder orchestrates template resolution using a Resolver.
// It validates inputs and delegates resolution, returning the
// assembled ResolveResult or a descriptive error.
type Builder struct {
	resolver Resolver
}

// New creates a Builder with the given Resolver.
// The resolver must not be nil.
func New(resolver Resolver) *Builder {
	if resolver == nil {
		panic("builder: nil resolver")
	}
	return &Builder{resolver: resolver}
}

// Build processes the given template content through the resolver.
// content is the raw template text to resolve.
// read is called by the resolver for each {{include}} or {content from}
// reference encountered.
//
// Build returns the resolved result, or an error if the content
// contains unresolvable directives.
func (b *Builder) Build(content string, read ReadFunc) (*ResolveResult, error) {
	if read == nil {
		return nil, fmt.Errorf("builder: read function is required")
	}

	result, err := b.resolver.Resolve(content, read)
	if err != nil {
		return nil, fmt.Errorf("builder: resolve: %w", err)
	}

	// Check for unresolved include errors in the output.
	if err := detectErrors(result.Content); err != nil {
		return result, err
	}

	return result, nil
}

// detectErrors scans resolved content for sentinel error markers left by
// the resolver when a read call fails during recursive resolution.
func detectErrors(content string) error {
	if idx := findErrorMarker(content); idx >= 0 {
		end := min(idx+140, len(content))
		msg := content[idx:end]
		return fmt.Errorf("builder: %s", msg)
	}
	return nil
}

// findErrorMarker returns the index of the first [[include error: or
// [[circular include: marker, or -1 if none is found.
func findErrorMarker(content string) int {
	for _, marker := range []string{"[[include error:", "[[circular include:"} {
		if idx := indexOf(content, marker); idx >= 0 {
			return idx
		}
	}
	return -1
}

// indexOf returns the index of substr in s, or -1 if not found.
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
