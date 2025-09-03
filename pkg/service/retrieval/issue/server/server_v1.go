package server

import (
	"context"

	"github.com/joyfuldevs/project-lumos/gen/go/retrieval/issue/v1"
)

type ServiceV1 interface {
	Retrieve(ctx context.Context, keys []string) ([]*issue.Issue, error)
}

type serverV1 struct {
	issue.UnimplementedIssueRetrievalServiceServer

	service ServiceV1
}

func (s *serverV1) Retrieve(ctx context.Context, req *issue.RetrieveRequest) (*issue.RetrieveResponse, error) {
	issues, err := s.service.Retrieve(ctx, req.IssueKeys)
	if err != nil {
		return nil, err
	}
	return &issue.RetrieveResponse{Issues: issues}, nil
}
