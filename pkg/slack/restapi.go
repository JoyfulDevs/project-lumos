package slack

type APIResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type OpenConnectionResponse struct {
	APIResponse

	// 소켓 모드 연결에 사용하는 Web Socket URL.
	URL string `json:"url"`
}

type MessageParseType string

const (
	MessageParseTypeNone     MessageParseType = "none"
	MessageParseTypeMarkdown MessageParseType = "mrkdwn"
	MessageParseTypeFull     MessageParseType = "full"
)

type PostMessageRequest struct {
	// An encoded ID that represents a channel, private group, or IM channel to send the message to.
	Channel string `json:"channel"`
	// How this field works and whether it is required depends on other fields you use in your API call.
	Text string `json:"text,omitempty"`

	Blocks []SlackBlock `json:"blocks,omitempty"`

	// URL to an image to use as the icon for this message.
	IconURL string `json:"icon_url,omitempty"`
	// Emoji to use as the icon for this message. Overrides icon_url.
	IconEmoji string `json:"icon_emoji,omitempty"`
	// Find and link user groups. No longer supports linking individual users.
	// use syntax shown in Mentioning Users instead.
	LinkNames bool `json:"link_names,omitempty"`
	// Disable Slack markup parsing by setting to false.
	Markdown bool `json:"mrkdwn"`
	// Change how messages are treated.
	//
	// By default, URLs will be hyperlinked. Set parse to none to remove the hyperlinks.
	//
	// The behavior of parse is different for text formatted with mrkdwn.
	// By default, or when parse is set to none, mrkdwn formatting is implemented.
	// To ignore mrkdwn formatting, set parse to full.
	Parse MessageParseType `json:"parse,omitempty"`
	// Provide another message's ts value to make this message a reply.
	// Avoid using a reply's ts value. use its parent instead.
	ThreadTimestamp Timestamp `json:"thread_ts,omitempty"`
	// Used in conjunction with thread_ts and indicates whether reply should be made visible to everyone
	// in the channel or conversation. Defaults to false.
	ReplyBroadcast bool `json:"reply_broadcast,omitempty"`
	// Pass true to enable unfurling of primarily text-based content.
	UnfurlLinks bool `json:"unfurl_links,omitempty"`
	// Pass false to disable unfurling of media content.
	UnfurlMedia bool `json:"unfurl_media,omitempty"`
	// Set your bot's user name.
	Username string `json:"username,omitempty"`
}
type SlackBlock struct {
	Type      string             `json:"type"`
	Text      *SlackText         `json:"text,omitempty"`
	BlockID   string             `json:"block_id,omitempty"`
	Elements  []any              `json:"elements,omitempty"`
	Accessory *SlackBlockElement `json:"accessory,omitempty"`
}

type SlackText struct {
	Type  string `json:"type"` // "plain_text" 또는 "mrkdwn"
	Text  string `json:"text"`
	Emoji *bool  `json:"emoji,omitempty"`
}

// SlackBlockElement는 Slack 블록 요소를 나타냅니다
type SlackBlockElement struct {
	Type     string     `json:"type"`
	Text     *SlackText `json:"text,omitempty"`
	Style    string     `json:"style,omitempty"`
	Value    string     `json:"value,omitempty"`
	ActionID string     `json:"action_id,omitempty"`
	URL      string     `json:"url,omitempty"`
}

func BoolPtr(b bool) *bool {
	return &b
}

type SlackContextElement struct {
	Type string `json:"type"` // "mrkdwn" 또는 "plain_text"
	Text string `json:"text"`
}

type PostMessageResponse struct {
	APIResponse

	Channel   string    `json:"channel"`
	Timestamp Timestamp `json:"ts"`
}

type AssistantSetStatusRequest struct {
	Channel         string    `json:"channel_id"`
	ThreadTimestamp Timestamp `json:"thread_ts"`
	Status          string    `json:"status"`
}

type AssistantSetStatusResponse struct {
	APIResponse

	Detail string `json:"detail,omitempty"`
}

type SuggestedPrompt struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type AssistantSetSuggestedPromptsRequest struct {
	Channel         string            `json:"channel_id"`
	ThreadTimestamp Timestamp         `json:"thread_ts"`
	Title           string            `json:"title,omitempty"`
	Prompts         []SuggestedPrompt `json:"prompts"`
}

type AssistantSetSuggestedPromptsResponse struct {
	APIResponse
}
