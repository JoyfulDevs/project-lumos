package eventsapi

import (
	"encoding/json"
)

type PayloadType string

const (
	PayloadTypeEventCallback PayloadType = "event_callback"
)

type Payload struct {
	Type PayloadType `json:"type"`

	OfEventCallback *EventCallback `json:"-"`
}

func (p *Payload) UnmarshalJSON(data []byte) error {
	type alias Payload

	raw := &alias{}
	if err := json.Unmarshal(data, raw); err != nil {
		return err
	}

	p.Type = raw.Type
	switch raw.Type {
	case PayloadTypeEventCallback:
		p.OfEventCallback = &EventCallback{}
		if err := json.Unmarshal(data, p.OfEventCallback); err != nil {
			return err
		}
	}

	return nil
}

type EventCallback struct {
	EventID string `json:"event_id"`
	Event   Event  `json:"event"`
}
