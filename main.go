package main

import (
	"context"
	"log"

	poster "github.com/alejandropriv/poster/postwriter"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/chains"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("No .env file loaded: %v", err)
	}
	topic := "Who is the current colombian president?"
	if err := researchStepConfig.SetConfig(map[string]any{
		"topic": topic,
	}); err != nil {
		log.Fatal(err)
	}
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

	imagePath, err := poster.GenerateImage(ctx, imageStepConfig, post)
	if err != nil {
		log.Fatalf("Failed to generate image: %v", err)
	}
	log.Printf("Generated image: %s", imagePath)
}
