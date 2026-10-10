package poster

// import (
// 	"context"
// 	"fmt"
// 	"regexp"

// 	"github.com/tmc/langchaingo/chains"
// 	"github.com/tmc/langchaingo/prompts"
// )

// var templateInputKey = regexp.MustCompile(`\{\{\s*\.([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)

// // NewChain builds a sequential chain in config order. inputKeys are the keys
// // supplied to the first stage. Later stages consume the preceding result.
// // A config with ImagePath set becomes the final JPEG generation stage.
// func NewChain(configs []ModelConfig, inputKeys []string) (*chains.SequentialChain, error) {
// 	if len(configs) == 0 {
// 		return nil, fmt.Errorf("cannot build an empty chain")
// 	}
// 	if len(inputKeys) == 0 {
// 		return nil, fmt.Errorf("initial input keys cannot be empty")
// 	}

// 	templates := make([]string, len(configs))
// 	stageInputs := make([][]string, len(configs))
// 	for i, config := range configs {
// 		templates[i] = config.PromptTemplate
// 		if i == 0 {
// 			stageInputs[i] = inputKeys
// 			continue
// 		}
// 		keys := uniqueTemplateKeys(config.PromptTemplate)
// 		if len(keys) != 1 {
// 			return nil, fmt.Errorf("config %d must use exactly one input key for the previous stage", i)
// 		}
// 		stageKey := fmt.Sprintf("stage%d", i)
// 		templates[i] = renameTemplateKey(config.PromptTemplate, keys[0], stageKey)
// 		stageInputs[i] = []string{stageKey}
// 	}

// 	steps := make([]chains.Chain, 0, len(configs))
// 	for i, config := range configs {
// 		outputKey := "output"
// 		if i < len(configs)-1 {
// 			outputKey = stageInputs[i+1][0]
// 		}
// 		if config.ImagePath != "" {
// 			if i != len(configs)-1 {
// 				return nil, fmt.Errorf("image config must be the last chain stage")
// 			}
// 			inputKey := stageInputs[i][0]
// 			imageStep := NewImage(config.ImagePath)
// 			image := chains.NewTransform(func(ctx context.Context, values map[string]any, _ ...chains.ChainCallOption) (map[string]any, error) {
// 				input, ok := values[inputKey].(string)
// 				if !ok {
// 					return nil, fmt.Errorf("image input %q must be a string", inputKey)
// 				}
// 				prompt := input
// 				if templates[i] != "" {
// 					var err error
// 					prompt, err = prompts.NewPromptTemplate(templates[i], stageInputs[i]).Format(values)
// 					if err != nil {
// 						return nil, err
// 					}
// 				}
// 				path, err := imageStep.Execute(ctx, prompt)
// 				if err != nil {
// 					return nil, err
// 				}
// 				return map[string]any{"imagePath": path}, nil
// 			}, stageInputs[i], []string{"imagePath"})
// 			steps = append(steps, image)
// 			continue
// 		}

// 		model, err := newModel(ModelConfig{LLMProvider: config.LLMProvider, LLMModel: config.LLMModel})
// 		if err != nil {
// 			return nil, fmt.Errorf("create model for config %d: %w", i, err)
// 		}
// 		llmChain := chains.NewLLMChain(model, prompts.NewPromptTemplate(templates[i], stageInputs[i]))
// 		llmChain.OutputKey = outputKey
// 		steps = append(steps, llmChain)
// 	}

// 	outputKey := steps[len(steps)-1].GetOutputKeys()[0]
// 	return chains.NewSequentialChain(steps, inputKeys, []string{outputKey})
// }

// func uniqueTemplateKeys(prompt string) []string {
// 	seen := make(map[string]bool)
// 	var keys []string
// 	for _, match := range templateInputKey.FindAllStringSubmatch(prompt, -1) {
// 		if !seen[match[1]] {
// 			seen[match[1]] = true
// 			keys = append(keys, match[1])
// 		}
// 	}
// 	return keys
// }

// func renameTemplateKey(prompt, oldKey, newKey string) string {
// 	keyPattern := regexp.MustCompile(`\{\{\s*\.` + regexp.QuoteMeta(oldKey) + `\s*\}\}`)
// 	return keyPattern.ReplaceAllString(prompt, "{{."+newKey+"}}")
// }
