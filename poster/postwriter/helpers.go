package poster

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
)

func getModel(ctx context.Context, provider, model string) (llms.Model, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai":
		return openai.New(openai.WithModel(model))
	case "googleai", "gemini":
		return googleai.New(ctx, googleai.WithDefaultModel(model))
	default:
		return nil, fmt.Errorf("unsupported text model provider %q", provider)
	}
}
