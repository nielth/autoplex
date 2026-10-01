FROM golang:1.23 AS build-stage
WORKDIR /app

COPY ./api/ ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o /api

FROM alpine AS final

# ffprobe is used by the disk balancer to read video bitrates
RUN apk add --no-cache ffmpeg

WORKDIR /

COPY --from=build-stage /api /api

# ENTRYPOINT ["tail", "-f", "/dev/null"]
ENTRYPOINT ["/api"]
