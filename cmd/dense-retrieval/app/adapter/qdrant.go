package adapter

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"

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
	type passage struct {
		Score   float32
		Key     string
		Title   string
		Content string
	}

	var (
		results  = make([]service.RetrieveResult, 0, params.Limit)
		passages = make(map[string]passage, params.Limit*3)
		limit    = uint64(params.Limit)
	)

	searchAndMerge := func(name string, limit uint64) {
		resp, err := q.client.Query(ctx, &qdrant.QueryPoints{
			CollectionName: name,
			Query:          qdrant.NewQueryDense(params.Vectors),
			Limit:          &limit,
			WithPayload:    qdrant.NewWithPayload(true),
		})

		if err != nil {
			slog.Warn("failed to query points", "collection", name, "error", err)
			return
		}

		for _, point := range resp {
			key := point.Payload["key"].GetStringValue()
			if p, exists := passages[key]; exists {
				p.Score = p.Score + point.Score
				passages[key] = p
				continue
			}
			title := point.Payload["title"].GetStringValue()
			content := point.Payload["content"].GetStringValue()
			passages[key] = passage{
				Score:   point.Score,
				Key:     key,
				Title:   title,
				Content: content,
			}
		}
	}

	searchAndMerge("large-jira-title", limit)
	searchAndMerge("large-jira-content", limit)
	searchAndMerge("large-jira-content-split", limit)

	keys := make([]string, 0, len(passages))
	for k := range passages {
		keys = append(keys, k)
	}

	// 내림차순 정렬
	slices.SortFunc(keys, func(a, b string) int {
		v := passages[a].Score - passages[b].Score
		switch {
		case v < 0:
			return 1
		case v > 0:
			return -1
		default:
			return 0
		}
	})

	for i := range max(int(params.Limit), len(keys)) {
		value := passages[keys[i]]
		data, err := json.Marshal(value)
		if err != nil {
			slog.Warn("failed to marshal payload", "error", err)
			continue
		}
		results = append(results, service.RetrieveResult{
			Score:   value.Score,
			Passage: data,
		})
	}

	return results, nil
}
