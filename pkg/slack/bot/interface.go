package bot

import (
	"context"

	"github.com/joyfuldevs/project-lumos/pkg/slack/event"
)

type EventHandler interface {
	HandleEventsAPI(ctx context.Context, payload *event.EventsAPIPayload)
}
