package filemerge

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type WriteResult struct {
	Changed bool
	Created bool
}

// WriteOptions configures the behavior of an atomic write operation.
type WriteOptions struct {
	// Perm is the file mode to apply. Defaults to 0644.
	Perm fs.FileMode
	// Force specifies whether to overwrite the file even if the
	// content is identical to what is already on disk.
	Force bool
}

// WriteFileAtomic writes content to path atomically. It skips the write if the
// file already exists and has identical content.
func WriteFileAtomic(path string, content []byte, perm fs.FileMode) (WriteResult, error) {
	return WriteFileAtomicWithOptions(path, content, WriteOptions{Perm: perm})
}

// WriteFileAtomicWithOptions writes content to path atomically using the
// provided options.
func WriteFileAtomicWithOptions(path string, content []byte, opts WriteOptions) (WriteResult, error) {
	perm := opts.Perm
	if perm == 0 {
		perm = 0o644
	}

	created := false
	existing, err := os.ReadFile(path)
	if err == nil {
		if !opts.Force && bytes.Equal(existing, content) {
			return WriteResult{}, nil
		}
	} else if !os.IsNotExist(err) {
		return WriteResult{}, fmt.Errorf("read existing file %q: %w", path, err)
	} else {
		created = true
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return WriteResult{}, fmt.Errorf("create parent directories for %q: %w", path, err)
	}
	// Ensure the directory is writable — it may have been created with
	// restricted permissions (e.g. 555) by a previous installer version or
	// the target agent itself. MkdirAll succeeds on existing dirs but does
	// not fix their permissions, causing os.CreateTemp to fail below.
	if err := os.Chmod(dir, 0o755); err != nil {
		return WriteResult{}, fmt.Errorf("set write permission on directory for %q: %w", path, err)
	}

	tmp, err := os.CreateTemp(dir, ".architect-ai-*.tmp")
	if err != nil {
		return WriteResult{}, fmt.Errorf("create temp file for %q: %w", path, err)
	}

	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return WriteResult{}, fmt.Errorf("write temp file for %q: %w", path, err)
	}

	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return WriteResult{}, fmt.Errorf("set permissions on temp file for %q: %w", path, err)
	}

	if err := tmp.Close(); err != nil {
		return WriteResult{}, fmt.Errorf("close temp file for %q: %w", path, err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return WriteResult{}, fmt.Errorf("replace %q atomically: %w", path, err)
	}

	cleanup = false
	return WriteResult{Changed: true, Created: created}, nil
}
