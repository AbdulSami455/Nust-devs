package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/abdulsami/nust-devs/internal/gamification"
	"github.com/abdulsami/nust-devs/internal/models"
)

func (r *SyncRepo) RecomputeGamification(ctx context.Context, devID string) error {
	rows, err := r.db.Query(ctx, `
		SELECT date::text, count FROM contribution_days
		WHERE developer_id = $1 AND date >= CURRENT_DATE - INTERVAL '365 days'
		ORDER BY date ASC`, devID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var days []gamification.ActiveDay
	for rows.Next() {
		var d gamification.ActiveDay
		if err := rows.Scan(&d.Date, &d.Count); err != nil {
			return err
		}
		days = append(days, d)
	}

	current, longest, multiplier := gamification.ComputeStreaks(days)

	var commits90d, prs, issues, reviews, stars int
	var activityScore float64
	err = r.db.QueryRow(ctx, `
		SELECT
			(SELECT COALESCE(SUM(count), 0) FROM contribution_days
			 WHERE developer_id = $1 AND date >= CURRENT_DATE - INTERVAL '90 days'),
			pr_contributions, issue_contributions, review_contributions,
			total_stars, activity_score
		FROM developers WHERE id = $1`, devID).
		Scan(&commits90d, &prs, &issues, &reviews, &stars, &activityScore)
	if err != nil {
		return err
	}

	xp := gamification.ComputeXP(commits90d, prs, issues, reviews, stars, activityScore)
	level := gamification.PowerLevelFromXP(xp)

	_, err = r.db.Exec(ctx, `
		UPDATE developers SET
			current_streak = $2,
			longest_streak = GREATEST(longest_streak, $3),
			streak_multiplier = $4,
			xp = $5,
			power_level = $6
		WHERE id = $1`,
		devID, current, longest, multiplier, xp, level,
	)
	return err
}

func (r *StatsRepo) GetStreakSummary(ctx context.Context) (*models.StreakSummary, error) {
	var s models.StreakSummary
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE current_streak >= 7)::int,
			COUNT(*) FILTER (WHERE current_streak >= 30)::int,
			COALESCE(MAX(current_streak), 0)::int
		FROM developers
		WHERE last_synced_at IS NOT NULL`).
		Scan(&s.DevsOn7PlusStreak, &s.DevsOn30PlusStreak, &s.LongestActiveStreak)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StatsRepo) ListDevOfMonthWinners(ctx context.Context, limit int) ([]models.DevOfMonthWinner, error) {
	if limit < 1 || limit > 50 {
		limit = 24
	}
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT w.year, w.month, w.score, w.activity_points, w.rank_gain, w.stars_gained,
		       %s
		FROM dev_of_month_winners w
		JOIN developers d ON d.id = w.developer_id
		ORDER BY w.year DESC, w.month DESC
		LIMIT $1`, developerCols), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.DevOfMonthWinner
	for rows.Next() {
		var w models.DevOfMonthWinner
		var d models.Developer
		if err := rows.Scan(
			&w.Year, &w.Month, &w.Score, &w.ActivityPoints, &w.RankGain, &w.StarsGained,
			&d.ID, &d.GithubUsername, &d.DisplayName, &d.AvatarURL, &d.Bio,
			&d.Location, &d.Company, &d.Website,
			&d.Followers, &d.Following, &d.PublicRepos, &d.TotalStars, &d.ActivityScore,
			&d.BuilderScore, &d.ContributorScore, &d.ReviewerScore, &d.CommunityScore,
			&d.PRContributions, &d.IssueContributions, &d.ReviewContributions,
			&d.ContributionPeriodStart, &d.ContributionPeriodEnd,
			&d.CurrentStreak, &d.LongestStreak, &d.StreakMultiplier, &d.XP, &d.PowerLevel,
			&d.VerificationStatus, &d.LastSyncedAt, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		w.Developer = d
		w.PowerTitle = gamification.PowerTitle(d.PowerLevel)
		out = append(out, w)
	}
	if out == nil {
		return []models.DevOfMonthWinner{}, nil
	}
	return out, nil
}

func (r *StatsRepo) GetWrapped(ctx context.Context, username string, year int) (*models.WrappedReport, error) {
	dev, err := r.GetDeveloperByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	var totalContributions int
	_ = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(count), 0)::int FROM contribution_days
		WHERE developer_id = $1 AND date >= $2 AND date < $3`,
		dev.ID, start.Format("2006-01-02"), end.Format("2006-01-02")).
		Scan(&totalContributions)

	var topRepoName string
	var topRepoStars int
	_ = r.db.QueryRow(ctx, `
		SELECT r.full_name, r.stars FROM repos r
		JOIN developer_repos dr ON dr.repo_id = r.id
		WHERE dr.developer_id = $1 AND NOT r.is_fork
		ORDER BY r.stars DESC LIMIT 1`, dev.ID).
		Scan(&topRepoName, &topRepoStars)

	langRows, err := r.db.Query(ctx, `
		SELECT rl.language, SUM(rl.bytes)::bigint
		FROM repo_languages rl
		JOIN developer_repos dr ON dr.repo_id = rl.repo_id
		JOIN repos r ON r.id = rl.repo_id
		WHERE dr.developer_id = $1 AND NOT r.is_fork
		GROUP BY rl.language
		ORDER BY 2 DESC
		LIMIT 5`, dev.ID)
	if err != nil {
		return nil, err
	}
	defer langRows.Close()
	var langs []models.NameCount
	for langRows.Next() {
		var l models.NameCount
		var bytes int64
		if err := langRows.Scan(&l.Name, &bytes); err != nil {
			return nil, err
		}
		l.Count = int(bytes)
		langs = append(langs, l)
	}

	rankStart, _ := r.rankNearDate(ctx, dev.ID, start)
	rankEnd, _ := r.rankNearDate(ctx, dev.ID, end.AddDate(0, 0, -1))
	rankChange := 0
	if rankStart > 0 && rankEnd > 0 {
		rankChange = rankStart - rankEnd
	}

	percentile, _ := r.activityPercentile(ctx, dev.ID)

	report := &models.WrappedReport{
		Year:               year,
		Username:           dev.GithubUsername,
		DisplayName:        dev.DisplayName,
		AvatarURL:          dev.AvatarURL,
		TotalContributions: totalContributions,
		TopRepo:            topRepoName,
		TopRepoStars:       topRepoStars,
		RankStart:          rankStart,
		RankEnd:            rankEnd,
		RankChange:         rankChange,
		ActivityPercentile: percentile,
		TopLanguages:       langs,
		PowerLevel:         dev.PowerLevel,
		PowerTitle:         gamification.PowerTitle(dev.PowerLevel),
		XP:                 dev.XP,
		CurrentStreak:      dev.CurrentStreak,
		LongestStreak:      dev.LongestStreak,
		PRContributions:    dev.PRContributions,
		TotalStars:         dev.TotalStars,
		PublicRepos:        dev.PublicRepos,
	}

	report.Highlights = gamification.BuildWrappedHighlights(
		report.TotalContributions, report.RankChange, report.TopRepo, report.TopRepoStars,
		report.CurrentStreak, report.ActivityPercentile, report.Year,
	)
	return report, nil
}

func (r *StatsRepo) rankNearDate(ctx context.Context, devID string, target time.Time) (int, error) {
	var score float64
	err := r.db.QueryRow(ctx, `
		SELECT activity_score FROM developer_snapshots
		WHERE developer_id = $1 AND snapshot_date <= $2
		ORDER BY snapshot_date DESC LIMIT 1`,
		devID, target.Format("2006-01-02")).Scan(&score)
	if err != nil {
		return 0, err
	}
	var rank int
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) + 1 FROM developers d
		WHERE d.last_synced_at IS NOT NULL AND d.activity_score > $1`, score).Scan(&rank)
	return rank, err
}

func (r *StatsRepo) activityPercentile(ctx context.Context, devID string) (int, error) {
	var score float64
	err := r.db.QueryRow(ctx, `SELECT activity_score FROM developers WHERE id = $1`, devID).Scan(&score)
	if err != nil {
		return 0, err
	}
	var below, total int
	err = r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE activity_score < $1),
			COUNT(*)
		FROM developers WHERE last_synced_at IS NOT NULL`, score).
		Scan(&below, &total)
	if err != nil || total == 0 {
		return 0, err
	}
	return (below * 100) / total, nil
}

func (r *StatsRepo) AwardDevOfMonth(ctx context.Context, year, month int) error {
	type candidate struct {
		devID          string
		activity       int
		rankGain       int
		starsGained    int
		score          float64
	}

	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	devs, err := r.getAllDevelopersSorted(ctx, "activity_score")
	if err != nil {
		return err
	}

	var best *candidate
	for _, d := range devs {
		var activity int
		_ = r.db.QueryRow(ctx, `
			SELECT COALESCE(SUM(count), 0)::int FROM contribution_days
			WHERE developer_id = $1 AND date >= $2 AND date < $3`,
			d.ID, start.Format("2006-01-02"), end.Format("2006-01-02")).Scan(&activity)

		rankStart, err1 := r.rankNearDate(ctx, d.ID, start)
		rankEnd, err2 := r.rankNearDate(ctx, d.ID, end.AddDate(0, 0, -1))
		rankGain := 0
		if err1 == nil && err2 == nil && rankStart > 0 && rankEnd > 0 {
			rankGain = rankStart - rankEnd
		}

		var starsStart, starsEnd int
		_ = r.db.QueryRow(ctx, `
			SELECT COALESCE(total_stars, 0) FROM developer_snapshots
			WHERE developer_id = $1 AND snapshot_date >= $2
			ORDER BY snapshot_date ASC LIMIT 1`, d.ID, start.Format("2006-01-02")).Scan(&starsStart)
		_ = r.db.QueryRow(ctx, `
			SELECT COALESCE(total_stars, 0) FROM developer_snapshots
			WHERE developer_id = $1 AND snapshot_date < $2
			ORDER BY snapshot_date DESC LIMIT 1`, d.ID, end.Format("2006-01-02")).Scan(&starsEnd)
		starsGained := starsEnd - starsStart
		if starsGained < 0 {
			starsGained = 0
		}

		score := float64(activity)*3 + float64(max(0, rankGain))*10 + float64(starsGained)*2
		if activity == 0 && rankGain <= 0 && starsGained == 0 {
			continue
		}
		c := candidate{d.ID, activity, rankGain, starsGained, score}
		if best == nil || c.score > best.score {
			best = &c
		}
	}

	if best == nil {
		return nil
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO dev_of_month_winners
			(developer_id, year, month, score, activity_points, rank_gain, stars_gained)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (year, month) DO UPDATE SET
			developer_id = EXCLUDED.developer_id,
			score = EXCLUDED.score,
			activity_points = EXCLUDED.activity_points,
			rank_gain = EXCLUDED.rank_gain,
			stars_gained = EXCLUDED.stars_gained`,
		best.devID, year, month, best.score, best.activity, best.rankGain, best.starsGained,
	)
	return err
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
