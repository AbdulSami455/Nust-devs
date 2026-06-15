package ai

import (
	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/repository"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

var allowedToolNames = map[string]struct{}{
	"get_top_developers":               {},
	"get_developer_profile":            {},
	"get_developer_repos":              {},
	"get_developer_contribution_stats": {},
	"get_developer_contributions":      {},
	"get_developer_wrapped":            {},
	"get_stats_overview":               {},
	"get_language_stats":               {},
	"get_oss_stats":                    {},
	"get_streak_summary":               {},
	"get_top_projects":                 {},
	"get_fastest_growing_projects":     {},
	"get_dev_of_month":                 {},
	"get_innovation_graph":             {},
	"get_recent_activity":              {},
	"get_community_activity":           {},
	"compare_developers":               {},
	"search_developers":                {},
}

type toolDeps struct {
	stats *repository.StatsRepo
}

func buildTools(stats *repository.StatsRepo) ([]tool.Tool, error) {
	d := &toolDeps{stats: stats}
	return d.allTools()
}

func (d *toolDeps) allTools() ([]tool.Tool, error) {
	type empty struct{}
	type usernameIn struct {
		Username string `json:"username"`
	}
	type wrappedIn struct {
		Username string `json:"username"`
		Year     int    `json:"year,omitempty"`
	}
	type leaderboardIn struct {
		SortBy string `json:"sort_by,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}
	type projectsIn struct {
		Language string `json:"language,omitempty"`
		Sort     string `json:"sort,omitempty"`
		Limit    int    `json:"limit,omitempty"`
	}
	type growingIn struct {
		Limit int `json:"limit,omitempty"`
	}
	type activityIn struct {
		Limit int `json:"limit,omitempty"`
	}
	type communityIn struct {
		Days int `json:"days,omitempty"`
	}
	type innovationIn struct {
		Granularity string `json:"granularity,omitempty"`
	}
	type compareIn struct {
		UsernameA string `json:"username_a"`
		UsernameB string `json:"username_b"`
	}
	type searchIn struct {
		MinStars   int `json:"min_stars,omitempty"`
		MinPRs     int `json:"min_prs,omitempty"`
		MinReviews int `json:"min_reviews,omitempty"`
		Limit      int `json:"limit,omitempty"`
	}

	must := func(t tool.Tool, err error) tool.Tool {
		if err != nil {
			panic(err)
		}
		return t
	}

	tools := []tool.Tool{
		must(functiontool.New(functiontool.Config{
			Name:        "get_top_developers",
			Description: "Get top NUST developers ranked by a score dimension.",
		}, func(ctx tool.Context, in leaderboardIn) (any, error) {
			sortBy := in.SortBy
			if sortBy == "" {
				sortBy = "activity_score"
			}
			limit := clamp(in.Limit, 1, 50)
			if limit == 0 {
				limit = 10
			}
			entries, err := d.stats.GetLeaderboardWithTrends(ctx, sortBy, "", 0, 1, limit)
			if err != nil {
				return nil, err
			}
			return entries, nil
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_developer_profile",
			Description: "Get full profile of a NUST developer: scores, XP, streak, bio, repos, stars.",
		}, func(ctx tool.Context, in usernameIn) (any, error) {
			if in.Username == "" {
				return map[string]string{"error": "username required"}, nil
			}
			dev, err := d.stats.GetDeveloperByUsername(ctx, in.Username)
			if err != nil {
				return map[string]string{"error": "not found"}, nil
			}
			return dev, nil
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_developer_repos",
			Description: "Get public repositories of a NUST developer.",
		}, func(ctx tool.Context, in usernameIn) (any, error) {
			if in.Username == "" {
				return map[string]string{"error": "username required"}, nil
			}
			dev, err := d.stats.GetDeveloperByUsername(ctx, in.Username)
			if err != nil {
				return map[string]string{"error": "not found"}, nil
			}
			return d.stats.GetDeveloperRepos(ctx, dev.ID)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_developer_contribution_stats",
			Description: "Get PR, issue, and review contributions for a developer by repository.",
		}, func(ctx tool.Context, in usernameIn) (any, error) {
			if in.Username == "" {
				return map[string]string{"error": "username required"}, nil
			}
			dev, err := d.stats.GetDeveloperByUsername(ctx, in.Username)
			if err != nil {
				return map[string]string{"error": "not found"}, nil
			}
			return d.stats.GetContributionStats(ctx, dev.ID)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_developer_contributions",
			Description: "Get 365-day contribution calendar summary for a developer.",
		}, func(ctx tool.Context, in usernameIn) (any, error) {
			if in.Username == "" {
				return map[string]string{"error": "username required"}, nil
			}
			dev, err := d.stats.GetDeveloperByUsername(ctx, in.Username)
			if err != nil {
				return map[string]string{"error": "not found"}, nil
			}
			days, err := d.stats.GetContributions(ctx, dev.ID)
			if err != nil {
				return nil, err
			}
			total, active, maxDay := 0, 0, 0
			for _, day := range days {
				total += day.Count
				if day.Count > 0 {
					active++
				}
				if day.Count > maxDay {
					maxDay = day.Count
				}
			}
			recent := days
			if len(days) > 30 {
				recent = days[len(days)-30:]
			}
			return map[string]any{
				"username":                 in.Username,
				"total_contributions_365d": total,
				"active_days":              active,
				"max_single_day":           maxDay,
				"last_30_days":             recent,
			}, nil
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_developer_wrapped",
			Description: "Get a developer's year-in-review (Wrapped) report.",
		}, func(ctx tool.Context, in wrappedIn) (any, error) {
			if in.Username == "" {
				return map[string]string{"error": "username required"}, nil
			}
			return d.stats.GetWrapped(ctx, in.Username, in.Year)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_stats_overview",
			Description: "Get platform-wide totals: developers, repos, stars, contributions.",
		}, func(ctx tool.Context, _ empty) (any, error) {
			return d.stats.GetOverview(ctx)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_language_stats",
			Description: "Get programming language breakdown across NUST developers.",
		}, func(ctx tool.Context, _ empty) (any, error) {
			return d.stats.GetLanguageStats(ctx)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_oss_stats",
			Description: "Get open-source statistics: original projects, forks, stars, contributors.",
		}, func(ctx tool.Context, _ empty) (any, error) {
			return d.stats.GetOSSStats(ctx)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_streak_summary",
			Description: "Get contribution streak stats across the community.",
		}, func(ctx tool.Context, _ empty) (any, error) {
			return d.stats.GetStreakSummary(ctx)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_top_projects",
			Description: "Get top open-source projects by NUST developers.",
		}, func(ctx tool.Context, in projectsIn) (any, error) {
			limit := clamp(in.Limit, 1, 50)
			if limit == 0 {
				limit = 10
			}
			sort := in.Sort
			if sort == "" {
				sort = "stars"
			}
			return d.stats.ListProjects(ctx, repository.ProjectFilter{
				Language: in.Language,
				Sort:     sort,
				Limit:    limit,
			})
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_fastest_growing_projects",
			Description: "Get fastest growing NUST projects by star/fork growth in the last 30 days.",
		}, func(ctx tool.Context, in growingIn) (any, error) {
			limit := clamp(in.Limit, 1, 30)
			if limit == 0 {
				limit = 10
			}
			return d.stats.GetFastestGrowingRepos(ctx, 30, limit)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_dev_of_month",
			Description: "Get Developer of the Month winners.",
		}, func(ctx tool.Context, _ empty) (any, error) {
			return d.stats.ListDevOfMonthWinners(ctx, 3)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_innovation_graph",
			Description: "Get innovation trends: pushes, new repos, active developers, net new stars.",
		}, func(ctx tool.Context, in innovationIn) (any, error) {
			g := in.Granularity
			if g != "monthly" && g != "quarterly" {
				g = "monthly"
			}
			return d.stats.GetInnovationGraph(ctx, g, 12)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_recent_activity",
			Description: "Get recent community activity events across NUST developers.",
		}, func(ctx tool.Context, in activityIn) (any, error) {
			limit := clamp(in.Limit, 1, 30)
			if limit == 0 {
				limit = 10
			}
			return d.stats.GetRecentActivity(ctx, limit)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_spotlight_developer",
			Description: "Get the currently featured (spotlight) NUST developer.",
		}, func(ctx tool.Context, _ empty) (any, error) {
			return d.stats.GetSpotlightDeveloper(ctx)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_community_activity",
			Description: "Get community-wide daily contribution counts for the last N days.",
		}, func(ctx tool.Context, in communityIn) (any, error) {
			days := clamp(in.Days, 7, 90)
			if days == 0 {
				days = 30
			}
			return d.stats.GetCommunityActivity(ctx, days)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "compare_developers",
			Description: "Compare two NUST developers side-by-side.",
		}, func(ctx tool.Context, in compareIn) (any, error) {
			if in.UsernameA == "" || in.UsernameB == "" {
				return map[string]string{"error": "both usernames required"}, nil
			}
			var devA, devB any
			if d1, err := d.stats.GetDeveloperByUsername(ctx, in.UsernameA); err == nil {
				devA = d1
			} else {
				devA = map[string]string{"error": "not found"}
			}
			if d2, err := d.stats.GetDeveloperByUsername(ctx, in.UsernameB); err == nil {
				devB = d2
			} else {
				devB = map[string]string{"error": "not found"}
			}
			return map[string]any{"developer_a": devA, "developer_b": devB}, nil
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "search_developers",
			Description: "Search NUST developers by min stars, PRs, or reviews.",
		}, func(ctx tool.Context, in searchIn) (any, error) {
			limit := clamp(in.Limit, 1, 30)
			if limit == 0 {
				limit = 10
			}
			entries, err := d.stats.GetLeaderboardWithTrends(ctx, "activity_score", "", 0, 1, 100)
			if err != nil {
				return nil, err
			}
			var results []models.LeaderboardEntry
			for _, e := range entries {
				if in.MinStars > 0 && e.TotalStars < in.MinStars {
					continue
				}
				if in.MinPRs > 0 && e.PRContributions < in.MinPRs {
					continue
				}
				if in.MinReviews > 0 && e.ReviewContributions < in.MinReviews {
					continue
				}
				results = append(results, e)
				if len(results) >= limit {
					break
				}
			}
			if results == nil {
				results = []models.LeaderboardEntry{}
			}
			return map[string]any{
				"query":   in,
				"results": results,
				"count":   len(results),
			}, nil
		})),
	}
	return tools, nil
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
