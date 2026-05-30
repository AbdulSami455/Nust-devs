package worker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TaskSyncDeveloper = "sync:developer"
	TaskSyncAll       = "sync:all"
	TaskSyncActive    = "sync:active"
)

type SyncDeveloperPayload struct {
	DeveloperID     string `json:"developer_id"`
	GithubUsername  string `json:"github_username"`
}

func NewSyncDeveloperTask(devID, username string) (*asynq.Task, error) {
	payload, err := json.Marshal(SyncDeveloperPayload{DeveloperID: devID, GithubUsername: username})
	if err != nil {
		return nil, fmt.Errorf("marshal sync task: %w", err)
	}
	return asynq.NewTask(
		TaskSyncDeveloper,
		payload,
		asynq.Timeout(10*time.Minute),
		asynq.MaxRetry(3),
	), nil
}
