package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/joyfuldevs/project-lumos/cmd/prototype/app"
)

var (
	// 배치 작업에 사용할 파일 경로.
	file string
	// Qdrant 서버 주소.
	address string
	// Qdrant 컬렉션 이름.
	name string
	// 벡터 차원 크기.
	dimension uint64 = 1024
)

var insertCmd = &cobra.Command{
	Use:   "insert",
	Short: "Insert embeddings into Qdrant collection",
	Long:  "Insert embeddings from a JSON file into a Qdrant collection",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		if ctx == nil {
			fmt.Println("error: no context available")
			return
		}

		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Println("error opening file:", err)
			return
		}

		embeddings := []app.Embedding{}
		if err := json.Unmarshal(data, &embeddings); err != nil {
			fmt.Println("error unmarshalling JSON:", err)
			return
		}

		if err := Insert(ctx, address, name, dimension, embeddings); err != nil {
			fmt.Println("error inserting data into Qdrant:", err)
			return
		}

		fmt.Println("Data successfully inserted into Qdrant collection:", name)
	},
}

func init() {
	insertCmd.Flags().StringVarP(&file, "file", "f", "", "Path to the file to process")
	insertCmd.Flags().StringVarP(&address, "address", "a", "localhost:6334", "Qdrant server address")
	insertCmd.Flags().StringVarP(&name, "name", "n", "default_collection", "Qdrant collection name")
	insertCmd.Flags().Uint64VarP(&dimension, "dimension", "d", 1024, "Vector dimension size")

	_ = insertCmd.MarkFlagRequired("file")

	app.AddCommand(insertCmd)
}
