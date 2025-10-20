package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joyfuldevs/project-lumos/pkg/slack/eventsapi"
	"github.com/joyfuldevs/project-lumos/pkg/slack/interactive"
)

type ConversationPair struct {
	UserMessage *eventsapi.Message `json:"user_message"`
	BotResponse string             `json:"bot_response"`
}

type FeedbackData struct {
	Conversation    *ConversationPair   `json:"conversation"`
	FeedbackMessage *eventsapi.Message  `json:"feedback_message"`
	Action          *interactive.Action `json:"action"`
	FeedbackTime    time.Time           `json:"feedback_time"`
	Metadata        map[string]string   `json:"metadata,omitempty"`
}

type S3Storage struct {
	client     *s3.Client
	bucketName string
}

func NewS3Storage(bucketName string) (*S3Storage, error) {
	// 환경변수에서 리전 확인, 없으면 기본값 사용
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-northeast-2" // 서울 리전 기본값
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Storage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (s *S3Storage) ProcessAndStoreFeedback(ctx context.Context, payload *interactive.BlockActionsPayload, conversation *ConversationPair) error {
	for _, action := range payload.Actions {
		if action.ActionID != "good" && action.ActionID != "bad" {
			continue
		}

		feedbackMessage := &eventsapi.Message{
			Channel:          payload.Channel.ID,
			User:             payload.User.ID,
			Text:             "",
			MessageTimestamp: "",
			EventTimestamp:   "",
			ThreadTimestamp:  conversation.UserMessage.ThreadTimestamp,
			ChannelType:      conversation.UserMessage.ChannelType,
			BotID:            "",
		}

		feedbackData := &FeedbackData{
			Conversation:    conversation,
			FeedbackMessage: feedbackMessage,
			Action:          &action,
			FeedbackTime:    time.Now(),
			Metadata: map[string]string{
				"trigger_id":         payload.TriggerID,
				"response_url":       payload.ResponseURL,
				"channel_id":         payload.Channel.ID,
				"channel_name":       payload.Channel.Name,
				"question_text":      conversation.UserMessage.Text,
				"response_text":      conversation.BotResponse,
				"feedback_user_name": payload.User.Name,
				"team_id":            payload.Team.ID,
				"team_domain":        payload.Team.Domain,
			},
		}

		if err := s.StoreFeedback(ctx, feedbackData); err != nil {
			slog.Error("failed to store feedback",
				slog.String("feedback_user_id", payload.User.ID),
				slog.String("feedback", action.ActionID),
				slog.String("question_user_id", conversation.UserMessage.User),
				slog.Any("error", err))
			return err
		}

		slog.Info("feedback stored successfully",
			slog.String("feedback_user_id", payload.User.ID),
			slog.String("feedback", action.ActionID),
			slog.String("question_user_id", conversation.UserMessage.User),
			slog.String("question_text", conversation.UserMessage.Text),
			slog.String("response_text", conversation.BotResponse),
			slog.String("channel_id", payload.Channel.ID))
	}

	return nil
}

func (s *S3Storage) StoreFeedback(ctx context.Context, feedback *FeedbackData) error {
	data, err := json.Marshal(feedback)
	if err != nil {
		return fmt.Errorf("failed to marshal feedback data: %w", err)
	}

	key := s.generateS3Key(feedback.FeedbackTime, feedback.FeedbackMessage.User, feedback.Action.ActionID)

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
		Metadata: map[string]string{
			"feedback-type":       feedback.Action.ActionID,
			"feedback-user-id":    feedback.FeedbackMessage.User,
			"feedback-channel-id": feedback.FeedbackMessage.Channel,
			"question-user-id":    feedback.Conversation.UserMessage.User,
			"question-channel-id": feedback.Conversation.UserMessage.Channel,
			"question-timestamp":  string(feedback.Conversation.UserMessage.MessageTimestamp),
			"question-thread-ts":  string(feedback.Conversation.UserMessage.ThreadTimestamp),
			"block-id":            feedback.Action.BlockID,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

func (s *S3Storage) generateS3Key(timestamp time.Time, userID, actionID string) string {
	return fmt.Sprintf("feedback/year=%d/month=%02d/day=%02d/hour=%02d/%s_%d_%s.json",
		timestamp.Year(),
		timestamp.Month(),
		timestamp.Day(),
		timestamp.Hour(),
		userID,
		timestamp.Unix(),
		actionID,
	)
}
