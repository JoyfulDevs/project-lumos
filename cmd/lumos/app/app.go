package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/devafterdark/project-lumos/pkg/slack"
	"github.com/devafterdark/project-lumos/pkg/slack/bot"
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

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	c := slack.NewClient(http.DefaultClient, appToken, botToken)
	resp, err := c.OpenConnection(ctx)
	if err != nil {
		return err
	}

	botHandler := NewBotHandler(appToken, botToken)
	bot := bot.NewBot(botHandler)

	return bot.Run(ctx, resp.URL)

}
