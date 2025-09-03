package event_test

import (
	"encoding/json"
	"testing"

	"github.com/joyfuldevs/project-lumos/pkg/slack/event"
)

func TestUnmarshalSlackEvent(t *testing.T) {
	testCases := []struct {
		desc       string
		jsonString string
	}{
		{
			desc:       "hello",
			jsonString: `{ "type": "hello", "num_connections": 1, "debug_info": { "host": "applink-4", "build_number": 0, "approximate_connection_time": 18060 }, "connection_info": { "app_id": "A09AP3HFHCH" } }`,
		},
		{
			desc:       "disconnect",
			jsonString: `{ "type": "disconnect", "reason": "link_disabled", "debug_info": { "host": "wss-111.slack.com" } }`,
		},
		{
			desc:       "events_api:event_callback:message",
			jsonString: `{ "envelope_id": "0367683f-3be8-4280-b339-36e3f6652bac", "payload": { "token": "one-long-verification-token", "team_id": "T061EG9R6", "api_app_id": "A0PNCHHK2", "event": { "type": "message", "channel": "D024BE91L", "user": "U2147483697", "text": "Hello hello can you hear me?", "ts": "1355517523.000005", "event_ts": "1355517523.000005", "channel_type": "im" }, "type": "event_callback", "authed_teams": [ "T061EG9R6" ], "event_id": "Ev0PV52K21", "event_time": 1355517523 }, "type": "events_api", "accepts_response_payload": false, "retry_attempt": 0, "retry_reason": "" }`,
		},
		{
			desc:       "events_api:event_callback:assistant_thread_started",
			jsonString: `{ "envelope_id": "0367683f-3be8-4280-b339-36e3f6652bac", "payload": { "token": "AUKWnaquTu8fLtxIcI8ImjoD", "team_id": "T04F7MWMD", "api_app_id": "A09AP3HFHCH", "event": { "type": "assistant_thread_started", "assistant_thread": { "user_id": "U04CJM7DTFX", "context": { "force_search": false }, "channel_id": "D099YAQN8KH", "thread_ts": "1755746532.930469" }, "event_ts": "1755746532.948562" }, "type": "event_callback", "event_id": "Ev09BA9R2SUV", "event_time": 1755746532, "authorizations": [ { "enterprise_id": null, "team_id": "T04F7MWMD", "user_id": "U09A9U6T9PX", "is_bot": true, "is_enterprise_install": false } ], "is_ext_shared_channel": false }, "type": "events_api", "accepts_response_payload": false, "retry_attempt": 1, "retry_reason": "timeout" }`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			var se event.SocketEvent
			if err := json.Unmarshal([]byte(tc.jsonString), &se); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			switch se.Type {
			case event.SocketEventTypeHello:
				if se.OfHello == nil {
					t.Fatalf("missing hello event")
				}
			case event.SocketEventTypeDisconnect:
				if se.OfDisconnect == nil {
					t.Fatalf("missing disconnect event")
				}
			case event.SocketEventTypeEventsAPI:
				eventsAPI := se.OfEventsAPI
				if eventsAPI == nil {
					t.Fatalf("missing events API event")
				}
				if eventsAPI.Payload == nil {
					t.Fatalf("missing events API payload")
				}
				switch eventsAPI.Payload.Type {
				case event.EventsAPITypeEventCallback:
					if eventsAPI.Payload.OfEventCallback == nil {
						t.Fatalf("missing event callback")
					}
					switch eventsAPI.Payload.OfEventCallback.Event.Type {
					case event.EventTypeMessage:
						if eventsAPI.Payload.OfEventCallback.Event.OfMessage == nil {
							t.Fatalf("missing message event")
						}
					case event.EventTypeAssistantThreadStarted:
						if eventsAPI.Payload.OfEventCallback.Event.OfAssistantThreadStarted == nil {
							t.Fatalf("missing assistant thread started event")
						}
					case event.EventTypeAssistantThreadContextChanged:
						if eventsAPI.Payload.OfEventCallback.Event.OfAssistantThreadContextChanged == nil {
							t.Fatalf("missing assistant thread context changed event")
						}
					}
				}
			default:
				t.Fatalf("unexpected event type: %v", se.Type)
			}
		})
	}
}
