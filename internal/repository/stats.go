package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsRepo struct {
	db *pgxpool.Pool
}

func NewStatsRepo(db *pgxpool.Pool) *StatsRepo {
	return &StatsRepo{db: db}
}

var developerCols = `
	id, github_username, display_name, avatar_url, bio, location, company, website,
	followers, following, public_repos, total_stars, activity_score,
	builder_score, contributor_score, reviewer_score, community_score,
	verification_status, last_synced_at, created_at, updated_at`

func scanPublicDeveloper(row interface {
	Scan(...any) error
}, d *models.Developer) error {
	return row.Scan(
		&d.ID, &d.GithubUsername, &d.DisplayName, &d.AvatarURL, &d.Bio,
		&d.Location, &d.Company, &d.Website,
		&d.Followers, &d.Following, &d.PublicRepos, &d.TotalStars, &d.ActivityScore,
		&d.BuilderScore, &d.ContributorScore, &d.ReviewerScore, &d.CommunityScore,
		&d.VerificationStatus, &d.LastSyncedAt, &d.CreatedAt, &d.UpdatedAt,
	)
}

func (r *StatsRepo) ListDevelopers(ctx context.Context, page, limit int) ([]models.Developer, error) {
	offset := (page - 1) * limit
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT %s FROM developers
		ORDER BY activity_score DESC
		LIMIT $1 OFFSET $2`, developerCols), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var devs []models.Developer
	for rows.Next() {
		var d models.Developer
		if err := scanPublicDeveloper(rows, &d); err != nil {
			return nil, err
		}
		devs = append(devs, d)
	}
	return devs, nil
}

func (r *StatsRepo) GetDeveloperByUsername(ctx context.Context, username string) (*models.Developer, error) {
	var d models.Developer
	err := scanPublicDeveloper(r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s FROM developers WHERE github_username = $1`, developerCols), username), &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func scanPublicRepo(row interface {
	Scan(...any) error
}, rp *models.PublicRepo) error {
	var pushedAt *time.Time
	err := row.Scan(
		&rp.ID, &rp.Name, &rp.FullName, &rp.Owner, &rp.Description, &rp.URL,
		&rp.Language, &rp.Stars, &rp.Forks, &rp.IsFork, &pushedAt,
	)
	if err != nil {
		return err
	}
	if pushedAt != nil {
		s := pushedAt.UTC().Format(time.RFC3339)
		rp.PushedAt = &s
	}
	return nil
}

const repoSelectCols = `
	r.id, r.name, r.full_name, r.owner, r.description, r.url, r.language, r.stars, r.forks, r.is_fork, r.pushed_at`

func (r *StatsRepo) GetDeveloperRepos(ctx context.Context, devID string) ([]models.PublicRepo, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT %s FROM repos r
		JOIN developer_repos dr ON dr.repo_id = r.id
		WHERE dr.developer_id = $1
		ORDER BY r.stars DESC`, repoSelectCols), devID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var repos []models.PublicRepo
	for rows.Next() {
		var rp models.PublicRepo
		if err := scanPublicRepo(rows, &rp); err != nil {
			return nil, err
		}
		repos = append(repos, rp)
	}
	return repos, nil
}

func (r *StatsRepo) GetContributions(ctx context.Context, devID string) ([]models.ContributionDay, error) {
	rows, err := r.db.Query(ctx, `
		SELECT date::text, count FROM contribution_days
		WHERE developer_id = $1 AND date >= CURRENT_DATE - INTERVAL '365 days'
		ORDER BY date ASC`, devID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var days []models.ContributionDay
	for rows.Next() {
		var d models.ContributionDay
		if err := rows.Scan(&d.Date, &d.Count); err != nil {
			return nil, err
		}
		days = append(days, d)
	}
	return days, nil
}

var validSortFields = map[string]string{
	"activity_score":    "activity_score",
	"total_stars":       "total_stars",
	"public_repos":      "public_repos",
	"followers":         "followers",
	"builder_score":     "builder_score",
	"contributor_score": "contributor_score",
	"reviewer_score":    "reviewer_score",
	"community_score":   "community_score",
}

func (r *StatsRepo) GetLeaderboard(ctx context.Context, sortBy string, page, limit int) ([]models.Developer, error) {
	col, ok := validSortFields[sortBy]
	if !ok {
		col = "activity_score"
	}
	offset := (page - 1) * limit
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT %s FROM developers
		ORDER BY %s DESC
		LIMIT $1 OFFSET $2`, developerCols, col), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var devs []models.Developer
	for rows.Next() {
		var d models.Developer
		if err := scanPublicDeveloper(rows, &d); err != nil {
			return nil, err
		}
		devs = append(devs, d)
	}
	return devs, nil
}

func (r *StatsRepo) GetTopProjects(ctx context.Context, limit int) ([]models.PublicRepo, error) {
	return r.ListProjects(ctx, ProjectFilter{Category: "original", Sort: "stars", Limit: limit})
}

type ProjectFilter struct {
	Category string
	Language string
	Sort     string
	Limit    int
}

func (r *StatsRepo) ListProjects(ctx context.Context, f ProjectFilter) ([]models.PublicRepo, error) {
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 30
	}
	where := []string{"EXISTS (SELECT 1 FROM developer_repos dr WHERE dr.repo_id = r.id)"}
	args := []any{}
	argN := 1

	switch f.Category {
	case "original":
		where = append(where, "r.is_fork = false")
	case "forks":
		where = append(where, "r.is_fork = true")
	}

	if f.Language != "" {
		where = append(where, fmt.Sprintf("r.language = $%d", argN))
		args = append(args, f.Language)
		argN++
	}

	order := "r.stars DESC"
	switch f.Sort {
	case "recent":
		order = "r.pushed_at DESC NULLS LAST"
	case "forks":
		order = "r.forks DESC"
	case "growth":
		order = "r.stars DESC"
	}

	args = append(args, f.Limit)
	query := fmt.Sprintf(`
		SELECT %s FROM repos r
		WHERE %s
		ORDER BY %s
		LIMIT $%d`, repoSelectCols, strings.Join(where, " AND "), order, argN)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var repos []models.PublicRepo
	for rows.Next() {
		var rp models.PublicRepo
		if err := scanPublicRepo(rows, &rp); err != nil {
			return nil, err
		}
		repos = append(repos, rp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if f.Sort == "growth" {
		return r.sortReposByGrowth(ctx, repos, f.Limit)
	}
	r.attachRepoGrowth(ctx, repos)
	return repos, nil
}

func (r *StatsRepo) GetOverview(ctx context.Context) (*models.Overview, error) {
	var o models.Overview
	err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM developers)::int,
			(SELECT COUNT(DISTINCT repo_id) FROM developer_repos)::int,
			(SELECT COALESCE(SUM(total_stars), 0) FROM developers)::int,
			(SELECT COALESCE(SUM(count), 0) FROM contribution_days)::bigint`).
		Scan(&o.TotalDevelopers, &o.TotalRepos, &o.TotalStars, &o.TotalContributions)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *StatsRepo) GetLanguageStats(ctx context.Context) ([]models.LanguageStat, error) {
	rows, err := r.db.Query(ctx, `
		SELECT language, SUM(bytes) as bytes, COUNT(DISTINCT repo_id) as repo_count
		FROM repo_languages
		GROUP BY language
		ORDER BY bytes DESC
		LIMIT 20`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats []models.LanguageStat
	for rows.Next() {
		var s models.LanguageStat
		if err := rows.Scan(&s.Language, &s.Bytes, &s.RepoCount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (r *StatsRepo) GetCommunityActivity(ctx context.Context, days int) ([]models.CommunityActivityDay, error) {
	if days < 1 || days > 365 {
		days = 30
	}
	rows, err := r.db.Query(ctx, `
		SELECT date::text, COALESCE(SUM(count), 0)::int
		FROM contribution_days
		WHERE date >= CURRENT_DATE - ($1::int * INTERVAL '1 day')
		GROUP BY date
		ORDER BY date ASC`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.CommunityActivityDay
	for rows.Next() {
		var d models.CommunityActivityDay
		if err := rows.Scan(&d.Date, &d.Count); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (r *StatsRepo) GetSpotlightDeveloper(ctx context.Context) (*models.Developer, error) {
	var d models.Developer
	err := scanPublicDeveloper(r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s FROM developers
		WHERE last_synced_at IS NOT NULL
		ORDER BY activity_score DESC
		LIMIT 1`, developerCols)), &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *StatsRepo) GetRecentActivity(ctx context.Context, limit int) ([]models.ActivityEvent, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	rows, err := r.db.Query(ctx, `
		SELECT d.github_username, r.full_name, r.pushed_at
		FROM repos r
		JOIN developer_repos dr ON dr.repo_id = r.id
		JOIN developers d ON d.id = dr.developer_id
		WHERE r.pushed_at IS NOT NULL
		ORDER BY r.pushed_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []models.ActivityEvent
	for rows.Next() {
		var username, repo string
		var pushedAt time.Time
		if err := rows.Scan(&username, &repo, &pushedAt); err != nil {
			return nil, err
		}
		events = append(events, models.ActivityEvent{
			Type:       "repo_updated",
			Username:   username,
			Repo:       repo,
			Message:    fmt.Sprintf("updated %s", repo),
			OccurredAt: pushedAt.UTC().Format(time.RFC3339),
		})
	}
	return events, nil
}

func (r *StatsRepo) GetOSSStats(ctx context.Context) (*models.OSSStats, error) {
	var s models.OSSStats
	var topLang *string
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE NOT r.is_fork),
			COUNT(*) FILTER (WHERE r.is_fork),
			COALESCE(SUM(r.stars) FILTER (WHERE NOT r.is_fork), 0)::int,
			COALESCE(SUM(r.forks), 0)::int,
			(SELECT COUNT(DISTINCT dr.developer_id) FROM developer_repos dr)::int
		FROM repos r
		WHERE EXISTS (SELECT 1 FROM developer_repos dr WHERE dr.repo_id = r.id)`).
		Scan(&s.OriginalProjects, &s.ForkProjects, &s.TotalStars, &s.TotalForksReceived, &s.Contributors)
	if err != nil {
		return nil, err
	}
	_ = r.db.QueryRow(ctx, `
		SELECT language FROM repos r
		WHERE EXISTS (SELECT 1 FROM developer_repos dr WHERE dr.repo_id = r.id)
		  AND r.language IS NOT NULL AND NOT r.is_fork
		GROUP BY language ORDER BY COUNT(*) DESC LIMIT 1`).Scan(&topLang)
	s.TopLanguage = topLang
	return &s, nil
}

func (r *StatsRepo) GetInnovationGraph(ctx context.Context, granularity string, periods int) (*models.InnovationGraph, error) {
	if granularity != "monthly" {
		granularity = "quarterly"
	}
	if periods < 4 || periods > 24 {
		periods = 8
	}

	since := time.Now().UTC()
	if granularity == "monthly" {
		since = since.AddDate(0, -(periods - 1), 0)
	} else {
		since = since.AddDate(0, -((periods-1)*3), 0)
	}
	since = time.Date(since.Year(), since.Month(), 1, 0, 0, 0, 0, time.UTC)

	periodKeys := buildPeriodKeys(granularity, periods)

	pushes, err := r.trendSeries(ctx, fmt.Sprintf(`
		SELECT %s AS period, COALESCE(SUM(count), 0)::int
		FROM contribution_days
		WHERE date >= $1
		GROUP BY date_trunc('%s', date)
		ORDER BY date_trunc('%s', date)`, periodKeyExpr(granularity, "date"), truncUnit(granularity), truncUnit(granularity)), since)
	if err != nil {
		return nil, err
	}

	repositories, err := r.trendSeries(ctx, fmt.Sprintf(`
		SELECT %s AS period, COUNT(DISTINCT r.id)::int
		FROM repos r
		WHERE EXISTS (SELECT 1 FROM developer_repos dr WHERE dr.repo_id = r.id)
		  AND r.pushed_at IS NOT NULL AND r.pushed_at >= $1
		GROUP BY date_trunc('%s', r.pushed_at)
		ORDER BY date_trunc('%s', r.pushed_at)`, periodKeyExpr(granularity, "r.pushed_at"), truncUnit(granularity), truncUnit(granularity)), since)
	if err != nil {
		return nil, err
	}

	developers, err := r.trendSeries(ctx, fmt.Sprintf(`
		SELECT %s AS period, COUNT(*)::int
		FROM developers
		WHERE created_at >= $1
		GROUP BY date_trunc('%s', created_at)
		ORDER BY date_trunc('%s', created_at)`, periodKeyExpr(granularity, "created_at"), truncUnit(granularity), truncUnit(granularity)), since)
	if err != nil {
		return nil, err
	}

	organizations, err := r.trendSeries(ctx, fmt.Sprintf(`
		SELECT %s AS period, COUNT(*)::int
		FROM developers
		WHERE created_at >= $1 AND company IS NOT NULL AND btrim(company) <> ''
		GROUP BY date_trunc('%s', created_at)
		ORDER BY date_trunc('%s', created_at)`, periodKeyExpr(granularity, "created_at"), truncUnit(granularity), truncUnit(granularity)), since)
	if err != nil {
		return nil, err
	}

	languages, err := r.nameCounts(ctx, `
		SELECT rl.language, COUNT(DISTINCT rl.repo_id)::int
		FROM repo_languages rl
		JOIN repos r ON r.id = rl.repo_id
		WHERE EXISTS (SELECT 1 FROM developer_repos dr WHERE dr.repo_id = r.id)
		GROUP BY rl.language
		ORDER BY 2 DESC
		LIMIT 12`)
	if err != nil {
		return nil, err
	}

	licenses, err := r.nameCounts(ctx, `
		SELECT COALESCE(NULLIF(btrim(r.license), ''), 'No License') AS license, COUNT(*)::int
		FROM repos r
		WHERE EXISTS (SELECT 1 FROM developer_repos dr WHERE dr.repo_id = r.id)
		GROUP BY 1
		ORDER BY 2 DESC
		LIMIT 12`)
	if err != nil {
		return nil, err
	}

	topOrgs, err := r.nameCounts(ctx, `
		SELECT COALESCE(NULLIF(btrim(company), ''), 'Independent') AS org, COUNT(*)::int
		FROM developers
		WHERE last_synced_at IS NOT NULL
		GROUP BY 1
		ORDER BY 2 DESC
		LIMIT 10`)
	if err != nil {
		return nil, err
	}

	topContribs, err := r.topContributors(ctx, 10)
	if err != nil {
		return nil, err
	}

	netNewStars, err := r.netNewStarsSeries(ctx, granularity, periodKeys)
	if err != nil {
		return nil, err
	}

	return &models.InnovationGraph{
		Granularity:      granularity,
		Pushes:           fillTrendSeries(periodKeys, pushes, granularity),
		Repositories:     fillTrendSeries(periodKeys, repositories, granularity),
		Developers:       fillTrendSeries(periodKeys, developers, granularity),
		Organizations:    fillTrendSeries(periodKeys, organizations, granularity),
		NetNewStars:      fillTrendSeries(periodKeys, netNewStars, granularity),
		Languages:        languages,
		Licenses:         licenses,
		TopOrganizations: topOrgs,
		TopContributors:  topContribs,
	}, nil
}

func (r *StatsRepo) trendSeries(ctx context.Context, query string, since time.Time) (map[string]int, error) {
	rows, err := r.db.Query(ctx, query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var period string
		var value int
		if err := rows.Scan(&period, &value); err != nil {
			return nil, err
		}
		out[period] = value
	}
	return out, nil
}

func truncUnit(granularity string) string {
	if granularity == "monthly" {
		return "month"
	}
	return "quarter"
}

func periodKeyExpr(granularity, column string) string {
	if granularity == "monthly" {
		return fmt.Sprintf("to_char(date_trunc('month', %s), 'YYYY-MM')", column)
	}
	return fmt.Sprintf("(extract(year from date_trunc('quarter', %s))::int || '-' || extract(quarter from date_trunc('quarter', %s))::int)", column, column)
}

func (r *StatsRepo) nameCounts(ctx context.Context, query string) ([]models.NameCount, error) {
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.NameCount
	for rows.Next() {
		var item models.NameCount
		if err := rows.Scan(&item.Name, &item.Count); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func (r *StatsRepo) topContributors(ctx context.Context, limit int) ([]models.ContributorStat, error) {
	rows, err := r.db.Query(ctx, `
		SELECT github_username, COALESCE(display_name, ''), activity_score, total_stars
		FROM developers
		WHERE last_synced_at IS NOT NULL
		ORDER BY activity_score DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ContributorStat
	for rows.Next() {
		var c models.ContributorStat
		if err := rows.Scan(&c.Username, &c.Name, &c.Score, &c.Stars); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func buildPeriodKeys(granularity string, count int) []string {
	now := time.Now().UTC()
	keys := make([]string, 0, count)
	if granularity == "monthly" {
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		start = start.AddDate(0, -(count - 1), 0)
		for i := 0; i < count; i++ {
			d := start.AddDate(0, i, 0)
			keys = append(keys, fmt.Sprintf("%d-%02d", d.Year(), int(d.Month())))
		}
		return keys
	}
	start := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	q := (int(now.Month())-1)/3 + 1
	start = start.AddDate(0, (q-count)*3, 0)
	for i := 0; i < count; i++ {
		d := start.AddDate(0, i*3, 0)
		qtr := (int(d.Month())-1)/3 + 1
		keys = append(keys, fmt.Sprintf("%d-%d", d.Year(), qtr))
	}
	return keys
}

func fillTrendSeries(keys []string, values map[string]int, granularity string) []models.TrendPoint {
	out := make([]models.TrendPoint, 0, len(keys))
	for _, key := range keys {
		out = append(out, models.TrendPoint{
			Period: key,
			Label:  periodLabel(key, granularity),
			Value:  values[key],
		})
	}
	return out
}

func periodLabel(key, granularity string) string {
	if granularity == "monthly" {
		var y, m int
		fmt.Sscanf(key, "%d-%d", &y, &m)
		return time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC).Format("Jan 2006")
	}
	var y, q int
	fmt.Sscanf(key, "%d-%d", &y, &q)
	return fmt.Sprintf("Q%d %d", q, y)
}
