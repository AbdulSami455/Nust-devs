package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const contributionsQuery = `
query($login: String!, $from: DateTime!, $to: DateTime!) {
  user(login: $login) {
    contributionsCollection(from: $from, to: $to) {
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

func (c *Client) GetContributions(ctx context.Context, username string) ([]ContributionDay, error) {
	to := time.Now()
	from := to.AddDate(-1, 0, 0)

	payload, _ := json.Marshal(map[string]any{
		"query": contributionsQuery,
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
				ContributionsCollection struct {
					ContributionCalendar struct {
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

	var days []ContributionDay
	for _, week := range result.Data.User.ContributionsCollection.ContributionCalendar.Weeks {
		days = append(days, week.Days...)
	}
	return days, nil
}
