package poster

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

type ResearchAgent struct {
	executor       *agents.Executor
	searchToolName string
}

// NewResearchAgent creates an agent that must search before answering.
func NewResearchAgent(ctx context.Context, config StepConfig) (*ResearchAgent, error) {
	llm, err := getModel(ctx, config.LLMProvider, config.LLMModel)
	if err != nil {
		return nil, fmt.Errorf("create research model: %w", err)
	}

	search, err := newSearchToolFromEnv()
	if err != nil {
		return nil, fmt.Errorf("create search tool: %w", err)
	}

	agent := agents.NewOpenAIFunctionsAgent(
		llm,
		[]tools.Tool{search},
		agents.NewOpenAIOption().WithSystemMessage(`
		You are a web research assistant.

		For every research request:
		1. Call the available web search tool before answering.
		2. Base factual claims on the search results.
		3. Cite the source URLs returned by the tool as Markdown links.
		4. Never invent URLs or claim to have read pages you did not access.
		5. If the search tool fails, report the failure honestly.
		`),
	)
	executor := agents.NewExecutor(agent, agents.WithMaxIterations(8), agents.WithReturnIntermediateSteps())
	return &ResearchAgent{executor: executor, searchToolName: search.Name()}, nil
}

func (researchAgent *ResearchAgent) Execute(ctx context.Context, prompt string) (string, error) {
	if researchAgent == nil || researchAgent.executor == nil || researchAgent.searchToolName == "" {
		return "", fmt.Errorf("research agent is not initialized")
	}
	result, err := researchAgent.executor.Call(ctx, map[string]any{"input": prompt})
	if err != nil {
		return "", fmt.Errorf("run research agent: %w", err)
	}

	steps, ok := result["intermediateSteps"].([]schema.AgentStep)
	if !ok {
		return "", fmt.Errorf("research agent returned no intermediate steps")
	}
	
	var sourceLines []string
	for _, step := range steps {
		if strings.EqualFold(step.Action.Tool, researchAgent.searchToolName) {
			output, ok := result["output"].(string)
			if !ok {
				return "", fmt.Errorf("research agent returned no string output")
			}
			for _, source := range extractSources(step.Observation) {
				sourceLines = append(sourceLines, fmt.Sprintf("- [%s](%s)", source.Title, source.URL))
			}
			if len(sourceLines) > 0 {
				return output + "\n\nSources:\n" + strings.Join(sourceLines, "\n"), nil
			}
			return output, nil
		}
	}
	return "", fmt.Errorf("research agent did not use the %q tool", researchAgent.searchToolName)
}

type webSearchTool struct {
	apiKey string
	client *http.Client
}
type searchSource struct {
	Title   string `json:"title"`
	URL     string `json:"link"`
	Snippet string `json:"snippet"`
}
type searchResponse struct {
	Organic []searchSource `json:"organic_results"`
}

func newSearchToolFromEnv() (*webSearchTool, error) {
	key := os.Getenv("SERPAPI_API_KEY")
	if key == "" {
		key = os.Getenv("SERPER_API_KEY")
	}
	if key == "" {
		return nil, fmt.Errorf("missing SERPAPI_API_KEY (or SERPER_API_KEY)")
	}
	return &webSearchTool{apiKey: key, client: http.DefaultClient}, nil
}

func (*webSearchTool) Name() string { return "GoogleSearch" }
func (*webSearchTool) Description() string {
	return "Search Google and return result titles, URLs, and snippets. Input is a search query."
}
func (tool *webSearchTool) Call(ctx context.Context, query string) (string, error) {
	u, _ := url.Parse("https://serpapi.com/search")
	params := u.Query()
	params.Set("q", query)
	params.Set("google_domain", "google.com")
	params.Set("gl", "us")
	params.Set("hl", "en")
	params.Set("api_key", tool.apiKey)
	u.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	resp, err := tool.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("SerpAPI returned HTTP %s", resp.Status)
	}
	var data searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode SerpAPI response: %w", err)
	}
	results := data.Organic[:0]
	for _, source := range data.Organic {
		if strings.HasPrefix(source.URL, "https://") || strings.HasPrefix(source.URL, "http://") {
			results = append(results, source)
		}
	}
	if len(results) == 0 {
		return "", fmt.Errorf("SerpAPI returned no organic results with source URLs")
	}
	encoded, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func extractSources(observation string) []searchSource {
	var sources []searchSource
	if err := json.Unmarshal([]byte(observation), &sources); err != nil {
		return nil
	}
	seen := make(map[string]bool)
	unique := sources[:0]
	for _, source := range sources {
		if source.URL != "" && !seen[source.URL] {
			seen[source.URL] = true
			if source.Title == "" {
				source.Title = source.URL
			}
			unique = append(unique, source)
		}
	}
	return unique
}
