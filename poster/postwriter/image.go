package poster

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const openAIImageGenerationURL = "https://api.openai.com/v1/images/generations"

type imageGenerationRequest struct {
	Model        string `json:"model"`
	Prompt       string `json:"prompt"`
	N            int    `json:"n"`
	OutputFormat string `json:"output_format"`
}

type imageGenerationResponse struct {
	Data []struct {
		Base64 string `json:"b64_json"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// GenerateImage creates an image from the final post and saves it to config.ImageOutputPath.
func GenerateImage(ctx context.Context, config StepConfig, post string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is not set")
	}
	return generateImage(ctx, config, post, apiKey, openAIImageGenerationURL, &http.Client{Timeout: 5 * time.Minute})
}

func generateImage(ctx context.Context, config StepConfig, post, apiKey, endpoint string, client *http.Client) (string, error) {
	if strings.TrimSpace(post) == "" {
		return "", fmt.Errorf("post is empty")
	}
	if strings.TrimSpace(config.LLMModel) == "" {
		return "", fmt.Errorf("image model is not configured")
	}
	if strings.TrimSpace(config.ImageOutputPath) == "" {
		return "", fmt.Errorf("image output path is not configured")
	}
	outputFormat := strings.TrimPrefix(strings.ToLower(filepath.Ext(config.ImageOutputPath)), ".")
	if outputFormat != "png" && outputFormat != "jpeg" && outputFormat != "jpg" && outputFormat != "webp" {
		return "", fmt.Errorf("image output path must end in .png, .jpg, .jpeg, or .webp")
	}
	if outputFormat == "jpg" {
		outputFormat = "jpeg"
	}

	prompt := strings.TrimSpace(config.PromptTemplate) + "\n\nPost and image direction:\n" + strings.TrimSpace(post)
	body, err := json.Marshal(imageGenerationRequest{
		Model:        config.LLMModel,
		Prompt:       prompt,
		N:            1,
		OutputFormat: outputFormat,
	})
	if err != nil {
		return "", fmt.Errorf("encode image request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create image request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("send image request: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 20<<20))
	if err != nil {
		return "", fmt.Errorf("read image response: %w", err)
	}
	var result imageGenerationResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", fmt.Errorf("decode image response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		if result.Error != nil {
			return "", fmt.Errorf("image generation failed: %s", result.Error.Message)
		}
		return "", fmt.Errorf("image generation failed with status %s", response.Status)
	}
	if len(result.Data) == 0 || result.Data[0].Base64 == "" {
		return "", fmt.Errorf("image response did not contain image data")
	}

	imageBytes, err := base64.StdEncoding.DecodeString(result.Data[0].Base64)
	if err != nil {
		return "", fmt.Errorf("decode generated image: %w", err)
	}
	if dir := filepath.Dir(config.ImageOutputPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("create image output directory: %w", err)
		}
	}
	if err := os.WriteFile(config.ImageOutputPath, imageBytes, 0644); err != nil {
		return "", fmt.Errorf("save generated image: %w", err)
	}
	return config.ImageOutputPath, nil
}
