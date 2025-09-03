package server

import (
	"context"

	passagev1 "github.com/joyfuldevs/project-lumos/gen/go/retrieval/passage/v1"
)

type ServiceV1 interface {
	Retrieve(ctx context.Context, query string, limit int32) ([]*passagev1.Passage, error)
}

type serverV1 struct {
	passagev1.UnimplementedPassageRetrievalServiceServer

	service ServiceV1
}

func (s *serverV1) Retrieve(ctx context.Context, req *passagev1.RetrieveRequest) (*passagev1.RetrieveResponse, error) {
	passages, err := s.service.Retrieve(ctx, req.Query, req.Limit)
	if err != nil {
		return nil, err
	}
	return &passagev1.RetrieveResponse{Passages: passages}, nil
}
