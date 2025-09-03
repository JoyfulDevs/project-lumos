package search

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/joyfuldevs/project-lumos/cmd/prototype/app"
)

var (
	// 검색 쿼리.
	query string
	// API 서버 주소.
	apiAddress string
	// Qdrant 서버 주소.
	dbAddress string
	// Qdrant 컬렉션 이름.
	dbName string
	// 출력 파일 경로.
	outputFile string
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search embeddings in Qdrant collection",
	Long:  "Search for embeddings in a Qdrant collection based on a query vector",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		if ctx == nil {
			fmt.Println("error no context available")
			return
		}

		vectors, err := Embedding(ctx, apiAddress, query)
		if err != nil {
			fmt.Println("error creating embedding:", err)
			return
		}

		results, err := Search(ctx, dbAddress, dbName, vectors)
		if err != nil {
			fmt.Println("error searching embeddings:", err)
			return
		}

		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			fmt.Println("error marshalling results:", err)
			return
		}

		if outputFile != "" {
			if err := os.WriteFile(outputFile, data, 0644); err != nil {
				fmt.Println("error writing output file:", err)
				return
			}
			fmt.Println("results written to", outputFile)
		} else {
			fmt.Println(string(data))
		}

	},
}

func init() {
	searchCmd.Flags().StringVarP(&query, "query", "q", "", "Query to search in the collection")
	searchCmd.Flags().StringVarP(&dbAddress, "db-address", "d", "localhost:6334", "Qdrant server address")
	searchCmd.Flags().StringVarP(&dbName, "db-name", "n", "default_collection", "Qdrant collection name")
	searchCmd.Flags().StringVarP(&apiAddress, "api-address", "a", "http://localhost:8080/v1", "API server address")
	searchCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path for results")

	_ = searchCmd.MarkFlagRequired("query")

	app.AddCommand(searchCmd)
}
