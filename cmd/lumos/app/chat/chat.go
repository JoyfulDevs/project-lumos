package chat

import (
	"context"

	"github.com/devafterdark/project-lumos/pkg/slack"
)

type Chat struct {
	// 대화가 이루어진 채널 ID.
	Channel string
	// 스레드의 타임스탬프.
	Timestamp slack.Timestamp
	// 스레드 내용.
	Thread []string

	ctx context.Context
}

func (c *Chat) Context() context.Context {
	return c.ctx
}

func (c *Chat) WithContext(ctx context.Context) *Chat {
	if ctx == nil {
		panic("nil context")
	}
	c2 := new(Chat)
	*c2 = *c
	c2.ctx = ctx
	return c2
}
