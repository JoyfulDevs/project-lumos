package chain

import (
	"context"

	"github.com/devafterdark/project-lumos/cmd/lumos/app/chat"
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

func WithResponseGeneration(handler chat.Handler) chat.HandlerFunc {
	return chat.HandlerFunc(func(chat *chat.Chat) {
		ctx := chat.Context()

		passages := PassagesFrom(ctx)
		if len(passages) == 0 {
			chat = chat.WithContext(WithResponse(ctx, "관련된 정보를 찾을 수 없습니다."))
			handler.HandleChat(chat)
			return
		}

		// TODO: implement response generation
		ctx = WithResponse(ctx, `"생성된 답변"`)

		handler.HandleChat(chat.WithContext(ctx))
	})
}
