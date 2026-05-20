package skill

import (
	"testing"
)

type MockSource struct {
	name       string
	shouldFind bool
	result     *ResearchResult
	called     bool
}

func (m *MockSource) Name() string {
	return m.name
}

func (m *MockSource) Search(query string) (*ResearchResult, error) {
	m.called = true
	if m.shouldFind {
		return m.result, nil
	}
	return &ResearchResult{Status: "not_found"}, nil
}

func TestResearcherRouter_EngramHit(t *testing.T) {
	engramRes := &ResearchResult{
		Status:     "found",
		Source:     "engram",
		Summary:    "Found in Engram",
		Confidence: "high",
	}

	engramMock := &MockSource{name: "engram", shouldFind: true, result: engramRes}
	rgMock := &MockSource{name: "ripgrep", shouldFind: false}
	c7Mock := &MockSource{name: "context7", shouldFind: false}

	router := &ResearchRouter{
		Sources: []ResearchSource{engramMock, rgMock, c7Mock},
	}

	res, err := router.Research("auth middleware")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Status != "found" || res.Source != "engram" {
		t.Errorf("expected engram hit, got %v", res)
	}

	if !engramMock.called {
		t.Error("expected engram mock to be called")
	}

	if rgMock.called {
		t.Error("expected ripgrep mock NOT to be called when engram hits")
	}
}

func TestResearcherRouter_RgFallback(t *testing.T) {
	rgRes := &ResearchResult{
		Status:     "found",
		Source:     "local_codebase",
		Summary:    "Found via ripgrep",
		Confidence: "medium",
	}

	engramMock := &MockSource{name: "engram", shouldFind: false}
	rgMock := &MockSource{name: "ripgrep", shouldFind: true, result: rgRes}
	c7Mock := &MockSource{name: "context7", shouldFind: false}

	router := &ResearchRouter{
		Sources: []ResearchSource{engramMock, rgMock, c7Mock},
	}

	res, err := router.Research("auth middleware")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Status != "found" || res.Source != "local_codebase" {
		t.Errorf("expected ripgrep hit, got %v", res)
	}

	if !engramMock.called {
		t.Error("expected engram mock to be called")
	}

	if !rgMock.called {
		t.Error("expected ripgrep mock to be called")
	}

	if c7Mock.called {
		t.Error("expected context7 mock NOT to be called when ripgrep hits")
	}
}

func TestResearcherRouter_FullChainMiss(t *testing.T) {
	engramMock := &MockSource{name: "engram", shouldFind: false}
	rgMock := &MockSource{name: "ripgrep", shouldFind: false}
	c7Mock := &MockSource{name: "context7", shouldFind: false}

	router := &ResearchRouter{
		Sources: []ResearchSource{engramMock, rgMock, c7Mock},
	}

	res, err := router.Research("auth middleware")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Status != "not_found" {
		t.Errorf("expected status 'not_found', got %s", res.Status)
	}

	if !engramMock.called || !rgMock.called || !c7Mock.called {
		t.Error("expected all sources to be searched on full chain miss")
	}
}
