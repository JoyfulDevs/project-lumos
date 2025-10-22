# Multi-stage build for Jira sync cronjob
# Stage 1: Build Go binaries
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binaries with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/jira-sync ./cmd/jira-sync
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/prototype ./cmd/prototype

# Stage 2: Runtime image
FROM python:3.11-slim

LABEL maintainer="lumos-team"
LABEL description="Jira issue sync and embedding cronjob"

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    jq \
    bash \
    && rm -rf /var/lib/apt/lists/*

# Copy Go binaries from builder
COPY --from=builder /bin/jira-sync /usr/local/bin/jira-sync
COPY --from=builder /bin/prototype /usr/local/bin/prototype

# Install Python dependencies for BM42
COPY cmd/jira-sync/python/requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir -r /tmp/requirements.txt && rm /tmp/requirements.txt

# Copy Python scripts for BM42
COPY cmd/jira-sync/python/ /app/python/

# Create data directory and app user
RUN mkdir -p /data && \
    useradd -m -u 1000 -s /bin/bash appuser && \
    chown -R appuser:appuser /app /data

# Make Python scripts executable
RUN chmod +x /app/python/*.py

# Switch to non-root user
USER appuser

WORKDIR /data

# Default command (will be overridden by Kubernetes)
CMD ["/bin/bash"]
