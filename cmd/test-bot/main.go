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
	"github.com/joyfuldevs/project-lumos/pkg/storage"
)

type Handler struct {
	appToken          string
	botToken          string
	feedbackStorage   *storage.S3Storage
	conversationCache map[string]*storage.ConversationPair
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

		// 대화 캐시에 사용자 메시지 저장
		cacheKey := h.generateCacheKey(e.Channel, string(e.ThreadTimestamp))
		if h.conversationCache[cacheKey] == nil {
			h.conversationCache[cacheKey] = &storage.ConversationPair{
				UserMessage: e,
			}
		}

		c := api.NewClient(http.DefaultClient, h.appToken, h.botToken)
		botResponse := "You said: " + ec.Event.OfMessage.Text
		_, err := c.PostMessage(ctx, &api.PostMessageRequest{
			Channel:         ec.Event.OfMessage.Channel,
			Text:            botResponse,
			ThreadTimestamp: ec.Event.OfMessage.ThreadTimestamp,
		})
		if err != nil {
			slog.Warn("failed to post message", slog.String("channel", ec.Event.OfMessage.Channel), slog.Any("error", err))
		} else {
			// 봇 응답을 캐시에 저장
			h.conversationCache[cacheKey].BotResponse = botResponse
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

		// 피드백 수집 및 S3 저장
		// 모든 가능한 캐시 키를 시도해서 대화를 찾음
		var conversation *storage.ConversationPair
		var foundCacheKey string

		// 1. 채널 ID만으로 시도 (DM이나 스레드가 아닌 경우)
		cacheKey1 := h.generateCacheKey(payload.OfBlockActions.Channel.ID, "")
		if conv := h.conversationCache[cacheKey1]; conv != nil {
			conversation = conv
			foundCacheKey = cacheKey1
		}

		// 2. 캐시에서 해당 채널로 시작하는 키를 모두 찾아보기
		if conversation == nil {
			for key, conv := range h.conversationCache {
				if len(key) > len(payload.OfBlockActions.Channel.ID) &&
					key[:len(payload.OfBlockActions.Channel.ID)] == payload.OfBlockActions.Channel.ID {
					conversation = conv
					foundCacheKey = key
					break
				}
			}
		}

		if conversation != nil && conversation.UserMessage != nil && conversation.BotResponse != "" {
			if err := h.feedbackStorage.ProcessAndStoreFeedback(ctx, payload.OfBlockActions, conversation); err != nil {
				slog.Error("failed to process feedback", slog.Any("error", err))
			} else {
				slog.Info("feedback processed successfully", slog.String("cache_key", foundCacheKey))
			}
		} else {
			slog.Warn("conversation not found in cache",
				slog.String("channel_id", payload.OfBlockActions.Channel.ID),
				slog.String("tried_key", cacheKey1),
				slog.Bool("conversation_exists", conversation != nil))

			// 디버깅: 현재 캐시의 모든 키 출력
			slog.Info("current cache contents:")
			for key, conv := range h.conversationCache {
				slog.Info("cache entry",
					slog.String("key", key),
					slog.Bool("has_user_message", conv.UserMessage != nil),
					slog.Bool("has_bot_response", conv.BotResponse != ""))
			}
		}

		for _, action := range payload.OfBlockActions.Actions {
			slog.Info("action", slog.String("id", action.ActionID))
		}
		rp := &interactive.ResponsePayload{
			ResponseType:    interactive.Ephemeral,
			Text:            "Feedback recorded! Thank you.",
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

func (h *Handler) generateCacheKey(channelID, threadTS string) string {
	if threadTS == "" {
		return channelID
	}
	return channelID + "_" + threadTS
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	appToken := os.Getenv("SLACK_APP_TOKEN")
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	s3BucketName := os.Getenv("S3_BUCKET_NAME")

	if appToken == "" || botToken == "" || s3BucketName == "" {
		slog.Error("SLACK_APP_TOKEN, SLACK_BOT_TOKEN, and S3_BUCKET_NAME must be set")
		return
	}

	// S3 스토리지 초기화
	s3Storage, err := storage.NewS3Storage(s3BucketName)
	if err != nil {
		slog.Error("failed to initialize S3 storage", slog.Any("error", err))
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

	handler := &Handler{
		appToken:          appToken,
		botToken:          botToken,
		feedbackStorage:   s3Storage,
		conversationCache: make(map[string]*storage.ConversationPair),
	}

	b := bot.NewBot(handler)
	if err := b.Run(ctx, resp.URL); err != nil {
		slog.Error("failed to run bot", slog.Any("error", err))
	}
}
