package embedding

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"

	"github.com/joyfuldevs/project-lumos/cmd/prototype/app"
	"github.com/joyfuldevs/project-lumos/pkg/jira"
)

var (
	// API 서버 주소.
	address string
	// 입력 파일 경로.
	input string
	// 출력 파일 경로.
	output string
)

var embeddingCmd = &cobra.Command{
	Use:   "embedding",
	Short: "Create embeddings using embedding models",
	Long:  "Commands to manage embeddings in Qdrant, including inserting and searching",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		if ctx == nil {
			fmt.Println("error: no context available")
			return
		}

		data, err := os.ReadFile(input)
		if err != nil {
			fmt.Println("error reading input file:", err)
			return
		}

		issues := []jira.Issue{}
		if err := json.Unmarshal(data, &issues); err != nil {
			fmt.Println("error unmarshalling JSON:", err)
			return
		}

		titleCh, contentCh := embedding(ctx, issues)

		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			embeddings := make([]*app.Embedding, 0, len(issues))
			for e := range titleCh {
				if e != nil {
					embeddings = append(embeddings, e)
				}
			}
			if ctx.Err() != nil {
				return
			}
			savePath := path.Join(filepath.Dir(output), "title_"+path.Base(output))
			if err := save(savePath, embeddings); err != nil {
				fmt.Println("error saving title embeddings:", err)
				return
			}
		}()
		go func() {
			defer wg.Done()
			// 이슈 본문의 내용을 나누어서 벡터를 생성하기 때문에 이슈의 10배수로 미리 슬라이스를 할당한다.
			embeddings := make([]*app.Embedding, 0, len(issues)*10)
			for e := range contentCh {
				if e != nil {
					embeddings = append(embeddings, e)
				}
			}
			if ctx.Err() != nil {
				return
			}
			savePath := path.Join(filepath.Dir(output), "content_"+path.Base(output))
			if err := save(savePath, embeddings); err != nil {
				fmt.Println("error saving content embeddings:", err)
				return
			}
		}()

		wg.Wait()
	},
}

func init() {
	embeddingCmd.Flags().StringVarP(&address, "address", "a", "http://localhost:8080/v1", "API server address")
	embeddingCmd.Flags().StringVarP(&input, "input", "i", "", "Input file path")
	embeddingCmd.Flags().StringVarP(&output, "output", "o", "embedding.json", "Output file path")

	_ = embeddingCmd.MarkFlagRequired("input")

	app.AddCommand(embeddingCmd)
}

func convert(in []float64) []float32 {
	out := make([]float32, len(in))
	for i, v := range in {
		out[i] = float32(v)
	}
	return out
}

func embedding(
	ctx context.Context,
	issues []jira.Issue,
) (titleCh chan *app.Embedding, contentCh chan *app.Embedding) {
	client := openai.NewClient(
		option.WithBaseURL(address),
	)

	titleCh = make(chan *app.Embedding, 1)
	contentCh = make(chan *app.Embedding, 1)

	perform := func(issueKey, text string) *app.Embedding {
		resp, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
			Input: openai.EmbeddingNewParamsInputUnion{
				OfString: openai.String(text),
			},
			EncodingFormat: openai.EmbeddingNewParamsEncodingFormatFloat,
		})
		if err != nil {
			fmt.Println("skipping issue", issueKey)
			fmt.Println("error creating embedding:", err)
			return nil
		}
		if len(resp.Data) < 1 {
			fmt.Println("no embeddings returned for issue", issueKey)
			return nil
		}
		return &app.Embedding{
			Payload: map[string]any{
				"key":   issueKey,
				"value": text,
			},
			Vectors: convert(resp.Data[0].Embedding),
		}
	}

	go func() {
		defer func() {
			close(titleCh)
			close(contentCh)
		}()

		for _, issue := range issues {
			if ctx.Err() != nil {
				fmt.Println("context cancelled, stopping processing")
				return
			}

			// 이슈 제목에 대한 벡터 생성.
			titleCh <- perform(issue.Key, issue.Fields.Title)

			// 이슈 본문에 대한 벡터 생성.
			contentCh <- perform(issue.Key, issue.Fields.Content)
		}
	}()

	return titleCh, contentCh
}

func save(file string, embeddings []*app.Embedding) error {
	if len(embeddings) == 0 {
		fmt.Println("no embeddings created")
		return nil
	}

	data, err := json.Marshal(embeddings)
	if err != nil {
		return err
	}

	if err := os.WriteFile(file, data, 0644); err != nil {
		return err
	}

	return nil
}
