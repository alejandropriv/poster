FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p /out \
	&& CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/poster .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
	&& addgroup -S poster \
	&& adduser -S -G poster -u 10001 poster \
	&& mkdir -p /app/POSTS \
	&& chown -R poster:poster /app

WORKDIR /app
COPY --from=build --chown=poster:poster /out/poster /app/poster

USER poster
EXPOSE 8080
ENTRYPOINT ["/app/poster"]
