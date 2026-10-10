package poster

import (
	"fmt"
	"time"

	"github.com/tmc/langchaingo/prompts"
)

// StepConfig selects a provider and model for one chain stage.
type StepConfig struct {
	LLMProvider     string
	LLMModel        string
	PromptTemplate  string
	ImageOutputPath string
	Prompt          string
}

func (config *StepConfig) SetConfig(inputParams map[string]any) error {
	prompt, err := setPrompt(config.PromptTemplate, inputParams)
	if err != nil {
		return fmt.Errorf("setting prompt: %w", err)
	}
	config.Prompt = prompt
	return nil
}

func setPrompt(promptTemplate string, inputParams map[string]any) (string, error) {

	stringTemplate := prompts.PromptTemplate{
		Template: promptTemplate,
		PartialVariables: map[string]any{
			"date": func() string {
				return time.Now().Local().Format("2006-01-02")
			},
		},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
	}

	prompt, err := stringTemplate.Format(inputParams)
	if err != nil {
		return "", fmt.Errorf("formatting prompt: %w", err)
	}

	return prompt, nil

}
