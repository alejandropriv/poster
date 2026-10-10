package main

import (
	"context"
	// "fmt"
	"log"
	//"time"

	poster "github.com/alejandropriv/poster/postwriter"
	"github.com/joho/godotenv"
	//"github.com/tmc/langchaingo/chains"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("No .env file loaded: %v", err)
	}
	if err := researchStepConfig.SetConfig(map[string]any{
		"topic": "What is the kubernetes latest version and what are the new features?",
	}); err != nil {
		log.Fatal(err)
	}

	println(researchStepConfig.Prompt)

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
	println(research)





	
	// research, err := chains.Run(context.Background(), researchAgent,
	// 	"Research the requested topic using Google Search and return concise notes with source links.")
	// if err != nil {
	// 	log.Fatalf("Research agent failed: %v", err)
	// }

	// configs := make([]poster.ModelConfig, 0, 3)
	// for _, path := range []string{"conf_write.txt", "conf_optimize.txt", "conf_image.txt"} {
	// 	config, err := poster.LoadLLMConfig(path)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	configs = append(configs, config)
	// }

	// inputKeys := []string{"research"}
	// workflow, err := poster.NewChain(configs, inputKeys)
	// if err != nil {
	// 	log.Fatalf("Failed to build content chain: %v", err)
	// }
	// result, err := chains.Call(context.Background(), workflow, map[string]any{
	// 	"research": research,
	// })
	// if err != nil {
	// 	log.Fatalf("Failed to run content chain: %v", err)
	// }
	// fmt.Printf("Generated image: %s\n", result["imagePath"])
}
