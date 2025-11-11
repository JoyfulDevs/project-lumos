package chain

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/joyfuldevs/project-lumos/cmd/lumos/app/chat"
	"github.com/joyfuldevs/project-lumos/pkg/slack/api"
)

type slackClientKeyType int

const slackClientKey slackClientKeyType = iota

func WithSlackClient(parent context.Context, client *api.Client) context.Context {
	return context.WithValue(parent, slackClientKey, client)
}

func SlackClientFrom(ctx context.Context) *api.Client {
	info, _ := ctx.Value(slackClientKey).(*api.Client)
	return info
}

func WithSlackClientInit(handler chat.Handler) chat.HandlerFunc {
	return chat.HandlerFunc(func(chat *chat.Chat) {
		ctx := chat.Context()

		appToken, ok := os.LookupEnv("SLACK_APP_TOKEN")
		if !ok {
			slog.Warn("SLACK_APP_TOKEN is not set")
		}
		botToken, ok := os.LookupEnv("SLACK_BOT_TOKEN")
		if !ok {
			slog.Warn("SLACK_BOT_TOKEN is not set")
		}

		slackClient := api.NewClient(http.DefaultClient, appToken, botToken)
		chat = chat.WithContext(WithSlackClient(ctx, slackClient))

		handler.HandleChat(chat)
	})
}

func WithAssistantStatus(handler chat.Handler, status string) chat.HandlerFunc {
	return chat.HandlerFunc(func(chat *chat.Chat) {
		ctx := chat.Context()
		slackClient := SlackClientFrom(ctx)

		if slackClient == nil {
			slog.Error("slack client is not initialized")
			return
		}

		_, err := slackClient.AssistantSetStatus(ctx, &api.AssistantSetStatusRequest{
			Channel:         chat.Channel,
			ThreadTimestamp: chat.Timestamp,
			Status:          status,
		})
		if err != nil {
			slog.Warn("failed to set slack status", "error", err)
		}

		handler.HandleChat(chat)
	})
}
