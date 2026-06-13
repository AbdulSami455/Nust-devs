package ai

import (
	"context"
	"fmt"

	adkmodel "google.golang.org/adk/model"

	"github.com/abdulsami/nust-devs/internal/config"
)

// NewLLM returns an ADK-compatible model backed by OpenRouter.
func NewLLM(_ context.Context, cfg *config.Config) (adkmodel.LLM, error) {
	if cfg.OpenRouterKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY is required")
	}
	return newOpenRouterModel(cfg.OpenRouterKey, cfg.AIModel), nil
}
