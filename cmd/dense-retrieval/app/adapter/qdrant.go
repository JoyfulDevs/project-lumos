package adapter

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/qdrant/go-client/qdrant"

	"github.com/joyfuldevs/project-lumos/cmd/dense-retrieval/app/service"
)

var _ service.VectorRetriever = (*QdrantClient)(nil)

type QdrantClient struct {
	client *qdrant.Client
}

func NewQdrantClient(host string) (*QdrantClient, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
	})
	if err != nil {
		return nil, err
	}

	return &QdrantClient{client: client}, nil
}

func (q *QdrantClient) Retrieve(ctx context.Context, params service.RetrieveParams) ([]service.RetrieveResult, error) {
	limit := uint64(params.Limit)
	resp, err := q.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: "content",
		Query:          qdrant.NewQuery(params.Vectors...),
		Limit:          &limit,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}

	results := make([]service.RetrieveResult, 0, len(resp))
	for _, point := range resp {
		data, err := json.Marshal(point.Payload)
		if err != nil {
			slog.Warn("failed to marshal payload", "error", err)
			continue
		}
		results = append(results, service.RetrieveResult{
			Score:   point.Score,
			Passage: data,
		})
	}

	return results, nil
}
