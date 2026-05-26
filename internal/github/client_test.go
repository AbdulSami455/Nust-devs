package github_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gh "github.com/abdulsami/nust-devs/internal/github"
)

// newTestClient returns a Client pointed at a fake server and a function to add handlers.
func newTestServer(mux *http.ServeMux) (*gh.Client, *httptest.Server) {
	srv := httptest.NewServer(mux)
	c := gh.NewClientWithBaseURL("fake-token", srv.URL, srv.URL)
	return c, srv
}

func TestGetUser(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/octocat", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"login":        "octocat",
			"name":         "The Octocat",
			"followers":    1000,
			"public_repos": 8,
		})
	})
	client, srv := newTestServer(mux)
	defer srv.Close()

	user, err := client.GetUser(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Login != "octocat" {
		t.Errorf("got login %q, want %q", user.Login, "octocat")
	}
	if user.Followers != 1000 {
		t.Errorf("got followers %d, want 1000", user.Followers)
	}
}

func TestGetRepos_Pagination(t *testing.T) {
	page := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/users/octocat/repos", func(w http.ResponseWriter, r *http.Request) {
		page++
		if page == 1 {
			// Return 100 repos to trigger a second page fetch
			repos := make([]map[string]any, 100)
			for i := range repos {
				repos[i] = map[string]any{"id": i, "name": "repo", "full_name": "octocat/repo", "pushed_at": time.Now()}
			}
			json.NewEncoder(w).Encode(repos)
		} else {
			json.NewEncoder(w).Encode([]map[string]any{{"id": 200, "name": "last", "full_name": "octocat/last", "pushed_at": time.Now()}})
		}
	})
	client, srv := newTestServer(mux)
	defer srv.Close()

	repos, err := client.GetRepos(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 101 {
		t.Errorf("got %d repos, want 101", len(repos))
	}
}

func TestRateLimiter_BlocksOnLowQuota(t *testing.T) {
	calls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/users/x", func(w http.ResponseWriter, r *http.Request) {
		calls++
		// First call: signal quota almost gone, reset in past so no real sleep
		if calls == 1 {
			w.Header().Set("x-ratelimit-remaining", "5")
			w.Header().Set("x-ratelimit-reset", "1") // epoch past
		}
		json.NewEncoder(w).Encode(map[string]any{"login": "x"})
	})
	client, srv := newTestServer(mux)
	defer srv.Close()

	// First call updates the limiter with remaining=5
	if _, err := client.GetUser(context.Background(), "x"); err != nil {
		t.Fatal(err)
	}
	// Second call should still succeed (reset is in the past, no real sleep needed)
	if _, err := client.GetUser(context.Background(), "x"); err != nil {
		t.Fatal(err)
	}
}
