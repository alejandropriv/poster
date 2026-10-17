# Poster Helm chart

This chart deploys the Poster container and creates a Kubernetes Secret for
`OPENAI_API_KEY` and `SERPAPI_API_KEY`. It expects an image containing this Go
application; the chart does not build or publish the image.

## Install

Build and publish the application image, then install the chart with its image
reference and API keys:

```sh
helm upgrade --install poster ./chart \
  --set image.repository=registry.example.com/poster \
  --set image.tag=latest \
  --set-string secret.openaiApiKey='YOUR_OPENAI_API_KEY' \
  --set-string secret.serpApiKey='YOUR_SERPAPI_API_KEY'
```

For production, avoid passing credentials on the command line. Create a Secret
separately and set `secret.existingSecret` to its name. The Secret must contain
keys named `OPENAI_API_KEY` and `SERPAPI_API_KEY`.

Image pull secrets, resource requests and limits, pod labels and annotations,
node selectors, affinity, and tolerations can be configured in `values.yaml`.

## API

The API listens on port 8080. Submit a JSON object with a `topic` string to
`POST /api/v1/posts`:

```sh
curl -X POST http://localhost:8080/api/v1/posts \
  -H 'Content-Type: application/json' \
  -d '{"topic":"When should a team use Kubernetes in production?"}'
```

The topic is whitespace-normalized, must contain 1 to 250 Unicode characters,
and may not contain control characters. The API returns the normalized topic,
generated post, and image path as JSON. `GET /healthz` is used by Kubernetes
probes.

Generated files are written to the container's local `POSTS` directory and are
not persisted across pod restarts unless storage is added.