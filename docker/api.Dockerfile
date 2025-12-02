# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates curl

WORKDIR /app

# Install golang-migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate

# Copy go mod files first for better caching
COPY api/go.mod api/go.sum ./
RUN go mod download

# Copy source code
COPY api/ ./

# Build the binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /seed ./cmd/seed

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /server .
COPY --from=builder /seed .

# Copy migrate tool from builder
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# Copy migrations
COPY api/migrations ./migrations

EXPOSE 8080

CMD ["./server"]
