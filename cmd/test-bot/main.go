package main

import (
	"context"
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
}

func (h *Handler) HandleEventsAPI(ctx context.Context, payload *event.EventsAPIPayload) {
	ec := payload.OfEventCallback
	if ec == nil {
		return
	}

	switch ec.Event.Type {
	case event.EventTypeMessage:
		if ec.Event.OfMessage.Text == "" {
			return
		}
		if ec.Event.OfMessage.User == ec.Event.OfMessage.ParentUserID {
			return
		}
		slog.Info("received message event", slog.String("text", ec.Event.OfMessage.Text))
		c := slack.NewClient(http.DefaultClient, h.appToken, h.botToken)
		_, err := c.PostMessage(ctx, &slack.PostMessageRequest{
			Channel:         ec.Event.OfMessage.Channel,
			Text:            "You said: " + ec.Event.OfMessage.Text,
			ThreadTimestamp: ec.Event.OfMessage.ThreadTimestamp,
		})
		if err != nil {
			slog.Warn("failed to post message", slog.String("channel", ec.Event.OfMessage.Channel), slog.Any("error", err))
		}

	case event.EventTypeAssistantThreadStarted:
		slog.Info("received assistant thread started event")
		channelID := ec.Event.OfAssistantThreadStarted.AssistantThread.ChannelID
		c := slack.NewClient(http.DefaultClient, h.appToken, h.botToken)
		err := retry.Do(ctx, func(ctx context.Context) error {
			_, err := c.AssistantSetStatus(ctx, &slack.AssistantSetStatusRequest{
				Channel:         channelID,
				Status:          "Preparing magic...",
				ThreadTimestamp: ec.Event.OfAssistantThreadStarted.AssistantThread.ThreadTimestamp,
			})
			return err
		})
		if err != nil {
			slog.Warn("failed to set status", slog.String("channel", channelID), slog.Any("error", err))
		}

		time.Sleep(3 * time.Second)

		err = retry.Do(ctx, func(ctx context.Context) error {
			_, err := c.PostMessage(ctx, &slack.PostMessageRequest{
				Channel:         channelID,
				Text:            "What spell should I cast?",
				ThreadTimestamp: ec.Event.OfAssistantThreadStarted.AssistantThread.ThreadTimestamp,
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

	c := slack.NewClient(http.DefaultClient, appToken, botToken)
	resp, err := c.OpenConnection(ctx)
	if err != nil {
		return
	}

	b := bot.NewBot(&Handler{appToken: appToken, botToken: botToken})
	if err := b.Run(ctx, resp.URL); err != nil {
		slog.Error("failed to run bot", slog.Any("error", err))
	}
}
