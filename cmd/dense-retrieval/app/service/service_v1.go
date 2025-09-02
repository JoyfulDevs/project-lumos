package service

import (
	"context"

	"github.com/devafterdark/project-lumos/gen/go/retrieval/passage/v1"
	"github.com/devafterdark/project-lumos/pkg/service/retrieval/passage/server"
)

var _ server.ServiceV1 = (*Service)(nil)

func (s *Service) Retrieve(ctx context.Context, query string, limit int32) ([]*passage.Passage, error) {
	vectors, err := s.Embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	results, err := s.VectorRetriever.Retrieve(ctx, RetrieveParams{
		Vectors: vectors,
		Limit:   limit,
	})
	if err != nil {
		return nil, err
	}

	passages := make([]*passage.Passage, 0, len(results))
	for _, result := range results {
		passages = append(passages, &passage.Passage{
			Score:   result.Score,
			Content: result.Passage,
		})
	}

	return passages, nil
}
