package poster

import (
	"context"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

type recordingResearchModel struct {
	prompt string
}

func (model *recordingResearchModel) GenerateContent(_ context.Context, messages []llms.MessageContent, _ ...llms.CallOption) (*llms.ContentResponse, error) {
	for _, message := range messages {
		for _, part := range message.Parts {
			if text, ok := part.(llms.TextContent); ok {
				model.prompt += text.Text
			}
		}
	}
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "answer with cited source"}}}, nil
}

func (*recordingResearchModel) Call(context.Context, string, ...llms.CallOption) (string, error) {
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
	const searchResult = `{"title":"Kubernetes v1.35","link":"https://kubernetes.io/blog/release/"}`
	model := &recordingResearchModel{}
	search := &fixedSearchTool{result: searchResult}
	agent := &ResearchAgent{llm: model, search: search}

	got, err := agent.Execute(context.Background(), query)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got != "answer with cited source" {
		t.Fatalf("Execute() = %q; want answer from model", got)
	}
	if search.query != query {
		t.Errorf("search query = %q; want %q", search.query, query)
	}
	if !strings.Contains(model.prompt, searchResult) {
		t.Errorf("model prompt %q does not contain search result %q", model.prompt, searchResult)
	}
}
