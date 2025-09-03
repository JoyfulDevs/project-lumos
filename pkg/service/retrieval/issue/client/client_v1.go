package client

import (
	"context"

	"github.com/joyfuldevs/project-lumos/gen/go/retrieval/issue/v1"
)

// RetrieveIssuesV1은 주어진 이슈 키 목록에 대한 이슈를 검색합니다.
func (c *Client) RetrievalIssuesV1(ctx context.Context, keys []string) ([]*issue.Issue, error) {
	req := &issue.RetrieveRequest{
		IssueKeys: keys,
	}
	resp, err := c.serviceV1.Retrieve(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Issues, nil
}
