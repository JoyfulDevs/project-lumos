package app

import (
	"context"
	"log/slog"

	"github.com/joyfuldevs/project-lumos/cmd/lumos/app/chain"
	"github.com/joyfuldevs/project-lumos/cmd/lumos/app/chat"
	"github.com/joyfuldevs/project-lumos/pkg/slack"
	"github.com/joyfuldevs/project-lumos/pkg/slack/event"
)

type BotHandler struct {
	slackClient *slack.Client
	chatHandler chat.Handler
}

func NewBotHandler(slackClient *slack.Client) *BotHandler {
	return &BotHandler{
		slackClient: slackClient,
		chatHandler: BuildChatHandlerChain(slackClient),
	}
}

func (b *BotHandler) HandleEventsAPI(ctx context.Context, payload *event.EventsAPIPayload) {
	if payload.Type != event.EventsAPITypeEventCallback {
		return
	}

	e := payload.OfEventCallback.Event
	switch e.Type {
	case event.EventTypeAssistantThreadStarted:
		_, err := b.slackClient.PostMessage(ctx, &slack.PostMessageRequest{
			Channel:         e.OfAssistantThreadStarted.AssistantThread.ChannelID,
			Text:            "안녕하세요! 무엇을 도와드릴까요?",
			ThreadTimestamp: e.OfAssistantThreadStarted.AssistantThread.ThreadTimestamp,
		})
		if err != nil {
			slog.Error("failed to post message", slog.Any("error", err))
		}
	case event.EventTypeAssistantThreadContextChanged:
		// TODO: Implement thread context changed handling
	case event.EventTypeMessage:
		if e.OfMessage.User == e.OfMessage.ParentUserID {
			// 봇의 메시지 기능을 사용하므로 ParentUserID는 봇의 ID 값을 가진다.
			// 따라서 이벤트를 생성한 User와 비교해 봇이 스스로에게 응답하지 않도록 한다.
			return
		}
		c := &chat.Chat{
			Channel:   e.OfMessage.Channel,
			Timestamp: e.OfMessage.ThreadTimestamp,
			Thread:    []string{e.OfMessage.Text},
		}
		c = c.WithContext(ctx)
		b.chatHandler.HandleChat(c)
	default:
		slog.Warn("unknown event type", slog.String("type", string(e.Type)))
	}
}

func BuildChatHandlerChain(slackClient *slack.Client) chat.Handler {
	handler := chain.ResponseHandler()

	// 메시지 생성 핸들러 설정.
	handler = chain.WithResponseGeneration(handler)
	handler = chain.WithAssistantStatus(handler, "generating response...")

	// 패시지 검색 핸들러 설정.
	handler = chain.WithPassageRetrieval(handler)
	handler = chain.WithAssistantStatus(handler, "retrieving passages...")

	// 슬랙 클라이언트 초기화 핸들러 설정.
	handler = chain.WithSlackClientInit(handler, slackClient)

	// 패닉 복구 핸들러 설정.
	handler = chain.WithPanicRecovery(handler)

	return handler
}
