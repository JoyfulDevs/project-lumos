package index

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/joyfuldevs/project-lumos/cmd/prototype/app"
	"github.com/joyfuldevs/project-lumos/cmd/prototype/app/hybrid"
)

var (
	bm42QdrantHost string
	bm42QdrantPort int
	bm42Collection string
	bm42InputFile  string
	bm42ConfigFile string
)

// BM42IndexCmd represents the bm42-index command
var BM42IndexCmd = &cobra.Command{
	Use:   "bm42-index",
	Short: "Index documents to Qdrant for BM42 search",
	Long: `Create BM42 sparse vector index in Qdrant for efficient semantic text search.

This command:
- Loads documents from JSON file
- Creates sparse vectors using BM42 (attention-based) algorithm with fastembed
- Uses Qdrant/bm42-all-minilm-l6-v2-attentions model
- Indexes them to Qdrant collection with IDF modifier`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Load config
		cfg, err := hybrid.LoadConfig(bm42ConfigFile)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		// Override with command line flags if provided
		if bm42QdrantHost == "" {
			bm42QdrantHost = cfg.Database.QdrantHost
		}
		if bm42QdrantPort == 0 {
			bm42QdrantPort = cfg.Database.QdrantPort
		}
		if bm42Collection == "" {
			bm42Collection = cfg.Database.BM42Collection
			if bm42Collection == "" {
				bm42Collection = "jira_bm42"
			}
		}

		// Parse host:port format if provided
		if strings.Contains(bm42QdrantHost, ":") {
			parts := strings.Split(bm42QdrantHost, ":")
			bm42QdrantHost = parts[0]
			if len(parts) > 1 {
				_, _ = fmt.Sscanf(parts[1], "%d", &bm42QdrantPort)
			}
		}

		fmt.Printf("[BM42 Index] Connecting to Qdrant at %s:%d\n", bm42QdrantHost, bm42QdrantPort)
		fmt.Printf("[BM42 Index] Collection: %s\n", bm42Collection)
		fmt.Printf("[BM42 Index] Input file: %s\n", bm42InputFile)

		// Index documents using BM42
		fmt.Println("[BM42 Index] Indexing documents using BM42 algorithm...")
		err = hybrid.IndexDocuments(ctx, bm42InputFile, bm42Collection, bm42QdrantHost, bm42QdrantPort)
		if err != nil {
			log.Fatalf("Failed to index documents: %v", err)
		}

		fmt.Println("[BM42 Index] Indexing completed successfully!")
	},
}

func init() {
	BM42IndexCmd.Flags().StringVar(&bm42QdrantHost, "qdrant-host", "", "Qdrant host (default: from config)")
	BM42IndexCmd.Flags().IntVar(&bm42QdrantPort, "qdrant-port", 0, "Qdrant port (default: from config)")
	BM42IndexCmd.Flags().StringVar(&bm42Collection, "collection", "", "Collection name (default: from config)")
	BM42IndexCmd.Flags().StringVarP(&bm42InputFile, "input", "i", "json/content_embedding.json", "Input JSON file with documents")
	BM42IndexCmd.Flags().StringVarP(&bm42ConfigFile, "config", "c", "config/config.yaml", "Config file path")

	// Register command
	app.AddCommand(BM42IndexCmd)
}
