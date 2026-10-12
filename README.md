# Poster

Runs one LangChain `SequentialChain` from an ordered slice of `LLMConfig` values. `inputKeys` lists the initial inputs (`topic`, `date`); each following config consumes the preceding result. The final config sets `ImagePath` and generates a JPEG with OpenAI's latest image model.

The research stage uses an OpenAI functions agent and SerpAPI's Google Search endpoint. Search results include their titles, snippets, and URLs, and the research output adds a `Sources` list from the returned URLs. Set `SERPAPI_API_KEY` in `.env` to enable search.

Set these values in `.env`:

```dotenv
GOOGLE_API_KEY=...
OPENAI_API_KEY=...
SERPAPI_API_KEY=...
IMAGE_OUTPUT=generated/post.jpg
```

Run the workflow with `go run .` and tests with `go test ./...`.
