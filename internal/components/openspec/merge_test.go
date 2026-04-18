package openspec

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeDelta(t *testing.T, projectRoot, changeName, domain, baseSHA, body string) string {
	t.Helper()
	dir := filepath.Join(projectRoot, "openspec", "changes", changeName, "specs", domain)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "spec.md")
	fm := DeltaFrontMatter{Wrap: DeltaWrap{
		BaseSHA:          baseSHA,
		BasePath:         filepath.Join("openspec", "specs", domain, "spec.md"),
		BaseCapturedAt:   time.Now().UTC(),
		Generator:        "sdd-spec",
		GeneratorVersion: 1,
	}}
	if err := WriteDeltaFrontMatter(path, fm, body); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeMain(t *testing.T, projectRoot, domain, content string) {
	t.Helper()
	dir := filepath.Join(projectRoot, "openspec", "specs", domain)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCheckConflict_NoConflict(t *testing.T) {
	root := t.TempDir()
	writeMain(t, root, "sale", "# Sale spec\nbase content\n")
	sha, _ := SHAOfFile(filepath.Join(root, "openspec", "specs", "sale", "spec.md"))
	writeDelta(t, root, "add-export", "sale", sha, "# delta body\n")

	report, err := CheckConflict(root, "add-export", "sale")
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	if report != nil {
		t.Fatalf("want nil report, got %+v", report)
	}
}

func TestCheckConflict_MainChanged(t *testing.T) {
	root := t.TempDir()
	writeMain(t, root, "sale", "# Sale spec\noriginal content\n")
	sha, _ := SHAOfFile(filepath.Join(root, "openspec", "specs", "sale", "spec.md"))
	writeDelta(t, root, "add-export", "sale", sha, "# delta body\n")
	// Someone else modified main after delta was written.
	writeMain(t, root, "sale", "# Sale spec\nMODIFIED content\n")

	report, err := CheckConflict(root, "add-export", "sale")
	if !errors.Is(err, ErrMergeConflict) {
		t.Fatalf("want ErrMergeConflict, got %v", err)
	}
	if report == nil {
		t.Fatal("want report, got nil")
	}
	if report.Domain != "sale" {
		t.Errorf("domain: got %q", report.Domain)
	}
}

func TestCheckConflict_NewCapability_NoMain(t *testing.T) {
	root := t.TempDir()
	writeDelta(t, root, "add-export", "export", SentinelNoBase, "# new capability\n")
	report, err := CheckConflict(root, "add-export", "export")
	if err != nil {
		t.Fatalf("want nil, got %v", err)
	}
	if report != nil {
		t.Fatalf("want nil report, got %+v", report)
	}
}

func TestCheckConflict_NewCapability_MainExists(t *testing.T) {
	root := t.TempDir()
	writeMain(t, root, "export", "# already exists\n")
	writeDelta(t, root, "add-export", "export", SentinelNoBase, "# new capability\n")
	report, err := CheckConflict(root, "add-export", "export")
	if !errors.Is(err, ErrUnexpectedExists) {
		t.Fatalf("want ErrUnexpectedExists, got %v", err)
	}
	if report == nil || report.Domain != "export" {
		t.Fatalf("bad report: %+v", report)
	}
}

func TestReadWriteRoundTrip(t *testing.T) {
	root := t.TempDir()
	path := writeDelta(t, root, "add-export", "sale", "abcdef", "# hello body\n\nthis is content.\n")
	fm, body, err := ReadDeltaFrontMatter(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if fm.Wrap.BaseSHA != "abcdef" {
		t.Errorf("base_sha round-trip: got %q", fm.Wrap.BaseSHA)
	}
	if body != "# hello body\n\nthis is content.\n" {
		t.Errorf("body round-trip lost data: %q", body)
	}
}

func TestReadMissingFrontMatter(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "no-fm.md")
	if err := os.WriteFile(p, []byte("# just a doc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadDeltaFrontMatter(p); !errors.Is(err, ErrNoFrontMatter) {
		t.Fatalf("want ErrNoFrontMatter, got %v", err)
	}
}

func TestWriteConflictReport_Emits(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "openspec", "changes", "add-export"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := WriteConflictReport(root, "add-export", []ConflictReport{{
		Domain:          "sale",
		DeltaPath:       filepath.Join(root, "openspec", "changes", "add-export", "specs", "sale", "spec.md"),
		MainPath:        filepath.Join(root, "openspec", "specs", "sale", "spec.md"),
		ExpectedSHA:     "abc",
		ActualSHA:       "def",
		DeltaCapturedAt: time.Now().UTC(),
		Reason:          "Main changed",
	}})
	if err != nil {
		t.Fatal(err)
	}
	bs, err := os.ReadFile(filepath.Join(root, "openspec", "changes", "add-export", "merge-conflict.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(bs)
	if !strings.Contains(content, "Domain `sale`") || !strings.Contains(content, "`abc`") {
		t.Errorf("report missing expected fields:\n%s", content)
	}
}
