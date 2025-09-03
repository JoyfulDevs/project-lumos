package server

import (
	"context"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	issuev1 "github.com/joyfuldevs/project-lumos/gen/go/retrieval/issue/v1"
)

type Server struct {
	options *serverOptions
}

func NewServer(opts ...Option) *Server {
	options := defaultServerOptions
	for _, opt := range opts {
		opt(&options)
	}
	return &Server{
		options: &options,
	}
}

func (s *Server) Serve(ctx context.Context) error {
	listener, err := net.Listen("tcp", net.JoinHostPort("", s.options.port))
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	if s.options.serviceV1 != nil {
		issuev1.RegisterIssueRetrievalServiceServer(grpcServer, &serverV1{
			service: s.options.serviceV1,
		})
	} else {
		slog.Warn("service v1 is not set, skipping registration of v1 service")
	}

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	return grpcServer.Serve(listener)
}
