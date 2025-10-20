package chain

import (
	"context"
	"log/slog"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

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

func WithChatClientInit(handler chat.Handler) chat.HandlerFunc {
	return chat.HandlerFunc(func(chat *chat.Chat) {
		ctx := chat.Context()

		baseURL, ok := os.LookupEnv("CHAT_API_URL")
		if !ok {
			slog.Warn("CHAT_API_URL is not set")
		}
		key, ok := os.LookupEnv("CHAT_API_KEY")
		if !ok {
			slog.Warn("CHAT_API_KEY is not set")
		}

		chatClient := openai.NewClient(
			option.WithBaseURL(baseURL),
			option.WithAPIKey(key),
		)
		chat = chat.WithContext(WithChatClient(ctx, &chatClient))

		handler.HandleChat(chat)
	})
}
