# Poster

Runs one LangChain `SequentialChain` from an ordered slice of `LLMConfig` values. `inputKeys` lists the initial inputs (`topic`, `date`); each following config consumes the preceding result. The final config sets `ImagePath` and generates a JPEG with OpenAI's latest image model.

The research stage loads `conf_research.txt` into an OpenAI functions agent. The agent can search Google through Serper; set `SERPER_API_KEY` in `.env` to enable it.

Set these values in `.env`:

```dotenv
GOOGLE_API_KEY=...
OPENAI_API_KEY=...
SERPER_API_KEY=...
IMAGE_OUTPUT=generated/post.jpg
```

Run the workflow with `go run .` and tests with `go test ./...`.
