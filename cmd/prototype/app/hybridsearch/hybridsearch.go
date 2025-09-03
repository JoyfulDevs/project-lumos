package hybridsearch

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/joyfuldevs/project-lumos/cmd/prototype/app"
	"github.com/joyfuldevs/project-lumos/cmd/prototype/app/hybrid"
)

var (
	hybridQuery      string
	hybridConfigFile string
	hybridOutputFile string
)

var hybridSearchCmd = &cobra.Command{
	Use:   "hybrid-search",
	Short: "Perform hybrid search combining BM42 and embedding methods",
	Long: `Hybrid search combines BM42 text search and vector similarity search
for improved accuracy. Results are weighted and merged from both methods.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		if ctx == nil {
			fmt.Println("error: no context available")
			return
		}

		// Load configuration
		cfg, err := loadConfiguration(hybridConfigFile)
		if err != nil {
			fmt.Printf("error loading configuration: %v\n", err)
			return
		}

		// Setup searchers
		hybridSearcher := hybrid.NewHybridSearcher(&hybrid.SearchConfig{
			EmbeddingWeight: cfg.Search.EmbeddingWeight,
			BM42Weight:      cfg.Search.BM42Weight,
			MaxResults:      cfg.Search.MaxResults,
			MinScore:        cfg.Search.MinScore,
		})

		// Configure overlap bonus if specified
		if cfg.Search.OverlapBonus > 0 {
			hybridSearcher.SetOverlapBonus(cfg.Search.OverlapBonus)
		}
		if cfg.Search.OverlapPriority {
			hybridSearcher.SetOverlapPriority(true)
		}

		// Add BM42 searcher
		if cfg.Database.QdrantHost != "" && cfg.Search.BM42Weight > 0 {
			collection := cfg.Database.BM42Collection
			if collection == "" {
				collection = "jira_bm42"
			}

			bm42Searcher, err := hybrid.NewBM42Searcher(
				cfg.Database.QdrantHost,
				cfg.Database.QdrantPort,
				collection,
			)
			if err != nil {
				fmt.Printf("warning: could not initialize BM42 searcher: %v\n", err)
			} else {
				hybridSearcher.AddSearcher("bm42", bm42Searcher, cfg.Search.BM42Weight)
			}
		}

		// Add embedding searcher (if available and weight > 0)
		if cfg.Database.QdrantHost != "" && cfg.Search.EmbeddingWeight > 0 {
			// Embedding uses gRPC port (6334), while BM42 uses HTTP port (6333)
			embSearcher, err := hybrid.NewEmbeddingSearcher(
				fmt.Sprintf("%s:%d", cfg.Database.QdrantHost, 6334),
				cfg.Database.CollectionName,
				cfg.API.EmbeddingServer,
			)
			if err != nil {
				fmt.Printf("warning: could not initialize embedding searcher: %v\n", err)
			} else {
				hybridSearcher.AddSearcher("embedding", embSearcher, cfg.Search.EmbeddingWeight)
			}
		}

		// Perform search
		results, err := hybridSearcher.Search(ctx, hybridQuery, cfg.Search.MaxResults)
		if err != nil {
			fmt.Printf("error performing search: %v\n", err)
			return
		}

		// Save to file if output flag is provided
		if hybridOutputFile != "" {
			if err := saveResultsToFile(results, hybridOutputFile); err != nil {
				fmt.Printf("error saving results to file: %v\n", err)
				return
			}
			fmt.Printf("Results saved to %s\n", hybridOutputFile)
		}

		// Display results
		displayResults(results)
	},
}

func init() {
	hybridSearchCmd.Flags().StringVarP(&hybridQuery, "query", "q", "", "Search query")
	hybridSearchCmd.Flags().StringVarP(&hybridConfigFile, "config", "c", "config/config.yaml", "Configuration file path")
	hybridSearchCmd.Flags().StringVarP(&hybridOutputFile, "output", "o", "", "Output file path for search results (JSON format)")

	_ = hybridSearchCmd.MarkFlagRequired("query")

	app.AddCommand(hybridSearchCmd)
}

func loadConfiguration(configFile string) (*hybrid.Config, error) {
	// Try to load from file
	cfg, err := hybrid.LoadConfig(configFile)
	if err != nil {
		// Fall back to default configuration
		fmt.Printf("warning: could not load config file %s, using defaults: %v\n", configFile, err)
		cfg = hybrid.GetDefaultConfig()
	}
	return cfg, nil
}

func saveResultsToFile(results []hybrid.SearchResult, filename string) error {
	// Create output structure
	output := struct {
		Query   string                `json:"query"`
		Count   int                   `json:"count"`
		Results []hybrid.SearchResult `json:"results"`
	}{
		Query:   hybridQuery,
		Count:   len(results),
		Results: results,
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func displayResults(results []hybrid.SearchResult) {
	if len(results) == 0 {
		fmt.Println("No results found")
		return
	}

	fmt.Printf("\nFound %d results:\n", len(results))
	fmt.Println(strings.Repeat("-", 80))

	for i, result := range results {
		fmt.Printf("%d. [%s] %s (Score: %.3f)\n",
			i+1, result.Key, result.Title, result.Score)

		// Show score contributions if available
		if contributions, ok := result.Payload["score_contributions"].(map[string]float32); ok {
			fmt.Printf("   Contributions: ")
			for source, score := range contributions {
				fmt.Printf("%s=%.3f ", source, score)
			}
			fmt.Println()
		}

		// Show snippet of content if available
		if result.Content != "" && len(result.Content) > 100 {
			fmt.Printf("   %s...\n", result.Content[:100])
		} else if result.Content != "" {
			fmt.Printf("   %s\n", result.Content)
		}

		fmt.Println()
	}
}
