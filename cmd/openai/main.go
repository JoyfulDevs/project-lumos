package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joyfuldevs/project-lumos/cmd/openai/app"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	app.Execute(ctx)
}
