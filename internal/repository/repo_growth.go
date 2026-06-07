package repository

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/abdulsami/nust-devs/internal/models"
)

const repoSparklineDays = 14
const repoGrowthPeriodDays = 30

type repoSnapshotRow struct {
	RepoID       string
	SnapshotDate time.Time
	Stars        int
	Forks        int
}

func (r *StatsRepo) fetchRepoSnapshots(ctx context.Context, repoIDs []string, since time.Time) (map[string][]repoSnapshotRow, error) {
	if len(repoIDs) == 0 {
		return map[string][]repoSnapshotRow{}, nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT repo_id, snapshot_date, stars, forks
		FROM repo_snapshots
		WHERE repo_id = ANY($1) AND snapshot_date >= $2
		ORDER BY repo_id, snapshot_date ASC`, repoIDs, since.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string][]repoSnapshotRow{}
	for rows.Next() {
		var s repoSnapshotRow
		if err := rows.Scan(&s.RepoID, &s.SnapshotDate, &s.Stars, &s.Forks); err != nil {
			return nil, err
		}
		out[s.RepoID] = append(out[s.RepoID], s)
	}
	return out, nil
}

func repoMetricAtDate(rows []repoSnapshotRow, target time.Time, metric string) (int, bool) {
	key := target.Format("2006-01-02")
	var best *repoSnapshotRow
	for i := range rows {
		d := rows[i].SnapshotDate.UTC()
		if d.After(target) {
			continue
		}
		if best == nil || d.After(best.SnapshotDate) {
			best = &rows[i]
		}
		if d.Format("2006-01-02") == key {
			best = &rows[i]
			break
		}
	}
	if best == nil {
		return 0, false
	}
	if metric == "forks" {
		return best.Forks, true
	}
	return best.Stars, true
}

func buildRepoSparkline(rows []repoSnapshotRow, since time.Time, days int, metric string) []models.SparkPoint {
	byDate := map[string]float64{}
	for _, s := range rows {
		v := float64(s.Stars)
		if metric == "forks" {
			v = float64(s.Forks)
		}
		byDate[s.SnapshotDate.Format("2006-01-02")] = v
	}
	out := make([]models.SparkPoint, 0, days)
	var last float64
	for i := 0; i < days; i++ {
		d := since.AddDate(0, 0, i).Format("2006-01-02")
		if v, ok := byDate[d]; ok {
			last = v
		}
		out = append(out, models.SparkPoint{Date: d, Value: last})
	}
	return out
}

func (r *StatsRepo) attachRepoGrowth(ctx context.Context, repos []models.PublicRepo) {
	if len(repos) == 0 {
		return
	}
	ids := make([]string, len(repos))
	for i := range repos {
		ids[i] = repos[i].ID
	}
	since := time.Now().UTC().AddDate(0, 0, -(repoSparklineDays - 1))
	pastTarget := time.Now().UTC().AddDate(0, 0, -repoGrowthPeriodDays)
	snapshots, err := r.fetchRepoSnapshots(ctx, ids, since)
	if err != nil {
		return
	}
	for i := range repos {
		rows := snapshots[repos[i].ID]
		pastStars, okS := repoMetricAtDate(rows, pastTarget, "stars")
		pastForks, okF := repoMetricAtDate(rows, pastTarget, "forks")
		if okS {
			delta := repos[i].Stars - pastStars
			repos[i].StarsGrowth30d = &delta
		}
		if okF {
			delta := repos[i].Forks - pastForks
			repos[i].ForksGrowth30d = &delta
		}
		repos[i].Sparkline = buildRepoSparkline(rows, since, repoSparklineDays, "stars")
	}
}

func (r *StatsRepo) sortReposByGrowth(ctx context.Context, repos []models.PublicRepo, limit int) ([]models.PublicRepo, error) {
	r.attachRepoGrowth(ctx, repos)
	sort.SliceStable(repos, func(i, j int) bool {
		gi := growthValue(repos[i].StarsGrowth30d)
		gj := growthValue(repos[j].StarsGrowth30d)
		if gi != gj {
			return gi > gj
		}
		return repos[i].Stars > repos[j].Stars
	})
	if len(repos) > limit {
		repos = repos[:limit]
	}
	return repos, nil
}

func growthValue(v *int) int {
	if v == nil {
		return -1 << 30
	}
	return *v
}

func (r *StatsRepo) GetFastestGrowingRepos(ctx context.Context, days, limit int) ([]models.PublicRepo, error) {
	if days < 7 || days > 90 {
		days = 30
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	repos, err := r.ListProjects(ctx, ProjectFilter{Category: "original", Sort: "stars", Limit: 100})
	if err != nil {
		return nil, err
	}
	// Recompute growth for custom period
	ids := make([]string, len(repos))
	for i := range repos {
		ids[i] = repos[i].ID
	}
	since := time.Now().UTC().AddDate(0, 0, -(repoSparklineDays - 1))
	pastTarget := time.Now().UTC().AddDate(0, 0, -days)
	snapshots, err := r.fetchRepoSnapshots(ctx, ids, since)
	if err != nil {
		return nil, err
	}
	for i := range repos {
		rows := snapshots[repos[i].ID]
		if pastStars, ok := repoMetricAtDate(rows, pastTarget, "stars"); ok {
			delta := repos[i].Stars - pastStars
			repos[i].StarsGrowth30d = &delta
		}
		if pastForks, ok := repoMetricAtDate(rows, pastTarget, "forks"); ok {
			delta := repos[i].Forks - pastForks
			repos[i].ForksGrowth30d = &delta
		}
		repos[i].Sparkline = buildRepoSparkline(rows, since, repoSparklineDays, "stars")
	}
	sort.SliceStable(repos, func(i, j int) bool {
		gi := growthValue(repos[i].StarsGrowth30d)
		gj := growthValue(repos[j].StarsGrowth30d)
		if gi != gj {
			return gi > gj
		}
		return repos[i].Stars > repos[j].Stars
	})
	if len(repos) > limit {
		repos = repos[:limit]
	}
	// drop entries with no growth data
	out := make([]models.PublicRepo, 0, limit)
	for _, rp := range repos {
		if rp.StarsGrowth30d != nil {
			out = append(out, rp)
		}
	}
	return out, nil
}

func (r *StatsRepo) GetRepoGrowth(ctx context.Context, repoID string, days int) ([]models.SparkPoint, error) {
	if days < 7 || days > 90 {
		days = 30
	}
	since := time.Now().UTC().AddDate(0, 0, -days)
	snapshots, err := r.fetchRepoSnapshots(ctx, []string{repoID}, since)
	if err != nil {
		return nil, err
	}
	return buildRepoSparkline(snapshots[repoID], since, days, "stars"), nil
}

func (r *StatsRepo) netNewStarsSeries(ctx context.Context, granularity string, periodKeys []string) (map[string]int, error) {
	unit := truncUnit(granularity)
	var keyExpr string
	if granularity == "monthly" {
		keyExpr = "to_char(lip.period, 'YYYY-MM')"
	} else {
		keyExpr = "(extract(year FROM lip.period)::int || '-' || extract(quarter FROM lip.period)::int)"
	}

	query := fmt.Sprintf(`
		WITH latest_in_period AS (
			SELECT DISTINCT ON (rs.repo_id, date_trunc('%s', rs.snapshot_date))
				date_trunc('%s', rs.snapshot_date) AS period,
				rs.repo_id,
				rs.stars
			FROM repo_snapshots rs
			JOIN repos r ON r.id = rs.repo_id
			WHERE EXISTS (SELECT 1 FROM developer_repos dr WHERE dr.repo_id = r.id)
			ORDER BY rs.repo_id, date_trunc('%s', rs.snapshot_date), rs.snapshot_date DESC
		),
		period_totals AS (
			SELECT period, SUM(stars)::int AS total_stars
			FROM latest_in_period
			GROUP BY period
			ORDER BY period
		)
		SELECT
			%s AS period_key,
			GREATEST(total_stars - COALESCE(LAG(total_stars) OVER (ORDER BY period), 0), 0)::int AS net_new
		FROM period_totals lip`, unit, unit, unit, keyExpr)

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]int{}
	for rows.Next() {
		var key string
		var val int
		if err := rows.Scan(&key, &val); err != nil {
			return nil, err
		}
		out[key] = val
	}
	return out, nil
}
