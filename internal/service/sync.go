package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	gh "github.com/abdulsami/nust-devs/internal/github"
	"github.com/abdulsami/nust-devs/internal/models"
	"github.com/abdulsami/nust-devs/internal/repository"
)

type SyncService struct {
	github   *gh.Client
	syncRepo *repository.SyncRepo
	devRepo  *repository.DeveloperRepo
}

func NewSyncService(github *gh.Client, syncRepo *repository.SyncRepo, devRepo *repository.DeveloperRepo) *SyncService {
	return &SyncService{github: github, syncRepo: syncRepo, devRepo: devRepo}
}

// SyncDeveloper runs the full pipeline: profile → repos → languages → contributions → snapshot.
func (s *SyncService) SyncDeveloper(ctx context.Context, dev *models.Developer) error {
	log := slog.With("developer", dev.GithubUsername)
	log.Info("sync started")

	// 1. Profile
	user, err := s.github.GetUser(ctx, dev.GithubUsername)
	if err != nil {
		return fmt.Errorf("fetch profile: %w", err)
	}
	if err := s.syncRepo.UpsertDeveloperProfile(ctx, dev.ID, user); err != nil {
		return fmt.Errorf("upsert profile: %w", err)
	}
	log.Info("profile synced")

	// 2. Repos
	repos, err := s.github.GetRepos(ctx, dev.GithubUsername)
	if err != nil {
		return fmt.Errorf("fetch repos: %w", err)
	}
	if err := s.syncRepo.UpdateTotalStars(ctx, dev.ID, repos); err != nil {
		return fmt.Errorf("update stars: %w", err)
	}

	for _, repo := range repos {
		repoID, err := s.syncRepo.UpsertRepo(ctx, repo)
		if err != nil {
			log.Warn("upsert repo failed", "repo", repo.FullName, "err", err)
			continue
		}
		if err := s.syncRepo.LinkDeveloperRepo(ctx, dev.ID, repoID); err != nil {
			log.Warn("link repo failed", "repo", repo.FullName, "err", err)
		}
		pushedAt := repo.PushedAt
		var pushedPtr *time.Time
		if !pushedAt.IsZero() {
			pushedPtr = &pushedAt
		}
		if err := s.syncRepo.WriteRepoSnapshot(ctx, repoID, repo.StargazersCount, repo.ForksCount, pushedPtr); err != nil {
			log.Warn("repo snapshot failed", "repo", repo.FullName, "err", err)
		}
	}

	// 3. Languages — top repos by stars only (keeps sync fast; 56 repos × API calls is slow)
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].StargazersCount > repos[j].StargazersCount
	})
	langLimit := 20
	langFetched := 0
	for _, repo := range repos {
		if repo.Fork {
			continue
		}
		if langFetched >= langLimit {
			break
		}
		langFetched++
		repoID, err := s.syncRepo.RepoIDByGithubID(ctx, repo.ID)
		if err != nil {
			log.Warn("repo id lookup failed", "repo", repo.FullName, "err", err)
			continue
		}
		langs, err := s.github.GetLanguages(ctx, user.Login, repo.Name)
		if err != nil {
			log.Warn("fetch languages failed", "repo", repo.FullName, "err", err)
			continue
		}
		if err := s.syncRepo.UpsertRepoLanguages(ctx, repoID, langs); err != nil {
			log.Warn("upsert languages failed", "repo", repo.FullName, "err", err)
		}
	}
	log.Info("repos synced", "count", len(repos))

	// 4. GraphQL stats — contributions + dimension score inputs (non-fatal)
	graphStats, err := s.github.GetUserGraphStats(ctx, dev.GithubUsername)
	if err != nil {
		log.Warn("fetch graphql stats skipped", "err", err)
	} else {
		if err := s.syncRepo.UpsertContributionDays(ctx, dev.ID, graphStats.Days); err != nil {
			log.Warn("upsert contributions failed", "err", err)
		} else {
			log.Info("contributions synced", "days", len(graphStats.Days))
		}
		if err := s.syncRepo.UpsertContributionStats(ctx, dev.ID, graphStats); err != nil {
			log.Warn("contribution stats failed", "err", err)
		} else {
			log.Info("contribution stats synced",
				"prs", graphStats.PRContributions,
				"issues", graphStats.IssueContributions,
				"reviews", graphStats.ReviewContributions,
				"repos", len(graphStats.ByRepository),
			)
		}
		if err := s.syncRepo.RecomputeDimensionScores(ctx, dev.ID, graphStats); err != nil {
			log.Warn("dimension scores failed", "err", err)
		} else {
			log.Info("dimension scores updated")
		}
	}

	// 5. Recompute activity score (uses contribution_days)
	if err := s.syncRepo.RecomputeActivityScore(ctx, dev.ID); err != nil {
		return fmt.Errorf("recompute score: %w", err)
	}

	if err := s.syncRepo.RecomputeGamification(ctx, dev.ID); err != nil {
		log.Warn("gamification recompute failed", "err", err)
	}

	// 6. Daily snapshot with all scores
	fresh, err := s.devRepo.GetByID(ctx, dev.ID)
	if err != nil {
		return fmt.Errorf("refresh developer: %w", err)
	}
	if err := s.syncRepo.WriteSnapshot(ctx, fresh); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}

	// 7. Mark synced
	if err := s.syncRepo.UpdateLastSynced(ctx, dev.ID); err != nil {
		return fmt.Errorf("update last_synced_at: %w", err)
	}

	log.Info("sync complete")
	return nil
}
