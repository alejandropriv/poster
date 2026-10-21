package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigurationUsesEnvironmentBeforeFileAndFallsBackToFile(t *testing.T) {
	envFile := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envFile, []byte("OPENAI_API_KEY=file-openai\nSERPAPI_API_KEY=file-serp\n"), 0600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	t.Setenv("OPENAI_API_KEY", "secret-openai")
	t.Setenv("SERPAPI_API_KEY", "")

	if err := loadConfiguration(envFile); err != nil {
		t.Fatalf("loadConfiguration() error = %v", err)
	}
	if got := os.Getenv("OPENAI_API_KEY"); got != "secret-openai" {
		t.Errorf("OPENAI_API_KEY = %q, want environment value", got)
	}
	if got := os.Getenv("SERPAPI_API_KEY"); got != "file-serp" {
		t.Errorf("SERPAPI_API_KEY = %q, want file fallback", got)
	}
}

func TestHealthzEndpoint(t *testing.T) {
	handler := newAPIHandler(func(context.Context, string) (createPostResponse, error) {
		t.Fatal("generator must not be called for health checks")
		return createPostResponse{}, nil
	})

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body)
	}
	if got := response.Body.String(); !strings.Contains(got, `"status":"ok"`) {
		t.Errorf("body = %q, want health status", got)
	}
}

func TestGeneratedImageEndpoint(t *testing.T) {
	imageDirectory := t.TempDir()
	imageBytes := []byte("generated png")
	if err := os.WriteFile(filepath.Join(imageDirectory, "post.png"), imageBytes, 0600); err != nil {
		t.Fatalf("write generated image: %v", err)
	}

	handler := generatedImageHandler(imageDirectory)
	request := httptest.NewRequest(http.MethodGet, "/POSTS/post.png", nil)
	request.SetPathValue("filename", "post.png")
	response := httptest.NewRecorder()
	handler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body)
	}
	if got := response.Body.Bytes(); string(got) != string(imageBytes) {
		t.Errorf("body = %q, want %q", got, imageBytes)
	}
	if got := response.Header().Get("Content-Type"); got != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", got)
	}
}

func TestLoadConfigurationAllowsMissingFileButRequiresKeys(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "openai")
	t.Setenv("SERPAPI_API_KEY", "serp")
	if err := loadConfiguration(filepath.Join(t.TempDir(), "missing.env")); err != nil {
		t.Fatalf("loadConfiguration() with environment values error = %v", err)
	}

	t.Setenv("SERPAPI_API_KEY", " ")
	err := loadConfiguration(filepath.Join(t.TempDir(), "missing.env"))
	if err == nil || !strings.Contains(err.Error(), "SERPAPI_API_KEY") {
		t.Fatalf("loadConfiguration() error = %v, want missing SERPAPI_API_KEY", err)
	}
}

func TestCreatePostEndpoint(t *testing.T) {
	var generatedTopic string
	handler := newAPIHandler(func(_ context.Context, topic string) (createPostResponse, error) {
		generatedTopic = topic
		return createPostResponse{Topic: topic, Post: "A generated post", ImagePath: "POSTS/post.png"}, nil
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(`{"topic":"  Kubernetes   in production  "}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusCreated, response.Body)
	}
	if generatedTopic != "Kubernetes in production" {
		t.Fatalf("generated topic = %q, want normalized topic", generatedTopic)
	}
	var payload createPostResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Post != "A generated post" || payload.ImagePath != "POSTS/post.png" {
		t.Fatalf("unexpected response: %+v", payload)
	}
}

func TestCreatePostRejectsInvalidTopicsAndRequests(t *testing.T) {
	handler := newAPIHandler(func(context.Context, string) (createPostResponse, error) {
		t.Fatal("generator must not be called for invalid requests")
		return createPostResponse{}, nil
	})
	tests := []struct {
		name        string
		body        string
		contentType string
		wantStatus  int
	}{
		{name: "missing topic", body: `{}`, contentType: "application/json", wantStatus: http.StatusBadRequest},
		{name: "blank topic", body: `{"topic":"   "}`, contentType: "application/json", wantStatus: http.StatusBadRequest},
		{name: "control character", body: "{\"topic\":\"hello\\u0000world\"}", contentType: "application/json", wantStatus: http.StatusBadRequest},
		{name: "too long", body: `{"topic":"` + strings.Repeat("a", maxTopicLength+1) + `"}`, contentType: "application/json", wantStatus: http.StatusBadRequest},
		{name: "unknown field", body: `{"topic":"hello","other":"value"}`, contentType: "application/json", wantStatus: http.StatusBadRequest},
		{name: "invalid JSON", body: `{"topic":`, contentType: "application/json", wantStatus: http.StatusBadRequest},
		{name: "wrong content type", body: `{"topic":"hello"}`, contentType: "text/plain", wantStatus: http.StatusUnsupportedMediaType},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(test.body))
			request.Header.Set("Content-Type", test.contentType)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", response.Code, test.wantStatus, response.Body)
			}
		})
	}
}

func TestSanitizeTopic(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "collapse whitespace", input: "  Go   for   production ", want: "Go for production"},
		{name: "unicode", input: "café ☕", want: "café ☕"},
		{name: "blank", input: " \t ", wantErr: true},
		{name: "control character", input: "topic\x7f", wantErr: true},
		{name: "too long", input: strings.Repeat("界", maxTopicLength+1), wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := sanitizeTopic(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("sanitizeTopic() error = %v, wantErr %t", err, test.wantErr)
			}
			if err == nil && got != test.want {
				t.Errorf("sanitizeTopic() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestTopicFilename(t *testing.T) {
	tests := []struct {
		name  string
		topic string
		limit int
		want  string
	}{
		{name: "slugify", topic: "  Go & Kubernetes: In Production! ", limit: 40, want: "go-kubernetes-in-production"},
		{name: "trim truncated separator", topic: "a very long topic", limit: 6, want: "a-very"},
		{name: "empty slug fallback", topic: "こんにちは", limit: 20, want: "post"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := topicFilename(test.topic, test.limit); got != test.want {
				t.Errorf("topicFilename() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSavePostCreatesDirectoryAndAppendsNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "post.md")
	if err := savePost(path, "post content"); err != nil {
		t.Fatalf("savePost() error = %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved post: %v", err)
	}
	if string(got) != "post content\n" {
		t.Errorf("saved post = %q, want %q", got, "post content\n")
	}
}
