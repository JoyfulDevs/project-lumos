# Stage 1: Python dependencies builder
FROM python:3.13-slim AS deps-builder

WORKDIR /build

# Install build dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Copy and install Python dependencies
COPY services/sparse-retrieval/requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

# Stage 2: Final runtime image
FROM python:3.13-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    libgomp1 \
    && rm -rf /var/lib/apt/lists/*

# Copy Python packages from deps-builder
COPY --from=deps-builder /root/.local /root/.local

# Copy already generated protobuf code from project
COPY gen/python ./gen/python

# Copy application code
COPY services/sparse-retrieval/ ./services/sparse-retrieval/

# Make sure scripts are in PATH
ENV PATH=/root/.local/bin:$PATH
ENV PYTHONPATH=/app:/app/gen/python:/root/.local/lib/python3.13/site-packages

# Set environment variables
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    GRPC_PORT=50051 \
    QDRANT_HOST=qdrant \
    QDRANT_PORT=6333 \
    COLLECTION_NAME=jira_bm42_full \
    MAX_WORKERS=10

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD python -c "import grpc; channel = grpc.insecure_channel('localhost:50051'); channel.close()" || exit 1

# Expose gRPC port
EXPOSE 50051

# Run the service
CMD ["python", "services/sparse-retrieval/main.py"]