package repository

import (
	"context"
	"fmt"

	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepo struct {
	db *pgxpool.Pool
}

func NewAdminRepo(db *pgxpool.Pool) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) GetByEmail(ctx context.Context, email string) (*models.AdminUser, error) {
	var a models.AdminUser
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at FROM admin_users WHERE email = $1`, email,
	).Scan(&a.ID, &a.Email, &a.PasswordHash, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get admin: %w", err)
	}
	return &a, nil
}

func (r *AdminRepo) Create(ctx context.Context, email, passwordHash string) (*models.AdminUser, error) {
	var a models.AdminUser
	err := r.db.QueryRow(ctx,
		`INSERT INTO admin_users (email, password_hash) VALUES ($1, $2)
		 RETURNING id, email, password_hash, created_at`,
		email, passwordHash,
	).Scan(&a.ID, &a.Email, &a.PasswordHash, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create admin: %w", err)
	}
	return &a, nil
}

func (r *AdminRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&n)
	return n, err
}
