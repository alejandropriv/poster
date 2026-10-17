package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	poster "github.com/alejandropriv/poster/postwriter"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/chains"
)

const maxTopicLength = 500

type createPostRequest struct {
	Topic string `json:"topic"`
}

type createPostResponse struct {
	Topic     string `json:"topic"`
	Post      string `json:"post"`
	ImagePath string `json:"imagePath"`
}

type postGenerator func(context.Context, string) (createPostResponse, error)

func main() {
	if err := loadConfiguration(".env"); err != nil {
		log.Fatal(err)
	}
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           newAPIHandler(generatePost),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("Poster API listening on %s", server.Addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func loadConfiguration(envFile string) error {
	fileValues, err := godotenv.Read(envFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read environment file: %w", err)
	}
	for key, value := range fileValues {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("load %s from environment file: %w", key, err)
			}
		}
	}
	for _, key := range []string{"OPENAI_API_KEY", "SERPAPI_API_KEY"} {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			return fmt.Errorf("missing required configuration %s", key)
		}
	}
	return nil
}

func newAPIHandler(generate postGenerator) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /	docker buildx build --platform linux/amd64 --tag post	docker buildx build --platform linux/amd64 --tag post	docker buildx build --platform linux/amd64 --tag post	docker buildx build --platform linux/amd64 --tag post	docker buildx build --platform linux/amd64 --tag poster-api:latest --load .", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/v1/posts", func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			writeAPIError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var request createPostRequest
		if err := decoder.Decode(&request); err != nil {
			writeDecodeError(w, err)
			return
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			writeAPIError(w, http.StatusBadRequest, "request must contain one JSON object")
			return
		}

		topic, err := sanitizeTopic(request.Topic)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, err.Error())
			return
		}

		result, err := generate(r.Context(), topic)
		if err != nil {
			log.Printf("Failed to generate post: %v", err)
			writeAPIError(w, http.StatusInternalServerError, "post generation failed")
			return
		}
		writeJSON(w, http.StatusCreated, result)
	})
	return mux
}

func generatePost(ctx context.Context, topic string) (createPostResponse, error) {
	researchConfig := researchStepConfig
	if err := researchConfig.SetConfig(map[string]any{"topic": topic}); err != nil {
		return createPostResponse{}, fmt.Errorf("prepare research prompt: %w", err)
	}

	researchAgent, err := poster.NewResearchAgent(ctx, researchConfig)
	if err != nil {
		return createPostResponse{}, fmt.Errorf("create research agent: %w", err)
	}
	research, err := researchAgent.Execute(ctx, researchConfig.Prompt)
	if err != nil {
		return createPostResponse{}, fmt.Errorf("run research agent: %w", err)
	}

	workflow, err := poster.NewSequentialChain(ctx, writeStepConfig, optimizeStepConfig)
	if err != nil {
		return createPostResponse{}, fmt.Errorf("build content chain: %w", err)
	}
	result, err := chains.Call(ctx, workflow, map[string]any{
		writeStepConfig.InputKey: research,
	})
	if err != nil {
		return createPostResponse{}, fmt.Errorf("run content chain: %w", err)
	}
	post, ok := result[optimizeStepConfig.OutputKey].(string)
	if !ok {
		return createPostResponse{}, fmt.Errorf("content chain returned no %q output", optimizeStepConfig.OutputKey)
	}

	filename := fmt.Sprintf("%s-%d", topicFilename(topic, 20), time.Now().UnixNano())
	postPath := filepath.Join("POSTS", filename+".md")
	if err := savePost(postPath, post); err != nil {
		return createPostResponse{}, fmt.Errorf("save post: %w", err)
	}

	imageConfig := imageStepConfig
	imageConfig.ImageOutputPath = filepath.Join("POSTS", filename+".png")
	imagePath, err := poster.GenerateImage(ctx, imageConfig, post)
	if err != nil {
		return createPostResponse{}, fmt.Errorf("generate image: %w", err)
	}
	return createPostResponse{Topic: topic, Post: post, ImagePath: imagePath}, nil
}

func sanitizeTopic(topic string) (string, error) {
	for _, character := range topic {
		if unicode.IsControl(character) {
			return "", fmt.Errorf("topic must not contain control characters")
		}
	}
	topic = strings.Join(strings.Fields(topic), " ")
	if topic == "" {
		return "", fmt.Errorf("topic is required")
	}
	if len([]rune(topic)) > maxTopicLength {
		return "", fmt.Errorf("topic must be no more than %d characters", maxTopicLength)
	}
	return topic, nil
}

func writeDecodeError(w http.ResponseWriter, err error) {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		writeAPIError(w, http.StatusRequestEntityTooLarge, "request body is too large")
		return
	}
	writeAPIError(w, http.StatusBadRequest, "invalid JSON request")
}

func writeAPIError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("Failed to encode API response: %v", err)
	}
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
