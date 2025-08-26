package chain

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/devafterdark/project-lumos/cmd/lumos/app/chat"
	"github.com/devafterdark/project-lumos/pkg/slack"
)

type slackClientKeyType int

const slackClientKey slackClientKeyType = iota

func WithSlackClient(parent context.Context, client *slack.Client) context.Context {
	return context.WithValue(parent, slackClientKey, client)
}

func SlackClientFrom(ctx context.Context) *slack.Client {
	info, _ := ctx.Value(slackClientKey).(*slack.Client)
	return info
}

func WithSlackClientInit(handler chat.Handler, appToken, botToken string) chat.HandlerFunc {
	return chat.HandlerFunc(func(chat *chat.Chat) {
		ctx := chat.Context()
		slackClient := slack.NewClient(http.DefaultClient, appToken, botToken)

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

		_, err := slackClient.AssistantSetStatus(ctx, &slack.AssistantSetStatusRequest{
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
