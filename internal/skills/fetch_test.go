package skills_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rd-mg/architect-ai/internal/skills"
)

func newTestServer(fn http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func TestFetchSkillMD_Success(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent header")
		}
		w.Header().Set("X-GitHub-SHA", "abc123")
		w.Write([]byte("# Test Skill\n\nsome content"))
	})
	defer srv.Close()

	skills.SetHTTPClient(&http.Client{Transport: &staticTransport{srv.URL}})
	t.Cleanup(func() { skills.SetHTTPClient(http.DefaultClient) })

	result, err := skills.FetchSkillMD(context.Background(), "owner", "repo", "skills/test")
	if err != nil {
		t.Fatalf("FetchSkillMD: %v", err)
	}
	if len(result.Content) == 0 {
		t.Error("expected non-empty Content")
	}
	if result.SHA != "abc123" {
		t.Errorf("SHA = %q, want %q", result.SHA, "abc123")
	}
}

func TestFetchSkillMD_NotFound(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer srv.Close()

	skills.SetHTTPClient(&http.Client{Transport: &staticTransport{srv.URL}})
	t.Cleanup(func() { skills.SetHTTPClient(http.DefaultClient) })

	_, err := skills.FetchSkillMD(context.Background(), "owner", "repo", "missing")
	if err != skills.ErrSkillNotFound {
		t.Errorf("expected ErrSkillNotFound, got %v", err)
	}
}

func TestFetchSkillMD_RateLimited(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	defer srv.Close()

	skills.SetHTTPClient(&http.Client{Transport: &staticTransport{srv.URL}})
	t.Cleanup(func() { skills.SetHTTPClient(http.DefaultClient) })

	_, err := skills.FetchSkillMD(context.Background(), "owner", "repo", "path")
	if err != skills.ErrRateLimited {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
}

func TestFetchSkillMD_TooManyRequests(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	defer srv.Close()

	skills.SetHTTPClient(&http.Client{Transport: &staticTransport{srv.URL}})
	t.Cleanup(func() { skills.SetHTTPClient(http.DefaultClient) })

	_, err := skills.FetchSkillMD(context.Background(), "owner", "repo", "path")
	if err != skills.ErrRateLimited {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
}

func TestFetchSkillMD_ServerError(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()

	skills.SetHTTPClient(&http.Client{Transport: &staticTransport{srv.URL}})
	t.Cleanup(func() { skills.SetHTTPClient(http.DefaultClient) })

	_, err := skills.FetchSkillMD(context.Background(), "owner", "repo", "path")
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestFetchSkillMD_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := skills.FetchSkillMD(ctx, "owner", "repo", "path")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestListSkillPathsInRepo_Success(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Errorf("Accept = %q, want application/vnd.github+json", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"tree": [
				{"path": "skills/gimp/SKILL.md", "type": "blob"},
				{"path": "skills/vim/SKILL.md", "type": "blob"},
				{"path": "skills/gimp/helper.sh", "type": "blob"},
				{"path": "README.md", "type": "blob"},
				{"path": "skills/gimp", "type": "tree"}
			]
		}`))
	})
	defer srv.Close()

	skills.SetHTTPClient(&http.Client{Transport: &staticTransport{srv.URL}})
	t.Cleanup(func() { skills.SetHTTPClient(http.DefaultClient) })

	paths, err := skills.ListSkillPathsInRepo(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("ListSkillPathsInRepo: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d: %v", len(paths), paths)
	}
	seen := map[string]bool{}
	for _, p := range paths {
		seen[p] = true
	}
	if !seen["skills/gimp"] {
		t.Errorf("missing skills/gimp, got %v", paths)
	}
	if !seen["skills/vim"] {
		t.Errorf("missing skills/vim, got %v", paths)
	}
}

func TestListSkillPathsInRepo_RootSKILLMD(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"tree": [
				{"path": "SKILL.md", "type": "blob"},
				{"path": "skills/gimp/SKILL.md", "type": "blob"}
			]
		}`))
	})
	defer srv.Close()

	skills.SetHTTPClient(&http.Client{Transport: &staticTransport{srv.URL}})
	t.Cleanup(func() { skills.SetHTTPClient(http.DefaultClient) })

	paths, err := skills.ListSkillPathsInRepo(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("ListSkillPathsInRepo: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d: %v", len(paths), paths)
	}
	foundRoot := false
	for _, p := range paths {
		if p == "" {
			foundRoot = true
			break
		}
	}
	if !foundRoot {
		t.Errorf("expected empty string for root SKILL.md, got %v", paths)
	}
}

func TestListSkillPathsInRepo_NoMatches(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tree": [{"path": "README.md", "type": "blob"}]}`))
	})
	defer srv.Close()

	skills.SetHTTPClient(&http.Client{Transport: &staticTransport{srv.URL}})
	t.Cleanup(func() { skills.SetHTTPClient(http.DefaultClient) })

	paths, err := skills.ListSkillPathsInRepo(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("ListSkillPathsInRepo: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(paths))
	}
}

func TestListSkillPathsInRepo_APIFailure(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()

	skills.SetHTTPClient(&http.Client{Transport: &staticTransport{srv.URL}})
	t.Cleanup(func() { skills.SetHTTPClient(http.DefaultClient) })

	_, err := skills.ListSkillPathsInRepo(context.Background(), "owner", "repo")
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

type staticTransport struct {
	baseURL string
}

func (t *staticTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rewritten, err := http.NewRequest(req.Method, t.baseURL, req.Body)
	if err != nil {
		return nil, err
	}
	rewritten.Header = req.Header
	rewritten.ContentLength = req.ContentLength
	return http.DefaultTransport.RoundTrip(rewritten)
}
