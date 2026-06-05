package repository

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/abdulsami/nust-devs/internal/models"
)

const sparklineDays = 14

type snapshotRow struct {
	DeveloperID   string
	SnapshotDate  time.Time
	ActivityScore float64
	TotalStars    int
	PublicRepos   int
	Followers     int
}

func metricFromDev(sortBy string, d *models.Developer) float64 {
	switch sortBy {
	case "total_stars":
		return float64(d.TotalStars)
	case "public_repos":
		return float64(d.PublicRepos)
	case "followers":
		return float64(d.Followers)
	default:
		return d.ActivityScore
	}
}

func metricFromSnapshot(sortBy string, s snapshotRow) float64 {
	switch sortBy {
	case "total_stars":
		return float64(s.TotalStars)
	case "public_repos":
		return float64(s.PublicRepos)
	case "followers":
		return float64(s.Followers)
	default:
		return s.ActivityScore
	}
}

func (r *StatsRepo) GetLeaderboardWithTrends(
	ctx context.Context,
	sortBy, view string,
	period, page, limit int,
) ([]models.LeaderboardEntry, error) {
	if period != 30 {
		period = 7
	}

	devs, err := r.getAllDevelopersSorted(ctx, sortBy)
	if err != nil {
		return nil, err
	}
	if len(devs) == 0 {
		return []models.LeaderboardEntry{}, nil
	}

	ids := make([]string, len(devs))
	for i := range devs {
		ids[i] = devs[i].ID
	}

	since := time.Now().UTC().AddDate(0, 0, -(sparklineDays - 1))
	snapshots, err := r.fetchSnapshots(ctx, ids, since)
	if err != nil {
		return nil, err
	}

	date7 := nearestSnapshotDate(snapshots, 7)
	date30 := nearestSnapshotDate(snapshots, 30)

	currentRanks := rankDevelopers(devs, sortBy)

	past7 := snapshotsAtDate(snapshots, date7)
	past30 := snapshotsAtDate(snapshots, date30)
	rank7 := rankSnapshots(past7, sortBy)
	rank30 := rankSnapshots(past30, sortBy)

	entries := make([]models.LeaderboardEntry, len(devs))
	for i, d := range devs {
		entry := models.LeaderboardEntry{
			Developer: d,
			Rank:      currentRanks[d.ID],
		}

		if date7 != nil {
			if past, ok := past7[d.ID]; ok {
				cur := metricFromDev(sortBy, &d)
				pastVal := metricFromSnapshot(sortBy, past)
				delta := round2(cur - pastVal)
				entry.ScoreDelta7d = &delta
				if pr, ok := rank7[d.ID]; ok {
					rd := pr - entry.Rank
					entry.RankDelta7d = &rd
				}
			}
		}
		if date30 != nil {
			if past, ok := past30[d.ID]; ok {
				cur := metricFromDev(sortBy, &d)
				pastVal := metricFromSnapshot(sortBy, past)
				delta := round2(cur - pastVal)
				entry.ScoreDelta30d = &delta
				if pr, ok := rank30[d.ID]; ok {
					rd := pr - entry.Rank
					entry.RankDelta30d = &rd
				}
			}
		}

		entry.Sparkline = buildSparkline(snapshots[d.ID], sortBy, since, sparklineDays)
		entries[i] = entry
	}

	if view == "rising" {
		rising := make([]models.LeaderboardEntry, 0, len(entries))
		for _, e := range entries {
			if deltaForPeriod(&e, period) > math.Inf(-2) {
				rising = append(rising, e)
			}
		}
		sort.SliceStable(rising, func(i, j int) bool {
			di := deltaForPeriod(&rising[i], period)
			dj := deltaForPeriod(&rising[j], period)
			if di != dj {
				return di > dj
			}
			return rising[i].Rank < rising[j].Rank
		})
		entries = rising
	}

	offset := (page - 1) * limit
	if offset >= len(entries) {
		return []models.LeaderboardEntry{}, nil
	}
	end := offset + limit
	if end > len(entries) {
		end = len(entries)
	}
	return entries[offset:end], nil
}

func deltaForPeriod(e *models.LeaderboardEntry, period int) float64 {
	if period == 30 && e.ScoreDelta30d != nil {
		return *e.ScoreDelta30d
	}
	if e.ScoreDelta7d != nil {
		return *e.ScoreDelta7d
	}
	return math.Inf(-1)
}

func (r *StatsRepo) getAllDevelopersSorted(ctx context.Context, sortBy string) ([]models.Developer, error) {
	col, ok := validSortFields[sortBy]
	if !ok {
		col = "activity_score"
	}
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT %s FROM developers
		WHERE last_synced_at IS NOT NULL
		ORDER BY %s DESC, github_username ASC`, developerCols, col))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var devs []models.Developer
	for rows.Next() {
		var d models.Developer
		if err := scanPublicDeveloper(rows, &d); err != nil {
			return nil, err
		}
		devs = append(devs, d)
	}
	return devs, nil
}

func (r *StatsRepo) fetchSnapshots(ctx context.Context, devIDs []string, since time.Time) (map[string][]snapshotRow, error) {
	rows, err := r.db.Query(ctx, `
		SELECT developer_id, snapshot_date, activity_score, total_stars, public_repos, followers
		FROM developer_snapshots
		WHERE developer_id = ANY($1) AND snapshot_date >= $2
		ORDER BY developer_id, snapshot_date ASC`, devIDs, since.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string][]snapshotRow{}
	for rows.Next() {
		var s snapshotRow
		if err := rows.Scan(&s.DeveloperID, &s.SnapshotDate, &s.ActivityScore, &s.TotalStars, &s.PublicRepos, &s.Followers); err != nil {
			return nil, err
		}
		out[s.DeveloperID] = append(out[s.DeveloperID], s)
	}
	return out, nil
}

func nearestSnapshotDate(byDev map[string][]snapshotRow, daysAgo int) *time.Time {
	target := time.Now().UTC().AddDate(0, 0, -daysAgo)
	var best *time.Time
	for _, rows := range byDev {
		for _, s := range rows {
			d := s.SnapshotDate.UTC()
			if d.After(target) {
				continue
			}
			if best == nil || d.After(*best) {
				t := d
				best = &t
			}
		}
	}
	return best
}

func snapshotsAtDate(byDev map[string][]snapshotRow, date *time.Time) map[string]snapshotRow {
	if date == nil {
		return nil
	}
	target := date.Format("2006-01-02")
	out := map[string]snapshotRow{}
	for devID, rows := range byDev {
		for _, s := range rows {
			if s.SnapshotDate.Format("2006-01-02") == target {
				out[devID] = s
				break
			}
		}
	}
	return out
}

func rankDevelopers(devs []models.Developer, sortBy string) map[string]int {
	type ranked struct {
		id     string
		metric float64
		name   string
	}
	items := make([]ranked, len(devs))
	for i, d := range devs {
		items[i] = ranked{id: d.ID, metric: metricFromDev(sortBy, &d), name: d.GithubUsername}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].metric != items[j].metric {
			return items[i].metric > items[j].metric
		}
		return items[i].name < items[j].name
	})
	out := map[string]int{}
	for i, it := range items {
		out[it.id] = i + 1
	}
	return out
}

func rankSnapshots(snaps map[string]snapshotRow, sortBy string) map[string]int {
	if len(snaps) == 0 {
		return nil
	}
	type ranked struct {
		id     string
		metric float64
	}
	items := make([]ranked, 0, len(snaps))
	for id, s := range snaps {
		items = append(items, ranked{id: id, metric: metricFromSnapshot(sortBy, s)})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].metric > items[j].metric
	})
	out := map[string]int{}
	for i, it := range items {
		out[it.id] = i + 1
	}
	return out
}

func buildSparkline(rows []snapshotRow, sortBy string, since time.Time, days int) []models.SparkPoint {
	byDate := map[string]float64{}
	for _, s := range rows {
		byDate[s.SnapshotDate.Format("2006-01-02")] = metricFromSnapshot(sortBy, s)
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

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
