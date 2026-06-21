package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/repository"
)

const comparisonSystemPrompt = `You are generating a side-by-side comparison card for two tracked NUST developers.

Return ONLY a JSON object in this exact format:
{
  "headline": "short comparison headline",
  "summary": "2-3 sentences comparing their strengths, scale, and focus",
  "takeaways": ["bullet 1", "bullet 2", "bullet 3"],
  "shared_strengths": ["shared strength 1", "shared strength 2"],
  "verdict": "one concise recommendation or distinction"
}

Rules:
- Use only the JSON snapshot below. Do not call tools.
- Keep takeaways specific and grounded in the numbers or repo data provided.
- If a field cannot be inferred confidently, leave it out of the supporting sentence rather than inventing it.
- Output only the JSON object, nothing else.`

type DeveloperComparison struct {
	Left            models.Developer `json:"left"`
	Right           models.Developer `json:"right"`
	LeftRank        *int             `json:"left_rank,omitempty"`
	RightRank       *int             `json:"right_rank,omitempty"`
	Headline        string           `json:"headline"`
	Summary         string           `json:"summary"`
	Takeaways       []string         `json:"takeaways"`
	SharedStrengths []string         `json:"shared_strengths"`
	Verdict         string           `json:"verdict"`
	Source          string           `json:"source"`
	ModelVersion    string           `json:"model_version"`
	GeneratedAt     time.Time        `json:"generated_at"`
}

type comparisonJSON struct {
	Headline        string   `json:"headline"`
	Summary         string   `json:"summary"`
	Takeaways       []string `json:"takeaways"`
	SharedStrengths []string `json:"shared_strengths"`
	Verdict         string   `json:"verdict"`
}

type compareRepo struct {
	Name        string  `json:"name"`
	FullName    string  `json:"full_name"`
	Description string  `json:"description,omitempty"`
	Language    *string `json:"language,omitempty"`
	Stars       int     `json:"stars"`
	Forks       int     `json:"forks"`
	IsFork      bool    `json:"is_fork"`
}

type compareProfile struct {
	Username            string        `json:"username"`
	DisplayName         *string       `json:"display_name,omitempty"`
	Rank                *int          `json:"rank,omitempty"`
	ActivityScore       float64       `json:"activity_score"`
	BuilderScore        float64       `json:"builder_score"`
	ContributorScore    float64       `json:"contributor_score"`
	ReviewerScore       float64       `json:"reviewer_score"`
	CommunityScore      float64       `json:"community_score"`
	PublicRepos         int           `json:"public_repos"`
	ReadmeRepos         int           `json:"readme_repos"`
	TotalStars          int           `json:"total_stars"`
	Followers           int           `json:"followers"`
	Following           int           `json:"following"`
	CurrentStreak       int           `json:"current_streak"`
	LongestStreak       int           `json:"longest_streak"`
	XP                  int           `json:"xp"`
	PowerLevel          int           `json:"power_level"`
	PRContributions     int           `json:"pr_contributions"`
	IssueContributions  int           `json:"issue_contributions"`
	ReviewContributions int           `json:"review_contributions"`
	ContributionTotal   int           `json:"contribution_total_365d"`
	ActiveDays          int           `json:"active_days_365d"`
	TopLanguage         string        `json:"top_language,omitempty"`
	TopRepos            []compareRepo `json:"top_repos,omitempty"`
	TopLanguages        []string      `json:"top_languages,omitempty"`
}

type comparisonSnapshot struct {
	Left  compareProfile `json:"left"`
	Right compareProfile `json:"right"`
}

// CompareService generates a side-by-side comparison for two developers.
type CompareService struct {
	chat  *ChatService
	stats *repository.StatsRepo
	model string
}

func NewCompareService(chat *ChatService, stats *repository.StatsRepo, model string) *CompareService {
	return &CompareService{chat: chat, stats: stats, model: model}
}

func (s *CompareService) Get(ctx context.Context, leftUsername, rightUsername string) (*DeveloperComparison, error) {
	slog.Info("comparison requested", "left", leftUsername, "right", rightUsername)
	left, right, err := s.fetchDevelopers(ctx, leftUsername, rightUsername)
	if err != nil {
		slog.Warn("comparison lookup failed", "left", leftUsername, "right", rightUsername, "err", err)
		return nil, err
	}

	ranks := s.fetchRanks(ctx)
	leftRank := rankForUsername(ranks, left.GithubUsername)
	rightRank := rankForUsername(ranks, right.GithubUsername)

	leftRepos, _ := s.stats.GetDeveloperRepos(ctx, left.ID)
	rightRepos, _ := s.stats.GetDeveloperRepos(ctx, right.ID)
	leftContribs, _ := s.stats.GetContributions(ctx, left.ID)
	rightContribs, _ := s.stats.GetContributions(ctx, right.ID)

	leftTopRepos, leftTopLanguage, leftLanguages := summarizeRepos(leftRepos)
	rightTopRepos, rightTopLanguage, rightLanguages := summarizeRepos(rightRepos)

	leftProfile := compareProfileFromData(*left, leftRank, leftTopRepos, leftTopLanguage, leftLanguages, leftContribs)
	rightProfile := compareProfileFromData(*right, rightRank, rightTopRepos, rightTopLanguage, rightLanguages, rightContribs)

	snapshot := comparisonSnapshot{Left: leftProfile, Right: rightProfile}

	aiResult, err := s.generate(ctx, snapshot)
	if err != nil {
		slog.Warn("comparison ai generation failed", "left", leftUsername, "right", rightUsername, "err", err)
		fallback := fallbackComparison(leftProfile, rightProfile)
		slog.Info("comparison fallback used", "left", leftUsername, "right", rightUsername)
		return &DeveloperComparison{
			Left:            *left,
			Right:           *right,
			LeftRank:        leftRank,
			RightRank:       rightRank,
			Headline:        fallback.Headline,
			Summary:         fallback.Summary,
			Takeaways:       fallback.Takeaways,
			SharedStrengths: fallback.SharedStrengths,
			Verdict:         fallback.Verdict,
			Source:          "fallback",
			ModelVersion:    s.model,
			GeneratedAt:     time.Now().UTC(),
		}, nil
	}

	slog.Info("comparison ai generated", "left", leftUsername, "right", rightUsername)
	return &DeveloperComparison{
		Left:            *left,
		Right:           *right,
		LeftRank:        leftRank,
		RightRank:       rightRank,
		Headline:        aiResult.Headline,
		Summary:         aiResult.Summary,
		Takeaways:       aiResult.Takeaways,
		SharedStrengths: aiResult.SharedStrengths,
		Verdict:         aiResult.Verdict,
		Source:          "ai",
		ModelVersion:    s.model,
		GeneratedAt:     time.Now().UTC(),
	}, nil
}

func (s *CompareService) fetchDevelopers(ctx context.Context, leftUsername, rightUsername string) (*models.Developer, *models.Developer, error) {
	left, err := s.stats.GetDeveloperByUsername(ctx, leftUsername)
	if err != nil {
		return nil, nil, fmt.Errorf("left developer not found")
	}
	right, err := s.stats.GetDeveloperByUsername(ctx, rightUsername)
	if err != nil {
		return nil, nil, fmt.Errorf("right developer not found")
	}
	return left, right, nil
}

func (s *CompareService) fetchRanks(ctx context.Context) map[string]int {
	entries, err := s.stats.GetLeaderboardWithTrends(ctx, "activity_score", "", 0, 1, 1000)
	if err != nil {
		return map[string]int{}
	}
	ranks := make(map[string]int, len(entries))
	for _, entry := range entries {
		ranks[entry.GithubUsername] = entry.Rank
	}
	return ranks
}

func rankForUsername(ranks map[string]int, username string) *int {
	if rank, ok := ranks[username]; ok {
		r := rank
		return &r
	}
	return nil
}

func summarizeRepos(repos []models.PublicRepo) ([]compareRepo, string, []string) {
	top := make([]compareRepo, 0, len(repos))
	languages := map[string]int{}
	for _, repo := range repos {
		if repo.Language != nil && *repo.Language != "" {
			languages[*repo.Language]++
		}
		top = append(top, compareRepo{
			Name:        repo.Name,
			FullName:    repo.FullName,
			Description: repo.Description,
			Language:    repo.Language,
			Stars:       repo.Stars,
			Forks:       repo.Forks,
			IsFork:      repo.IsFork,
		})
	}
	sort.SliceStable(top, func(i, j int) bool {
		if top[i].Stars != top[j].Stars {
			return top[i].Stars > top[j].Stars
		}
		return top[i].FullName < top[j].FullName
	})
	if len(top) > 3 {
		top = top[:3]
	}

	var topLanguage string
	if len(languages) > 0 {
		type pair struct {
			name  string
			count int
		}
		pairs := make([]pair, 0, len(languages))
		for name, count := range languages {
			pairs = append(pairs, pair{name: name, count: count})
		}
		sort.SliceStable(pairs, func(i, j int) bool {
			if pairs[i].count != pairs[j].count {
				return pairs[i].count > pairs[j].count
			}
			return pairs[i].name < pairs[j].name
		})
		topLanguage = pairs[0].name
	}

	sharedLanguages := make([]string, 0)
	for language, count := range languages {
		if count > 0 {
			sharedLanguages = append(sharedLanguages, language)
		}
	}
	sort.Strings(sharedLanguages)
	if len(sharedLanguages) > 4 {
		sharedLanguages = sharedLanguages[:4]
	}

	return top, topLanguage, sharedLanguages
}

func compareProfileFromData(
	dev models.Developer,
	rank *int,
	topRepos []compareRepo,
	topLanguage string,
	topLanguages []string,
	contribs []models.ContributionDay,
) compareProfile {
	totalContribs, activeDays := 0, 0
	for _, day := range contribs {
		totalContribs += day.Count
		if day.Count > 0 {
			activeDays++
		}
	}

	return compareProfile{
		Username:            dev.GithubUsername,
		DisplayName:         dev.DisplayName,
		Rank:                rank,
		ActivityScore:       dev.ActivityScore,
		BuilderScore:        dev.BuilderScore,
		ContributorScore:    dev.ContributorScore,
		ReviewerScore:       dev.ReviewerScore,
		CommunityScore:      dev.CommunityScore,
		PublicRepos:         dev.PublicRepos,
		ReadmeRepos:         dev.ReadmeRepos,
		TotalStars:          dev.TotalStars,
		Followers:           dev.Followers,
		Following:           dev.Following,
		CurrentStreak:       dev.CurrentStreak,
		LongestStreak:       dev.LongestStreak,
		XP:                  dev.XP,
		PowerLevel:          dev.PowerLevel,
		PRContributions:     dev.PRContributions,
		IssueContributions:  dev.IssueContributions,
		ReviewContributions: dev.ReviewContributions,
		ContributionTotal:   totalContribs,
		ActiveDays:          activeDays,
		TopLanguage:         topLanguage,
		TopRepos:            topRepos,
		TopLanguages:        topLanguages,
	}
}

func (s *CompareService) generate(ctx context.Context, snapshot comparisonSnapshot) (*comparisonJSON, error) {
	slog.Info("comparison generation started", "left", snapshot.Left.Username, "right", snapshot.Right.Username)
	snapshotJSON, err := truncateJSON(snapshot)
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(
		"Generate a side-by-side comparison for these two NUST developers.\n"+
			"Use only the JSON snapshot below. Do not call tools.\n\n%s\n\n%s",
		comparisonSystemPrompt, snapshotJSON,
	)

	raw, err := s.chat.RunSync(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return parseComparisonJSON(raw)
}

func parseComparisonJSON(raw string) (*comparisonJSON, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		parts := strings.SplitN(raw, "```", 3)
		if len(parts) >= 2 {
			raw = parts[1]
			raw = strings.TrimPrefix(raw, "json")
			raw = strings.TrimSpace(raw)
		}
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON object found in response")
	}
	raw = raw[start : end+1]

	var out comparisonJSON
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	if out.Headline == "" {
		out.Headline = "Developer comparison"
	}
	if out.Summary == "" {
		return nil, fmt.Errorf("incomplete comparison JSON")
	}
	if len(out.Takeaways) == 0 {
		out.Takeaways = []string{}
	}
	if len(out.SharedStrengths) == 0 {
		out.SharedStrengths = []string{}
	}
	return &out, nil
}

func fallbackComparison(left, right compareProfile) comparisonJSON {
	takeaways := []string{}
	if left.ActivityScore != right.ActivityScore {
		if left.ActivityScore > right.ActivityScore {
			takeaways = append(takeaways, fmt.Sprintf("%s leads on activity score.", left.Username))
		} else {
			takeaways = append(takeaways, fmt.Sprintf("%s leads on activity score.", right.Username))
		}
	}
	if left.TotalStars != right.TotalStars {
		if left.TotalStars > right.TotalStars {
			takeaways = append(takeaways, fmt.Sprintf("%s has more total stars.", left.Username))
		} else {
			takeaways = append(takeaways, fmt.Sprintf("%s has more total stars.", right.Username))
		}
	}
	if left.PublicRepos != right.PublicRepos {
		if left.PublicRepos > right.PublicRepos {
			takeaways = append(takeaways, fmt.Sprintf("%s has a broader public repo footprint.", left.Username))
		} else {
			takeaways = append(takeaways, fmt.Sprintf("%s has a broader public repo footprint.", right.Username))
		}
	}
	if left.ReviewContributions != right.ReviewContributions {
		if left.ReviewContributions > right.ReviewContributions {
			takeaways = append(takeaways, fmt.Sprintf("%s is stronger on reviews.", left.Username))
		} else {
			takeaways = append(takeaways, fmt.Sprintf("%s is stronger on reviews.", right.Username))
		}
	}
	if len(takeaways) == 0 {
		takeaways = append(takeaways, "Both developers are closely matched across the tracked metrics.")
	}

	shared := []string{}
	if left.TopLanguage != "" && left.TopLanguage == right.TopLanguage {
		shared = append(shared, left.TopLanguage)
	}
	if len(shared) == 0 {
		shared = append(shared, "public GitHub activity")
	}

	summary := fmt.Sprintf(
		"%s and %s are close on the tracked surface area, but they lean toward different strengths.",
		left.Username, right.Username,
	)
	if left.ActivityScore > right.ActivityScore {
		summary = fmt.Sprintf(
			"%s is ahead on overall activity, while %s shows a different profile of strengths and scale.",
			left.Username, right.Username,
		)
	} else if right.ActivityScore > left.ActivityScore {
		summary = fmt.Sprintf(
			"%s is ahead on overall activity, while %s shows a different profile of strengths and scale.",
			right.Username, left.Username,
		)
	}

	verdict := "Pick the developer whose strongest metrics better match the use case: activity, reach, or review depth."

	return comparisonJSON{
		Headline:        fmt.Sprintf("%s vs %s", left.Username, right.Username),
		Summary:         summary,
		Takeaways:       takeaways,
		SharedStrengths: shared,
		Verdict:         verdict,
	}
}
