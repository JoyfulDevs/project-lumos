package app

import (
	"context"
	"log/slog"

	"github.com/joyfuldevs/project-lumos/cmd/lumos/app/chain"
	"github.com/joyfuldevs/project-lumos/cmd/lumos/app/chat"
	"github.com/joyfuldevs/project-lumos/pkg/slack/api"
	"github.com/joyfuldevs/project-lumos/pkg/slack/eventsapi"
	"github.com/joyfuldevs/project-lumos/pkg/slack/interactive"
)

type BotHandler struct {
	slackClient *api.Client
	chatHandler chat.Handler
}

func NewBotHandler(slackClient *api.Client) *BotHandler {
	return &BotHandler{
		slackClient: slackClient,
		chatHandler: BuildChatHandlerChain(slackClient),
	}
}

func (b *BotHandler) HandleEventsAPI(ctx context.Context, payload *eventsapi.Payload) {
	if payload.Type != eventsapi.PayloadTypeEventCallback {
		return
	}

	e := payload.OfEventCallback.Event
	switch e.Type {
	case eventsapi.EventTypeAssistantThreadStarted:
		_, err := b.slackClient.PostMessage(ctx, &api.PostMessageRequest{
			Channel:         e.OfAssistantThreadStarted.AssistantThread.ChannelID,
			Text:            "안녕하세요! 무엇을 도와드릴까요?",
			ThreadTimestamp: e.OfAssistantThreadStarted.AssistantThread.ThreadTimestamp,
		})
		if err != nil {
			slog.Error("failed to post message", slog.Any("error", err))
		}
	case eventsapi.EventTypeAssistantThreadContextChanged:
		// TODO: Implement thread context changed handling
	case eventsapi.EventTypeMessage:
		if e.OfMessage.BotID != "" {
			// 봇이 보낸 메시지는 무시한다.
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

func (b *BotHandler) HandleInteractive(ctx context.Context, payload *interactive.Payload) {
	// TODO: implement interactive handler
	switch payload.Type {
	case interactive.PayloadTypeBlockActions:
	case interactive.PayloadTypeMessageActions:
	case interactive.PayloadTypeViewClosed:
	case interactive.PayloadTypeViewSubmission:
	default:
		slog.Warn("unknown interactive payload", slog.String("type", string(payload.Type)))
	}
}

func BuildChatHandlerChain(slackClient *api.Client) chat.Handler {
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
