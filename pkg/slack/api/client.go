package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const baseURL = "https://slack.com/api"

type Client struct {
	client *http.Client

	AppToken string
	BotToken string
}

func NewClient(clitn *http.Client, appToken, botToken string) *Client {
	return &Client{
		client:   clitn,
		AppToken: appToken,
		BotToken: botToken,
	}
}

// Generate a temporary Socket Mode WebSocket URL that your app can connect to
// in order to receive events and interactive payloads over.
func (c *Client) OpenConnection(ctx context.Context) (*OpenConnectionResponse, error) {
	path := "apps.connections.open"

	r, err := c.newRequest(ctx, "POST", c.AppToken, path, nil)
	if err != nil {
		return nil, err
	}

	data, err := c.sendRequest(r)
	if err != nil {
		return nil, err
	}

	result := &OpenConnectionResponse{}
	if err := json.Unmarshal(data, result); err != nil {
		return nil, err
	}

	if !result.OK || result.Error != "" {
		return nil, errors.New(result.Error)
	}

	return result, nil
}

// Sends a message to a channel.
func (c *Client) PostMessage(
	ctx context.Context,
	req *PostMessageRequest,
) (*PostMessageResponse, error) {
	path := "chat.postMessage"

	r, err := c.newRequest(ctx, "POST", c.BotToken, path, req)
	if err != nil {
		return nil, err
	}

	data, err := c.sendRequest(r)
	if err != nil {
		return nil, err
	}

	result := &PostMessageResponse{}
	if err := json.Unmarshal(data, result); err != nil {
		return nil, err
	}

	if !result.OK || result.Error != "" {
		return nil, errors.New(result.Error)
	}

	return result, nil
}

// Set the status for an AI assistant thread.
func (c *Client) AssistantSetStatus(
	ctx context.Context,
	req *AssistantSetStatusRequest,
) (*AssistantSetStatusResponse, error) {
	path := "assistant.threads.setStatus"

	r, err := c.newRequest(ctx, "POST", c.BotToken, path, req)
	if err != nil {
		return nil, err
	}

	data, err := c.sendRequest(r)
	if err != nil {
		return nil, err
	}

	result := &AssistantSetStatusResponse{}
	if err := json.Unmarshal(data, result); err != nil {
		return nil, err
	}

	if !result.OK || result.Error != "" {
		return nil, errors.New(result.Error)
	}

	return result, nil
}

// Set suggested prompts for the given assistant thread.
func (c *Client) AssistantSetSuggestedPrompts(
	ctx context.Context,
	req *AssistantSetSuggestedPromptsRequest,
) (*AssistantSetSuggestedPromptsResponse, error) {
	path := "assistant.threads.setSuggestedPrompts"

	r, err := c.newRequest(ctx, "POST", c.BotToken, path, req)
	if err != nil {
		return nil, err
	}

	data, err := c.sendRequest(r)
	if err != nil {
		return nil, err
	}

	result := &AssistantSetSuggestedPromptsResponse{}
	if err := json.Unmarshal(data, result); err != nil {
		return nil, err
	}

	if !result.OK || result.Error != "" {
		return nil, errors.New(result.Error)
	}

	return result, nil
}

// Fetches a conversation's history of messages and events.
func (c *Client) ConversationsHistory(
	ctx context.Context,
	req *ConversationsHistoryRequest,
) (*ConversationsHistoryResponse, error) {
	path := "conversations.history"

	r, err := c.newRequest(ctx, "POST", c.BotToken, path, req)
	if err != nil {
		return nil, err
	}

	data, err := c.sendRequest(r)
	if err != nil {
		return nil, err
	}

	result := &ConversationsHistoryResponse{}
	if err := json.Unmarshal(data, result); err != nil {
		return nil, err
	}

	if !result.OK || result.Error != "" {
		return nil, errors.New(result.Error)
	}

	return result, nil
}

func (c *Client) newRequest(
	ctx context.Context,
	method string,
	token string,
	path string,
	body any,
) (*http.Request, error) {
	bodyReader := func() io.Reader {
		if body == nil {
			return nil
		}
		data, err := json.Marshal(body)
		if err != nil {
			return nil
		}
		return bytes.NewReader(data)
	}()

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s/%s", baseURL, path), bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	return req, nil
}

func (c *Client) sendRequest(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("failed to close response body", slog.Any("error", err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
