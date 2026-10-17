package poster

import (
	"testing"
	"time"
)

func TestStepConfigSetPrompt(t *testing.T) {
	config := StepConfig{PromptTemplate: "Topic: {{.topic}}\nDate: {{.date}}"}
	wantDate := time.Now().Local().Format("2006-01-02")

	if err := config.SetConfig(map[string]any{"topic": "Go templates"}); err != nil {
		t.Fatalf("SetPrompt() error = %v", err)
	}

	want := "Topic: Go templates\nDate: " + wantDate
	if config.Prompt != want {
		t.Errorf("Prompt = %q; want %q", config.Prompt, want)
	}
}

func TestStepConfigSetPromptInvalidTemplate(t *testing.T) {
	config := StepConfig{
		PromptTemplate: "Topic: {{.topic",
		Prompt:         "previous prompt",
	}

	err := config.SetConfig(map[string]any{"topic": "Go templates"})
	if err == nil {
		t.Fatal("SetPrompt() error = nil; want malformed-template error")
	}
	if config.Prompt != "previous prompt" {
		t.Errorf("Prompt = %q after formatting error; want previous prompt", config.Prompt)
	}
}

// func TestImageStepWritesJPEGResponse(t *testing.T) {
// 	wantImage := []byte("fake jpeg bytes")
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		var request imageRequest
// 		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
// 			t.Errorf("decode request: %v", err)
// 		}
// 		if request.Model != latestOpenAIImageModel || request.OutputFormat != "jpeg" {
// 			t.Errorf("request = %#v; want latest model and JPEG output", request)
// 		}
// 		if request.Prompt != "optimized post" {
// 			t.Errorf("prompt = %q; want optimized post", request.Prompt)
// 		}
// 		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{{
// 			"b64_json": base64.StdEncoding.EncodeToString(wantImage),
// 		}}})
// 	}))
// 	defer server.Close()
// 	t.Setenv("OPENAI_API_KEY", "test-key")

// 	outputPath := filepath.Join(t.TempDir(), "post.jpg")
// 	step := NewImage(outputPath)
// 	step.endpoint = server.URL
// 	step.client = server.Client()

// 	got, err := step.Execute(context.Background(), "optimized post")
// 	if err != nil {
// 		t.Fatalf("Execute() error = %v", err)
// 	}
// 	if got != outputPath {
// 		t.Fatalf("Execute() = %q; want %q", got, outputPath)
// 	}
// 	gotImage, err := os.ReadFile(outputPath)
// 	if err != nil {
// 		t.Fatalf("read image: %v", err)
// 	}
// 	if string(gotImage) != string(wantImage) {
// 		t.Fatalf("image = %q; want %q", gotImage, wantImage)
// 	}
// }
