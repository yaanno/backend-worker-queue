FROM golang:1.22-alpine AS builder

WORKDIR /app

# Download dependencies first (better caching)
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o /app/backend ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates curl
COPY --from=builder /app/backend .

WORKDIR /app

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

CMD ["/backend"]
