package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/joyfuldevs/project-lumos/pkg/slack/api"
	"github.com/joyfuldevs/project-lumos/pkg/slack/bot"
)

func Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slackClient, err := slackClientFromEnv()
	if err != nil {
		return err
	}

	resp, err := slackClient.OpenConnection(ctx)
	if err != nil {
		return err
	}

	openaiClient, err := openaiClientFromEnv()
	if err != nil {
		return err
	}

	botHandler := NewBotHandler(slackClient, openaiClient)
	bot := bot.NewBot(botHandler)

	return bot.Run(ctx, resp.URL)
}

func slackClientFromEnv() (*api.Client, error) {
	appToken, ok := os.LookupEnv("SLACK_APP_TOKEN")
	if !ok {
		return nil, errors.New("SLACK_APP_TOKEN is not set")
	}
	botToken, ok := os.LookupEnv("SLACK_BOT_TOKEN")
	if !ok {
		return nil, errors.New("SLACK_BOT_TOKEN is not set")
	}

	client := api.NewClient(http.DefaultClient, appToken, botToken)
	return client, nil
}

func openaiClientFromEnv() (*openai.Client, error) {
	apiURL, ok := os.LookupEnv("OPENAI_API_URL")
	if !ok {
		return nil, errors.New("OPENAI_API_URL is not set")
	}
	apiKey, ok := os.LookupEnv("OPENAI_API_KEY")
	if !ok {
		return nil, errors.New("OPENAI_API_KEY is not set")
	}

	client := openai.NewClient(
		option.WithBaseURL(apiURL),
		option.WithAPIKey(apiKey),
	)
	return &client, nil
}
