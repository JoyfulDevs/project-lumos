package app

import (
	"context"
	"log/slog"

	"github.com/openai/openai-go"

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

func NewBotHandler(slackClient *api.Client, openaiClient *openai.Client) *BotHandler {
	handler := BuildChatHandlerChain(slackClient, openaiClient)

	return &BotHandler{
		slackClient: slackClient,
		chatHandler: handler,
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
		// 봇이 보낸 메시지는 무시한다.
		if e.OfMessage.BotID != "" || e.OfMessage.User == "" {
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

func BuildChatHandlerChain(slackClient *api.Client, openaiClient *openai.Client) chat.Handler {
	handler := chain.ChatResponse()

	// 메시지 생성 핸들러 설정.
	handler = chain.ResponseGeneration(handler)
	handler = chain.AssistantStatusUpdate(handler, "가 마법을 부리는 중...")

	// 패시지 검색 핸들러 설정.
	handler = chain.PassageRetrieval(handler)
	handler = chain.AssistantStatusUpdate(handler, "가 주문을 외우는 중...")

	// OpenAI 클라이언트 초기화.
	handler = chain.WithChatClientInit(handler, openaiClient)
	// Slack 클라이언트 초기화.
	handler = chain.WithSlackClientInit(handler, slackClient)

	// 패닉 복구 핸들러 설정.
	handler = chain.PanicRecovery(handler)

	return handler
}
