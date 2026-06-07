package github

import "testing"

func TestMergeRepoContributions(t *testing.T) {
	pr := []repoContribGroup{{
		Repository: struct {
			NameWithOwner string `json:"nameWithOwner"`
			URL           string `json:"url"`
		}{NameWithOwner: "ossf/scorecard", URL: "https://github.com/ossf/scorecard"},
		Contributions: struct {
			TotalCount int `json:"totalCount"`
		}{TotalCount: 12},
	}}
	issues := []repoContribGroup{{
		Repository: struct {
			NameWithOwner string `json:"nameWithOwner"`
			URL           string `json:"url"`
		}{NameWithOwner: "ossf/scorecard", URL: "https://github.com/ossf/scorecard"},
		Contributions: struct {
			TotalCount int `json:"totalCount"`
		}{TotalCount: 3},
	}}
	reviews := []repoContribGroup{{
		Repository: struct {
			NameWithOwner string `json:"nameWithOwner"`
			URL           string `json:"url"`
		}{NameWithOwner: "kubernetes/kubernetes", URL: "https://github.com/kubernetes/kubernetes"},
		Contributions: struct {
			TotalCount int `json:"totalCount"`
		}{TotalCount: 5},
	}}

	out := mergeRepoContributions(pr, issues, reviews)
	if len(out) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(out))
	}

	byName := map[string]RepoContribution{}
	for _, r := range out {
		byName[r.FullName] = r
	}
	sc := byName["ossf/scorecard"]
	if sc.PRs != 12 || sc.Issues != 3 || sc.Reviews != 0 {
		t.Fatalf("unexpected scorecard stats: %+v", sc)
	}
	k8s := byName["kubernetes/kubernetes"]
	if k8s.Reviews != 5 {
		t.Fatalf("unexpected k8s stats: %+v", k8s)
	}
}
