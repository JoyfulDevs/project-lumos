package main

import (
	"log/slog"
	"os"

	"github.com/devafterdark/project-lumos/cmd/lumos/app"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	slog.Info("lumos bot service starting")
	if err := app.Run(); err != nil {
		slog.Error("failed to run lumos bot", slog.Any("error", err))
	}
	slog.Info("lumos bot service finished")
}
