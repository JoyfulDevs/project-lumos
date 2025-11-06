package chain

import (
	"log/slog"

	"github.com/joyfuldevs/project-lumos/cmd/lumos/app/chat"
	"github.com/joyfuldevs/project-lumos/pkg/slack/api"
)

func ChatResponse() chat.HandlerFunc {
	return chat.HandlerFunc(func(chat *chat.Chat) {
		ctx := chat.Context()
		client := SlackClientFrom(ctx)
		if client == nil {
			slog.Error("slack client is not initialized")
			return
		}

		response := ResponseFrom(ctx)
		if response == "" {
			response = "답변을 생성하지 못했습니다.\n관리자에게 문의해주세요."
		}

		_, err := client.PostMessage(ctx, &api.PostMessageRequest{
			Channel:         chat.Channel,
			Text:            response,
			ThreadTimestamp: chat.Timestamp,
		})
		if err != nil {
			slog.Error("failed to post message", slog.Any("error", err))
		}
	})
}

func PanicRecovery(handler chat.Handler) chat.HandlerFunc {
	return chat.HandlerFunc(func(chat *chat.Chat) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered", slog.Any("error", r))
			}
		}()

		handler.HandleChat(chat)
	})
}
