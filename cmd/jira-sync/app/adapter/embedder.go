package adapter

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

// Embedder generates embeddings from issues
type Embedder struct {
	apiURL string
}

// NewEmbedder creates a new embedder
func NewEmbedder(apiURL string) *Embedder {
	return &Embedder{apiURL: apiURL}
}

// GenerateEmbeddings generates embeddings using the prototype binary
func (e *Embedder) GenerateEmbeddings(inputPath, outputPath string) error {
	slog.Info("generating embeddings",
		slog.String("input", inputPath),
		slog.String("output", outputPath))

	cmd := exec.Command("prototype", "embedding",
		"--input", inputPath,
		"--output", outputPath,
		"--api-url", e.apiURL)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("prototype embedding failed: %w", err)
	}

	slog.Info("embeddings generated successfully")
	return nil
}
