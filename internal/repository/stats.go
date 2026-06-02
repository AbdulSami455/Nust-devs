package repository

import (
	"context"
	"fmt"

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
	verification_status, last_synced_at, created_at, updated_at`

func scanPublicDeveloper(row interface {
	Scan(...any) error
}, d *models.Developer) error {
	return row.Scan(
		&d.ID, &d.GithubUsername, &d.DisplayName, &d.AvatarURL, &d.Bio,
		&d.Location, &d.Company, &d.Website,
		&d.Followers, &d.Following, &d.PublicRepos, &d.TotalStars, &d.ActivityScore,
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

func (r *StatsRepo) GetDeveloperRepos(ctx context.Context, devID string) ([]models.PublicRepo, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.id, r.name, r.full_name, r.description, r.url, r.language, r.stars, r.forks, r.is_fork
		FROM repos r
		JOIN developer_repos dr ON dr.repo_id = r.id
		WHERE dr.developer_id = $1
		ORDER BY r.stars DESC`, devID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var repos []models.PublicRepo
	for rows.Next() {
		var rp models.PublicRepo
		if err := rows.Scan(&rp.ID, &rp.Name, &rp.FullName, &rp.Description, &rp.URL,
			&rp.Language, &rp.Stars, &rp.Forks, &rp.IsFork); err != nil {
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
	"activity_score": "activity_score",
	"total_stars":    "total_stars",
	"public_repos":   "public_repos",
	"followers":      "followers",
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
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT ON (r.id) r.id, r.name, r.full_name, r.description, r.url, r.language, r.stars, r.forks, r.is_fork
		FROM repos r
		JOIN developer_repos dr ON dr.repo_id = r.id
		WHERE r.is_fork = false
		ORDER BY r.id, r.stars DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var repos []models.PublicRepo
	for rows.Next() {
		var rp models.PublicRepo
		if err := rows.Scan(&rp.ID, &rp.Name, &rp.FullName, &rp.Description, &rp.URL,
			&rp.Language, &rp.Stars, &rp.Forks, &rp.IsFork); err != nil {
			return nil, err
		}
		repos = append(repos, rp)
	}
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
