package bot

import (
	"context"

	"github.com/joyfuldevs/project-lumos/pkg/slack/eventsapi"
	"github.com/joyfuldevs/project-lumos/pkg/slack/interactive"
)

type EventHandler interface {
	HandleEventsAPI(ctx context.Context, payload *eventsapi.Payload)
	HandleInteractive(ctx context.Context, payload *interactive.Payload)
}
