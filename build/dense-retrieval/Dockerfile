FROM golang:1.25 AS builder

WORKDIR /app

COPY . .
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" ./cmd/dense-retrieval

FROM istio/distroless

COPY --from=builder /app/dense-retrieval /app/

ENTRYPOINT ["/app/dense-retrieval"]
