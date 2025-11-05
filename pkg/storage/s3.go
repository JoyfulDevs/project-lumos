package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joyfuldevs/project-lumos/pkg/slack/interactive"
)

// 단순한 대화 구조
type Conversation struct {
	UserMessage     string    `json:"user_message"`
	UserID          string    `json:"user_id"`
	UserMessageTime time.Time `json:"user_message_time"`
	BotResponse     string    `json:"bot_response,omitempty"`
	BotResponseTime time.Time `json:"bot_response_time,omitempty"`
}

// 이전 피드백 이후의 대화들을 포함한 피드백 데이터 (채널 단위)
type FeedbackWithConversations struct {
	ChannelID        string          `json:"channel_id"`
	FeedbackType     string          `json:"feedback_type"`
	FeedbackUser     string          `json:"feedback_user"`
	FeedbackTime     time.Time       `json:"feedback_time"`
	LastFeedbackTime *time.Time      `json:"last_feedback_time,omitempty"`      // 이전 피드백 시간
	TimeSinceLastMin float64         `json:"time_since_last_minutes,omitempty"` // 이전 피드백부터 경과 시간(분)
	Conversations    []*Conversation `json:"conversations"`
	TotalMessages    int             `json:"total_messages"`
}

// S3 저장소 클라이언트
type S3Storage struct {
	client     *s3.Client
	bucketName string
}

// S3 클라이언트 초기화
func NewS3Storage(bucketName, region string) (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Storage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// 피드백 + 대화 내용을 S3에 저장
func (s *S3Storage) SaveFeedbackWithConversations(
	ctx context.Context,
	payload *interactive.BlockActionsPayload,
	channelID string,
	_ string,
	feedbackType string,
	conversations []*Conversation,
	feedbackTime time.Time,
	lastFeedbackTime time.Time,
) error {
	// 총 메시지 수 계산
	totalMessages := 0
	for _, conv := range conversations {
		if conv.UserMessage != "" {
			totalMessages++
		}
		if conv.BotResponse != "" {
			totalMessages++
		}
	}

	feedbackData := &FeedbackWithConversations{
		ChannelID:     channelID,
		FeedbackType:  feedbackType,
		FeedbackUser:  payload.User.ID,
		FeedbackTime:  feedbackTime,
		Conversations: conversations,
		TotalMessages: totalMessages,
	}

	// 이전 피드백 시간 계산
	if !lastFeedbackTime.IsZero() {
		feedbackData.LastFeedbackTime = &lastFeedbackTime
		feedbackData.TimeSinceLastMin = feedbackTime.Sub(lastFeedbackTime).Minutes()
	}

	// JSON 직렬화
	data, err := json.Marshal(feedbackData)
	if err != nil {
		return fmt.Errorf("failed to marshal feedback data: %w", err)
	}

	// S3 키 생성
	key := fmt.Sprintf(
		"feedback/year=%d/month=%02d/day=%02d/%s_%s_%d.json",
		feedbackTime.Year(),
		feedbackTime.Month(),
		feedbackTime.Day(),
		channelID,
		feedbackType,
		feedbackTime.Unix(),
	)

	// 메타데이터 구성
	metadata := map[string]string{
		"feedback-type":      feedbackType,
		"channel-id":         channelID,
		"feedback-user":      payload.User.ID,
		"conversation-count": fmt.Sprintf("%d", len(conversations)),
		"total-messages":     fmt.Sprintf("%d", totalMessages),
		"feedback-time":      feedbackTime.Format(time.RFC3339),
	}

	if !lastFeedbackTime.IsZero() {
		metadata["last-feedback-time"] = lastFeedbackTime.Format(time.RFC3339)
		metadata["minutes-since-last"] = fmt.Sprintf("%.1f", feedbackData.TimeSinceLastMin)
	} else {
		metadata["is-first-feedback"] = "true"
	}

	if payload.User.Name != "" {
		metadata["feedback-user-name"] = payload.User.Name
	}
	if payload.Channel != nil && payload.Channel.Name != "" {
		metadata["channel-name"] = payload.Channel.Name
	}

	// S3 업로드
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
		Metadata:    metadata,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	slog.Info("channel feedback saved",
		slog.String("channel_id", channelID),
		slog.String("feedback_type", feedbackType),
		slog.Int("conversation_count", len(conversations)),
		slog.Int("total_messages", totalMessages),
		slog.Float64("minutes_since_last", feedbackData.TimeSinceLastMin),
		slog.String("s3_key", key))

	return nil
}

// 특정 채널의 마지막 피드백 시간 조회
func (s *S3Storage) GetLastFeedbackTime(ctx context.Context, channelID, _ string) (time.Time, error) {
	prefix := "feedback/"

	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	}

	var latestTime time.Time
	paginator := s3.NewListObjectsV2Paginator(s.client, listInput)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			if !strings.Contains(key, channelID) {
				continue
			}
			if obj.LastModified != nil && obj.LastModified.After(latestTime) {
				latestTime = *obj.LastModified
			}
		}
	}

	if latestTime.IsZero() {
		slog.Debug("no previous feedback found", slog.String("channel_id", channelID))
	} else {
		slog.Debug("found last feedback time",
			slog.String("channel_id", channelID),
			slog.String("last_feedback_time", latestTime.Format(time.RFC3339)))
	}

	return latestTime, nil
}
