FROM golang:1.23 AS build-stage
WORKDIR /app

COPY ./api/ ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /api

FROM alpine AS final

WORKDIR /

COPY --from=build-stage /api /api

USER nonroot:nonroot

ENTRYPOINT ["/api"]
