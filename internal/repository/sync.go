package repository

import (
	"context"
	"fmt"
	"time"

	gh "github.com/abdulsami/nust-devs/internal/github"
	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SyncRepo struct {
	db *pgxpool.Pool
}

func NewSyncRepo(db *pgxpool.Pool) *SyncRepo {
	return &SyncRepo{db: db}
}

func (r *SyncRepo) UpsertDeveloperProfile(ctx context.Context, devID string, u *gh.User) error {
	_, err := r.db.Exec(ctx, `
		UPDATE developers SET
			avatar_url   = $2,
			bio          = $3,
			location     = $4,
			company      = $5,
			website      = $6,
			followers    = $7,
			following    = $8,
			public_repos = $9,
			updated_at   = NOW()
		WHERE id = $1`,
		devID, u.AvatarURL, u.Bio, u.Location, u.Company, u.Blog,
		u.Followers, u.Following, u.PublicRepos,
	)
	return err
}

func (r *SyncRepo) UpsertRepo(ctx context.Context, repo gh.Repo) (string, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO repos (github_id, owner, name, full_name, description, url, language, stars, forks, is_fork, pushed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (github_id) DO UPDATE SET
			description = EXCLUDED.description,
			language    = EXCLUDED.language,
			stars       = EXCLUDED.stars,
			forks       = EXCLUDED.forks,
			pushed_at   = EXCLUDED.pushed_at,
			updated_at  = NOW()
		RETURNING id`,
		repo.ID, ownerOf(repo.FullName), repo.Name, repo.FullName,
		repo.Description, repo.HTMLURL, repo.Language,
		repo.StargazersCount, repo.ForksCount, repo.Fork, repo.PushedAt,
	).Scan(&id)
	return id, err
}

func (r *SyncRepo) RepoIDByGithubID(ctx context.Context, githubID int64) (string, error) {
	var id string
	err := r.db.QueryRow(ctx, `SELECT id FROM repos WHERE github_id = $1`, githubID).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("repo lookup github_id=%d: %w", githubID, err)
	}
	return id, nil
}

func (r *SyncRepo) LinkDeveloperRepo(ctx context.Context, devID, repoID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO developer_repos (developer_id, repo_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`, devID, repoID)
	return err
}

func (r *SyncRepo) UpsertRepoLanguages(ctx context.Context, repoID string, langs map[string]int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM repo_languages WHERE repo_id = $1`, repoID)
	if err != nil {
		return err
	}
	for lang, bytes := range langs {
		if _, err := r.db.Exec(ctx,
			`INSERT INTO repo_languages (repo_id, language, bytes) VALUES ($1, $2, $3)`,
			repoID, lang, bytes,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *SyncRepo) UpsertContributionDays(ctx context.Context, devID string, days []gh.ContributionDay) error {
	for _, d := range days {
		if _, err := r.db.Exec(ctx, `
			INSERT INTO contribution_days (developer_id, date, count) VALUES ($1, $2, $3)
			ON CONFLICT (developer_id, date) DO UPDATE SET count = EXCLUDED.count`,
			devID, d.Date, d.Count,
		); err != nil {
			return fmt.Errorf("upsert contribution day %s: %w", d.Date, err)
		}
	}
	return nil
}

func (r *SyncRepo) WriteSnapshot(ctx context.Context, dev *models.Developer) error {
	today := time.Now().Format("2006-01-02")
	_, err := r.db.Exec(ctx, `
		INSERT INTO developer_snapshots (developer_id, snapshot_date, public_repos, total_stars, followers, activity_score)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (developer_id, snapshot_date) DO UPDATE SET
			public_repos   = EXCLUDED.public_repos,
			total_stars    = EXCLUDED.total_stars,
			followers      = EXCLUDED.followers,
			activity_score = EXCLUDED.activity_score`,
		dev.ID, today, dev.PublicRepos, dev.TotalStars, dev.Followers, dev.ActivityScore,
	)
	return err
}

func (r *SyncRepo) UpdateLastSynced(ctx context.Context, devID string) error {
	_, err := r.db.Exec(ctx, `UPDATE developers SET last_synced_at = NOW() WHERE id = $1`, devID)
	return err
}

func (r *SyncRepo) RecomputeActivityScore(ctx context.Context, devID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE developers SET activity_score = (
			(SELECT COALESCE(SUM(count), 0) FROM contribution_days
			 WHERE developer_id = $1 AND date >= CURRENT_DATE - INTERVAL '90 days') * 3
			+ (public_repos * 2)
			+ (total_stars * 0.1)
			+ (SELECT COALESCE(SUM(count), 0) FROM contribution_days
			   WHERE developer_id = $1 AND date >= CURRENT_DATE - INTERVAL '30 days') * 5
		)
		WHERE id = $1`, devID)
	return err
}

func (r *SyncRepo) UpdateTotalStars(ctx context.Context, devID string, repos []gh.Repo) error {
	total := 0
	for _, repo := range repos {
		total += repo.StargazersCount
	}
	_, err := r.db.Exec(ctx, `UPDATE developers SET total_stars = $2 WHERE id = $1`, devID, total)
	return err
}

func ownerOf(fullName string) string {
	for i, c := range fullName {
		if c == '/' {
			return fullName[:i]
		}
	}
	return fullName
}
