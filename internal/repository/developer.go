package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDeveloperDuplicate = errors.New("developer already exists")

type DeveloperRepo struct {
	db *pgxpool.Pool
}

func NewDeveloperRepo(db *pgxpool.Pool) *DeveloperRepo {
	return &DeveloperRepo{db: db}
}

func (r *DeveloperRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM developers WHERE lower(github_username) = lower($1))`, username).
		Scan(&exists)
	return exists, err
}

func (r *DeveloperRepo) Create(ctx context.Context, in models.CreateDeveloperInput) (*models.Developer, error) {
	exists, err := r.ExistsByUsername(ctx, in.GithubUsername)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDeveloperDuplicate
	}

	var d models.Developer
	err = r.db.QueryRow(ctx, `
		INSERT INTO developers (github_username, email, display_name, notes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, github_username, email, display_name, notes,
		          avatar_url, bio, location, company, website,
		          followers, following, public_repos, total_stars,
		          activity_score, verification_status, last_synced_at, created_at, updated_at`,
		in.GithubUsername, in.Email, in.DisplayName, in.Notes,
	).Scan(
		&d.ID, &d.GithubUsername, &d.Email, &d.DisplayName, &d.Notes,
		&d.AvatarURL, &d.Bio, &d.Location, &d.Company, &d.Website,
		&d.Followers, &d.Following, &d.PublicRepos, &d.TotalStars,
		&d.ActivityScore, &d.VerificationStatus, &d.LastSyncedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDeveloperDuplicate
		}
		return nil, fmt.Errorf("create developer: %w", err)
	}
	return &d, nil
}

func (r *DeveloperRepo) List(ctx context.Context) ([]models.Developer, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, github_username, email, display_name, notes,
		       avatar_url, bio, location, company, website,
		       followers, following, public_repos, total_stars,
		       activity_score, verification_status, last_synced_at, created_at, updated_at
		FROM developers
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list developers: %w", err)
	}
	defer rows.Close()

	var devs []models.Developer
	for rows.Next() {
		var d models.Developer
		if err := rows.Scan(
			&d.ID, &d.GithubUsername, &d.Email, &d.DisplayName, &d.Notes,
			&d.AvatarURL, &d.Bio, &d.Location, &d.Company, &d.Website,
			&d.Followers, &d.Following, &d.PublicRepos, &d.TotalStars,
			&d.ActivityScore, &d.VerificationStatus, &d.LastSyncedAt, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan developer: %w", err)
		}
		devs = append(devs, d)
	}
	return devs, nil
}

func (r *DeveloperRepo) GetByID(ctx context.Context, id string) (*models.Developer, error) {
	var d models.Developer
	err := r.db.QueryRow(ctx, `
		SELECT id, github_username, email, display_name, notes,
		       avatar_url, bio, location, company, website,
		       followers, following, public_repos, total_stars,
		       activity_score, builder_score, contributor_score, reviewer_score, community_score,
		       pr_contributions, issue_contributions, review_contributions,
		       contribution_period_start::text, contribution_period_end::text,
		       verification_status, last_synced_at, created_at, updated_at
		FROM developers WHERE id = $1`, id,
	).Scan(
		&d.ID, &d.GithubUsername, &d.Email, &d.DisplayName, &d.Notes,
		&d.AvatarURL, &d.Bio, &d.Location, &d.Company, &d.Website,
		&d.Followers, &d.Following, &d.PublicRepos, &d.TotalStars,
		&d.ActivityScore, &d.BuilderScore, &d.ContributorScore, &d.ReviewerScore, &d.CommunityScore,
		&d.PRContributions, &d.IssueContributions, &d.ReviewContributions,
		&d.ContributionPeriodStart, &d.ContributionPeriodEnd,
		&d.VerificationStatus, &d.LastSyncedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get developer: %w", err)
	}
	return &d, nil
}

func (r *DeveloperRepo) Update(ctx context.Context, id string, in models.UpdateDeveloperInput) (*models.Developer, error) {
	var d models.Developer
	err := r.db.QueryRow(ctx, `
		UPDATE developers SET
		  email        = COALESCE($2, email),
		  display_name = COALESCE($3, display_name),
		  notes        = COALESCE($4, notes),
		  updated_at   = NOW()
		WHERE id = $1
		RETURNING id, github_username, email, display_name, notes,
		          avatar_url, bio, location, company, website,
		          followers, following, public_repos, total_stars,
		          activity_score, verification_status, last_synced_at, created_at, updated_at`,
		id, in.Email, in.DisplayName, in.Notes,
	).Scan(
		&d.ID, &d.GithubUsername, &d.Email, &d.DisplayName, &d.Notes,
		&d.AvatarURL, &d.Bio, &d.Location, &d.Company, &d.Website,
		&d.Followers, &d.Following, &d.PublicRepos, &d.TotalStars,
		&d.ActivityScore, &d.VerificationStatus, &d.LastSyncedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update developer: %w", err)
	}
	return &d, nil
}

func (r *DeveloperRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM developers WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete developer: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("developer not found")
	}
	return nil
}
