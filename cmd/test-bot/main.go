package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joyfuldevs/project-lumos/pkg/slack"
	"github.com/joyfuldevs/project-lumos/pkg/slack/bot"
	"github.com/joyfuldevs/project-lumos/pkg/slack/event"
)

type Handler struct {
	appToken        string
	botToken        string
	feedbackHandler *event.FeedbackHandler
	botUserID       string
	client          *slack.Client
}

func (h *Handler) HandleEventsAPI(ctx context.Context, payload *event.EventsAPIPayload) {
	ec := payload.OfEventCallback
	if ec == nil {
		return
	}

	switch ec.Event.Type {
	case event.EventTypeMessage:
		h.handleMessageEvent(ctx, ec.Event.OfMessage)

	case event.EventTypeAssistantThreadStarted:
		h.handleAssistantThreadStarted(ctx, ec.Event.OfAssistantThreadStarted)

	case event.EventTypeAssistantThreadContextChanged:
		slog.Info("received assistant thread context changed event")
	}
}

// isBotMessage는 메시지가 봇에서 온 것인지 확인합니다
func (h *Handler) isBotMessage(event *event.MessageEvent) bool {
	// 1. 봇 User ID와 일치하는지 확인
	if h.botUserID != "" && event.User == h.botUserID {
		slog.Info("bot message detected by user ID", slog.String("user", event.User))
		return true
	}

	// 2. bot_id 필드가 있는지 확인
	if event.BotID != "" {
		slog.Info("bot message detected by bot ID", slog.String("bot_id", event.BotID))
		return true
	}

	// 3. subtype이 bot_message인지 확인
	if event.Subtype == "bot_message" {
		slog.Info("bot message detected by subtype", slog.String("subtype", event.Subtype))
		return true
	}

	// 4. 사용자가 비어있고 봇 username이 있는 경우
	if event.User == "" && event.Username != "" {
		slog.Info("bot message detected by username", slog.String("username", event.Username))
		return true
	}

	slog.Debug("not a bot message",
		slog.String("user", event.User),
		slog.String("bot_id", event.BotID),
		slog.String("subtype", event.Subtype),
		slog.String("username", event.Username),
		slog.String("text", event.Text))
	return false
}

func (h *Handler) handleMessageEvent(ctx context.Context, messageEvent *event.MessageEvent) {
	if messageEvent.Text == "" {
		return
	}
	if messageEvent.User == messageEvent.ParentUserID {
		return
	}

	// 모든 봇 메시지는 무시 (이벤트 루프 방지 및 깔끔한 처리)
	if h.isBotMessage(messageEvent) {
		slog.Info("ignoring all bot messages",
			slog.String("text", messageEvent.Text),
			slog.String("user", messageEvent.User),
			slog.String("bot_id", messageEvent.BotID))
		return
	}

	// 사용자 메시지에만 응답하고 피드백 버튼 추가
	slog.Info("received user message event", slog.String("text", messageEvent.Text))

	_, err := h.client.PostMessage(ctx, &slack.PostMessageRequest{
		Channel:         messageEvent.Channel,
		Text:            "You said: " + messageEvent.Text,
		ThreadTimestamp: messageEvent.ThreadTimestamp,
	})
	if err != nil {
		slog.Warn("failed to post message", slog.String("channel", messageEvent.Channel), slog.Any("error", err))
		return
	}

	// 사용자 메시지에 응답한 직후 피드백 버튼 전송
	slog.Info("sending feedback buttons after user message response", slog.String("channel", messageEvent.Channel))
	if err := h.feedbackHandler.SendFeedbackButtons(messageEvent.Channel); err != nil {
		slog.Warn("failed to send feedback buttons", slog.String("channel", messageEvent.Channel), slog.Any("error", err))
	}
}

func (h *Handler) handleAssistantThreadStarted(ctx context.Context, event *event.AssistantThreadStartedEvent) {
	slog.Info("received assistant thread started event")
	// AI 어시스턴트 기능을 사용하지 않고 일반 메시지로 처리
	channelID := event.AssistantThread.ChannelID

	// 일반 메시지로 인사 전송 (스레드 없음)
	_, err := h.client.PostMessage(ctx, &slack.PostMessageRequest{
		Channel: channelID,
		Text:    "What spell should I cast?",
		// ThreadTimestamp 없음 - 일반 메시지
	})
	if err != nil {
		slog.Warn("failed to post message", slog.String("channel", channelID), slog.Any("error", err))
	}

	// AI 어시스턴트 첫 인사 메시지에는 피드백 버튼을 추가하지 않음
	slog.Info("AI assistant thread started - no feedback buttons for initial greeting")
}

// getBotUserID는 봇의 User ID를 가져옵니다
func getBotUserID(ctx context.Context, appToken, botToken string) (string, error) {
	// 환경변수에서 먼저 확인
	if botUserID := os.Getenv("SLACK_BOT_USER_ID"); botUserID != "" {
		slog.Info("using bot user ID from environment variable", slog.String("bot_user_id", botUserID))
		return botUserID, nil
	}

	// API를 통해 봇 User ID 가져오기
	client := slack.NewClient(http.DefaultClient, appToken, botToken)
	resp, err := client.AuthTest(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get bot user ID from API: %w", err)
	}

	slog.Info("retrieved bot user ID from API", slog.String("bot_user_id", resp.UserID))
	return resp.UserID, nil
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	appToken := os.Getenv("SLACK_APP_TOKEN")
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if appToken == "" || botToken == "" {
		slog.Error("SLACK_APP_TOKEN and SLACK_BOT_TOKEN must be set")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	// 봇 User ID 자동 가져오기
	botUserID, err := getBotUserID(ctx, appToken, botToken)
	if err != nil {
		slog.Error("failed to get bot user ID", slog.Any("error", err))
		return
	}

	// 클라이언트 및 피드백 핸들러 초기화
	client := slack.NewClient(http.DefaultClient, appToken, botToken)
	feedbackHandler := event.NewFeedbackHandler(botToken)

	resp, err := client.OpenConnection(ctx)
	if err != nil {
		slog.Error("failed to open connection", slog.Any("error", err))
		return
	}

	handler := &Handler{
		appToken:        appToken,
		botToken:        botToken,
		feedbackHandler: feedbackHandler,
		botUserID:       botUserID,
		client:          client,
	}

	// 기존 봇 사용
	b := bot.NewBot(handler)
	if err := b.Run(ctx, resp.URL); err != nil {
		slog.Error("failed to run bot", slog.Any("error", err))
	}
}
