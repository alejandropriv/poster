package poster

import (
	"context"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

type scriptedResearchModel struct {
	calls       int
	toolResults string
}

func (model *scriptedResearchModel) GenerateContent(_ context.Context, messages []llms.MessageContent, _ ...llms.CallOption) (*llms.ContentResponse, error) {
	model.calls++
	if model.calls == 1 {
		return &llms.ContentResponse{Choices: []*llms.ContentChoice{{
			ToolCalls: []llms.ToolCall{{
				ID:   "search-call",
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      "search",
					Arguments: `{"__arg1":"What is the latest Kubernetes version?"}`,
				},
			}},
		}}}, nil
	}
	for _, message := range messages {
		for _, part := range message.Parts {
			if result, ok := part.(llms.ToolCallResponse); ok {
				model.toolResults += result.Content
			}
		}
	}
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "answer with cited source"}}}, nil
}

func (*scriptedResearchModel) Call(context.Context, string, ...llms.CallOption) (string, error) {
	return "", nil
}

type fixedSearchTool struct {
	result string
	query  string
}

func (*fixedSearchTool) Name() string { return "search" }

func (*fixedSearchTool) Description() string { return "Search the web" }

func (tool *fixedSearchTool) Call(_ context.Context, query string) (string, error) {
	tool.query = query
	return tool.result, nil
}

func TestResearchAgentUsesSearchResults(t *testing.T) {
	const query = "What is the latest Kubernetes version?"
	const searchResult = `[{"title":"Kubernetes v1.35","link":"https://kubernetes.io/blog/release/","snippet":"Release information"}]`
	model := &scriptedResearchModel{}
	search := &fixedSearchTool{result: searchResult}
	functionAgent := agents.NewOpenAIFunctionsAgent(
		model,
		[]tools.Tool{search},
		agents.NewOpenAIOption().WithSystemMessage("Always search before answering."),
	)
	agent := &ResearchAgent{
		executor:       agents.NewExecutor(functionAgent, agents.WithMaxIterations(3), agents.WithReturnIntermediateSteps()),
		searchToolName: search.Name(),
	}

	got, err := agent.Execute(context.Background(), query)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.HasPrefix(got, "answer with cited source") {
		t.Fatalf("Execute() = %q; want answer from model", got)
	}
	if !strings.Contains(got, "[Kubernetes v1.35](https://kubernetes.io/blog/release/)") {
		t.Errorf("Execute() did not include a source link: %q", got)
	}
	if search.query != query {
		t.Errorf("search query = %q; want %q", search.query, query)
	}
	if !strings.Contains(model.toolResults, searchResult) {
		t.Errorf("model tool results %q do not contain search result %q", model.toolResults, searchResult)
	}
}
