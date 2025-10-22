package main

import (
	"log/slog"
	"os"

	"github.com/joyfuldevs/project-lumos/cmd/jira-sync/app"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	rootCmd := &cobra.Command{
		Use:     "jira-sync",
		Short:   "Jira issue synchronization tool with incremental updates",
		Long:    `A unified CLI tool for synchronizing Jira issues to Qdrant vector database.`,
		Version: version,
	}

	// Global flags
	var (
		dataDir  string
		stateDir string
	)

	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", getEnv("DATA_DIR", "/data"), "Data directory for temporary files")
	rootCmd.PersistentFlags().StringVar(&stateDir, "state-dir", getEnv("STATE_DIR", "/state"), "State directory for persistent data")

	// Add subcommands
	rootCmd.AddCommand(newFullCommand(&dataDir, &stateDir))
	rootCmd.AddCommand(newIncrementalCommand(&dataDir, &stateDir))

	if err := rootCmd.Execute(); err != nil {
		slog.Error("command failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func newFullCommand(dataDir, stateDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "full",
		Short: "Run full synchronization (all issues)",
		Long: `Performs a complete synchronization of all Jira issues.
This command will:
1. Collect all issues from Jira
2. Generate embeddings
3. Upload to Qdrant (both dense and sparse vectors)
4. Save timestamp for future incremental updates`,
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("Jira Sync starting", slog.String("mode", "full"))
			if err := app.RunFull(*dataDir, *stateDir); err != nil {
				slog.Error("full sync failed", slog.Any("error", err))
				return err
			}
			slog.Info("Jira Sync finished successfully")
			return nil
		},
	}
}

func newIncrementalCommand(dataDir, stateDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "incremental",
		Short: "Run incremental synchronization (updated issues only)",
		Long: `Performs an incremental synchronization of changed Jira issues.
This command will:
1. Check last sync timestamp
2. Collect only updated/new issues since last sync
3. Generate embeddings for changed issues
4. Upsert to Qdrant (update existing, insert new)
5. Update timestamp

If no previous sync found, falls back to full synchronization.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			slog.Info("Jira Sync starting", slog.String("mode", "incremental"))
			if err := app.RunIncremental(*dataDir, *stateDir); err != nil {
				slog.Error("incremental sync failed", slog.Any("error", err))
				return err
			}
			slog.Info("Jira Sync finished successfully")
			return nil
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
