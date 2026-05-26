package service

import (
	"context"
	"fmt"
	"log/slog"

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

		// 3. Languages (only for non-forks to save API quota)
		if !repo.Fork {
			langs, err := s.github.GetLanguages(ctx, user.Login, repo.Name)
			if err != nil {
				log.Warn("fetch languages failed", "repo", repo.FullName, "err", err)
				continue
			}
			if err := s.syncRepo.UpsertRepoLanguages(ctx, repoID, langs); err != nil {
				log.Warn("upsert languages failed", "repo", repo.FullName, "err", err)
			}
		}
	}
	log.Info("repos synced", "count", len(repos))

	// 4. Contributions
	days, err := s.github.GetContributions(ctx, dev.GithubUsername)
	if err != nil {
		return fmt.Errorf("fetch contributions: %w", err)
	}
	if err := s.syncRepo.UpsertContributionDays(ctx, dev.ID, days); err != nil {
		return fmt.Errorf("upsert contributions: %w", err)
	}
	log.Info("contributions synced", "days", len(days))

	// 5. Refresh dev record for snapshot
	fresh, err := s.devRepo.GetByID(ctx, dev.ID)
	if err != nil {
		return fmt.Errorf("refresh developer: %w", err)
	}
	if err := s.syncRepo.WriteSnapshot(ctx, fresh); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}

	// 6. Mark synced
	if err := s.syncRepo.UpdateLastSynced(ctx, dev.ID); err != nil {
		return fmt.Errorf("update last_synced_at: %w", err)
	}

	log.Info("sync complete")
	return nil
}
