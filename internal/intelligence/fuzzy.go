package intelligence

import (
	"sort"
	"strings"
)

type FuzzyMatcher struct {
	caseSensitive bool
}

func NewFuzzyMatcher() *FuzzyMatcher {
	return &FuzzyMatcher{
		caseSensitive: false,
	}
}

type MatchResult struct {
	Score     float64 `json:"score"`
	Highlight string  `json:"highlight"`
}

type ScoredCompletion struct {
	Completion  Completion
	FuzzyScore  float64
	MatchResult *MatchResult
}

func (fm *FuzzyMatcher) FilterAndRank(query string, completions []Completion) []Completion {
	if query == "" {
		return completions
	}

	scored := make([]ScoredCompletion, 0)
	for _, completion := range completions {
		if result := fm.Match(query, completion.Text); result.Score > 0 {
			scored = append(scored, ScoredCompletion{
				Completion:  completion,
				FuzzyScore:  result.Score,
				MatchResult: result,
			})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		scoreI := scored[i].Completion.Score + scored[i].FuzzyScore*50
		scoreJ := scored[j].Completion.Score + scored[j].FuzzyScore*50
		return scoreI > scoreJ
	})

	result := make([]Completion, len(scored))
	for i, sc := range scored {
		completion := sc.Completion
		completion.Score = sc.Completion.Score + sc.FuzzyScore*50
		if completion.Metadata == nil {
			completion.Metadata = make(map[string]any)
		}
		completion.Metadata["fuzzy_score"] = sc.FuzzyScore
		completion.Metadata["highlight"] = sc.MatchResult.Highlight
		result[i] = completion
	}

	return result
}

func (fm *FuzzyMatcher) Match(query, target string) *MatchResult {
	if query == "" {
		return &MatchResult{Score: 100.0, Highlight: target}
	}

	if strings.EqualFold(query, target) {
		return &MatchResult{Score: 100.0, Highlight: target}
	}

	if strings.HasPrefix(strings.ToLower(target), strings.ToLower(query)) {
		return &MatchResult{
			Score:     90.0 - float64(len(target)-len(query)),
			Highlight: target,
		}
	}

	if strings.Contains(strings.ToLower(target), strings.ToLower(query)) {
		return &MatchResult{
			Score:     70.0 - float64(len(target)-len(query)),
			Highlight: target,
		}
	}

	return &MatchResult{Score: 0.0, Highlight: target}
}

func (fm *FuzzyMatcher) FindBestMatches(query string, candidates []string, maxResults int) []string {
	if query == "" {
		if len(candidates) <= maxResults {
			return candidates
		}
		return candidates[:maxResults]
	}

	type ScoredCandidate struct {
		Text  string
		Score float64
	}

	scored := make([]ScoredCandidate, 0)
	for _, candidate := range candidates {
		if result := fm.Match(query, candidate); result.Score > 0 {
			scored = append(scored, ScoredCandidate{
				Text:  candidate,
				Score: result.Score,
			})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	results := make([]string, 0)
	for i := 0; i < len(scored) && i < maxResults; i++ {
		results = append(results, scored[i].Text)
	}

	return results
}
