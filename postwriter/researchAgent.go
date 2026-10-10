package poster

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/serpapi"
)

type ResearchAgent struct {
	llm    llms.Model
	search tools.Tool
}

// NewResearchAgent creates an agent whose instructions and model come from the
// research config. Search runs before the agent drafts its response.
func NewResearchAgent(ctx context.Context, config StepConfig) (*ResearchAgent, error) {
	llm, err := getModel(ctx, config.LLMProvider, config.LLMModel)
	if err != nil {
		return nil, fmt.Errorf("create research model: %w", err)
	}

	search, err := serpapi.New()
	if err != nil {
		return nil, fmt.Errorf("create Serper tool: %w", err)
	}

	return &ResearchAgent{llm: llm, search: search}, nil
}

func (researchAgent *ResearchAgent) Execute(ctx context.Context, prompt string) (string, error) {
	if researchAgent == nil || researchAgent.llm == nil || researchAgent.search == nil {
		return "", fmt.Errorf("research agent is not initialized")
	}

	searchResults, err := researchAgent.search.Call(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("search for research prompt: %w", err)
	}

	modelPrompt := fmt.Sprintf(`User request:
%s

Use the following live web-search results as the evidence for your answer:
%s

Answer the request using facts supported by these results. Cite the source URLs in your answer. If the results do not establish an answer, say so instead of guessing.`, prompt, searchResults)
	result, err := llms.GenerateFromSinglePrompt(ctx, researchAgent.llm, modelPrompt)
	if err != nil {
		return "", fmt.Errorf("generate research answer: %w", err)
	}
	return result, nil
}
