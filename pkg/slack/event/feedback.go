package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// FeedbackHandler는 AI 어시스턴트 이벤트에 대한 피드백 버튼을 처리합니다
type FeedbackHandler struct {
	SlackBotToken string
	HTTPClient    *http.Client
}

// NewFeedbackHandler는 새로운 FeedbackHandler를 생성합니다
func NewFeedbackHandler(slackBotToken string) *FeedbackHandler {
	return &FeedbackHandler{
		SlackBotToken: slackBotToken,
		HTTPClient:    &http.Client{},
	}
}

// SendFeedbackButtons는 지정된 채널에 피드백 버튼을 보냅니다 (외부에서 호출용)
func (h *FeedbackHandler) SendFeedbackButtons(channelID string) error {
	return h.sendFeedbackMessage(channelID)
}

// HandleSocketEvent는 소켓 이벤트를 처리하고 필요시 피드백 버튼을 보냅니다
func (h *FeedbackHandler) HandleSocketEvent(socketEvent *SocketEvent) error {
	if socketEvent.Type != SocketEventTypeEventsAPI {
		return nil
	}

	eventsAPI := socketEvent.OfEventsAPI
	if eventsAPI == nil || eventsAPI.Payload == nil {
		return nil
	}

	if eventsAPI.Payload.Type != EventsAPITypeEventCallback {
		return nil
	}

	eventCallback := eventsAPI.Payload.OfEventCallback
	if eventCallback == nil {
		return nil
	}

	switch eventCallback.Event.Type {
	case EventTypeAssistantThreadStarted:
		return h.handleAssistantThreadStarted(eventCallback.Event.OfAssistantThreadStarted)
	case EventTypeAssistantThreadContextChanged:
		// 필요시 컨텍스트 변경 이벤트에서도 피드백 버튼 표시
		return h.handleAssistantThreadContextChanged(eventCallback.Event.OfAssistantThreadContextChanged)
	}

	return nil
}

// handleAssistantThreadStarted는 어시스턴트 스레드 시작 시 피드백 버튼을 보냅니다
func (h *FeedbackHandler) handleAssistantThreadStarted(event *AssistantThreadStartedEvent) error {
	if event == nil {
		return nil
	}

	channelID := event.AssistantThread.ChannelID
	if channelID == "" {
		return fmt.Errorf("missing channel ID in assistant thread started event")
	}

	return h.sendFeedbackMessage(channelID)
}

// handleAssistantThreadContextChanged는 어시스턴트 스레드 컨텍스트 변경 시 처리합니다
func (h *FeedbackHandler) handleAssistantThreadContextChanged(event *AssistantThreadContextChangedEvent) error {
	if event == nil {
		return nil
	}

	// 필요에 따라 컨텍스트 변경 시에도 피드백 버튼을 보낼 수 있습니다
	// channelID := event.AssistantThread.ChannelID
	// return h.sendFeedbackMessage(channelID)

	return nil
}

// sendFeedbackMessage는 지정된 채널에 피드백 버튼이 포함된 메시지를 보냅니다
func (h *FeedbackHandler) sendFeedbackMessage(channelID string) error {
	message := h.createFeedbackBlocks(channelID)
	return h.postMessage(message)
}

// createFeedbackBlocks는 피드백 버튼 블록을 생성합니다
func (h *FeedbackHandler) createFeedbackBlocks(channelID string) *SlackMessage {
	return &SlackMessage{
		Channel: channelID,
		Blocks: []SlackBlock{
			{
				Type: "section",
				Text: &SlackText{
					Type: "mrkdwn",
					Text: "plz give us feedback",
				},
			},
			{
				Type:    "actions",
				BlockID: "actionblock789",
				Elements: []SlackBlockElement{
					{
						Type: "button",
						Text: &SlackText{
							Type: "plain_text",
							Text: "good",
						},
						Style:    "primary",
						Value:    "feedback_good",
						ActionID: "feedback_good",
					},
					{
						Type: "button",
						Text: &SlackText{
							Type: "plain_text",
							Text: "bad",
						},
						Style:    "danger",
						Value:    "feedback_bad",
						ActionID: "feedback_bad",
					},
				},
			},
		},
	}
}

// postMessage는 Slack API를 통해 메시지를 보냅니다
func (h *FeedbackHandler) postMessage(message *SlackMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.SlackBotToken))

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

// SlackMessage는 Slack에 보낼 메시지의 구조체입니다
type SlackMessage struct {
	Channel string       `json:"channel"`
	Blocks  []SlackBlock `json:"blocks"`
}

// SlackBlock은 Slack Block Kit의 블록을 나타냅니다
type SlackBlock struct {
	Type      string              `json:"type"`
	Text      *SlackText          `json:"text,omitempty"`
	Accessory *SlackAccessory     `json:"accessory,omitempty"`
	BlockID   string              `json:"block_id,omitempty"`
	Elements  []SlackBlockElement `json:"elements,omitempty"`
}

// SlackText는 Slack 텍스트 요소를 나타냅니다
type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// SlackAccessory는 Slack 액세서리 요소를 나타냅니다
type SlackAccessory struct {
	Type     string     `json:"type"`
	Text     *SlackText `json:"text,omitempty"`
	Value    string     `json:"value,omitempty"`
	ActionID string     `json:"action_id,omitempty"`
}

// SlackBlockElement는 Slack 블록 요소를 나타냅니다
type SlackBlockElement struct {
	Type     string     `json:"type"`
	Text     *SlackText `json:"text,omitempty"`
	Style    string     `json:"style,omitempty"`
	Value    string     `json:"value,omitempty"`
	ActionID string     `json:"action_id,omitempty"`
}
