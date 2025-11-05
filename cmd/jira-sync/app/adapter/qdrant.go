package adapter

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

// QdrantUploader handles uploading vectors to Qdrant
type QdrantUploader struct {
	host           string
	port           string
	collection     string
	bm42Collection string
}

// NewQdrantUploader creates a new Qdrant uploader
func NewQdrantUploader(host, port, collection, bm42Collection string) *QdrantUploader {
	return &QdrantUploader{
		host:           host,
		port:           port,
		collection:     collection,
		bm42Collection: bm42Collection,
	}
}

// UpsertDenseVectors uploads dense vectors to Qdrant
func (q *QdrantUploader) UpsertDenseVectors(embeddingPath string) error {
	slog.Info("upserting dense vectors to Qdrant")

	qdrantAddr := fmt.Sprintf("%s:%s", q.host, q.port)

	cmd := exec.Command("prototype", "insert",
		"--file", embeddingPath,
		"--qdrant-host", qdrantAddr,
		"--collection", q.collection)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dense vector upsert failed: %w", err)
	}

	slog.Info("dense vectors upserted successfully")
	return nil
}

// UpsertSparseVectors uploads BM42 sparse vectors to Qdrant
func (q *QdrantUploader) UpsertSparseVectors(issuesPath string) error {
	slog.Info("upserting BM42 sparse vectors to Qdrant")

	cmd := exec.Command("python3", "/app/python/bm42_indexer.py",
		"--input", issuesPath,
		"--host", q.host,
		"--port", q.port,
		"--collection", q.bm42Collection)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("BM42 indexing failed: %w", err)
	}

	slog.Info("BM42 vectors upserted successfully")
	return nil
}
