package ai

import (
	"fmt"

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
	"get_developer_snapshot":           {},
	"get_leaderboard_snapshot":         {},
	"get_community_snapshot":           {},
	"get_project_snapshot":             {},
}

type toolDeps struct {
	stats *repository.StatsRepo
}

func (d *toolDeps) developerSnapshot(ctx tool.Context, username string) (map[string]any, error) {
	dev, err := d.stats.GetDeveloperByUsername(ctx, username)
	if err != nil {
		return map[string]any{"error": "not found"}, nil
	}

	leaderboard, err := d.stats.GetLeaderboardWithTrends(ctx, "activity_score", "", 0, 1, 100)
	if err != nil {
		return nil, err
	}
	var rank *int
	for _, entry := range leaderboard {
		if entry.GithubUsername == username {
			r := entry.Rank
			rank = &r
			break
		}
	}

	repos, err := d.stats.GetDeveloperRepos(ctx, dev.ID)
	if err != nil {
		return nil, err
	}
	contribDays, err := d.stats.GetContributions(ctx, dev.ID)
	if err != nil {
		return nil, err
	}
	contribStats, err := d.stats.GetContributionStats(ctx, dev.ID)
	if err != nil {
		return nil, err
	}

	topRepos, topLanguage, languages := summarizeRepos(repos)
	totalContribs, activeDays, maxDay := contributionTotals(contribDays)
	quickFacts := []string{
		fmt.Sprintf("%d public repos", dev.PublicRepos),
		fmt.Sprintf("%d stars", dev.TotalStars),
		fmt.Sprintf("%d active contribution days", activeDays),
	}
	if rank != nil {
		quickFacts = append([]string{fmt.Sprintf("rank #%d", *rank)}, quickFacts...)
	}
	if topLanguage != "" {
		quickFacts = append(quickFacts, fmt.Sprintf("top language: %s", topLanguage))
	}

	return map[string]any{
		"developer":                dev,
		"rank":                     rank,
		"summary":                  developerSummaryText(*dev, rank, totalContribs, activeDays, maxDay, topLanguage),
		"quick_facts":              quickFacts,
		"languages":                languages,
		"top_repos":                topRepos,
		"repos":                    repos,
		"contribution_stats":       contribStats,
		"total_contributions":      totalContribs,
		"active_contribution_days": activeDays,
		"max_single_day":           maxDay,
	}, nil
}

func contributionTotals(days []models.ContributionDay) (total, active, maxDay int) {
	for _, day := range days {
		total += day.Count
		if day.Count > 0 {
			active++
		}
		if day.Count > maxDay {
			maxDay = day.Count
		}
	}
	return total, active, maxDay
}

func developerSummaryText(dev models.Developer, rank *int, totalContribs, activeDays, maxDay int, topLanguage string) string {
	name := dev.DisplayName
	if name == nil || *name == "" {
		name = &dev.GithubUsername
	}
	summary := fmt.Sprintf("%s has %d public repos, %d stars, and %d contributions across %d active days.",
		*name, dev.PublicRepos, dev.TotalStars, totalContribs, activeDays)
	if rank != nil {
		summary = fmt.Sprintf("%s They sit at rank #%d on the activity leaderboard.", summary, *rank)
	}
	if topLanguage != "" {
		summary = fmt.Sprintf("%s Their most common language is %s.", summary, topLanguage)
	}
	if maxDay > 0 {
		summary = fmt.Sprintf("%s Peak day: %d contributions.", summary, maxDay)
	}
	return summary
}

func leaderboardSummary(entries []models.LeaderboardEntry, sortBy string) map[string]any {
	top := entries
	if len(top) > 3 {
		top = top[:3]
	}
	leaders := []map[string]any{}
	for _, e := range top {
		leaders = append(leaders, map[string]any{
			"rank":           e.Rank,
			"username":       e.GithubUsername,
			"display_name":   e.DisplayName,
			"activity_score": e.ActivityScore,
			"total_stars":    e.TotalStars,
			"public_repos":   e.PublicRepos,
			"followers":      e.Followers,
			"rank_delta_7d":  e.RankDelta7d,
			"rank_delta_30d": e.RankDelta30d,
			"sparkline":      e.Sparkline,
		})
	}

	return map[string]any{
		"sort_by":       sortBy,
		"count":         len(entries),
		"leaders":       leaders,
		"top_3":         top,
		"trend_summary": fmt.Sprintf("Top %d developers ranked by %s.", len(entries), sortBy),
	}
}

func communitySnapshotData(ctx tool.Context, stats *repository.StatsRepo, days int) (map[string]any, error) {
	overview, err := stats.GetOverview(ctx)
	if err != nil {
		return nil, err
	}
	languages, err := stats.GetLanguageStats(ctx)
	if err != nil {
		return nil, err
	}
	oss, err := stats.GetOSSStats(ctx)
	if err != nil {
		return nil, err
	}
	streaks, err := stats.GetStreakSummary(ctx)
	if err != nil {
		return nil, err
	}
	activity, err := stats.GetCommunityActivity(ctx, days)
	if err != nil {
		return nil, err
	}
	recent, err := stats.GetRecentActivity(ctx, 10)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"overview":           overview,
		"languages":          languages,
		"open_source":        oss,
		"streaks":            streaks,
		"community_activity": activity,
		"recent_activity":    recent,
		"summary":            fmt.Sprintf("%d developers, %d repos, and %d total stars tracked across the platform.", overview.TotalDevelopers, overview.TotalRepos, overview.TotalStars),
	}, nil
}

func projectSnapshotData(ctx tool.Context, stats *repository.StatsRepo, category, language, sortBy string, limit int) (map[string]any, error) {
	projects, err := stats.ListProjects(ctx, repository.ProjectFilter{
		Category: category,
		Language: language,
		Sort:     sortBy,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}
	fastest, err := stats.GetFastestGrowingRepos(ctx, 30, limit)
	if err != nil {
		return nil, err
	}

	topRepos := projects
	if len(topRepos) > 5 {
		topRepos = topRepos[:5]
	}
	return map[string]any{
		"projects":            projects,
		"fastest_growing":     fastest,
		"top_projects":        topRepos,
		"requested_limit":     limit,
		"returned_count":      len(projects),
		"has_partial_results": len(projects) < limit,
		"summary":             fmt.Sprintf("%d projects returned for category=%s, language=%s, sort=%s.", len(projects), category, language, sortBy),
	}, nil
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
			if len(entries) == 0 {
				return map[string]any{
					"sort_by": sortBy,
					"count":   0,
					"leaders": []models.LeaderboardEntry{},
				}, nil
			}
			return leaderboardSummary(entries, sortBy), nil
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_developer_profile",
			Description: "Get an enriched profile snapshot of a NUST developer including rank, repos, and contribution summary.",
		}, func(ctx tool.Context, in usernameIn) (any, error) {
			if in.Username == "" {
				return map[string]string{"error": "username required"}, nil
			}
			return d.developerSnapshot(ctx, in.Username)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_developer_snapshot",
			Description: "Get a dense developer context pack: profile, rank, repos, contribution stats, and quick facts.",
		}, func(ctx tool.Context, in usernameIn) (any, error) {
			if in.Username == "" {
				return map[string]string{"error": "username required"}, nil
			}
			return d.developerSnapshot(ctx, in.Username)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_developer_repos",
			Description: "Get public repositories of a NUST developer with a compact repo summary.",
		}, func(ctx tool.Context, in usernameIn) (any, error) {
			if in.Username == "" {
				return map[string]string{"error": "username required"}, nil
			}
			dev, err := d.stats.GetDeveloperByUsername(ctx, in.Username)
			if err != nil {
				return map[string]string{"error": "not found"}, nil
			}
			repos, err := d.stats.GetDeveloperRepos(ctx, dev.ID)
			if err != nil {
				return nil, err
			}
			topRepos, topLanguage, _ := summarizeRepos(repos)
			originals, forks := 0, 0
			for _, repo := range repos {
				if repo.IsFork {
					forks++
				} else {
					originals++
				}
			}
			return map[string]any{
				"developer":      dev,
				"top_language":   topLanguage,
				"original_count": originals,
				"fork_count":     forks,
				"top_repos":      topRepos,
				"repos":          repos,
				"summary":        fmt.Sprintf("%s has %d repos, %d original projects, and %d forks.", dev.GithubUsername, len(repos), originals, forks),
			}, nil
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_developer_contribution_stats",
			Description: "Get PR, issue, and review contributions for a developer with repository-level summary.",
		}, func(ctx tool.Context, in usernameIn) (any, error) {
			if in.Username == "" {
				return map[string]string{"error": "username required"}, nil
			}
			dev, err := d.stats.GetDeveloperByUsername(ctx, in.Username)
			if err != nil {
				return map[string]string{"error": "not found"}, nil
			}
			stats, err := d.stats.GetContributionStats(ctx, dev.ID)
			if err != nil {
				return nil, err
			}
			topRepo := ""
			if len(stats.ByRepository) > 0 {
				topRepo = stats.ByRepository[0].RepoFullName
			}
			return map[string]any{
				"developer":      dev,
				"stats":          stats,
				"top_repository": topRepo,
				"summary":        fmt.Sprintf("%s has %d PRs, %d issues, and %d reviews in the tracked period.", dev.GithubUsername, stats.PullRequests, stats.Issues, stats.Reviews),
			}, nil
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_developer_contributions",
			Description: "Get 365-day contribution calendar summary for a developer with streak-friendly rollups.",
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
			total, active, maxDay := contributionTotals(days)
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
				"summary":                  fmt.Sprintf("%s recorded %d contributions across %d active days.", in.Username, total, active),
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
			Description: "Get top open-source projects by NUST developers with a compact summary.",
		}, func(ctx tool.Context, in projectsIn) (any, error) {
			limit := clamp(in.Limit, 1, 50)
			if limit == 0 {
				limit = 10
			}
			sort := in.Sort
			if sort == "" {
				sort = "stars"
			}
			projects, err := d.stats.ListProjects(ctx, repository.ProjectFilter{
				Language: in.Language,
				Sort:     sort,
				Limit:    limit,
			})
			if err != nil {
				return nil, err
			}
			top := projects
			if len(top) > 5 {
				top = top[:5]
			}
			return map[string]any{
				"projects":     projects,
				"top_projects": top,
				"summary":      fmt.Sprintf("%d projects found, sorted by %s.", len(projects), sort),
			}, nil
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_fastest_growing_projects",
			Description: "Get fastest growing NUST projects by star/fork growth in the last 30 days with a summary.",
		}, func(ctx tool.Context, in growingIn) (any, error) {
			limit := clamp(in.Limit, 1, 30)
			if limit == 0 {
				limit = 10
			}
			projects, err := d.stats.GetFastestGrowingRepos(ctx, 30, limit)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"projects": projects,
				"summary":  fmt.Sprintf("Top %d fastest-growing projects in the last 30 days.", len(projects)),
			}, nil
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
			Description: "Get recent community activity events across NUST developers with a compact summary.",
		}, func(ctx tool.Context, in activityIn) (any, error) {
			limit := clamp(in.Limit, 1, 30)
			if limit == 0 {
				limit = 10
			}
			events, err := d.stats.GetRecentActivity(ctx, limit)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"events":  events,
				"summary": fmt.Sprintf("Fetched %d recent activity events.", len(events)),
			}, nil
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
			Name:        "get_leaderboard_snapshot",
			Description: "Get a dense leaderboard snapshot with top developers, trends, and summary context.",
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
			return map[string]any{
				"sort_by": sortBy,
				"count":   len(entries),
				"leaders": entries,
				"summary": fmt.Sprintf("Leaderboard snapshot for %s with %d developers.", sortBy, len(entries)),
			}, nil
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_community_snapshot",
			Description: "Get a dense platform-wide community snapshot: overview, languages, OSS, streaks, and recent activity.",
		}, func(ctx tool.Context, _ empty) (any, error) {
			return communitySnapshotData(ctx, d.stats, 30)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "get_project_snapshot",
			Description: "Get a dense project snapshot with top repos, fastest growers, and summary context.",
		}, func(ctx tool.Context, in projectsIn) (any, error) {
			category := "original"
			if in.Sort == "forks" {
				category = "forks"
			}
			sortBy := in.Sort
			if sortBy == "" {
				sortBy = "stars"
			}
			limit := clamp(in.Limit, 1, 50)
			if limit == 0 {
				limit = 10
			}
			return projectSnapshotData(ctx, d.stats, category, in.Language, sortBy, limit)
		})),

		must(functiontool.New(functiontool.Config{
			Name:        "compare_developers",
			Description: "Compare two NUST developers side-by-side with metric deltas and summaries.",
		}, func(ctx tool.Context, in compareIn) (any, error) {
			if in.UsernameA == "" || in.UsernameB == "" {
				return map[string]string{"error": "both usernames required"}, nil
			}
			a, errA := d.stats.GetDeveloperByUsername(ctx, in.UsernameA)
			b, errB := d.stats.GetDeveloperByUsername(ctx, in.UsernameB)
			if errA != nil || errB != nil {
				return map[string]any{
					"error": "one or both developers not found",
				}, nil
			}
			aRepos, _ := d.stats.GetDeveloperRepos(ctx, a.ID)
			bRepos, _ := d.stats.GetDeveloperRepos(ctx, b.ID)
			aContribs, _ := d.stats.GetContributions(ctx, a.ID)
			bContribs, _ := d.stats.GetContributions(ctx, b.ID)
			aTopRepos, aTopLanguage, _ := summarizeRepos(aRepos)
			bTopRepos, bTopLanguage, _ := summarizeRepos(bRepos)
			aTotal, aActive, _ := contributionTotals(aContribs)
			bTotal, bActive, _ := contributionTotals(bContribs)
			winner := map[string]string{}
			if a.ActivityScore > b.ActivityScore {
				winner["activity_score"] = in.UsernameA
			} else if b.ActivityScore > a.ActivityScore {
				winner["activity_score"] = in.UsernameB
			}
			if a.TotalStars > b.TotalStars {
				winner["total_stars"] = in.UsernameA
			} else if b.TotalStars > a.TotalStars {
				winner["total_stars"] = in.UsernameB
			}
			if a.PublicRepos > b.PublicRepos {
				winner["public_repos"] = in.UsernameA
			} else if b.PublicRepos > a.PublicRepos {
				winner["public_repos"] = in.UsernameB
			}
			return map[string]any{
				"developer_a":              a,
				"developer_b":              b,
				"developer_a_top_repos":    aTopRepos,
				"developer_b_top_repos":    bTopRepos,
				"developer_a_top_language": aTopLanguage,
				"developer_b_top_language": bTopLanguage,
				"developer_a_contributions": map[string]any{
					"total":       aTotal,
					"active_days": aActive,
				},
				"developer_b_contributions": map[string]any{
					"total":       bTotal,
					"active_days": bActive,
				},
				"winner_by_metric": winner,
				"summary":          fmt.Sprintf("Compared %s and %s across profile, repo, and contribution data.", in.UsernameA, in.UsernameB),
			}, nil
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
				"summary": fmt.Sprintf("Found %d developers matching the current filters.", len(results)),
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
