package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	poster "github.com/alejandropriv/poster/postwriter"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/chains"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("No .env file loaded: %v", err)
	}
	topic := "wWhen is worth it to have kubernetes in production instead of a simple vm or vm with docker"
	if err := researchStepConfig.SetConfig(map[string]any{
		"topic": topic,
	}); err != nil {
		log.Fatal(err)
	}
	filenameLength := 20
	filename := topicFilename(topic, filenameLength)
	ctx := context.Background()
	researchAgent, err := poster.NewResearchAgent(ctx, researchStepConfig)
	if err != nil {
		log.Fatalf("Failed to create research agent: %v", err)
	}
	println(researchStepConfig.Prompt)
	research, err := researchAgent.Execute(ctx, researchStepConfig.Prompt)
	if err != nil {
		log.Fatalf("Failed to run research agent: %v", err)
	}

	workflow, err := poster.NewSequentialChain(ctx, writeStepConfig, optimizeStepConfig)
	if err != nil {
		log.Fatalf("Failed to build content chain: %v", err)
	}
	result, err := chains.Call(ctx, workflow, map[string]any{
		writeStepConfig.InputKey: research,
	})
	if err != nil {
		log.Fatalf("Failed to run content chain: %v", err)
	}
	post, ok := result[optimizeStepConfig.OutputKey].(string)
	if !ok {
		log.Fatalf("Content chain returned no %q output", optimizeStepConfig.OutputKey)
	}
	println(post)
	postPath := filepath.Join("POSTS", filename+".md")
	if err := savePost(postPath, post); err != nil {
		log.Fatalf("Failed to save post: %v", err)
	}
	log.Printf("Saved post: %s", postPath)

	imageStepConfig.ImageOutputPath = filepath.Join("POSTS", filename+".png")
	imagePath, err := poster.GenerateImage(ctx, imageStepConfig, post)
	if err != nil {
		log.Fatalf("Failed to generate image: %v", err)
	}
	log.Printf("Generated image: %s", imagePath)
}

var filenameSeparator = regexp.MustCompile(`[^a-z0-9]+`)

func topicFilename(topic string, filenameLength int) string {
	slug := filenameSeparator.ReplaceAllString(strings.ToLower(strings.TrimSpace(topic)), "-")
	slug = strings.Trim(slug, "-")

	if len(slug) > filenameLength {
		slug = strings.Trim(slug[:filenameLength], "-")
	}
	if slug == "" {
		return "post"
	}
	return slug
}

func savePost(path, post string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(post+"\n"), 0644)
}
