package poster

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
)

func getModel(context context.Context, llProvider string, llModel string) (llms.Model, error) {

	switch strings.ToLower(strings.TrimSpace(llProvider)) {
	case "openai":
		return openai.New(openai.WithModel(llModel))
	case "googleai", "gemini":
		return googleai.New(context, googleai.WithDefaultModel(llModel))
	default:
		return nil, fmt.Errorf("unsupported text model provider %q", llProvider)
	}
}
