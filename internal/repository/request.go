package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDeveloperExists = errors.New("developer already registered")
	ErrRequestPending  = errors.New("a pending request already exists for this username")
)

type RequestRepo struct {
	db *pgxpool.Pool
}

func NewRequestRepo(db *pgxpool.Pool) *RequestRepo {
	return &RequestRepo{db: db}
}

func (r *RequestRepo) DeveloperExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM developers WHERE lower(github_username) = lower($1))`, username).
		Scan(&exists)
	return exists, err
}

func (r *RequestRepo) PendingRequestExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM developer_requests
			WHERE lower(github_username) = lower($1) AND status = 'pending'
		)`, username).Scan(&exists)
	return exists, err
}

func (r *RequestRepo) Create(ctx context.Context, in models.SubmitProfileRequestInput) (*models.DeveloperRequest, error) {
	exists, err := r.DeveloperExists(ctx, in.GithubUsername)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDeveloperExists
	}
	pending, err := r.PendingRequestExists(ctx, in.GithubUsername)
	if err != nil {
		return nil, err
	}
	if pending {
		return nil, ErrRequestPending
	}

	var req models.DeveloperRequest
	err = r.db.QueryRow(ctx, `
		INSERT INTO developer_requests (github_username, email, display_name, batch, course, message)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, github_username, email, display_name, batch, course, message, status, admin_notes, reviewed_at, created_at`,
		in.GithubUsername, in.Email, in.DisplayName, in.Batch, in.Course, in.Message,
	).Scan(
		&req.ID, &req.GithubUsername, &req.Email, &req.DisplayName, &req.Batch, &req.Course, &req.Message,
		&req.Status, &req.AdminNotes, &req.ReviewedAt, &req.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	return &req, nil
}

func (r *RequestRepo) List(ctx context.Context, status string) ([]models.DeveloperRequest, error) {
	var rows pgx.Rows
	var err error
	if status == "" {
		rows, err = r.db.Query(ctx, `
			SELECT id, github_username, email, display_name, batch, course, message, status, admin_notes, reviewed_at, created_at
			FROM developer_requests ORDER BY created_at DESC`)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, github_username, email, display_name, batch, course, message, status, admin_notes, reviewed_at, created_at
			FROM developer_requests WHERE status = $1 ORDER BY created_at DESC`, status)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.DeveloperRequest
	for rows.Next() {
		var req models.DeveloperRequest
		if err := rows.Scan(
			&req.ID, &req.GithubUsername, &req.Email, &req.DisplayName, &req.Batch, &req.Course, &req.Message,
			&req.Status, &req.AdminNotes, &req.ReviewedAt, &req.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, req)
	}
	return out, nil
}

func (r *RequestRepo) GetByID(ctx context.Context, id string) (*models.DeveloperRequest, error) {
	var req models.DeveloperRequest
	err := r.db.QueryRow(ctx, `
		SELECT id, github_username, email, display_name, batch, course, message, status, admin_notes, reviewed_at, created_at
		FROM developer_requests WHERE id = $1`, id,
	).Scan(
		&req.ID, &req.GithubUsername, &req.Email, &req.DisplayName, &req.Batch, &req.Course, &req.Message,
		&req.Status, &req.AdminNotes, &req.ReviewedAt, &req.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("request not found")
		}
		return nil, err
	}
	return &req, nil
}

func (r *RequestRepo) SetStatus(ctx context.Context, id, status string, adminNotes *string) (*models.DeveloperRequest, error) {
	var req models.DeveloperRequest
	err := r.db.QueryRow(ctx, `
		UPDATE developer_requests SET
			status = $2,
			admin_notes = COALESCE($3, admin_notes),
			reviewed_at = NOW()
		WHERE id = $1 AND status = 'pending'
		RETURNING id, github_username, email, display_name, batch, course, message, status, admin_notes, reviewed_at, created_at`,
		id, status, adminNotes,
	).Scan(
		&req.ID, &req.GithubUsername, &req.Email, &req.DisplayName, &req.Batch, &req.Course, &req.Message,
		&req.Status, &req.AdminNotes, &req.ReviewedAt, &req.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("request not found or already reviewed")
		}
		return nil, err
	}
	return &req, nil
}
