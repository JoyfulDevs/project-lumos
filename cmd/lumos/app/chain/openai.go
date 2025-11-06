package chain

import (
	"context"

	"github.com/openai/openai-go"

	"github.com/joyfuldevs/project-lumos/cmd/lumos/app/chat"
)

type chatClientKeyType int

const chatClientKey chatClientKeyType = iota

func WithChatClient(parent context.Context, client *openai.Client) context.Context {
	return context.WithValue(parent, chatClientKey, client)
}

func ChatClientFrom(ctx context.Context) *openai.Client {
	info, _ := ctx.Value(chatClientKey).(*openai.Client)
	return info
}

func WithChatClientInit(handler chat.Handler, client *openai.Client) chat.HandlerFunc {
	return chat.HandlerFunc(func(chat *chat.Chat) {
		ctx := chat.Context()
		chat = chat.WithContext(WithChatClient(ctx, client))
		handler.HandleChat(chat)
	})
}
