package skills

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"
)

var (
	ErrSkillNotFound = errors.New("SKILL.md not found at the specified path")
	ErrRateLimited   = errors.New("GitHub API rate limited (60 req/hour unauthenticated)")
)

const (
	rawURL    = "https://raw.githubusercontent.com/%s/%s/HEAD/%s/SKILL.md"
	treeURL   = "https://api.github.com/repos/%s/%s/git/trees/HEAD?recursive=1"
	userAgent = "architect-ai (https://github.com/rd-mg/architect-ai)"
	timeout   = 15 * time.Second
)

type FetchResult struct {
	Content []byte
	SHA     string // best-effort; empty if GitHub does not return it
}

// FetchSkillMD fetches SKILL.md from github.com/{owner}/{repo}/HEAD/{path}/SKILL.md.
// path may be empty (repo root) or a subdirectory path (e.g. "skills/gimp").
func FetchSkillMD(ctx context.Context, owner, repo, path string) (FetchResult, error) {
	url := fmt.Sprintf(rawURL, owner, repo, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return FetchResult{}, err
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return FetchResult{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return FetchResult{}, ErrSkillNotFound
	case http.StatusForbidden, http.StatusTooManyRequests:
		return FetchResult{}, ErrRateLimited
	default:
		return FetchResult{}, fmt.Errorf("GitHub returned HTTP %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB cap
	if err != nil {
		return FetchResult{}, err
	}
	return FetchResult{Content: body, SHA: resp.Header.Get("X-GitHub-SHA")}, nil
}

// ListSkillPathsInRepo returns the parent dir of every SKILL.md in the repo tree.
func ListSkillPathsInRepo(ctx context.Context, owner, repo string) ([]string, error) {
	url := fmt.Sprintf(treeURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tree API returned HTTP %d", resp.StatusCode)
	}

	var tree struct {
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
		} `json:"tree"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var paths []string
	for _, item := range tree.Tree {
		if item.Type == "blob" && filepath.Base(item.Path) == "SKILL.md" {
			dir := filepath.Dir(item.Path)
			if dir == "." {
				dir = ""
			}
			if !seen[dir] {
				seen[dir] = true
				paths = append(paths, dir)
			}
		}
	}
	return paths, nil
}
