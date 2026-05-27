// Package builder provides template resolution for assembling agent
// configuration from reusable fragments. It supports include directives,
// content-from references, and hash-token substitution.
package builder

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Include, content-from, and hash-token patterns.
var (
	includeRe = regexp.MustCompile(`\{\{[ \t]*include[ \t]+"([^"]+)"[ \t]*\}\}`)
	contentRe = regexp.MustCompile(`\{content from ([^\}]+)\}`)
	hashRe    = regexp.MustCompile(`\{([A-Z0-9_]+_HASH)\}`)
)

// ErrCircularInclude is returned when a template chain includes itself.
type ErrCircularInclude struct {
	Chain []string
}

func (e *ErrCircularInclude) Error() string {
	return fmt.Sprintf("circular include detected: %s", strings.Join(e.Chain, " → "))
}

// ReadFunc reads a template or asset by path and returns its content.
// When the path is unknown, ReadFunc must return an error.
type ReadFunc func(path string) (string, error)

// ResolveResult holds the fully resolved content and the list of source files
// that contributed to it (deduplicated, in resolve order).
type ResolveResult struct {
	Content string
	Sources []string
}

// Resolver resolves templates by expanding include directives, content-from
// placeholders, and hash tokens.
type Resolver interface {
	// Resolve processes content, resolving all directives via read.
	// The read function is called for each {{include}} or {content from}
	// reference encountered.
	Resolve(content string, read ReadFunc) (*ResolveResult, error)
}

// NewResolver creates a Resolver that processes includes, content-from
// references, and hash-token substitutions.
func NewResolver() Resolver {
	return &resolver{}
}

type resolver struct{}

func (r *resolver) Resolve(content string, read ReadFunc) (*ResolveResult, error) {
	if content == "" {
		return &ResolveResult{}, nil
	}

	res := &resolveCtx{seen: map[string]bool{}}
	out, err := r.resolvePass(content, read, res)
	res.Content = out
	if err != nil {
		return &res.ResolveResult, err
	}
	return &res.ResolveResult, nil
}

// resolveCtx accumulates sources and guards against circular includes.
type resolveCtx struct {
	ResolveResult
	chain []string
	seen  map[string]bool // guard against repeated sources in output
}

// resolvePass runs one resolution pass over content, recursively expanding
// includes and resolving content-from and hash-token placeholders.
// On error the returned string holds whatever content was resolved so far.
func (r *resolver) resolvePass(content string, read ReadFunc, ctx *resolveCtx) (string, error) {
	// Phase 1: expand {{include "path"}} directives.
	out, err := r.expandIncludes(content, read, ctx)
	if err != nil {
		return out, err
	}

	// Phase 2: resolve {content from path} placeholders.
	out, err = r.resolveContentFrom(out, read, ctx)
	if err != nil {
		return out, err
	}

	// Phase 3: substitute {TOKEN_HASH} placeholders.
	out = r.substituteHashes(out)

	return out, nil
}

// expandIncludes replaces {{include "path"}} with the resolved content of that
// file. Includes are resolved recursively.
func (r *resolver) expandIncludes(content string, read ReadFunc, ctx *resolveCtx) (string, error) {
	return includeRe.ReplaceAllStringFunc(content, func(match string) string {
		matches := includeRe.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		path := matches[1]

		if slices.Contains(ctx.chain, path) {
			ctx.chain = append(ctx.chain, path)
			// Return the error text in place — the outer flow
			// detects non-empty chains with issues.
			return fmt.Sprintf("[[circular include: %s]]", strings.Join(ctx.chain, " → "))
		}

		ctx.chain = append(ctx.chain, path)

		included, err := read(path)
		if err != nil {
			return fmt.Sprintf("[[include error: %s]]", err)
		}

		if !ctx.seen[path] {
			ctx.Sources = append(ctx.Sources, path)
			ctx.seen[path] = true
		}

		resolved, err := r.resolvePass(included, read, ctx)
		if err != nil {
			return fmt.Sprintf("[[include error: %s]]", err)
		}

		ctx.chain = ctx.chain[:len(ctx.chain)-1]
		return resolved
	}), nil
}

// resolveContentFrom replaces {content from path} with the raw content read
// from that path. These are NOT resolved recursively.
func (r *resolver) resolveContentFrom(content string, read ReadFunc, ctx *resolveCtx) (string, error) {
	var lastErr error
	out := contentRe.ReplaceAllStringFunc(content, func(match string) string {
		matches := contentRe.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		path := strings.TrimSpace(matches[1])
		data, err := read(path)
		if err != nil {
			lastErr = fmt.Errorf("{content from %s}: %w", path, err)
			return match
		}
		if !ctx.seen[path] {
			ctx.Sources = append(ctx.Sources, path)
			ctx.seen[path] = true
		}
		return data
	})
	return out, lastErr
}

// substituteHashes replaces {TOKEN_HASH} placeholders with the SHA-256 hex
// digest of the resolved content up to that point.
func (r *resolver) substituteHashes(content string) string {
	return hashRe.ReplaceAllStringFunc(content, func(match string) string {
		matches := hashRe.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		sum := sha256.Sum256([]byte(content))
		return fmt.Sprintf("%s%x", matches[1], sum)
	})
}
