package openspec

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type HybridWriteStatus string

const (
	HybridWriteCompleted HybridWriteStatus = "completed"
	HybridWritePartial   HybridWriteStatus = "partially_completed"
	HybridWriteBlocked   HybridWriteStatus = "blocked"
)

type HybridWriteResult struct {
	Status        HybridWriteStatus
	EngramOK      bool
	FilesystemOK  bool
	EngramErr     error
	FsErr         error
	BlockedReason string
}

type HybridArtifact struct {
	TopicKey     string
	FilePath     string
	Content      string
	Title        string
	ArtifactType string
}

type EngramWriteFn func(ctx context.Context, topicKey, title, artifactType, content string) error

type FilesystemWriteFn func(ctx context.Context, path, content string) error

type HybridWriter struct {
	EngramFn     EngramWriteFn
	FilesystemFn FilesystemWriteFn
	Timeout      time.Duration
}

func (hw *HybridWriter) Write(ctx context.Context, artifact HybridArtifact) HybridWriteResult {
	if artifact.TopicKey == "" {
		return HybridWriteResult{
			Status:        HybridWriteBlocked,
			BlockedReason: "hybrid_write_precondition: TopicKey must not be empty",
		}
	}
	if artifact.FilePath == "" {
		return HybridWriteResult{
			Status:        HybridWriteBlocked,
			BlockedReason: "hybrid_write_precondition: FilePath must not be empty",
		}
	}
	if artifact.Content == "" {
		return HybridWriteResult{
			Status:        HybridWriteBlocked,
			BlockedReason: "hybrid_write_precondition: Content must not be empty",
		}
	}

	timeout := hw.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	writeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		engramErr error
		fsErr     error
		wg        sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		engramErr = hw.EngramFn(writeCtx,
			artifact.TopicKey,
			artifact.Title,
			artifact.ArtifactType,
			artifact.Content,
		)
	}()

	go func() {
		defer wg.Done()
		fsErr = hw.FilesystemFn(writeCtx, artifact.FilePath, artifact.Content)
	}()

	wg.Wait()

	engramOK := engramErr == nil
	fsOK := fsErr == nil

	switch {
	case engramOK && fsOK:
		return HybridWriteResult{
			Status:       HybridWriteCompleted,
			EngramOK:     true,
			FilesystemOK: true,
		}
	case engramOK && !fsOK:
		return HybridWriteResult{
			Status:        HybridWritePartial,
			EngramOK:      true,
			FilesystemOK:  false,
			FsErr:         fsErr,
			BlockedReason: fmt.Sprintf("filesystem_unavailable: %v", fsErr),
		}
	case !engramOK && fsOK:
		return HybridWriteResult{
			Status:        HybridWritePartial,
			EngramOK:      false,
			FilesystemOK:  true,
			EngramErr:     engramErr,
			BlockedReason: fmt.Sprintf("engram_unavailable: %v", engramErr),
		}
	default:
		return HybridWriteResult{
			Status:        HybridWriteBlocked,
			EngramOK:      false,
			FilesystemOK:  false,
			EngramErr:     engramErr,
			FsErr:         fsErr,
			BlockedReason: fmt.Sprintf("both_stores_failed: engram=%v fs=%v", engramErr, fsErr),
		}
	}
}

func (r HybridWriteResult) IsComplete() bool {
	return r.Status == HybridWriteCompleted
}

func (r HybridWriteResult) ResultContractStatus() string {
	switch r.Status {
	case HybridWriteCompleted:
		return "completed"
	case HybridWritePartial:
		return "partially_completed"
	default:
		return "blocked"
	}
}

func AtomicFilesystemWrite(ctx context.Context, path, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent dirs for %s: %w", path, err)
	}

	tmp := fmt.Sprintf("%s.%d.%d.tmp", path, os.Getpid(), time.Now().UnixNano())
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write tmp %s: %w", tmp, err)
	}

	if f, err := os.OpenFile(tmp, os.O_RDWR, 0o644); err == nil {
		_ = f.Sync()
		_ = f.Close()
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename %s → %s: %w", tmp, path, err)
	}
	return nil
}
