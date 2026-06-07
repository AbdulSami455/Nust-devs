package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/abdulsami/nust-devs/internal/cache"
	"github.com/abdulsami/nust-devs/internal/repository"
)

type PublicHandler struct {
	stats *repository.StatsRepo
	cache *cache.Cache
}

func NewPublicHandler(stats *repository.StatsRepo, cache *cache.Cache) *PublicHandler {
	return &PublicHandler{stats: stats, cache: cache}
}

func (h *PublicHandler) cachedJSON(w http.ResponseWriter, r *http.Request, key string, ttl time.Duration, fetch func() (any, error)) {
	var dest any
	if hit, _ := h.cache.GetJSON(r.Context(), key, &dest); hit {
		writeJSON(w, http.StatusOK, dest)
		return
	}
	val, err := fetch()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	_ = h.cache.SetJSON(r.Context(), key, val, ttl)
	writeJSON(w, http.StatusOK, val)
}

func (h *PublicHandler) ListDevelopers(w http.ResponseWriter, r *http.Request) {
	page, limit := pagination(r)
	key := fmt.Sprintf("developers:list:%d:%d", page, limit)
	h.cachedJSON(w, r, key, 5*time.Minute, func() (any, error) {
		devs, err := h.stats.ListDevelopers(r.Context(), page, limit)
		if err != nil {
			return nil, err
		}
		if devs == nil {
			return []struct{}{}, nil
		}
		return devs, nil
	})
}

func (h *PublicHandler) GetDeveloper(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	key := "developer:" + username
	h.cachedJSON(w, r, key, 5*time.Minute, func() (any, error) {
		dev, err := h.stats.GetDeveloperByUsername(r.Context(), username)
		if err != nil {
			return nil, err
		}
		return dev, nil
	})
}

func (h *PublicHandler) GetDeveloperRepos(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	dev, err := h.stats.GetDeveloperByUsername(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusNotFound, "developer not found")
		return
	}
	key := "developer:" + username + ":repos"
	h.cachedJSON(w, r, key, 10*time.Minute, func() (any, error) {
		return h.stats.GetDeveloperRepos(r.Context(), dev.ID)
	})
}

func (h *PublicHandler) GetDeveloperContributions(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	dev, err := h.stats.GetDeveloperByUsername(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusNotFound, "developer not found")
		return
	}
	key := "developer:" + username + ":contributions"
	h.cachedJSON(w, r, key, 30*time.Minute, func() (any, error) {
		return h.stats.GetContributions(r.Context(), dev.ID)
	})
}

func (h *PublicHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	sortBy := r.URL.Query().Get("sort_by")
	view := r.URL.Query().Get("view")
	period, _ := strconv.Atoi(r.URL.Query().Get("period"))
	page, limit := pagination(r)
	key := fmt.Sprintf("leaderboard:%s:%s:%d:%d:%d", sortBy, view, period, page, limit)
	h.cachedJSON(w, r, key, 5*time.Minute, func() (any, error) {
		entries, err := h.stats.GetLeaderboardWithTrends(r.Context(), sortBy, view, period, page, limit)
		if err != nil {
			return nil, err
		}
		if entries == nil {
			return []struct{}{}, nil
		}
		return entries, nil
	})
}

func (h *PublicHandler) GetTopProjects(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	language := r.URL.Query().Get("language")
	sortBy := r.URL.Query().Get("sort")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 30
	}
	key := fmt.Sprintf("projects:list:%s:%s:%s:%d", category, language, sortBy, limit)
	h.cachedJSON(w, r, key, 10*time.Minute, func() (any, error) {
		repos, err := h.stats.ListProjects(r.Context(), repository.ProjectFilter{
			Category: category,
			Language: language,
			Sort:     sortBy,
			Limit:    limit,
		})
		if err != nil {
			return nil, err
		}
		if repos == nil {
			return []struct{}{}, nil
		}
		return repos, nil
	})
}

func (h *PublicHandler) GetFastestGrowingProjects(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	key := fmt.Sprintf("projects:fastest-growing:%d:%d", days, limit)
	h.cachedJSON(w, r, key, 10*time.Minute, func() (any, error) {
		repos, err := h.stats.GetFastestGrowingRepos(r.Context(), days, limit)
		if err != nil {
			return nil, err
		}
		if repos == nil {
			return []struct{}{}, nil
		}
		return repos, nil
	})
}

func (h *PublicHandler) GetRepoGrowth(w http.ResponseWriter, r *http.Request) {
	repoID := r.PathValue("id")
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	key := fmt.Sprintf("repos:growth:%s:%d", repoID, days)
	h.cachedJSON(w, r, key, 15*time.Minute, func() (any, error) {
		return h.stats.GetRepoGrowth(r.Context(), repoID, days)
	})
}

func (h *PublicHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	h.cachedJSON(w, r, "stats:overview", 5*time.Minute, func() (any, error) {
		return h.stats.GetOverview(r.Context())
	})
}

func (h *PublicHandler) GetLanguages(w http.ResponseWriter, r *http.Request) {
	h.cachedJSON(w, r, "stats:languages", 15*time.Minute, func() (any, error) {
		return h.stats.GetLanguageStats(r.Context())
	})
}

func (h *PublicHandler) GetCommunityActivity(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days < 1 {
		days = 30
	}
	key := fmt.Sprintf("stats:community-activity:%d", days)
	h.cachedJSON(w, r, key, 10*time.Minute, func() (any, error) {
		return h.stats.GetCommunityActivity(r.Context(), days)
	})
}

func (h *PublicHandler) GetSpotlight(w http.ResponseWriter, r *http.Request) {
	h.cachedJSON(w, r, "developers:spotlight", 5*time.Minute, func() (any, error) {
		return h.stats.GetSpotlightDeveloper(r.Context())
	})
}

func (h *PublicHandler) GetRecentActivity(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	key := fmt.Sprintf("activity:recent:%d", limit)
	h.cachedJSON(w, r, key, 2*time.Minute, func() (any, error) {
		events, err := h.stats.GetRecentActivity(r.Context(), limit)
		if err != nil {
			return nil, err
		}
		if events == nil {
			return []struct{}{}, nil
		}
		return events, nil
	})
}

func (h *PublicHandler) GetOSSStats(w http.ResponseWriter, r *http.Request) {
	h.cachedJSON(w, r, "stats:open-source", 10*time.Minute, func() (any, error) {
		return h.stats.GetOSSStats(r.Context())
	})
}

func (h *PublicHandler) GetInnovationGraph(w http.ResponseWriter, r *http.Request) {
	granularity := r.URL.Query().Get("granularity")
	periods, _ := strconv.Atoi(r.URL.Query().Get("periods"))
	key := fmt.Sprintf("stats:innovation:%s:%d", granularity, periods)
	h.cachedJSON(w, r, key, 15*time.Minute, func() (any, error) {
		return h.stats.GetInnovationGraph(r.Context(), granularity, periods)
	})
}

func pagination(r *http.Request) (page, limit int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return
}
