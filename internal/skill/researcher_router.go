package skill

type ResearchResult struct {
	Status      string
	Source      string
	Summary     string
	KeyFindings []string
	Evidence    []string
	Confidence  string
}

type ResearchSource interface {
	Name() string
	Search(query string) (*ResearchResult, error)
}

type ResearchRouter struct {
	Sources []ResearchSource
}

func (r *ResearchRouter) Research(query string) (*ResearchResult, error) {
	for _, source := range r.Sources {
		res, err := source.Search(query)
		if err == nil && res != nil && res.Status == "found" {
			return res, nil
		}
	}
	return &ResearchResult{
		Status:     "not_found",
		Confidence: "low",
		Summary:    "All sources missed.",
	}, nil
}
