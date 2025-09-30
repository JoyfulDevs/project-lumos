package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joyfuldevs/project-lumos/pkg/retry"
	"github.com/joyfuldevs/project-lumos/pkg/slack"
	"github.com/joyfuldevs/project-lumos/pkg/slack/bot"
	"github.com/joyfuldevs/project-lumos/pkg/slack/event"
)

type Handler struct {
	appToken string
	botToken string
	client   *slack.Client
}

func (h *Handler) HandleEventsAPI(ctx context.Context, payload *event.EventsAPIPayload) {
	ec := payload.OfEventCallback
	if ec == nil {
		return
	}

	switch ec.Event.Type {
	case event.EventTypeMessage:
		msg := ec.Event.OfMessage

		// 디버그: BotID와 User 출력
		slog.Info("MESSAGE EVENT",
			slog.String("user", msg.User),
			slog.String("bot_id", msg.BotID))

		// 봇 메시지 필터링
		if msg.BotID != "" {
			return
		}

		if msg.Text == "" {
			return
		}
		if msg.User == msg.ParentUserID {
			return
		}

		slog.Info("USER MESSAGE - RESPONDING", slog.String("text", msg.Text))

		// 스레드 타임스탬프 결정 (응답 전에)
		threadTS := msg.ThreadTimestamp

		// Block Kit을 사용한 응답 메시지 with 피드백 버튼
		resp, err := h.client.PostMessage(ctx, &slack.PostMessageRequest{
			Channel:         msg.Channel,
			Text:            "You said: " + msg.Text, // 폴백 텍스트
			ThreadTimestamp: threadTS,
			Blocks: []slack.SlackBlock{
				{
					Type: "section",
					Text: &slack.SlackText{
						Type: "mrkdwn",
						Text: fmt.Sprintf("*You said:* %s", msg.Text),
					},
				},
				{
					Type: "divider",
				},
				{
					Type: "section",
					Text: &slack.SlackText{
						Type: "mrkdwn",
						Text: "plz give us feedback",
					},
				},
				{
					Type:    "actions",
					BlockID: "actionblock789",
					Elements: []interface{}{
						slack.SlackBlockElement{
							Type: "button",
							Text: &slack.SlackText{
								Type:  "plain_text",
								Text:  "good",
								Emoji: slack.BoolPtr(true),
							},
							Style:    "primary",
							Value:    "feedback_good",
							ActionID: "feedback_good",
						},
						slack.SlackBlockElement{
							Type: "button",
							Text: &slack.SlackText{
								Type:  "plain_text",
								Text:  "bad",
								Emoji: slack.BoolPtr(true),
							},
							Style:    "danger",
							Value:    "feedback_bad",
							ActionID: "feedback_bad",
						},
					},
				},
			},
		})
		if err != nil {
			slog.Warn("failed to post message", slog.String("channel", msg.Channel), slog.Any("error", err))
			return
		}

		slog.Info("sent message with feedback buttons", slog.String("ts", string(resp.Timestamp)))

	case event.EventTypeAssistantThreadStarted:
		slog.Info("received assistant thread started event")
		channelID := ec.Event.OfAssistantThreadStarted.AssistantThread.ChannelID
		threadTS := ec.Event.OfAssistantThreadStarted.AssistantThread.ThreadTimestamp

		err := retry.Do(ctx, func(ctx context.Context) error {
			_, err := h.client.AssistantSetStatus(ctx, &slack.AssistantSetStatusRequest{
				Channel:         channelID,
				Status:          "Preparing magic...",
				ThreadTimestamp: threadTS,
			})
			return err
		})
		if err != nil {
			slog.Warn("failed to set status", slog.String("channel", channelID), slog.Any("error", err))
		}

		time.Sleep(3 * time.Second)

		// Block Kit을 사용한 인사 메시지
		err = retry.Do(ctx, func(ctx context.Context) error {
			_, err := h.client.PostMessage(ctx, &slack.PostMessageRequest{
				Channel:         channelID,
				Text:            "What spell should I cast?", // 폴백 텍스트
				ThreadTimestamp: threadTS,
				Blocks: []slack.SlackBlock{
					{
						Type: "section",
						Text: &slack.SlackText{
							Type: "mrkdwn",
							Text: "✨ *What spell should I cast?*",
						},
					},
					{
						Type: "context",
						Elements: []interface{}{
							slack.SlackContextElement{
								Type: "mrkdwn",
								Text: "_Type your message to get started..._",
							},
						},
					},
				},
			})
			return err
		})
		if err != nil {
			slog.Warn("failed to post message", slog.String("channel", channelID), slog.Any("error", err))
		}

	case event.EventTypeAssistantThreadContextChanged:
		slog.Info("received assistant thread context changed event")
	}
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

	client := slack.NewClient(http.DefaultClient, appToken, botToken)
	resp, err := client.OpenConnection(ctx)
	if err != nil {
		slog.Error("failed to open connection", slog.Any("error", err))
		return
	}

	handler := &Handler{
		appToken: appToken,
		botToken: botToken,
		client:   client,
	}

	b := bot.NewBot(handler)
	if err := b.Run(ctx, resp.URL); err != nil {
		slog.Error("failed to run bot", slog.Any("error", err))
	}
}
