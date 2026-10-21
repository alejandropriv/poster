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
