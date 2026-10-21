package poster

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type imageRoundTripper struct{ handler http.Handler }

func (roundTrip imageRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	roundTrip.handler.ServeHTTP(recorder, request)
	return recorder.Result(), nil
}

func TestGenerateImageSendsRequestAndSavesImage(t *testing.T) {
	wantImage := []byte("fake image bytes")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q, want bearer token", got)
		}
		var request imageGenerationRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
			return
		}
		if request.Model != "image-model" || request.OutputFormat != "jpeg" || request.N != 1 {
			t.Errorf("request = %+v, want image-model JPEG request", request)
		}
		if request.Prompt != "Create a diagram\n\nPost and image direction:\nA short post" {
			t.Errorf("prompt = %q, want configured prompt and post", request.Prompt)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{{
			"b64_json": base64.StdEncoding.EncodeToString(wantImage),
		}}})
	})
	client := &http.Client{Transport: imageRoundTripper{handler: handler}}

	outputPath := filepath.Join(t.TempDir(), "nested", "post.jpg")
	config := StepConfig{LLMModel: "image-model", PromptTemplate: " Create a diagram ", ImageOutputPath: outputPath}
	got, err := generateImage(context.Background(), config, " A short post ", "test-key", "https://images.example/generate", client)
	if err != nil {
		t.Fatalf("generateImage() error = %v", err)
	}
	if got != outputPath {
		t.Fatalf("generateImage() = %q, want %q", got, outputPath)
	}
	image, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read generated image: %v", err)
	}
	if string(image) != string(wantImage) {
		t.Errorf("image bytes = %q, want %q", image, wantImage)
	}
}

func TestGenerateImageRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name   string
		config StepConfig
		post   string
	}{
		{name: "empty post", config: StepConfig{LLMModel: "model", ImageOutputPath: "post.png"}},
		{name: "missing model", config: StepConfig{ImageOutputPath: "post.png"}, post: "post"},
		{name: "missing output path", config: StepConfig{LLMModel: "model"}, post: "post"},
		{name: "unsupported extension", config: StepConfig{LLMModel: "model", ImageOutputPath: "post.gif"}, post: "post"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := generateImage(context.Background(), test.config, test.post, "key", "http://unused.invalid", http.DefaultClient)
			if err == nil {
				t.Fatal("generateImage() error = nil, want validation error")
			}
		})
	}
}

func TestGenerateImageHandlesAPIErrorsAndBadResponses(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        string
		wantInError string
	}{
		{name: "API error", status: http.StatusBadRequest, body: `{"error":{"message":"invalid prompt"}}`, wantInError: "invalid prompt"},
		{name: "non-success without API error", status: http.StatusServiceUnavailable, body: `{}`, wantInError: "503"},
		{name: "invalid JSON", status: http.StatusOK, body: `not json`, wantInError: "decode image response"},
		{name: "missing image data", status: http.StatusOK, body: `{"data":[]}`, wantInError: "did not contain image data"},
		{name: "invalid base64", status: http.StatusOK, body: `{"data":[{"b64_json":"!"}]}`, wantInError: "decode generated image"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			})
			client := &http.Client{Transport: imageRoundTripper{handler: handler}}

			config := StepConfig{LLMModel: "model", ImageOutputPath: filepath.Join(t.TempDir(), "post.png")}
			_, err := generateImage(context.Background(), config, "post", "key", "https://images.example/generate", client)
			if err == nil || !strings.Contains(err.Error(), test.wantInError) {
				t.Fatalf("generateImage() error = %v, want containing %q", err, test.wantInError)
			}
		})
	}
}

func TestGenerateImageRequiresAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	_, err := GenerateImage(context.Background(), StepConfig{}, "post")
	if err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("GenerateImage() error = %v, want missing API key error", err)
	}
}
