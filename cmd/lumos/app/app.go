package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joyfuldevs/project-lumos/pkg/slack"
	"github.com/joyfuldevs/project-lumos/pkg/slack/bot"
)

func Run() error {
	appToken, ok := os.LookupEnv("SLACK_APP_TOKEN")
	if !ok {
		return errors.New("SLACK_APP_TOKEN is not set")
	}
	botToken, ok := os.LookupEnv("SLACK_BOT_TOKEN")
	if !ok {
		return errors.New("SLACK_BOT_TOKEN is not set")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	c := slack.NewClient(http.DefaultClient, appToken, botToken)
	resp, err := c.OpenConnection(ctx)
	if err != nil {
		return err
	}

	botHandler := NewBotHandler(c)
	bot := bot.NewBot(botHandler)

	return bot.Run(ctx, resp.URL)

}
