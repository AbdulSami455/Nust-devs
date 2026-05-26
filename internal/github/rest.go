package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetUser(ctx context.Context, username string) (*User, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.restBase+"/users/"+username, nil)
	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github: get user %s: status %d", username, resp.StatusCode)
	}
	var u User
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("github: decode user: %w", err)
	}
	return &u, nil
}

func (c *Client) GetRepos(ctx context.Context, username string) ([]Repo, error) {
	var all []Repo
	for page := 1; ; page++ {
		url := fmt.Sprintf("%s/users/%s/repos?per_page=100&page=%d&sort=updated", c.restBase, username, page)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := c.do(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("github: get repos %s: status %d", username, resp.StatusCode)
		}
		var batch []Repo
		if err := json.NewDecoder(resp.Body).Decode(&batch); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("github: decode repos: %w", err)
		}
		resp.Body.Close()
		all = append(all, batch...)
		if len(batch) < 100 {
			break
		}
	}
	return all, nil
}

func (c *Client) GetLanguages(ctx context.Context, owner, repo string) (map[string]int64, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/languages", c.restBase, owner, repo)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github: get languages %s/%s: status %d", owner, repo, resp.StatusCode)
	}
	var langs map[string]int64
	if err := json.NewDecoder(resp.Body).Decode(&langs); err != nil {
		return nil, fmt.Errorf("github: decode languages: %w", err)
	}
	return langs, nil
}
