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
	page, limit := pagination(r)
	key := fmt.Sprintf("leaderboard:%s:%d:%d", sortBy, page, limit)
	h.cachedJSON(w, r, key, 5*time.Minute, func() (any, error) {
		return h.stats.GetLeaderboard(r.Context(), sortBy, page, limit)
	})
}

func (h *PublicHandler) GetTopProjects(w http.ResponseWriter, r *http.Request) {
	h.cachedJSON(w, r, "projects:top", 10*time.Minute, func() (any, error) {
		return h.stats.GetTopProjects(r.Context(), 30)
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
