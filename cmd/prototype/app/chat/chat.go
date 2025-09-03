package chat

import (
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"

	"github.com/joyfuldevs/project-lumos/cmd/prototype/app"
)

var (
	// 대화에 사용할 컨텍스트 파일 경로.
	contextFile string
	// API 서버 주소.
	apiAddress string
	// 사용자 질문.
	query string
)

var (
	chatCmd = &cobra.Command{
		Use:   "chat",
		Short: "chat with the AI",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			if ctx == nil {
				fmt.Println("error no context available")
				return
			}

			contextData, err := os.ReadFile(contextFile)
			if err != nil {
				fmt.Println("error reading context file:", err)
				return
			}

			client := openai.NewClient(
				option.WithBaseURL(apiAddress),
			)
			resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
				Messages: []openai.ChatCompletionMessageParamUnion{
					{
						OfSystem: &openai.ChatCompletionSystemMessageParam{
							Content: openai.ChatCompletionSystemMessageParamContentUnion{
								OfString: openai.String("/no_think"),
							},
						},
					},
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.ChatCompletionUserMessageParamContentUnion{
								OfString: openai.String("참고 자료:"),
							},
						},
					},
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.ChatCompletionUserMessageParamContentUnion{
								OfString: openai.String(string(contextData)),
							},
						},
					},
					{
						OfUser: &openai.ChatCompletionUserMessageParam{
							Content: openai.ChatCompletionUserMessageParamContentUnion{
								OfString: openai.String("참고 자료를 바탕으로 다음 질문에 답해주세요. : " + query),
							},
						},
					},
				},
			})
			if err != nil {
				fmt.Println("error creating completion:", err)
				return
			}

			fmt.Println(resp.Choices[0].Message.Content)
		},
	}
)

func init() {
	chatCmd.Flags().StringVarP(&contextFile, "context", "c", "", "Path to the input file containing chat context")
	chatCmd.Flags().StringVarP(&apiAddress, "api-address", "a", "http://localhost:8080/v1", "API server address")
	chatCmd.Flags().StringVarP(&query, "query", "q", "", "User query")

	_ = chatCmd.MarkFlagRequired("context")
	_ = chatCmd.MarkFlagRequired("query")

	app.AddCommand(chatCmd)
}
