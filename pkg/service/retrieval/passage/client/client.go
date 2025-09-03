package client

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	passagev1 "github.com/joyfuldevs/project-lumos/gen/go/retrieval/passage/v1"
)

// Client는 패시지 검색 서비스를 위한 클라이언트 API입니다.
type Client struct {
	options *clientOptions

	grpcClient *grpc.ClientConn
	serviceV1  passagev1.PassageRetrievalServiceClient
}

// NewClient는 새로운 패시지 검색 서비스 클라이언트를 생성합니다.
func NewClient(opts ...Option) (*Client, error) {
	options := defaultClientOptions
	for _, opt := range opts {
		opt(&options)
	}

	grpcClient, err := grpc.NewClient(
		net.JoinHostPort(options.host, options.port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		options:    &options,
		grpcClient: grpcClient,
		serviceV1:  passagev1.NewPassageRetrievalServiceClient(grpcClient),
	}, nil
}

// Close는 패시지 검색 서비스에 대한 연결을 종료합니다.
func (c *Client) Close() error {
	return c.grpcClient.Close()
}
