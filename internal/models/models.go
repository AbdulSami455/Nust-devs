package models

import "time"

type Developer struct {
	ID                      string     `json:"id"`
	GithubUsername          string     `json:"github_username"`
	Email                   *string    `json:"email,omitempty"`
	DisplayName             *string    `json:"display_name,omitempty"`
	Notes                   *string    `json:"notes,omitempty"`
	AvatarURL               *string    `json:"avatar_url,omitempty"`
	Bio                     *string    `json:"bio,omitempty"`
	Location                *string    `json:"location,omitempty"`
	Company                 *string    `json:"company,omitempty"`
	Website                 *string    `json:"website,omitempty"`
	Followers               int        `json:"followers"`
	Following               int        `json:"following"`
	PublicRepos             int        `json:"public_repos"`
	ReadmeRepos             int        `json:"readme_repos"`
	TotalStars              int        `json:"total_stars"`
	ActivityScore           float64    `json:"activity_score"`
	BuilderScore            float64    `json:"builder_score"`
	ContributorScore        float64    `json:"contributor_score"`
	ReviewerScore           float64    `json:"reviewer_score"`
	CommunityScore          float64    `json:"community_score"`
	PRContributions         int        `json:"pr_contributions"`
	IssueContributions      int        `json:"issue_contributions"`
	ReviewContributions     int        `json:"review_contributions"`
	ContributionPeriodStart *string    `json:"contribution_period_start,omitempty"`
	ContributionPeriodEnd   *string    `json:"contribution_period_end,omitempty"`
	CurrentStreak           int        `json:"current_streak"`
	LongestStreak           int        `json:"longest_streak"`
	StreakMultiplier        float64    `json:"streak_multiplier"`
	XP                      int        `json:"xp"`
	PowerLevel              int        `json:"power_level"`
	VerificationStatus      string     `json:"verification_status"`
	LastSyncedAt            *time.Time `json:"last_synced_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
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

type DeveloperRequest struct {
	ID             string     `json:"id"`
	GithubUsername string     `json:"github_username"`
	Email          *string    `json:"email,omitempty"`
	DisplayName    *string    `json:"display_name,omitempty"`
	Batch          *string    `json:"batch,omitempty"`
	Course         *string    `json:"course,omitempty"`
	Message        *string    `json:"message,omitempty"`
	Status         string     `json:"status"`
	AdminNotes     *string    `json:"admin_notes,omitempty"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type SubmitProfileRequestInput struct {
	GithubUsername string  `json:"github_username"`
	Email          *string `json:"email,omitempty"`
	DisplayName    *string `json:"display_name,omitempty"`
	Batch          *string `json:"batch,omitempty"`
	Course         *string `json:"course,omitempty"`
	Message        *string `json:"message,omitempty"`
}

type ReviewProfileRequestInput struct {
	AdminNotes *string `json:"admin_notes,omitempty"`
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
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	FullName       string       `json:"full_name"`
	Owner          string       `json:"owner"`
	Description    string       `json:"description"`
	URL            string       `json:"url"`
	Language       *string      `json:"language,omitempty"`
	Stars          int          `json:"stars"`
	Forks          int          `json:"forks"`
	IsFork         bool         `json:"is_fork"`
	PushedAt       *string      `json:"pushed_at,omitempty"`
	StarsGrowth30d *int         `json:"stars_growth_30d,omitempty"`
	ForksGrowth30d *int         `json:"forks_growth_30d,omitempty"`
	Sparkline      []SparkPoint `json:"sparkline,omitempty"`
}

type ProjectSummary struct {
	RepoID       string    `json:"repo_id"`
	Headline     string    `json:"headline"`
	Summary      string    `json:"summary"`
	ModelVersion string    `json:"model_version"`
	GeneratedAt  time.Time `json:"generated_at"`
}

type RankInsight struct {
	DeveloperID  string    `json:"developer_id"`
	Headline     string    `json:"headline"`
	Summary      string    `json:"summary"`
	Highlights   []string  `json:"highlights"`
	ModelVersion string    `json:"model_version"`
	GeneratedAt  time.Time `json:"generated_at"`
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

type RepoContributionStat struct {
	RepoFullName string `json:"repo_full_name"`
	RepoURL      string `json:"repo_url"`
	PullRequests int    `json:"pull_requests"`
	Issues       int    `json:"issues"`
	Reviews      int    `json:"reviews"`
	Total        int    `json:"total"`
}

type ContributionStats struct {
	PeriodStart  string                 `json:"period_start"`
	PeriodEnd    string                 `json:"period_end"`
	PullRequests int                    `json:"pull_requests"`
	Issues       int                    `json:"issues"`
	Reviews      int                    `json:"reviews"`
	ByRepository []RepoContributionStat `json:"by_repository"`
}

type CommunityActivityDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type TrendPoint struct {
	Period string `json:"period"`
	Label  string `json:"label"`
	Value  int    `json:"value"`
}

type NameCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type ContributorStat struct {
	Username string  `json:"username"`
	Name     string  `json:"name,omitempty"`
	Score    float64 `json:"score"`
	Stars    int     `json:"stars"`
}

type SparkPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

type StreakSummary struct {
	DevsOn7PlusStreak   int `json:"devs_on_7plus_streak"`
	DevsOn30PlusStreak  int `json:"devs_on_30plus_streak"`
	LongestActiveStreak int `json:"longest_active_streak"`
}

type DevOfMonthWinner struct {
	Year           int       `json:"year"`
	Month          int       `json:"month"`
	Score          float64   `json:"score"`
	ActivityPoints int       `json:"activity_points"`
	RankGain       int       `json:"rank_gain"`
	StarsGained    int       `json:"stars_gained"`
	PowerTitle     string    `json:"power_title,omitempty"`
	Developer      Developer `json:"developer"`
}

type WrappedReport struct {
	Year               int         `json:"year"`
	Username           string      `json:"username"`
	DisplayName        *string     `json:"display_name,omitempty"`
	AvatarURL          *string     `json:"avatar_url,omitempty"`
	TotalContributions int         `json:"total_contributions"`
	TopRepo            string      `json:"top_repo,omitempty"`
	TopRepoStars       int         `json:"top_repo_stars"`
	RankStart          int         `json:"rank_start"`
	RankEnd            int         `json:"rank_end"`
	RankChange         int         `json:"rank_change"`
	ActivityPercentile int         `json:"activity_percentile"`
	TopLanguages       []NameCount `json:"top_languages"`
	PowerLevel         int         `json:"power_level"`
	PowerTitle         string      `json:"power_title"`
	XP                 int         `json:"xp"`
	CurrentStreak      int         `json:"current_streak"`
	LongestStreak      int         `json:"longest_streak"`
	PRContributions    int         `json:"pr_contributions"`
	TotalStars         int         `json:"total_stars"`
	PublicRepos        int         `json:"public_repos"`
	Highlights         []string    `json:"highlights"`
}

type LeaderboardEntry struct {
	Developer
	Rank          int          `json:"rank"`
	RankDelta7d   *int         `json:"rank_delta_7d,omitempty"`
	RankDelta30d  *int         `json:"rank_delta_30d,omitempty"`
	ScoreDelta7d  *float64     `json:"score_delta_7d,omitempty"`
	ScoreDelta30d *float64     `json:"score_delta_30d,omitempty"`
	Sparkline     []SparkPoint `json:"sparkline,omitempty"`
}

type InnovationGraph struct {
	Granularity      string            `json:"granularity"`
	Pushes           []TrendPoint      `json:"pushes"`
	Repositories     []TrendPoint      `json:"repositories"`
	Developers       []TrendPoint      `json:"developers"`
	Organizations    []TrendPoint      `json:"organizations"`
	NetNewStars      []TrendPoint      `json:"net_new_stars"`
	Languages        []NameCount       `json:"languages"`
	Licenses         []NameCount       `json:"licenses"`
	TopOrganizations []NameCount       `json:"top_organizations"`
	TopContributors  []ContributorStat `json:"top_contributors"`
}

type AuditLog struct {
	ID           string         `json:"id"`
	ActorType    string         `json:"actor_type"`
	ActorID      string         `json:"actor_id,omitempty"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id,omitempty"`
	Method       string         `json:"method,omitempty"`
	Path         string         `json:"path,omitempty"`
	StatusCode   int            `json:"status_code"`
	IP           string         `json:"ip,omitempty"`
	UserAgent    string         `json:"user_agent,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

type AgentRun struct {
	ID            string     `json:"id"`
	SessionID     string     `json:"session_id"`
	AgentName     string     `json:"agent_name"`
	UserMessage   string     `json:"user_message"`
	InputHash     string     `json:"input_hash"`
	Status        string     `json:"status"`
	IP            string     `json:"ip,omitempty"`
	UserAgent     string     `json:"user_agent,omitempty"`
	ToolCalls     int        `json:"tool_calls"`
	LatencyMS     int        `json:"latency_ms"`
	ErrorMessage  string     `json:"error_message,omitempty"`
	ResponseChars int        `json:"response_chars"`
	CreatedAt     time.Time  `json:"created_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
}

type AgentRunEvent struct {
	ID        string         `json:"id"`
	RunID     string         `json:"run_id"`
	EventType string         `json:"event_type"`
	ToolName  string         `json:"tool_name,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	LatencyMS int            `json:"latency_ms"`
	Success   bool           `json:"success"`
	CreatedAt time.Time      `json:"created_at"`
}

type ObservabilityOverview struct {
	TotalAuditLogs      int        `json:"total_audit_logs"`
	AgentRuns24h        int        `json:"agent_runs_24h"`
	AgentSuccessRate24h float64    `json:"agent_success_rate_24h"`
	AvgAgentLatencyMS   int        `json:"avg_agent_latency_ms"`
	ActiveAgentRuns     int        `json:"active_agent_runs"`
	LastAgentRunAt      *time.Time `json:"last_agent_run_at,omitempty"`
}

type ObservabilityResponse struct {
	Overview     ObservabilityOverview `json:"overview"`
	RecentLogs   []AuditLog            `json:"recent_logs"`
	RecentRuns   []AgentRun            `json:"recent_runs"`
	RecentEvents []AgentRunEvent       `json:"recent_events"`
}
