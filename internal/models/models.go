package models

import "time"

type Developer struct {
	ID                 string     `json:"id"`
	GithubUsername     string     `json:"github_username"`
	Email              *string    `json:"email,omitempty"`
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
	Email          *string `json:"email,omitempty"`
	DisplayName    *string `json:"display_name,omitempty"`
	Notes          *string `json:"notes,omitempty"`
}

type UpdateDeveloperInput struct {
	Email       *string `json:"email,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type Overview struct {
	TotalDevelopers    int   `json:"total_developers"`
	TotalRepos         int   `json:"total_repos"`
	TotalStars         int   `json:"total_stars"`
	TotalContributions int64 `json:"total_contributions"`
}

type LanguageStat struct {
	Language  string `json:"language"`
	Bytes     int64  `json:"bytes"`
	RepoCount int    `json:"repo_count"`
}

type PublicRepo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	FullName    string  `json:"full_name"`
	Owner       string  `json:"owner"`
	Description string  `json:"description"`
	URL         string  `json:"url"`
	Language    *string `json:"language,omitempty"`
	Stars       int     `json:"stars"`
	Forks       int     `json:"forks"`
	IsFork      bool    `json:"is_fork"`
	PushedAt    *string `json:"pushed_at,omitempty"`
}

type ActivityEvent struct {
	Type       string `json:"type"`
	Username   string `json:"username"`
	Repo       string `json:"repo,omitempty"`
	Message    string `json:"message"`
	OccurredAt string `json:"occurred_at"`
}

type OSSStats struct {
	OriginalProjects   int     `json:"original_projects"`
	ForkProjects       int     `json:"fork_projects"`
	TotalStars         int     `json:"total_stars"`
	TotalForksReceived int     `json:"total_forks_received"`
	Contributors       int     `json:"contributors"`
	TopLanguage        *string `json:"top_language,omitempty"`
}

type ContributionDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type CommunityActivityDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
