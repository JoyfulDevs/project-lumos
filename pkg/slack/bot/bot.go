package bot

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gorilla/websocket"

	"github.com/devafterdark/project-lumos/pkg/slack/event"
)

type Bot struct {
	handler EventHandler
}

func NewBot(handler EventHandler) *Bot {
	return &Bot{handler: handler}
}

func (b *Bot) Run(ctx context.Context, url string) error {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Warn("failed to close websocket connection", slog.Any("error", err))
		}
	}()

	slog.Info("websocket connection established")

	connCtx, connCancel := context.WithCancel(ctx)
	defer connCancel()

	for {
		select {
		case <-connCtx.Done():
			return nil
		case e, ok := <-receiveEvent(conn):
			if !ok {
				return nil
			}
			switch e.Type {
			case event.SocketEventTypeHello:
				slog.Info("received hello event")
			case event.SocketEventTypeDisconnect:
				slog.Info("received disconnect event")
				connCancel()
			case event.SocketEventTypeEventsAPI:
				resp := map[string]any{"envelope_id": e.OfEventsAPI.EnvelopeID}
				if err := conn.WriteJSON(resp); err != nil {
					slog.Warn("failed to respond to events api", slog.Any("error", err))
				}
				b.handler.HandleEventsAPI(connCtx, e.OfEventsAPI.Payload)
			default:
				slog.Warn("received unknown event type", slog.String("raw", string(e.Raw)))
			}
		}
	}
}

func receiveEvent(conn *websocket.Conn) chan event.SocketEvent {
	ch := make(chan event.SocketEvent, 1)
	go func() {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				slog.Error("failed to read websocket message", slog.Any("error", err))
			}
			close(ch)
			return
		}
		var e event.SocketEvent
		if err := json.Unmarshal(msg, &e); err != nil {
			slog.Error("failed to unmarshal websocket message", slog.Any("error", err))
			close(ch)
			return
		}
		ch <- e
	}()
	return ch
}
