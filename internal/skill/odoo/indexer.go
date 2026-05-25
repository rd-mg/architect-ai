package odoo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

// EngramClient abstracts mem_save operations
type EngramClient interface {
	SaveIdempotent(topicKey, content, project string) error
}

// IndexResult reports per-guide indexing outcome
type IndexResult struct {
	TopicKey string
	Action   string // "created" | "updated" | "skipped" | "failed"
	Bytes    int
	Error    error
}

// OdooIndexer indexes Odoo reference guides into Engram during skill-registry
type OdooIndexer struct {
	OverlayDir  string
	OdooVersion string
	Engram      EngramClient
	Workers     int // goroutine pool size
}

// IndexAll indexes all Odoo guides with goroutine parallelism
// Uses idempotent SaveIdempotent — same topic_key = UPDATE not INSERT
func (idx *OdooIndexer) IndexAll(project string) ([]IndexResult, error) {
	guides, err := idx.collectGuides()
	if err != nil {
		return nil, fmt.Errorf("collect guides: %w", err)
	}

	results := make([]IndexResult, 0, len(guides))
	var mu sync.Mutex

	g := new(errgroup.Group)
	sem := make(chan struct{}, idx.workers())

	for key, path := range guides {
		key, path := key, path
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			content, err := os.ReadFile(path)
			if err != nil {
				mu.Lock()
				results = append(results, IndexResult{
					TopicKey: key, Action: "failed", Error: err,
				})
				mu.Unlock()
				return nil // non-fatal — continue other guides
			}

			saveErr := idx.Engram.SaveIdempotent(key, string(content), project)
			action := "updated"
			if saveErr != nil {
				action = "failed"
			}

			mu.Lock()
			results = append(results, IndexResult{
				TopicKey: key,
				Action:   action,
				Bytes:    len(content),
				Error:    saveErr,
			})
			mu.Unlock()
			return nil
		})
	}

	return results, g.Wait()
}

// collectGuides builds map of topic_key → file_path for all Odoo reference guides
func (idx *OdooIndexer) collectGuides() (map[string]string, error) {
	guides := make(map[string]string)
	ver := idx.OdooVersion

	dirs := map[string]string{
		filepath.Join(idx.OverlayDir, "skills", fmt.Sprintf("odoo-%s.0", ver)):   fmt.Sprintf("knowledge/odoo-%s/reference", ver),
		filepath.Join(idx.OverlayDir, "skills", fmt.Sprintf("patterns-%s", ver)): fmt.Sprintf("knowledge/odoo-%s/patterns", ver),
		filepath.Join(idx.OverlayDir, "skills", "patterns-agnostic"):             "knowledge/odoo-agnostic/reference",
		filepath.Join(idx.OverlayDir, "skills", "patterns-ddd"):                  "knowledge/odoo-agnostic/ddd",
	}

	// Migration guides (previous version → current)
	prevVer := fmt.Sprintf("%d", versionInt(ver)-1)
	migDir := filepath.Join(idx.OverlayDir, "skills", fmt.Sprintf("migration-%s-%s", prevVer, ver))
	dirs[migDir] = fmt.Sprintf("knowledge/odoo-migration/%s-%s", prevVer, ver)

	for dir, prefix := range dirs {
		walkMDs(dir, prefix, guides) // non-fatal per-dir
	}
	return guides, nil
}

func walkMDs(dir, prefix string, out map[string]string) {
	filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error { //nolint
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		slug := strings.TrimSuffix(d.Name(), ".md")
		out[prefix+"/"+slug] = p
		return nil
	})
}

func (idx *OdooIndexer) workers() int {
	if idx.Workers > 0 {
		return idx.Workers
	}
	return 4 // default goroutine pool
}

func versionInt(s string) int {
	var n int
	_, _ = fmt.Sscan(s, &n)
	return n
}
