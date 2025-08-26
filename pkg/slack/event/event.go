package event

import (
	"encoding/json"

	"github.com/devafterdark/project-lumos/pkg/slack"
)

type EventType string

const (
	EventTypeMessage                       EventType = "message"
	EventTypeAssistantThreadStarted        EventType = "assistant_thread_started"
	EventTypeAssistantThreadContextChanged EventType = "assistant_thread_context_changed"
)

type Event struct {
	Type EventType `json:"type"`

	OfMessage                       *MessageEvent                       `json:"-"`
	OfAssistantThreadStarted        *AssistantThreadStartedEvent        `json:"-"`
	OfAssistantThreadContextChanged *AssistantThreadContextChangedEvent `json:"-"`
}

func (e *Event) UnmarshalJSON(data []byte) error {
	type alias Event

	raw := &alias{}
	if err := json.Unmarshal(data, raw); err != nil {
		return err
	}

	e.Type = raw.Type
	switch raw.Type {
	case EventTypeMessage:
		e.OfMessage = &MessageEvent{}
		if err := json.Unmarshal(data, e.OfMessage); err != nil {
			return err
		}
	case EventTypeAssistantThreadStarted:
		e.OfAssistantThreadStarted = &AssistantThreadStartedEvent{}
		if err := json.Unmarshal(data, e.OfAssistantThreadStarted); err != nil {
			return err
		}
	case EventTypeAssistantThreadContextChanged:
		e.OfAssistantThreadContextChanged = &AssistantThreadContextChangedEvent{}
		if err := json.Unmarshal(data, e.OfAssistantThreadContextChanged); err != nil {
			return err
		}
	}

	return nil
}

type MessageEvent struct {
	Channel          string          `json:"channel"`
	User             string          `json:"user"`
	ParentUserID     string          `json:"parent_user_id,omitempty"`
	Text             string          `json:"text"`
	MessageTimestamp slack.Timestamp `json:"ts"`
	EventTimestamp   slack.Timestamp `json:"event_ts"`
	ThreadTimestamp  slack.Timestamp `json:"thread_ts"`
	ChannelType      string          `json:"channel_type"`
}

type AssistantThreadContext struct {
	ChannelID    string `json:"channel_id"`
	TeamID       string `json:"team_id"`
	EnterpriseID string `json:"enterprise_id"`
}

type AssistantThread struct {
	Context         AssistantThreadContext `json:"context"`
	UserID          string                 `json:"user_id"`
	ChannelID       string                 `json:"channel_id"`
	ThreadTimestamp slack.Timestamp        `json:"thread_ts"`
}

type AssistantThreadStartedEvent struct {
	EventTimestamp  slack.Timestamp `json:"event_ts"`
	AssistantThread AssistantThread `json:"assistant_thread"`
}

type AssistantThreadContextChangedEvent struct {
	EventTimestamp  slack.Timestamp `json:"event_ts"`
	AssistantThread AssistantThread `json:"assistant_thread"`
}
