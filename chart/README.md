# Poster Helm chart

This chart deploys the Poster Go API and Nginx web frontend, and creates a
Kubernetes Secret for `OPENAI_API_KEY` and `SERPAPI_API_KEY`. It expects both
images to be built and published; the chart does not build images.

## Install

Build and publish both images, then install the chart with their image
references and API keys:

```sh
docker buildx build --platform linux/amd64 -t registry.example.com/poster-api:0.1.0 --push poster
docker buildx build --platform linux/amd64 -f nginx/Dockerfile -t registry.example.com/nginx:0.1.0 --push .

helm upgrade --install poster ./chart \
  --set image.repository=registry.example.com \
  --set image.tag=0.1.0 \
  --set nginx.image.repository=registry.example.com \
  --set nginx.image.tag=0.1.0 \
  --set-string secret.openaiApiKey='YOUR_OPENAI_API_KEY' \
  --set-string secret.serpApiKey='YOUR_SERPAPI_API_KEY'
```

For production, avoid passing credentials on the command line. Create a Secret
separately and set `secret.existingSecret` to its name. The Secret must contain
keys named `OPENAI_API_KEY` and `SERPAPI_API_KEY`.

The Nginx Service is the frontend entry point (ClusterIP by default); configure
`nginx.service.type` as `LoadBalancer` or `NodePort` when appropriate. Nginx
proxies `/api/` to the internal Go API Service. Image pull secrets, resource
requests and limits, pod labels and annotations, node selectors, affinity, and
tolerations can be configured in `values.yaml`.

## API

The API listens internally on port 8080. The frontend is served by Nginx on
port 80, which also proxies API requests. Submit a JSON object with a `topic`
string to `POST /api/v1/posts`:

```sh
curl -X POST http://localhost:8080/api/v1/posts \
  -H 'Content-Type: application/json' \
  -d '{"topic":"When should a team use Kubernetes in production?"}'
```

The topic is whitespace-normalized, must contain 1 to 500 Unicode characters,
and may not contain control characters. The API returns the normalized topic,
generated post, and image path as JSON. `GET /healthz` is used by Kubernetes
probes.

Generated files are written to the container's local `POSTS` directory and are
not persisted across pod restarts unless storage is added.