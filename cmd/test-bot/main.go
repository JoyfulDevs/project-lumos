package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joyfuldevs/project-lumos/pkg/retry"
	"github.com/joyfuldevs/project-lumos/pkg/slack/api"
	"github.com/joyfuldevs/project-lumos/pkg/slack/blockkit"
	"github.com/joyfuldevs/project-lumos/pkg/slack/bot"
	"github.com/joyfuldevs/project-lumos/pkg/slack/eventsapi"
	"github.com/joyfuldevs/project-lumos/pkg/slack/interactive"
)

type Handler struct {
	appToken string
	botToken string
}

func (h *Handler) HandleEventsAPI(ctx context.Context, payload *eventsapi.Payload) {
	ec := payload.OfEventCallback
	if ec == nil {
		return
	}

	switch ec.Event.Type {
	case eventsapi.EventTypeMessage:
		e := ec.Event.OfMessage
		if e.Text == "" {
			return
		}
		if e.BotID != "" {
			return
		}

		slog.Info("received message event", slog.String("text", ec.Event.OfMessage.Text))

		c := api.NewClient(http.DefaultClient, h.appToken, h.botToken)
		_, err := c.PostMessage(ctx, &api.PostMessageRequest{
			Channel:         ec.Event.OfMessage.Channel,
			Text:            "You said: " + ec.Event.OfMessage.Text,
			ThreadTimestamp: ec.Event.OfMessage.ThreadTimestamp,
		})
		if err != nil {
			slog.Warn("failed to post message", slog.String("channel", ec.Event.OfMessage.Channel), slog.Any("error", err))
		}

		blocks := []*blockkit.Block{
			blockkit.NewBlockWithActionBlock(&blockkit.ActionBlock{
				Elements: []*blockkit.BlockElement{
					blockkit.NewBlockElementWithButtonElement(&blockkit.ButtonElement{
						Text:     blockkit.NewPlainText("Good", false),
						ActionID: "good",
						Style:    blockkit.ButtonStylePrimary,
					}),
					blockkit.NewBlockElementWithButtonElement(&blockkit.ButtonElement{
						Text:     blockkit.NewPlainText("Bad", false),
						ActionID: "bad",
						Style:    blockkit.ButtonStyleDanger,
					}),
				},
			}),
		}

		_, err = c.PostMessage(ctx, &api.PostMessageRequest{
			Channel:         ec.Event.OfMessage.Channel,
			Blocks:          blocks,
			ThreadTimestamp: ec.Event.OfMessage.ThreadTimestamp,
		})
		if err != nil {
			slog.Warn("failed to post message", slog.String("channel", ec.Event.OfMessage.Channel), slog.Any("error", err))
		}

	case eventsapi.EventTypeAssistantThreadStarted:
		slog.Info("received assistant thread started event")
		channelID := ec.Event.OfAssistantThreadStarted.AssistantThread.ChannelID
		c := api.NewClient(http.DefaultClient, h.appToken, h.botToken)
		err := retry.Do(ctx, func(ctx context.Context) error {
			_, err := c.AssistantSetStatus(ctx, &api.AssistantSetStatusRequest{
				Channel: channelID,
				Status:  "Preparing magic...",
			})
			return err
		})
		if err != nil {
			slog.Warn("failed to set status", slog.String("channel", channelID), slog.Any("error", err))
		}

		time.Sleep(3 * time.Second)

		err = retry.Do(ctx, func(ctx context.Context) error {
			_, err := c.PostMessage(ctx, &api.PostMessageRequest{
				Channel:         channelID,
				Text:            "What spell should I cast?",
				ThreadTimestamp: ec.Event.OfAssistantThreadStarted.AssistantThread.ThreadTimestamp,
			})
			return err
		})
		if err != nil {
			slog.Warn("failed to post message", slog.String("channel", channelID), slog.Any("error", err))
		}

	case eventsapi.EventTypeAssistantThreadContextChanged:
		slog.Info("received assistant thread context changed event")
	default:
		slog.Warn("unknown events api payload", slog.String("type", string(payload.Type)))
	}
}

func (h *Handler) HandleInteractive(ctx context.Context, payload *interactive.Payload) {
	switch payload.Type {
	case interactive.PayloadTypeBlockActions:
		slog.Info("received block actions")
		for _, action := range payload.OfBlockActions.Actions {
			slog.Info("action", slog.String("id", action.ActionID))
		}
		rp := &interactive.ResponsePayload{
			ResponseType:    interactive.Ephemeral,
			Text:            "Submitted!",
			ReplaceOriginal: true,
		}
		body, err := json.Marshal(rp)
		if err != nil {
			slog.Error("failed to marshal response", slog.Any("error", err))
			return
		}
		req, err := http.NewRequest("POST", payload.OfBlockActions.ResponseURL, bytes.NewReader(body))
		if err != nil {
			slog.Error("failed to create request", slog.Any("error", err))
			return
		}
		if _, err := http.DefaultClient.Do(req); err != nil {
			slog.Info("failed to send response", slog.Any("error", err))
		}

	case interactive.PayloadTypeMessageActions:
		slog.Info("received message actions")
	case interactive.PayloadTypeViewClosed:
		slog.Info("received view closed")
	case interactive.PayloadTypeViewSubmission:
		slog.Info("received view submission")
	default:
		slog.Warn("unknown interactive payload", slog.String("type", string(payload.Type)))
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

	c := api.NewClient(http.DefaultClient, appToken, botToken)
	resp, err := c.OpenConnection(ctx)
	if err != nil {
		return
	}

	b := bot.NewBot(&Handler{appToken: appToken, botToken: botToken})
	if err := b.Run(ctx, resp.URL); err != nil {
		slog.Error("failed to run bot", slog.Any("error", err))
	}
}
