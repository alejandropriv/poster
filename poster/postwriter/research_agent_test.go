package poster

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

type searchRoundTripper struct{ handler http.Handler }

func (roundTrip searchRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	roundTrip.handler.ServeHTTP(recorder, request)
	return recorder.Result(), nil
}

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

func TestWebSearchToolFiltersInvalidURLsAndSetsParameters(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		query := r.URL.Query()
		for key, want := range map[string]string{
			"q": "Go concurrency", "google_domain": "google.com", "gl": "us", "hl": "en", "api_key": "test-key",
		} {
			if got := query.Get(key); got != want {
				t.Errorf("query parameter %s = %q, want %q", key, got, want)
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"organic_results": []map[string]string{
			{"title": "Go blog", "link": "https://go.dev/blog/", "snippet": "Official blog"},
			{"title": "Bad URL", "link": "javascript:alert(1)", "snippet": "Ignore"},
			{"title": "HTTP source", "link": "http://example.com/", "snippet": "Allowed"},
		}})
	})

	tool := &webSearchTool{apiKey: "test-key", client: &http.Client{Transport: searchRoundTripper{handler}}, endpoint: "https://search.example/"}
	got, err := tool.Call(context.Background(), "Go concurrency")
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	var results []searchSource
	if err := json.Unmarshal([]byte(got), &results); err != nil {
		t.Fatalf("decode search results: %v", err)
	}
	if len(results) != 2 || results[0].URL != "https://go.dev/blog/" || results[1].URL != "http://example.com/" {
		t.Errorf("results = %+v, want only HTTP(S) sources", results)
	}
}

func TestWebSearchToolReportsFailures(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        string
		wantInError string
	}{
		{name: "HTTP failure", status: http.StatusForbidden, body: `forbidden`, wantInError: "HTTP 403"},
		{name: "invalid JSON", status: http.StatusOK, body: `not json`, wantInError: "decode SerpAPI response"},
		{name: "no valid sources", status: http.StatusOK, body: `{"organic_results":[{"title":"bad","link":"file:///tmp/post"}]}`, wantInError: "no organic results"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			})

			tool := &webSearchTool{apiKey: "key", client: &http.Client{Transport: searchRoundTripper{handler}}, endpoint: "https://search.example/"}
			_, err := tool.Call(context.Background(), "query")
			if err == nil || !strings.Contains(err.Error(), test.wantInError) {
				t.Fatalf("Call() error = %v, want containing %q", err, test.wantInError)
			}
		})
	}
}

func TestExtractSourcesDeduplicatesAndUsesURLForMissingTitle(t *testing.T) {
	sources := extractSources(`[
		{"title":"First","link":"https://example.com/"},
		{"title":"Duplicate","link":"https://example.com/"},
		{"title":"","link":"https://other.example/"},
		{"title":"No URL"}
	]`)
	if len(sources) != 2 {
		t.Fatalf("extractSources() returned %d sources, want 2: %+v", len(sources), sources)
	}
	if sources[0].Title != "First" || sources[1].Title != "https://other.example/" {
		t.Errorf("sources = %+v, want first entry preserved and missing title filled from URL", sources)
	}
	if got := extractSources("not json"); got != nil {
		t.Errorf("extractSources(invalid JSON) = %+v, want nil", got)
	}
}

func TestWebSearchToolRejectsMalformedEndpoint(t *testing.T) {
	tool := &webSearchTool{apiKey: "key", client: http.DefaultClient, endpoint: "://bad"}
	if _, err := tool.Call(context.Background(), "query"); err == nil || !strings.Contains(err.Error(), "parse SerpAPI endpoint") {
		t.Fatalf("Call() error = %v, want invalid endpoint error", err)
	}
}
