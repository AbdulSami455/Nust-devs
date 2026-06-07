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

type UserGraphStats struct {
	Days                []ContributionDay
	IssueContributions  int
	PRContributions     int
	ReviewContributions int
	OrgCount            int
	ReposWithReadme     int
	ReleaseCount        int
}

func (c *Client) GetUserGraphStats(ctx context.Context, username string) (*UserGraphStats, error) {
	to := time.Now()
	from := to.AddDate(-1, 0, 0)

	payload, _ := json.Marshal(map[string]any{
		"query": userStatsQuery,
		"variables": map[string]any{
			"login": username,
			"from":  from.UTC().Format(time.RFC3339),
			"to":    to.UTC().Format(time.RFC3339),
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
						IsFork        bool   `json:"isFork"`
						StargazerCount int   `json:"stargazerCount"`
						Description   string `json:"description"`
						Releases      struct {
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
					ContributionCalendar                struct {
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
	stats := &UserGraphStats{
		IssueContributions:  u.ContributionsCollection.TotalIssueContributions,
		PRContributions:     u.ContributionsCollection.TotalPullRequestContributions,
		ReviewContributions: u.ContributionsCollection.TotalPullRequestReviewContributions,
		OrgCount:            u.Organizations.TotalCount,
	}

	for _, week := range u.ContributionsCollection.ContributionCalendar.Weeks {
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

	_ = u.Followers.TotalCount // followers synced via REST profile
	return stats, nil
}

// GetContributions returns contribution calendar days for the last year.
func (c *Client) GetContributions(ctx context.Context, username string) ([]ContributionDay, error) {
	stats, err := c.GetUserGraphStats(ctx, username)
	if err != nil {
		return nil, err
	}
	return stats.Days, nil
}
