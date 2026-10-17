package poster

import (
	"context"
	"fmt"
	"time"

	langchains "github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/prompts"
)

// NewSequentialChain builds a sequential chain from its ordered step configs.
func NewSequentialChain(ctx context.Context, steps ...StepConfig) (*langchains.SequentialChain, error) {
	if len(steps) == 0 {
		return nil, fmt.Errorf("sequential chain requires at least one step")
	}

	chainSteps := make([]langchains.Chain, 0, len(steps))
	for i, step := range steps {
		if step.InputKey == "" || step.OutputKey == "" {
			return nil, fmt.Errorf("chain step %d must define input and output keys", i)
		}

		model, err := getModel(ctx, step.LLMProvider, step.LLMModel)
		if err != nil {
			return nil, fmt.Errorf("create model for chain step %d: %w", i, err)
		}
		prompt := prompts.NewPromptTemplate(step.PromptTemplate, []string{step.InputKey})
		prompt.PartialVariables = map[string]any{
			"date": func() string { return time.Now().Local().Format("2006-01-02") },
		}
		llmChain := langchains.NewLLMChain(model, prompt)
		llmChain.OutputKey = step.OutputKey
		chainSteps = append(chainSteps, llmChain)
	}

	return langchains.NewSequentialChain(
		chainSteps,
		[]string{steps[0].InputKey},
		[]string{steps[len(steps)-1].OutputKey},
	)
}
