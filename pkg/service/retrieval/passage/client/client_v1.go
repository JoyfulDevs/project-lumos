package client

import (
	"context"

	"github.com/joyfuldevs/project-lumos/gen/go/retrieval/passage/v1"
)

// RetrievePassagesV1은 주어진 쿼리를 기반으로 최대 limit 개수만큼 패시지를 검색합니다.
func (c *Client) RetrievePassagesV1(ctx context.Context, query string, limit int32) ([]*passage.Passage, error) {
	req := &passage.RetrieveRequest{
		Query: query,
		Limit: limit,
	}
	resp, err := c.serviceV1.Retrieve(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Passages, nil
}
