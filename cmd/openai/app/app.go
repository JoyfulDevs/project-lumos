package app

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/joyfuldevs/project-lumos/cmd/openai/app/appctx"
	"github.com/joyfuldevs/project-lumos/cmd/openai/app/batch"
	"github.com/joyfuldevs/project-lumos/cmd/openai/app/file"
)

var rootCmd = &cobra.Command{
	Use:          "openai",
	Short:        "OpenAI utilities",
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(batch.Cmd)
	rootCmd.AddCommand(file.Cmd)
}

func Execute(ctx context.Context) {
	key, ok := os.LookupEnv("OPENAI_API_KEY")
	if !ok {
		fmt.Println("OPENAI_API_KEY is not set")
		os.Exit(1)
	}
	url, ok := os.LookupEnv("OPENAI_API_URL")
	if !ok {
		fmt.Println("OPENAI_API_URL is not set")
		os.Exit(1)
	}

	ctx = appctx.WithOpenAIAPIKey(ctx, key)
	ctx = appctx.WithOpenAIAPIURL(ctx, url)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
