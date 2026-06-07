package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const userStatsQuery = `
query($login: String!, $from: DateTime!, $to: DateTime!) {
  user(login: $login) {
    followers { totalCount }
    organizations(first: 1) { totalCount }
    repositories(ownerAffiliations: OWNER, first: 100, orderBy: {field: STARGAZERS, direction: DESC}) {
      nodes {
        isFork
        stargazerCount
        description
        releases { totalCount }
        object(expression: "HEAD:README.md") {
          ... on Blob { text }
        }
      }
    }
    contributionsCollection(from: $from, to: $to) {
      totalIssueContributions
      totalPullRequestContributions
      totalPullRequestReviewContributions
      pullRequestContributionsByRepository(maxRepositories: 50) {
        repository { nameWithOwner url }
        contributions { totalCount }
      }
      issueContributionsByRepository(maxRepositories: 50) {
        repository { nameWithOwner url }
        contributions { totalCount }
      }
      pullRequestReviewContributionsByRepository(maxRepositories: 50) {
        repository { nameWithOwner url }
        contributions { totalCount }
      }
      contributionCalendar {
        weeks {
          contributionDays {
            date
            contributionCount
          }
        }
      }
    }
  }
}`

type RepoContribution struct {
	FullName string
	URL      string
	PRs      int
	Issues   int
	Reviews  int
}

type UserGraphStats struct {
	PeriodStart         time.Time
	PeriodEnd           time.Time
	Days                []ContributionDay
	IssueContributions  int
	PRContributions     int
	ReviewContributions int
	OrgCount            int
	ReposWithReadme     int
	ReleaseCount        int
	ByRepository        []RepoContribution
}

func (c *Client) GetUserGraphStats(ctx context.Context, username string) (*UserGraphStats, error) {
	to := time.Now().UTC()
	from := to.AddDate(-1, 0, 0)

	payload, _ := json.Marshal(map[string]any{
		"query": userStatsQuery,
		"variables": map[string]any{
			"login": username,
			"from":  from.Format(time.RFC3339),
			"to":    to.Format(time.RFC3339),
		},
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.graphqlBase, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			User struct {
				Followers struct {
					TotalCount int `json:"totalCount"`
				} `json:"followers"`
				Organizations struct {
					TotalCount int `json:"totalCount"`
				} `json:"organizations"`
				Repositories struct {
					Nodes []struct {
						IsFork         bool   `json:"isFork"`
						StargazerCount int    `json:"stargazerCount"`
						Description    string `json:"description"`
						Releases       struct {
							TotalCount int `json:"totalCount"`
						} `json:"releases"`
						Object *struct {
							Text string `json:"text"`
						} `json:"object"`
					} `json:"nodes"`
				} `json:"repositories"`
				ContributionsCollection struct {
					TotalIssueContributions             int `json:"totalIssueContributions"`
					TotalPullRequestContributions       int `json:"totalPullRequestContributions"`
					TotalPullRequestReviewContributions int `json:"totalPullRequestReviewContributions"`
					PullRequestContributionsByRepository  []repoContribGroup `json:"pullRequestContributionsByRepository"`
					IssueContributionsByRepository        []repoContribGroup `json:"issueContributionsByRepository"`
					PullRequestReviewContributionsByRepository []repoContribGroup `json:"pullRequestReviewContributionsByRepository"`
					ContributionCalendar                  struct {
						Weeks []struct {
							Days []ContributionDay `json:"contributionDays"`
						} `json:"weeks"`
					} `json:"contributionCalendar"`
				} `json:"contributionsCollection"`
			} `json:"user"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("github: decode graphql: %w", err)
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("github graphql: %s", result.Errors[0].Message)
	}

	u := result.Data.User
	cc := u.ContributionsCollection
	stats := &UserGraphStats{
		PeriodStart:         from,
		PeriodEnd:           to,
		IssueContributions:  cc.TotalIssueContributions,
		PRContributions:     cc.TotalPullRequestContributions,
		ReviewContributions: cc.TotalPullRequestReviewContributions,
		OrgCount:            u.Organizations.TotalCount,
	}

	for _, week := range cc.ContributionCalendar.Weeks {
		stats.Days = append(stats.Days, week.Days...)
	}

	for _, repo := range u.Repositories.Nodes {
		if repo.IsFork {
			continue
		}
		if repo.Object != nil && len(repo.Object.Text) > 20 {
			stats.ReposWithReadme++
		}
		stats.ReleaseCount += repo.Releases.TotalCount
	}

	stats.ByRepository = mergeRepoContributions(
		cc.PullRequestContributionsByRepository,
		cc.IssueContributionsByRepository,
		cc.PullRequestReviewContributionsByRepository,
	)

	_ = u.Followers.TotalCount
	return stats, nil
}

type repoContribGroup struct {
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
		URL           string `json:"url"`
	} `json:"repository"`
	Contributions struct {
		TotalCount int `json:"totalCount"`
	} `json:"contributions"`
}

func mergeRepoContributions(prGroups, issueGroups, reviewGroups []repoContribGroup) []RepoContribution {
	byName := map[string]*RepoContribution{}

	apply := func(groups []repoContribGroup, kind string) {
		for _, g := range groups {
			name := g.Repository.NameWithOwner
			if name == "" {
				continue
			}
			entry, ok := byName[name]
			if !ok {
				entry = &RepoContribution{FullName: name, URL: g.Repository.URL}
				byName[name] = entry
			}
			if entry.URL == "" {
				entry.URL = g.Repository.URL
			}
			switch kind {
			case "pr":
				entry.PRs = g.Contributions.TotalCount
			case "issue":
				entry.Issues = g.Contributions.TotalCount
			case "review":
				entry.Reviews = g.Contributions.TotalCount
			}
		}
	}

	apply(prGroups, "pr")
	apply(issueGroups, "issue")
	apply(reviewGroups, "review")

	out := make([]RepoContribution, 0, len(byName))
	for _, v := range byName {
		out = append(out, *v)
	}
	return out
}

// GetContributions returns contribution calendar days for the last year.
func (c *Client) GetContributions(ctx context.Context, username string) ([]ContributionDay, error) {
	stats, err := c.GetUserGraphStats(ctx, username)
	if err != nil {
		return nil, err
	}
	return stats.Days, nil
}
