FROM golang:1.23-bookworm AS build
WORKDIR /app

COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /api

FROM debian:bookworm-slim

# ca-certificates for Plex, TorrentLeech and TVmaze over HTTPS,
# ffmpeg for ffprobe, which the disk balancer uses to read video bitrates
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates ffmpeg \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /api /api

EXPOSE 8080
ENTRYPOINT ["/api"]
