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

Run the REST API with `go run .`; it listens on port `8080` by default, or on
the port specified by `PORT`. Create a post with:

```sh
curl -X POST http://localhost:8080/api/v1/posts \
	-H 'Content-Type: application/json' \
	-d '{"topic":"When should a team use Kubernetes in production?"}'
```

The request must contain a JSON `topic` string of 1 to 250 characters. Leading
and repeated whitespace is normalized; control characters and unknown JSON
fields are rejected. The response includes the normalized topic, generated
post, and image path. `GET /healthz` returns a health status.

Run tests with `go test ./...`.
