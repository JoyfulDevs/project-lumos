package interactive

import (
	"encoding/json"

	"github.com/joyfuldevs/project-lumos/pkg/slack"
)

type PayloadType string

const (
	// Received when a user clicks a Block Kit interactive component.
	PayloadTypeBlockActions PayloadType = "block_actions"
	// Received when an app action in the message menu is used.
	PayloadTypeMessageActions PayloadType = "message_actions"
	// Received when a modal is canceled.
	PayloadTypeViewClosed PayloadType = "view_closed"
	// Received when a modal is submitted.
	PayloadTypeViewSubmission PayloadType = "view_submission"
)

type Payload struct {
	// An interactive element in a block will have a type of block_actions,
	// whereas an interactive element in a message attachment will have a type of interactive_message.
	Type PayloadType `json:"type"`

	OfBlockActions   *BlockActionsPayload   `json:"-"`
	OfMessageActions *MessageActionsPayload `json:"-"`
	OfViewClosed     *ViewClosedPayload     `json:"-"`
	OfViewSubmission *ViewSubmissionPayload `json:"-"`
}

func (p *Payload) UnmarshalJSON(data []byte) error {
	type alias Payload

	raw := &alias{}
	if err := json.Unmarshal(data, raw); err != nil {
		return err
	}

	p.Type = raw.Type
	switch raw.Type {
	case PayloadTypeBlockActions:
		p.OfBlockActions = &BlockActionsPayload{}
		if err := json.Unmarshal(data, p.OfBlockActions); err != nil {
			return err
		}
	case PayloadTypeMessageActions:
		p.OfMessageActions = &MessageActionsPayload{}
		if err := json.Unmarshal(data, p.OfMessageActions); err != nil {
			return err
		}
	case PayloadTypeViewClosed:
		p.OfViewClosed = &ViewClosedPayload{}
		if err := json.Unmarshal(data, p.OfViewClosed); err != nil {
			return err
		}
	case PayloadTypeViewSubmission:
		p.OfViewSubmission = &ViewSubmissionPayload{}
		if err := json.Unmarshal(data, p.OfViewSubmission); err != nil {
			return err
		}
	}

	return nil
}

// Received when a user clicks a Block Kit interactive component.
type BlockActionsPayload struct {
	// The user who interacted to trigger this request.
	User *slack.User `json:"user"`
	// The workspace the app is installed on. Null if the app is org-installed.
	Team *slack.Team `json:"team"`
	// The channel where this block action took place.
	Channel *slack.Channel `json:"channel,omitempty"`
	// A short-lived ID that can be used to open modals.
	TriggerID string `json:"trigger_id,omitempty"`
	// A short-lived webhook that can be used to send messages in response to interactions.
	ResponseURL string `json:"response_url,omitempty"`
	// Contains data from the specific interactive component that was used.
	// App surfaces can contain blocks with multiple interactive components,
	// and each of those components can have multiple values selected by users.
	Actions   []Action `json:"actions,omitempty"`
	MessageTS string   `json:"message_ts,omitempty"`

	Container struct {
		ThreadTS  string `json:"thread_ts,omitempty"`
		MessageTS string `json:"message_ts,omitempty"`
	} `json:"container,omitempty"`

	Message struct {
		ThreadTS string `json:"thread_ts,omitempty"`
		TS       string `json:"ts,omitempty"`
	} `json:"message,omitempty"`
}

// Received when an app action in the message menu is used.
type MessageActionsPayload struct {
}

// Received when a modal is canceled.
type ViewClosedPayload struct {
}

// Received when a modal is submitted.
type ViewSubmissionPayload struct {
}
