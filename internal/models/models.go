package models

import "time"

type Developer struct {
	ID                 string     `json:"id"`
	GithubUsername     string     `json:"github_username"`
	Email              string     `json:"email"`
	DisplayName        *string    `json:"display_name,omitempty"`
	Notes              *string    `json:"notes,omitempty"`
	AvatarURL          *string    `json:"avatar_url,omitempty"`
	Bio                *string    `json:"bio,omitempty"`
	Location           *string    `json:"location,omitempty"`
	Company            *string    `json:"company,omitempty"`
	Website            *string    `json:"website,omitempty"`
	Followers          int        `json:"followers"`
	Following          int        `json:"following"`
	PublicRepos        int        `json:"public_repos"`
	TotalStars         int        `json:"total_stars"`
	ActivityScore      float64    `json:"activity_score"`
	VerificationStatus string     `json:"verification_status"`
	LastSyncedAt       *time.Time `json:"last_synced_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type AdminUser struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateDeveloperInput struct {
	GithubUsername string  `json:"github_username"`
	Email          string  `json:"email"`
	DisplayName    *string `json:"display_name,omitempty"`
	Notes          *string `json:"notes,omitempty"`
}

type UpdateDeveloperInput struct {
	Email       *string `json:"email,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}
