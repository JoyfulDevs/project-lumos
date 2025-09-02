package chain

import (
	"context"
	"log/slog"

	"github.com/devafterdark/project-lumos/cmd/lumos/app/chat"
	"github.com/devafterdark/project-lumos/gen/go/retrieval/passage/v1"
	"github.com/devafterdark/project-lumos/pkg/service/retrieval/passage/client"
)

type Passage = passage.Passage

type passageKeyType int

const passageKey passageKeyType = iota

func WithPassages(parent context.Context, passages ...*Passage) context.Context {
	return context.WithValue(parent, passageKey, passages)
}

func PassagesFrom(ctx context.Context) []*Passage {
	info, _ := ctx.Value(passageKey).([]*Passage)
	return info
}

func WithPassageRetrieval(handler chat.Handler) chat.HandlerFunc {
	return chat.HandlerFunc(func(chat *chat.Chat) {
		ctx := chat.Context()

		query := chat.Thread[len(chat.Thread)-1]
		passages := make([]*Passage, 0, 20)
		passages = append(passages, denseRetrieval(ctx, query, 10)...)
		passages = append(passages, sparseRetrieval(ctx, query, 10)...)

		// TODO: implement score based ranking

		chat = chat.WithContext(WithPassages(ctx, passages...))

		handler.HandleChat(chat)
	})
}

func denseRetrieval(ctx context.Context, query string, limit int32) []*Passage {
	drClient, err := client.NewClient(
		client.WithHost("dense-retrieval-service"),
	)
	if err != nil {
		slog.Error("failed to create dense retrieval client", slog.Any("error", err))
		return nil
	}

	passages, err := drClient.RetrievePassagesV1(ctx, query, limit)
	if err != nil {
		slog.Error("failed to retrieve passages", slog.Any("error", err))
		return nil
	}

	return passages
}

func sparseRetrieval(ctx context.Context, query string, limit int32) []*Passage {
	srClient, err := client.NewClient(
		client.WithHost("sparse-retrieval-service"),
	)
	if err != nil {
		slog.Error("failed to create sparse retrieval client", slog.Any("error", err))
		return nil
	}

	passages, err := srClient.RetrievePassagesV1(ctx, query, limit)
	if err != nil {
		slog.Error("failed to retrieve passages", slog.Any("error", err))
		return nil
	}

	return passages
}
