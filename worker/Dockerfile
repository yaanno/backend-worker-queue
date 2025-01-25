FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy shared module
COPY shared /shared

# Download dependencies first (better caching)
COPY worker/go.mod worker/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy source code
COPY worker/ .

# Build the application
RUN go build -o /app/worker ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates curl
COPY --from=builder /app/worker .

WORKDIR /app

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

CMD ["/worker"]
