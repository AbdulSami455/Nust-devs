package repository

import (
	"context"
	"sort"

	gh "github.com/abdulsami/nust-devs/internal/github"
	"github.com/abdulsami/nust-devs/internal/models"
)

func (r *SyncRepo) UpsertContributionStats(ctx context.Context, devID string, g *gh.UserGraphStats) error {
	periodStart := g.PeriodStart.Format("2006-01-02")
	periodEnd := g.PeriodEnd.Format("2006-01-02")

	_, err := r.db.Exec(ctx, `
		UPDATE developers SET
			pr_contributions = $2,
			issue_contributions = $3,
			review_contributions = $4,
			contribution_period_start = $5,
			contribution_period_end = $6
		WHERE id = $1`,
		devID, g.PRContributions, g.IssueContributions, g.ReviewContributions,
		periodStart, periodEnd,
	)
	if err != nil {
		return err
	}

	if _, err := r.db.Exec(ctx, `DELETE FROM developer_external_contributions WHERE developer_id = $1`, devID); err != nil {
		return err
	}

	for _, repo := range g.ByRepository {
		if repo.PRs == 0 && repo.Issues == 0 && repo.Reviews == 0 {
			continue
		}
		_, err := r.db.Exec(ctx, `
			INSERT INTO developer_external_contributions
				(developer_id, repo_full_name, repo_url, pr_count, issue_count, review_count)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			devID, repo.FullName, repo.URL, repo.PRs, repo.Issues, repo.Reviews,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *StatsRepo) GetContributionStats(ctx context.Context, devID string) (*models.ContributionStats, error) {
	var stats models.ContributionStats
	var periodStart, periodEnd *string
	err := r.db.QueryRow(ctx, `
		SELECT pr_contributions, issue_contributions, review_contributions,
		       contribution_period_start::text, contribution_period_end::text
		FROM developers WHERE id = $1`, devID).
		Scan(&stats.PullRequests, &stats.Issues, &stats.Reviews, &periodStart, &periodEnd)
	if err != nil {
		return nil, err
	}
	if periodStart != nil {
		stats.PeriodStart = *periodStart
	}
	if periodEnd != nil {
		stats.PeriodEnd = *periodEnd
	}

	rows, err := r.db.Query(ctx, `
		SELECT repo_full_name, repo_url, pr_count, issue_count, review_count
		FROM developer_external_contributions
		WHERE developer_id = $1
		ORDER BY (pr_count + issue_count + review_count) DESC, repo_full_name ASC
		LIMIT 50`, devID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.RepoContributionStat
		if err := rows.Scan(&item.RepoFullName, &item.RepoURL, &item.PullRequests, &item.Issues, &item.Reviews); err != nil {
			return nil, err
		}
		item.Total = item.PullRequests + item.Issues + item.Reviews
		stats.ByRepository = append(stats.ByRepository, item)
	}

	if stats.ByRepository == nil {
		stats.ByRepository = []models.RepoContributionStat{}
	}

	sort.SliceStable(stats.ByRepository, func(i, j int) bool {
		if stats.ByRepository[i].Total != stats.ByRepository[j].Total {
			return stats.ByRepository[i].Total > stats.ByRepository[j].Total
		}
		return stats.ByRepository[i].RepoFullName < stats.ByRepository[j].RepoFullName
	})

	return &stats, nil
}
