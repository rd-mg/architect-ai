// Package openspec provides runtime helpers for the OpenSpec persistence mode.
package openspec

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SentinelNoBase is the value of base_sha for a delta that creates a brand-new
// main spec (no prior file expected).
const SentinelNoBase = "0"

// DeltaFrontMatter is the schema embedded at the top of every delta spec.
type DeltaFrontMatter struct {
	Wrap DeltaWrap `yaml:"openspec_delta"`
}

type DeltaWrap struct {
	BaseSHA          string    `yaml:"base_sha"`
	BasePath         string    `yaml:"base_path"`
	BaseCapturedAt   time.Time `yaml:"base_captured_at"`
	Generator        string    `yaml:"generator"`
	GeneratorVersion int       `yaml:"generator_version"`
}

var (
	ErrNoFrontMatter    = errors.New("delta spec missing front-matter")
	ErrFrontMatterParse = errors.New("delta spec front-matter parse failed")
	ErrMergeConflict    = errors.New("merge conflict: base SHA mismatch")
	ErrUnexpectedExists = errors.New("merge conflict: new capability already exists")
)

// ConflictReport is returned when a conflict is detected.
type ConflictReport struct {
	Domain          string
	DeltaPath       string
	MainPath        string
	ExpectedSHA     string
	ActualSHA       string
	DeltaCapturedAt time.Time
	Reason          string // human-readable explanation
}

// SHAOfFile returns the hex SHA-256 of a file's bytes, or SentinelNoBase if
// the file does not exist.
func SHAOfFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return SentinelNoBase, nil
		}
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ReadDeltaFrontMatter parses the `---`-delimited YAML block at the start of
// a delta spec file. Returns the parsed front-matter + the body (content
// after the closing `---\n`).
func ReadDeltaFrontMatter(path string) (DeltaFrontMatter, string, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return DeltaFrontMatter{}, "", err
	}
	// Support both LF and CRLF. Must start with --- followed by newline.
	if !bytes.HasPrefix(bs, []byte("---\n")) && !bytes.HasPrefix(bs, []byte("---\r\n")) {
		return DeltaFrontMatter{}, "", ErrNoFrontMatter
	}
	
	// Find the closing --- on its own line. We look for \n---\n or \r\n---\r\n etc.
	// To be robust, we find the first occurrence of \n---\n after index 4.
	idx := bytes.Index(bs[4:], []byte("\n---\n"))
	if idx < 0 {
		// Try CRLF variant.
		idx = bytes.Index(bs[4:], []byte("\r\n---\r\n"))
		if idx < 0 {
			return DeltaFrontMatter{}, "", ErrNoFrontMatter
		}
	}
	
	yamlBlock := bs[4 : 4+idx]
	// body starts after the closing separator (e.g. \n---\n is 5 chars).
	// We skip the newline(s) to get to the actual content.
	bodyStart := 4 + idx + 5
	if bodyStart > len(bs) {
		bodyStart = len(bs)
	}
	body := string(bs[bodyStart:])

	var fm DeltaFrontMatter
	dec := yaml.NewDecoder(bytes.NewReader(yamlBlock))
	dec.KnownFields(true)
	if err := dec.Decode(&fm); err != nil {
		return DeltaFrontMatter{}, "", fmt.Errorf("%w: %v", ErrFrontMatterParse, err)
	}
	return fm, body, nil
}

// WriteDeltaFrontMatter prepends a YAML front-matter block to `body` and
// writes to `path` atomically.
func WriteDeltaFrontMatter(path string, fm DeltaFrontMatter, body string) error {
	out, err := yaml.Marshal(fm)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(out)
	buf.WriteString("---\n")
	buf.WriteString(body)
	
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// CheckConflict computes current main-spec SHA, compares to the delta's
// declared base, returns a *ConflictReport (nil if no conflict).
// `projectRoot` is the absolute path to the repo root.
// `changeName` and `domain` identify the delta.
func CheckConflict(projectRoot, changeName, domain string) (*ConflictReport, error) {
	deltaPath := filepath.Join(projectRoot, "openspec", "changes", changeName, "specs", domain, "spec.md")
	mainPath := filepath.Join(projectRoot, "openspec", "specs", domain, "spec.md")
	
	fm, _, err := ReadDeltaFrontMatter(deltaPath)
	if err != nil {
		return nil, err
	}
	
	currentSHA, err := SHAOfFile(mainPath)
	if err != nil {
		return nil, err
	}
	
	expected := strings.TrimSpace(fm.Wrap.BaseSHA)
	switch {
	case expected == SentinelNoBase && currentSHA != SentinelNoBase:
		return &ConflictReport{
			Domain:          domain,
			DeltaPath:       deltaPath,
			MainPath:        mainPath,
			ExpectedSHA:     "(new capability — no base expected)",
			ActualSHA:       currentSHA,
			DeltaCapturedAt: fm.Wrap.BaseCapturedAt,
			Reason:          "Delta was authored as a new capability, but a main spec for this domain already exists. Another change created it first.",
		}, ErrUnexpectedExists
	case expected != SentinelNoBase && expected != currentSHA:
		return &ConflictReport{
			Domain:          domain,
			DeltaPath:       deltaPath,
			MainPath:        mainPath,
			ExpectedSHA:     expected,
			ActualSHA:       currentSHA,
			DeltaCapturedAt: fm.Wrap.BaseCapturedAt,
			Reason:          "Main spec has changed since this delta was authored. Another change archived in the meantime.",
		}, ErrMergeConflict
	}
	return nil, nil
}

// WriteConflictReport renders a human-readable markdown conflict file into
// the change folder.
func WriteConflictReport(projectRoot, changeName string, reports []ConflictReport) error {
	path := filepath.Join(projectRoot, "openspec", "changes", changeName, "merge-conflict.md")
	var buf bytes.Buffer
	buf.WriteString("# Merge Conflict Report\n\n")
	fmt.Fprintf(&buf, "Change: `%s`\n", changeName)
	fmt.Fprintf(&buf, "Generated: %s\n\n", time.Now().UTC().Format(time.RFC3339))
	buf.WriteString("`sdd-archive` refused to merge the delta specs below because the main\nspec changed after this delta was authored. Resolve manually; see\n`docs/openspec-merge-conflict.md`.\n\n")
	
	for _, r := range reports {
		fmt.Fprintf(&buf, "## Domain `%s`\n\n", r.Domain)
		fmt.Fprintf(&buf, "- Delta file: `%s`\n", rel(projectRoot, r.DeltaPath))
		fmt.Fprintf(&buf, "- Main file:  `%s`\n", rel(projectRoot, r.MainPath))
		fmt.Fprintf(&buf, "- Expected base SHA: `%s`\n", r.ExpectedSHA)
		fmt.Fprintf(&buf, "- Actual  base SHA: `%s`\n", r.ActualSHA)
		fmt.Fprintf(&buf, "- Delta captured at: `%s`\n\n", r.DeltaCapturedAt.Format(time.RFC3339))
		fmt.Fprintf(&buf, "**%s**\n\n", r.Reason)
		buf.WriteString("### Recovery\n\n1. `git log openspec/specs/" + r.Domain + "/spec.md` — find what changed.\n2. Re-run `sdd-spec` for this delta to rebase onto current main.\n3. Or edit the delta manually and update `openspec_delta.base_sha` to the new SHA.\n4. Re-run `architect-ai sdd-archive-preflight " + changeName + "`.\n\n")
	}
	
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func rel(root, abs string) string {
	r, err := filepath.Rel(root, abs)
	if err != nil {
		return abs
	}
	return r
}
