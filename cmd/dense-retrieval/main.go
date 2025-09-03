package main

import (
	"log/slog"
	"os"

	"github.com/joyfuldevs/project-lumos/cmd/dense-retrieval/app"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	slog.Info("Dense Retrieval Service starting")
	if err := app.Run(); err != nil {
		slog.Error("failed to run dense retrieval service", slog.Any("error", err))
	}
	slog.Info("Dense Retrieval Service finished")
}
