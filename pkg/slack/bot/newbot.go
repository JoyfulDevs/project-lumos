package bot

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gorilla/websocket"

	"github.com/joyfuldevs/project-lumos/pkg/slack/event"
)

// EnhancedEventHandler는 기존 EventHandler를 확장합니다
type EnhancedEventHandler interface {
	EventHandler
	HandleSocketEvent(socketEvent *event.SocketEvent) // Socket 이벤트 처리 메서드 추가
}

// EnhancedBot은 기존 Bot을 확장하여 Socket 이벤트도 처리합니다
type EnhancedBot struct {
	handler EnhancedEventHandler
}

// NewEnhancedBot은 새로운 EnhancedBot을 생성합니다
func NewEnhancedBot(handler EnhancedEventHandler) *EnhancedBot {
	return &EnhancedBot{
		handler: handler,
	}
}

// Run은 기존 Bot의 Run 메서드를 확장하여 Socket 이벤트도 처리합니다
func (b *EnhancedBot) Run(ctx context.Context, url string) error {
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
		case e, ok := <-b.receiveEventEnhanced(conn):
			if !ok {
				return nil
			}

			// Socket 이벤트를 먼저 핸들러에 전달 (봇 메시지 감지용)
			b.handler.HandleSocketEvent(&e)

			// 기존 이벤트 처리 로직
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

// receiveEventEnhanced는 기존 receiveEvent를 개선하여 더 안정적으로 이벤트를 받습니다
func (b *EnhancedBot) receiveEventEnhanced(conn *websocket.Conn) chan event.SocketEvent {
	ch := make(chan event.SocketEvent, 10) // 버퍼 크기 증가
	go func() {
		defer close(ch)

		_, msg, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				slog.Error("failed to read websocket message", slog.Any("error", err))
			}
			return
		}

		var e event.SocketEvent
		if err := json.Unmarshal(msg, &e); err != nil {
			slog.Error("failed to unmarshal websocket message", slog.Any("error", err))
			return
		}

		ch <- e
	}()
	return ch
}
