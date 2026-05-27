package builder

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

// fakeFS implements ReadFunc over a static map for use in tests.
type fakeFS map[string]string

func (fs fakeFS) Read(path string) (string, error) {
	content, ok := fs[path]
	if !ok {
		return "", fmt.Errorf("file not found: %s", path)
	}
	return content, nil
}

func TestNewResolver(t *testing.T) {
	r := NewResolver()
	if r == nil {
		t.Fatal("NewResolver() returned nil")
	}
}

func TestResolver_EmptyContent(t *testing.T) {
	r := NewResolver()
	result, err := r.Resolve("", nil)
	if err != nil {
		t.Fatalf("Resolve('') unexpected error: %v", err)
	}
	if result.Content != "" {
		t.Errorf("Resolve('') content = %q, want empty", result.Content)
	}
}

func TestResolver_NoDirectives(t *testing.T) {
	r := NewResolver()
	fs := fakeFS{}

	input := "plain text without directives"
	result, err := r.Resolve(input, fs.Read)
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if result.Content != input {
		t.Errorf("Resolve() = %q, want %q", result.Content, input)
	}
	if len(result.Sources) != 0 {
		t.Errorf("Resolve() sources = %v, want empty", result.Sources)
	}
}

func TestResolver_IncludeSimple(t *testing.T) {
	r := NewResolver()
	fs := fakeFS{
		"header.md": "# Header",
	}

	input := "before\n{{include \"header.md\"}}\nafter"
	result, err := r.Resolve(input, fs.Read)
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}

	want := "before\n# Header\nafter"
	if result.Content != want {
		t.Errorf("Resolve() = %q, want %q", result.Content, want)
	}
	if !slices.Contains(result.Sources, "header.md") {
		t.Errorf("Resolve() sources = %v, want to contain header.md", result.Sources)
	}
}

func TestResolver_IncludeNested(t *testing.T) {
	r := NewResolver()
	fs := fakeFS{
		"outer.md":  "outer\n{{include \"inner.md\"}}",
		"inner.md":  "inner\n{{include \"leaf.md\"}}",
		"leaf.md":   "leaf content",
	}

	input := "start\n{{include \"outer.md\"}}\nend"
	result, err := r.Resolve(input, fs.Read)
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}

	want := "start\nouter\ninner\nleaf content\nend"
	if result.Content != want {
		t.Errorf("Resolve() = %q, want %q", result.Content, want)
	}

	if len(result.Sources) != 3 {
		t.Errorf("Resolve() sources = %v, want 3 entries", result.Sources)
	}
}

func TestResolver_IncludeCircular(t *testing.T) {
	r := NewResolver()
	fs := fakeFS{
		"a.md": "{{include \"b.md\"}}",
		"b.md": "{{include \"a.md\"}}",
	}

	input := "{{include \"a.md\"}}"
	result, err := r.Resolve(input, fs.Read)
	if err != nil {
		t.Fatalf("Resolve() should not error on circular include; got: %v", err)
	}

	// The resolver handles circular includes gracefully by inserting a
	// sentinel marker in the output.
	if !strings.Contains(result.Content, "circular include") {
		t.Errorf("Resolve() content = %q, want it to mention circular include", result.Content)
	}
}

func TestResolver_ContentFrom(t *testing.T) {
	r := NewResolver()
	fs := fakeFS{
		"data.yaml": "key: value\nnested:\n  sub: 42\n",
	}

	input := "config:\n{content from data.yaml}\n---"
	result, err := r.Resolve(input, fs.Read)
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}

	want := "config:\nkey: value\nnested:\n  sub: 42\n\n---"
	if result.Content != want {
		t.Errorf("Resolve() = %q, want %q", result.Content, want)
	}
	if !slices.Contains(result.Sources, "data.yaml") {
		t.Errorf("Resolve() sources = %v, want to contain data.yaml", result.Sources)
	}
}

func TestResolver_ContentFromNotFound(t *testing.T) {
	r := NewResolver()
	fs := fakeFS{}

	input := "load {content from missing.yaml}"
	result, err := r.Resolve(input, fs.Read)
	if err == nil {
		t.Fatal("Resolve() expected error for missing content-from target")
	}
	if !strings.Contains(err.Error(), "missing.yaml") {
		t.Errorf("Resolve() error = %v, want mention of missing.yaml", err)
	}
	// The unresolved placeholder remains in the output.
	if !strings.Contains(result.Content, "missing.yaml") {
		t.Errorf("Resolve() content = %q, want to contain missing.yaml", result.Content)
	}
}

func TestResolver_HashSubstitution(t *testing.T) {
	r := NewResolver()
	fs := fakeFS{}

	input := "content to hash\n{VERSION_HASH}"
	result, err := r.Resolve(input, fs.Read)
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}

	// The hash placeholder should be replaced with VERSION_HASH followed
	// by a hex digest (64 hex chars = SHA-256).
	if !strings.HasPrefix(result.Content, "content to hash\nVERSION_HASH") {
		t.Errorf("Resolve() content = %q, want VERSION_HASH prefix", result.Content)
	}

	// The total length should be "content to hash\n" + "VERSION_HASH" + 64 hex chars.
	wantLen := len("content to hash\nVERSION_HASH") + 64
	if len(result.Content) != wantLen {
		t.Errorf("Resolve() content length = %d, want %d", len(result.Content), wantLen)
	}
}

func TestResolver_IncludeNotFound(t *testing.T) {
	r := NewResolver()
	fs := fakeFS{}

	input := "{{include \"nonexistent.md\"}}"
	result, err := r.Resolve(input, fs.Read)
	if err != nil {
		t.Fatalf("Resolve() should not error on include error; got: %v", err)
	}

	if !strings.Contains(result.Content, "include error") {
		t.Errorf("Resolve() content = %q, want include error marker", result.Content)
	}
}

func TestResolver_MixedDirectives(t *testing.T) {
	r := NewResolver()
	fs := fakeFS{
		"header.yaml": "project: test",
		"body.txt":    "body content {CONTENT_HASH}",
	}

	input := "{{include \"header.yaml\"}}\n{content from body.txt}\nfooter"
	result, err := r.Resolve(input, fs.Read)
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}

	if !strings.Contains(result.Content, "project: test") {
		t.Errorf("Resolve() content missing header, got: %q", result.Content)
	}
	if !strings.Contains(result.Content, "body content") {
		t.Errorf("Resolve() content missing body content, got: %q", result.Content)
	}
	if !strings.Contains(result.Content, "CONTENT_HASH") {
		t.Errorf("Resolve() content missing CONTENT_HASH, got: %q", result.Content)
	}
	if !strings.Contains(result.Content, "footer") {
		t.Errorf("Resolve() content missing footer, got: %q", result.Content)
	}
	if len(result.Sources) != 2 {
		t.Errorf("Resolve() sources = %v, want 2 sources", result.Sources)
	}
}

func TestBuilder_NewPanicsOnNilResolver(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("New(nil) should panic")
		}
	}()
	New(nil)
}

func TestBuilder_Build(t *testing.T) {
	fs := fakeFS{
		"part.md": "included content\n",
	}

	input := "prefix\n{{include \"part.md\"}}\nsuffix"
	r := NewResolver()
	b := New(r)

	result, err := b.Build(input, fs.Read)
	if err != nil {
		t.Fatalf("Build() unexpected error: %v", err)
	}

	want := "prefix\nincluded content\n\nsuffix"
	if result.Content != want {
		t.Errorf("Build() = %q, want %q", result.Content, want)
	}
}

func TestBuilder_BuildNilRead(t *testing.T) {
	r := NewResolver()
	b := New(r)

	_, err := b.Build("content", nil)
	if err == nil {
		t.Fatal("Build() expected error for nil read function")
	}
	if !strings.Contains(err.Error(), "read function") {
		t.Errorf("Build() error = %v, want mention of read function", err)
	}
}

func TestBuilder_BuildDetectsErrors(t *testing.T) {
	fs := fakeFS{}

	r := NewResolver()
	b := New(r)

	// Include a file that doesn't exist — the resolver inserts an error
	// marker, and Build should detect it.
	_, err := b.Build("{{include \"missing.md\"}}", fs.Read)
	if err == nil {
		t.Fatal("Build() expected error for missing include")
	}
	if !strings.Contains(err.Error(), "include error") {
		t.Errorf("Build() error = %v, want include error", err)
	}
}

func TestBuilder_BuildDetectsCircularIncludes(t *testing.T) {
	fs := fakeFS{
		"a.md": "{{include \"b.md\"}}",
		"b.md": "{{include \"a.md\"}}",
	}

	r := NewResolver()
	b := New(r)

	_, err := b.Build("{{include \"a.md\"}}", fs.Read)
	if err == nil {
		t.Fatal("Build() expected error for circular include")
	}
	if !strings.Contains(err.Error(), "circular include") {
		t.Errorf("Build() error = %v, want circular include", err)
	}
}

func TestResolver_DeduplicatesSources(t *testing.T) {
	r := NewResolver()
	fs := fakeFS{
		"shared.yaml": "shared: true",
	}

	// Include the same file twice.
	input := "{{include \"shared.yaml\"}}\n{{include \"shared.yaml\"}}"
	result, err := r.Resolve(input, fs.Read)
	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}

	if len(result.Sources) != 1 {
		t.Errorf("Resolve() sources = %v, want exactly 1 entry (deduplicated)", result.Sources)
	}
}


