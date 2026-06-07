package gamification

import (
	"fmt"
	"sort"
	"time"
)

type ActiveDay struct {
	Date  string
	Count int
}

func ComputeStreaks(days []ActiveDay) (current, longest int, multiplier float64) {
	active := map[string]bool{}
	for _, d := range days {
		if d.Count > 0 {
			active[d.Date] = true
		}
	}
	if len(active) == 0 {
		return 0, 0, 1.0
	}

	sorted := make([]string, 0, len(active))
	for d := range active {
		sorted = append(sorted, d)
	}
	sort.Strings(sorted)

	longest = 1
	run := 1
	for i := 1; i < len(sorted); i++ {
		prev, _ := time.Parse("2006-01-02", sorted[i-1])
		cur, _ := time.Parse("2006-01-02", sorted[i])
		if cur.Sub(prev) == 24*time.Hour {
			run++
			if run > longest {
				longest = run
			}
		} else {
			run = 1
		}
	}

	today := time.Now().UTC().Format("2006-01-02")
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")

	start := ""
	if active[today] {
		start = today
	} else if active[yesterday] {
		start = yesterday
	}

	if start != "" {
		d, _ := time.Parse("2006-01-02", start)
		for active[d.Format("2006-01-02")] {
			current++
			d = d.AddDate(0, 0, -1)
		}
	}

	multiplier = 1.0
	switch {
	case current >= 30:
		multiplier = 2.0
	case current >= 7:
		multiplier = 1.2
	}
	return current, longest, multiplier
}

func ComputeXP(commits90d, prs, issues, reviews, stars int, activityScore float64) int {
	return commits90d*3 + prs*10 + issues*5 + reviews*8 + stars*2 + int(activityScore)
}

func PowerLevelFromXP(xp int) int {
	level := 1
	for l := 2; l <= 50; l++ {
		if xp < XPRequired(l) {
			break
		}
		level = l
	}
	return level
}

func XPRequired(level int) int {
	if level <= 1 {
		return 0
	}
	l := level - 1
	return l * l * 50
}

func PowerTitle(level int) string {
	switch {
	case level >= 41:
		return "Legend"
	case level >= 26:
		return "Maintainer"
	case level >= 11:
		return "Builder"
	default:
		return "Contributor"
	}
}

func BuildWrappedHighlights(totalContributions, rankChange int, topRepo string, topRepoStars, currentStreak, percentile, year int) []string {
	var out []string
	if totalContributions > 0 {
		out = append(out, fmt.Sprintf("%d contributions in %d", totalContributions, year))
	}
	if rankChange > 0 {
		out = append(out, fmt.Sprintf("Climbed %d places on the leaderboard", rankChange))
	} else if rankChange < 0 {
		out = append(out, fmt.Sprintf("Still competing — rank shifted by %d", rankChange))
	}
	if topRepo != "" {
		out = append(out, fmt.Sprintf("Top project: %s (%d stars)", topRepo, topRepoStars))
	}
	if currentStreak >= 7 {
		out = append(out, fmt.Sprintf("On a %d-day contribution streak", currentStreak))
	}
	if percentile >= 90 {
		out = append(out, "Top 10% of tracked NUST developers")
	} else if percentile >= 75 {
		out = append(out, "Top 25% of tracked NUST developers")
	}
	if len(out) == 0 {
		out = append(out, "Keep shipping — your wrapped gets richer with every sync")
	}
	return out
}
