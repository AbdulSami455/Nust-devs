package github

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const (
	baseURL    = "https://api.github.com"
	graphqlURL = "https://api.github.com/graphql"
)

type Client struct {
	http        *http.Client
	token       string
	limiter     *rateLimiter
	restBase    string
	graphqlBase string
}

func NewClient(token string) *Client {
	return NewClientWithBaseURL(token, baseURL, graphqlURL)
}

// NewClientWithBaseURL is used in tests to point the client at a fake server.
func NewClientWithBaseURL(token, restBase, gqlBase string) *Client {
	return &Client{
		http:       &http.Client{Timeout: 30 * time.Second},
		token:      token,
		limiter:    newRateLimiter(),
		restBase:   restBase,
		graphqlBase: gqlBase,
	}
}

func (c *Client) RateLimitState() (remaining int, resetAt time.Time) {
	return c.limiter.State()
}

func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	c.limiter.wait()

	backoff := time.Second
	for attempt := range 5 {
		resp, err := c.http.Do(req.Clone(ctx))
		if err != nil {
			return nil, fmt.Errorf("github request: %w", err)
		}
		c.limiter.update(resp)

		if resp.StatusCode == 403 || resp.StatusCode == 429 {
			resp.Body.Close()
			slog.Warn("github rate limited, retrying", "attempt", attempt+1, "backoff", backoff)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("github: exhausted retries for %s", req.URL.Path)
}
