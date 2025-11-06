package chain

import (
	"context"
	"log/slog"

	"github.com/openai/openai-go"

	"github.com/joyfuldevs/project-lumos/cmd/lumos/app/chat"
)

type responseKeyType int

const responseKey responseKeyType = iota

func WithResponse(parent context.Context, response string) context.Context {
	return context.WithValue(parent, responseKey, response)
}

func ResponseFrom(ctx context.Context) string {
	info, _ := ctx.Value(responseKey).(string)
	return info
}

func ResponseGeneration(handler chat.Handler) chat.HandlerFunc {
	return chat.HandlerFunc(func(chat *chat.Chat) {
		ctx := chat.Context()

		passages := PassagesFrom(ctx)
		if len(passages) == 0 {
			chat = chat.WithContext(WithResponse(ctx, "관련된 정보를 찾을 수 없습니다."))
			handler.HandleChat(chat)
			return
		}

		client := ChatClientFrom(ctx)
		if client == nil {
			chat = chat.WithContext(WithResponse(ctx, "답변 생성에 필요한 서비스가 준비되지 않았습니다."))
			handler.HandleChat(chat)
			return
		}

		query := chat.Thread[len(chat.Thread)-1]

		messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(passages)+4)
		messages = append(messages, openai.ChatCompletionMessageParamUnion{
			OfSystem: &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{
					OfString: openai.String("참고 자료:"),
				},
			},
		})

		for _, passage := range passages {
			messages = append(messages, openai.ChatCompletionMessageParamUnion{
				OfSystem: &openai.ChatCompletionSystemMessageParam{
					Content: openai.ChatCompletionSystemMessageParamContentUnion{
						OfString: openai.String(string(passage.Content)),
					},
				},
			})
		}

		messages = append(messages, openai.ChatCompletionMessageParamUnion{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String("참고 자료를 바탕으로 다음 질문에 답해주세요. : " + query),
				},
			},
		})

		resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Messages: messages,
			Model:    "gpt-5",
			TopP:     openai.Float(0.8),
		})

		if err != nil || len(resp.Choices) == 0 {
			slog.Error("failed to generate response", "error", err)
			chat = chat.WithContext(WithResponse(ctx, "답변 생성에 실패했습니다."))
			handler.HandleChat(chat)
			return
		}

		ctx = WithResponse(ctx, resp.Choices[0].Message.Content)
		handler.HandleChat(chat.WithContext(ctx))
	})
}
