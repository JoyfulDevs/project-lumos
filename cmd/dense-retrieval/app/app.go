package app

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/joyfuldevs/project-lumos/cmd/dense-retrieval/app/adapter"
	"github.com/joyfuldevs/project-lumos/cmd/dense-retrieval/app/service"
	"github.com/joyfuldevs/project-lumos/pkg/service/retrieval/passage/server"
)

func Run() error {
	qdrantHost, ok := os.LookupEnv("QDRANT_HOST")
	if !ok {
		return errors.New("QDRANT_HOST is not set")
	}
	retriever, err := adapter.NewQdrantClient(qdrantHost)
	if err != nil {
		return err
	}
	embeddingURL, ok := os.LookupEnv("EMBEDDING_API_URL")
	if !ok {
		return errors.New("EMBEDDING_API_URL is not set")
	}
	embedder := adapter.NewOpenAIClient(embeddingURL)

	svc := service.NewService(retriever, embedder)

	s := server.NewServer(
		server.WithServiceV1(svc),
	)

	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		cancel()
	}()

	return s.Serve(ctx)
}
