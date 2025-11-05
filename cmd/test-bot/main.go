package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joyfuldevs/project-lumos/pkg/retry"
	"github.com/joyfuldevs/project-lumos/pkg/slack"
	"github.com/joyfuldevs/project-lumos/pkg/slack/api"
	"github.com/joyfuldevs/project-lumos/pkg/slack/blockkit"
	"github.com/joyfuldevs/project-lumos/pkg/slack/bot"
	"github.com/joyfuldevs/project-lumos/pkg/slack/eventsapi"
	"github.com/joyfuldevs/project-lumos/pkg/slack/interactive"
	"github.com/joyfuldevs/project-lumos/pkg/storage"
)

type Handler struct {
	appToken string
	botToken string
	storage  *storage.S3Storage
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

			channelID := payload.OfBlockActions.Channel.ID

			userID := payload.OfBlockActions.User.ID

			threadTS := payload.OfBlockActions.Container.ThreadTS

			if threadTS == "" {
				threadTS = payload.OfBlockActions.Message.ThreadTS
			}
			if threadTS == "" {
				threadTS = payload.OfBlockActions.MessageTS
			}

			if threadTS == "" {
				threadTS = payload.OfBlockActions.MessageTS
				slog.Info("no thread_ts found — using message_ts as fallback",
					slog.String("channel", payload.OfBlockActions.Channel.ID),
					slog.String("thread_ts", threadTS))
			}

			// 마지막 피드백 시간 조회
			lastFeedbackTime, err := h.storage.GetLastFeedbackTime(ctx, channelID, userID)
			if err != nil {
				slog.Warn("failed to get last feedback time", slog.Any("error", err))
			}

			// Slack 대화 가져오기 (이전 피드백 이후만)
			client := api.NewClient(http.DefaultClient, h.appToken, h.botToken)

			var oldestTS slack.Timestamp
			if !lastFeedbackTime.IsZero() {
				// Slack timestamp 형식은 "1234567890.000000"
				oldestTS = slack.Timestamp(fmt.Sprintf("%d.000000", lastFeedbackTime.Unix()))
			}

			history, err := client.ConversationsHistory(ctx, &api.ConversationsHistoryRequest{
				Channel:   channelID,
				Oldest:    oldestTS, // 이전 피드백 이후부터
				Inclusive: false,
				Limit:     200,
			})
			if err != nil {
				slog.Error("failed to fetch conversation history", slog.Any("error", err))
				continue
			}

			// 메시지 중 현재 thread에 해당하는 것만 필터링
			var conversations []*storage.Conversation
			for _, msg := range history.Messages {
				if msg.ThreadTimestamp != "" && string(msg.ThreadTimestamp) != threadTS {
					continue
				}

				// Slack timestamp ("1730812612.583989") → time.Time 변환
				tsParts := strings.Split(string(msg.MessageTimestamp), ".")
				var msgTime time.Time

				if len(tsParts) > 0 {
					sec, _ := strconv.ParseInt(tsParts[0], 10, 64)
					nsec := int64(0)
					if len(tsParts) == 2 {
						frac := tsParts[1]
						if len(frac) > 9 {
							frac = frac[:9] // 나노초는 최대 9자리
						}
						// 오른쪽에 0을 채워서 9자리로 맞춤
						for len(frac) < 9 {
							frac += "0"
						}
						nsec, _ = strconv.ParseInt(frac, 10, 64)
					}
					msgTime = time.Unix(sec, nsec)
				}
				conversations = append(conversations, &storage.Conversation{
					UserMessage:     msg.Text,
					UserID:          msg.User,
					UserMessageTime: msgTime,
				})
			}

			slog.Info("fetched filtered conversations",
				slog.Int("count", len(conversations)),
				slog.String("thread_ts", threadTS),
				slog.String("channel", channelID))

			// S3에 저장
			now := time.Now()
			err = h.storage.SaveFeedbackWithConversations(
				ctx,
				payload.OfBlockActions,
				channelID,
				userID,
				action.ActionID,
				conversations,
				now,
				lastFeedbackTime,
			)
			if err != nil {
				slog.Error("failed to save feedback with conversations", slog.Any("error", err))
			}
		}

		// 루프 끝난 뒤 응답 처리
		rp := &interactive.ResponsePayload{
			ResponseType:    interactive.Ephemeral,
			Text:            "Feedback received and stored to S3!",
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
			slog.Warn("failed to send interactive response", slog.Any("error", err))
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
	go func() { <-sig; cancel() }()

	bucket := os.Getenv("S3_BUCKET_NAME")
	region := os.Getenv("AWS_REGION")
	s3storage, err := storage.NewS3Storage(bucket, region)
	if err != nil {
		slog.Error("failed to initialize s3 storage", slog.Any("error", err))
		return
	}

	client := api.NewClient(http.DefaultClient, appToken, botToken)
	resp, err := client.OpenConnection(ctx)
	if err != nil {
		slog.Error("failed to open connection", slog.Any("error", err))
		return
	}

	handler := &Handler{appToken: appToken, botToken: botToken, storage: s3storage}
	b := bot.NewBot(handler)

	if err := b.Run(ctx, resp.URL); err != nil {
		slog.Error("failed to run bot", slog.Any("error", err))
	}
}
