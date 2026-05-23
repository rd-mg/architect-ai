# Design: Phase 5 - Engram Context Guardian v3: ByteRover + Branch A/B + Skill Tiers

## Architecture
Engram integration is a two-layer system: (1) tool tiering controls which `mem_*` functions each agent can call, enforced via platform permissions, and (2) Context Guardian v3 acts as a proactive monitor that detects context pressure and triggers compression before the model degrades.

Memory retrieval uses the ByteRover Hierarchical Context Pattern (L1: Working, L2: Episodic, L3: Semantic, L4: Archive).
The 3-tool Poka-Yoke (search/save/get_observation) is MAINTAINED. The expansion is **additive selective** — only tools with clear semantic value at each level are added. `mem_suggest_topic_key` is the most important addition: it prevents the key drift problem via explicit collision handling.

context-mode and Engram are complementary, NOT competing: context-mode = session buffer (Branch B), Engram = durable architecture memory (Branch A). Policy: NEVER use `ctx_index` as substitute for `mem_save`, NEVER use `ctx_search` as substitute for `mem_search`.

## FODA Matrix

| | Detail |
|---|---|
| **F** | 3-tool Poka-Yoke prevents hallucinations from truncated fragments. context-guardian v2 has auto-trigger at 50% tokens. |
| **O** | `mem_suggest_topic_key` prevents key drift without breaking Poka-Yoke. Compress fallback makes context-guardian fully autonomous. |
| **D** | Without `mem_suggest_topic_key`, agents save with arbitrary keys that fragment taxonomy. Without compress fallback, context-guardian depends on user configuring hook manually. |
| **A** | Adding too many tools to L2 subagents can confuse the model and waste tokens. Maintain minimum sufficiency principle. |

## FMEA Matrix
| Component | Failure Mode | Effect | Likelihood | Severity | RPN | Mitigation |
|---|---|---|---|---|---|---|
| Engram MCP | Server unavailable | Branch A persistence fails | 1 | 4 | 4 | Branch B fallback (context-mode) activates automatically. |
| Context Guardian | Pressure detection too late | Model already degraded before compress | 2 | 3 | 6 | Conservative threshold (50%). Trigger early via Branch A. |
| Proactive Save | Agent forgets to save | Decision/discovery lost between sessions | 2 | 3 | 6 | Protocol-level mandate. Self-check after every task. |
| topic_key Collision | Two topics use same key | One overwrites the other | 1 | 3 | 3 | `mem_suggest_topic_key` protocol with similarity handling (>0.85 = update). |
| Odoo Indexing | Skill registry run twice | Duplicate Odoo guides in Engram | 1 | 2 | 2 | Idempotent indexing (`SaveIdempotent`) via Go indexer. |

## Go Implementation — Idempotent Indexing (Odoo)

### `internal/skill/odoo/indexer.go`

```go
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
        filepath.Join(idx.OverlayDir, "skills", fmt.Sprintf("odoo-%s.0", ver)):      fmt.Sprintf("knowledge/odoo-%s/reference", ver),
        filepath.Join(idx.OverlayDir, "skills", fmt.Sprintf("patterns-%s", ver)):    fmt.Sprintf("knowledge/odoo-%s/patterns", ver),
        filepath.Join(idx.OverlayDir, "skills", "patterns-agnostic"):                "knowledge/odoo-agnostic/reference",
        filepath.Join(idx.OverlayDir, "skills", "patterns-ddd"):                     "knowledge/odoo-agnostic/ddd",
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
    fmt.Sscan(s, &n)
    return n
}
```

### `internal/skill/odoo/indexer_test.go`

```go
package odoo

import (
    "errors"
    "os"
    "path/filepath"
    "sync"
    "testing"
)

// mockEngram implements EngramClient for testing
type mockEngram struct {
    mu      sync.Mutex
    saves   map[string]string
    failKey string // simulate failure for this key
}

func (m *mockEngram) SaveIdempotent(key, content, project string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    if key == m.failKey {
        return errors.New("simulated Engram failure")
    }
    if m.saves == nil {
        m.saves = make(map[string]string)
    }
    m.saves[key] = content
    return nil
}

// ... Additional tests ...
```

## Contracts
- `DetectCompressCapability(platform, projectDir)` → `CompressCapability`.
- `ResolveAction(cap)` → `CompressAction` (`"hook"` | `"native_compress"` | `"manual_summary"`).
- Priority order: Hook > Native > Manual.
- Engram tier map: see Tool Distribution Table in spec.

## Key Decisions
- **50% threshold**: Context pressure detection at 50% estimated usage, not 80%. Better to compress early than have a degraded model.
- **Mandatory session summary**: Not "recommended" — MANDATORY. Skipping means next session starts blind.
- **topic_key for upserts**: Evolving topics use stable keys so each topic has exactly ONE observation.
- **context-mode maintained**: It solves a different problem (session anti-flooding) than Engram (persistent memory). Both are needed.
- **Hook > Native > Manual**: Custom hooks give most control; native compress is good enough; manual summary is last resort but still works.
