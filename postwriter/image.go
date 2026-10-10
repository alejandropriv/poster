package poster

// import (
// 	"bytes"
// 	"context"
// 	"encoding/base64"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"time"
// )

// const openAIImagesURL = "https://api.openai.com/v1/images/generations"
// const latestOpenAIImageModel = "gpt-image-2.5-sunburst"

// type imageRequest struct {
// 	Model        string `json:"model"`
// 	Prompt       string `json:"prompt"`
// 	OutputFormat string `json:"output_format"`
// }

// type imageResponse struct {
// 	Data []struct {
// 		Base64 string `json:"b64_json"`
// 	} `json:"data"`
// 	Error *struct {
// 		Message string `json:"message"`
// 	} `json:"error"`
// }

// // Image generates an image from the preceding step's text and saves it locally.
// type Image struct {
// 	model      string
// 	outputPath string
// 	endpoint   string
// 	client     *http.Client
// }

// func NewImage(outputPath string) Image {
// 	return Image{
// 		model:      latestOpenAIImageModel,
// 		outputPath: outputPath,
// 		endpoint:   openAIImagesURL,
// 		client:     &http.Client{Timeout: 5 * time.Minute},
// 	}
// }

// func (step Image) Execute(ctx context.Context, prompt string) (string, error) {
// 	if strings.TrimSpace(prompt) == "" {
// 		return "", fmt.Errorf("image prompt is empty")
// 	}
// 	if strings.TrimSpace(step.outputPath) == "" {
// 		return "", fmt.Errorf("image output path is empty")
// 	}
// 	extension := strings.ToLower(filepath.Ext(step.outputPath))
// 	if extension != ".jpg" && extension != ".jpeg" {
// 		return "", fmt.Errorf("image output path must end in .jpg or .jpeg")
// 	}
// 	apiKey := os.Getenv("OPENAI_API_KEY")
// 	if apiKey == "" {
// 		return "", fmt.Errorf("OPENAI_API_KEY is not set")
// 	}

// 	body, err := json.Marshal(imageRequest{
// 		Model:        step.model,
// 		Prompt:       prompt,
// 		OutputFormat: "jpeg",
// 	})
// 	if err != nil {
// 		return "", err
// 	}
// 	request, err := http.NewRequestWithContext(ctx, http.MethodPost, step.endpoint, bytes.NewReader(body))
// 	if err != nil {
// 		return "", err
// 	}
// 	request.Header.Set("Authorization", "Bearer "+apiKey)
// 	request.Header.Set("Content-Type", "application/json")

// 	response, err := step.client.Do(request)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer response.Body.Close()

// 	var result imageResponse
// 	if err := json.NewDecoder(io.LimitReader(response.Body, 10<<20)).Decode(&result); err != nil {
// 		return "", fmt.Errorf("decode image response: %w", err)
// 	}
// 	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
// 		if result.Error != nil {
// 			return "", fmt.Errorf("image generation failed: %s", result.Error.Message)
// 		}
// 		return "", fmt.Errorf("image generation failed with status %s", response.Status)
// 	}
// 	if len(result.Data) == 0 || result.Data[0].Base64 == "" {
// 		return "", fmt.Errorf("image response did not contain image data")
// 	}

// 	imageBytes, err := base64.StdEncoding.DecodeString(result.Data[0].Base64)
// 	if err != nil {
// 		return "", fmt.Errorf("decode generated image: %w", err)
// 	}
// 	if dir := filepath.Dir(step.outputPath); dir != "." {
// 		if err := os.MkdirAll(dir, 0755); err != nil {
// 			return "", fmt.Errorf("create image output directory: %w", err)
// 		}
// 	}
// 	if err := os.WriteFile(step.outputPath, imageBytes, 0644); err != nil {
// 		return "", fmt.Errorf("write generated image: %w", err)
// 	}
// 	return step.outputPath, nil
// }
