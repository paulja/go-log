package server

import (
	"context"

	api "github.com/paulja/go-log/api/v1"
	"github.com/paulja/go-log/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Config struct {
	CommitLog CommitLog
}

var _ api.LogServer = (*server)(nil)

func NewGRPCServer(config *Config) (*grpc.Server, error) {
	tls, err := tls.ServerConfig()
	if err != nil {
		return nil, err
	}
	gsrv := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(tls)),
	)
	srv, err := newgrpcServer(config)
	if err != nil {
		return nil, err
	}
	api.RegisterLogServer(gsrv, srv)
	return gsrv, nil
}

type server struct {
	api.UnimplementedLogServer
	*Config
}

func newgrpcServer(c *Config) (*server, error) {
	return &server{Config: c}, nil
}

func (s *server) Produce(
	ctx context.Context,
	req *api.ProduceRequest,
) (
	*api.ProduceResponse,
	error,
) {
	offset, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &api.ProduceResponse{
		Offset: offset,
	}, nil
}

func (s *server) Consume(
	ctx context.Context,
	req *api.ConsumeRequest,
) (
	*api.ConsumeResponse,
	error,
) {
	record, err := s.CommitLog.Read(req.Offset)
	if err != nil {
		return nil, err
	}
	return &api.ConsumeResponse{
		Record: record,
	}, nil
}

func (s *server) ProduceStream(stream api.Log_ProduceStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		res, err := s.Produce(stream.Context(), req)
		if err != nil {
			return err
		}
		if err = stream.Send(res); err != nil {
			return err
		}
	}
}

func (s *server) ConsumeStream(
	req *api.ConsumeRequest,
	stream api.Log_ConsumeStreamServer,
) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			res, err := s.Consume(stream.Context(), req)
			switch err.(type) {
			case nil:
			case api.ErrOffsetOutOfRange:
				continue
			default:
				return err
			}
			if err = stream.Send(res); err != nil {
				return err
			}
			req.Offset++
		}
	}
}

type CommitLog interface {
	Append(*api.Record) (uint64, error)
	Read(uint64) (*api.Record, error)
}
